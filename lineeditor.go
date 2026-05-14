// lineeditor.go — readline-style line editing for terminal prompts.
// Copyright (C) 2025 R. S. Doiel
package termlib

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

// ErrInterrupted is returned by LineEditor.Prompt when the user presses Ctrl+C.
var ErrInterrupted = errors.New("interrupted")

/** LineEditor provides readline-style line editing at a terminal prompt.
 *
 * Supported keys:
 *   Left / Right arrows  — move cursor within the current line
 *   Home / End           — jump to start or end of line
 *   Up / Down arrows     — cycle through command history (only on first line)
 *   Backspace            — delete the character before the cursor
 *   Tab                  — complete the current word using Completer (if set);
 *                          first Tab lists all matches and fills the longest
 *                          common prefix; subsequent Tabs cycle through matches
 *   Ctrl+A               — move to beginning of current line
 *   Ctrl+E               — move to end of current line
 *   Ctrl+J               — insert a newline for multi-line input; Enter submits
 *   Ctrl+K               — kill (delete) from cursor to end of current line
 *   Ctrl+C               — cancel input; returns ErrInterrupted
 *   Ctrl+D               — EOF on an empty buffer; delete character under cursor otherwise
 *   Ctrl+X Ctrl+E        — open $EDITOR (falling back to $VISUAL then vi) to
 *                          compose or edit the prompt; when the editor exits the
 *                          saved content is returned as the line result
 *
 * Multi-line input: Ctrl+J appends a newline to the buffer and moves to the
 * next visual line (displayed with a "...  " continuation prompt). Enter
 * (Ctrl+M) submits the entire buffer, including embedded newlines. History
 * navigation is disabled once the buffer contains a newline. Backspace across
 * a newline merges the current line back onto the previous one.
 *
 * When stdin is not a TTY (e.g. piped input in tests), Prompt falls back to
 * plain line reading without raw-mode terminal manipulation.
 *
 * Example:
 *   le := termlib.NewLineEditor(os.Stdin, os.Stdout)
 *   line, err := le.Prompt("myapp > ")
 *   if err == termlib.ErrInterrupted || err == io.EOF { ... }
 *   le.AppendHistory(line)
 */
type LineEditor struct {
	in        *os.File
	out       io.Writer
	history   []string
	histBuf   string              // draft saved while navigating history
	Completer func(line string) []string // optional; receives text up to cursor, returns word candidates
}

/** NewLineEditor creates a LineEditor that reads from in and writes to out.
 *
 * Parameters:
 *   in  (*os.File)  — input file, typically os.Stdin.
 *   out (io.Writer) — output destination, typically os.Stdout.
 *
 * Returns:
 *   *LineEditor — ready to use; call Prompt to read a line.
 *
 * Example:
 *   le := termlib.NewLineEditor(os.Stdin, os.Stdout)
 */
func NewLineEditor(in *os.File, out io.Writer) *LineEditor {
	return &LineEditor{in: in, out: out}
}

/** AppendHistory adds line to the history list if it is non-empty and
 * differs from the most recent entry. Duplicate consecutive entries are
 * silently dropped.
 *
 * Parameters:
 *   line (string) — the line to record.
 *
 * Example:
 *   le.AppendHistory(input)
 */
func (le *LineEditor) AppendHistory(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	if len(le.history) > 0 && le.history[len(le.history)-1] == line {
		return
	}
	le.history = append(le.history, line)
}

/** Prompt displays prompt, then reads and returns one edited line.
 * The terminal is placed in raw mode for the duration of the call and
 * restored before returning. If raw mode is unavailable (e.g. stdin is a
 * pipe) the call falls back to plain unbuffered line reading.
 *
 * Long lines scroll horizontally rather than wrapping: the display shows a
 * window over the buffer that pans to keep the cursor visible. Left/Right
 * arrows let the user navigate to any part of the line.
 *
 * Parameters:
 *   prompt (string) — text printed before the cursor; must contain no ANSI
 *                     escape sequences (their widths are not tracked).
 *
 * Returns:
 *   string — the line the user typed, without the trailing newline.
 *   error  — ErrInterrupted on Ctrl+C, io.EOF on Ctrl+D with empty input,
 *             or an I/O error from the underlying file.
 *
 * Example:
 *   line, err := le.Prompt("harvey > ")
 */
func (le *LineEditor) Prompt(prompt string) (string, error) {
	fd := int(le.in.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return le.fallback(prompt)
	}
	defer term.Restore(fd, oldState)

	termWidth, _, err := term.GetSize(fd)
	if err != nil || termWidth <= 0 {
		termWidth = 80
	}

	const contPrompt = "...  " // shown on lines 2+ of multi-line input

	io.WriteString(le.out, prompt)

	buf := []rune{}
	pos := 0
	viewOffset := 0        // horizontal scroll offset, relative to current line's start
	histIdx := len(le.history)
	lineCount := 0         // number of '\n' characters currently in buf

	// currentLineStart returns the buf index where the current visual line begins
	// (one past the last '\n' before pos, or 0 if none).
	currentLineStart := func() int {
		for i := pos - 1; i >= 0; i-- {
			if buf[i] == '\n' {
				return i + 1
			}
		}
		return 0
	}

	// redraw repaints only the current visual line. Previous lines are frozen on
	// screen above. The horizontal viewport pans automatically to keep pos in view.
	redraw := func() {
		lineStart := currentLineStart()

		// Find end of current line (stop at the next '\n', if any).
		lineEnd := len(buf)
		for i := lineStart; i < len(buf); i++ {
			if buf[i] == '\n' {
				lineEnd = i
				break
			}
		}

		curPrompt := prompt
		if lineCount > 0 {
			curPrompt = contPrompt
		}
		curPromptLen := utf8.RuneCountInString(curPrompt)
		vw := termWidth - curPromptLen
		if vw < 1 {
			vw = 1
		}

		localPos := pos - lineStart // cursor position within the current line

		// Pan viewport to keep cursor visible.
		if localPos < viewOffset {
			viewOffset = localPos
		} else if localPos >= viewOffset+vw {
			viewOffset = localPos - vw + 1
		}

		dispStart := lineStart + viewOffset
		dispEnd := dispStart + vw
		if dispEnd > lineEnd {
			dispEnd = lineEnd
		}

		io.WriteString(le.out, "\r")
		io.WriteString(le.out, curPrompt)
		io.WriteString(le.out, string(buf[dispStart:dispEnd]))
		io.WriteString(le.out, "\033[K") // clear to end of line
		if back := dispEnd - pos; back > 0 {
			fmt.Fprintf(le.out, "\033[%dD", back)
		}
	}

	b := make([]byte, 1)
	ctrlXPending := false

	// Tab completion state — reset whenever a non-Tab key is pressed.
	var tabMatches []string
	var tabWordStart int // rune index in buf where the word being completed begins
	var tabIdx int       // next match index for cycling
	lastWasTab := false

	for {
		if _, err := le.in.Read(b); err != nil {
			return string(buf), err
		}
		ch := b[0]

		// Ctrl+X chord: wait for the second key.
		if ctrlXPending {
			ctrlXPending = false
			lastWasTab = false
			if ch == 0x05 { // Ctrl+E — open $EDITOR
				term.Restore(fd, oldState)
				result, edErr := le.openEditor(buf)
				if edErr == nil {
					io.WriteString(le.out, "\r\n")
					return result, nil
				}
				// Editor failed — re-enter raw mode and continue editing.
				if _, merr := term.MakeRaw(fd); merr != nil {
					return string(buf), merr
				}
				fmt.Fprintf(le.out, "\r\n  (editor: %v)\r\n", edErr)
				io.WriteString(le.out, prompt)
				redraw()
			}
			// Unrecognised Ctrl+X chord — silently discard both keys.
			continue
		}

		// Snapshot and reset tab state; the Tab case will set lastWasTab back to true.
		prevWasTab := lastWasTab
		lastWasTab = false

		switch {
		case ch == 0x09: // Tab — complete the current word using Completer
			if le.Completer == nil {
				break
			}
			lastWasTab = true
			// Find the start of the word being completed (last whitespace before cursor).
			wordStart := 0
			for i := pos - 1; i >= 0; i-- {
				if buf[i] == ' ' || buf[i] == '\t' {
					wordStart = i + 1
					break
				}
			}
			if !prevWasTab {
				// First Tab: compute a fresh candidate list.
				tabMatches = le.Completer(string(buf[:pos]))
				tabWordStart = wordStart
				tabIdx = 0
			}
			if len(tabMatches) == 0 {
				break // no candidates
			}
			if !prevWasTab && len(tabMatches) > 1 {
				// First Tab with multiple matches: print the list, then fill common prefix.
				io.WriteString(le.out, "\r\n")
				for _, m := range tabMatches {
					fmt.Fprintf(le.out, "  %s\r\n", m)
				}
				io.WriteString(le.out, prompt)
				prefix := leCommonPrefix(tabMatches)
				completion := []rune(prefix)
				buf = append(append(append([]rune{}, buf[:tabWordStart]...), completion...), buf[pos:]...)
				pos = tabWordStart + len(completion)
			} else {
				// Single match, or subsequent Tab: insert/cycle to the next candidate.
				completion := []rune(tabMatches[tabIdx%len(tabMatches)])
				buf = append(append(append([]rune{}, buf[:tabWordStart]...), completion...), buf[pos:]...)
				pos = tabWordStart + len(completion)
				tabIdx++
				if len(tabMatches) == 1 {
					tabMatches = nil // done; reset for next word
					lastWasTab = false
				}
			}
			redraw()

		case ch == '\r': // Enter — submit the full (possibly multi-line) buffer
			io.WriteString(le.out, "\r\n")
			return string(buf), nil

		case ch == 0x0a: // Ctrl+J — insert newline (begin next input line)
			buf = leInsertRune(buf, pos, '\n')
			pos++
			lineCount++
			viewOffset = 0
			io.WriteString(le.out, "\r\n")
			redraw()

		case ch == 0x03: // Ctrl+C
			io.WriteString(le.out, "\r\n")
			return "", ErrInterrupted

		case ch == 0x04: // Ctrl+D — EOF on empty buffer; delete under cursor otherwise
			if len(buf) == 0 {
				io.WriteString(le.out, "\r\n")
				return "", io.EOF
			}
			// Don't delete across a newline boundary.
			if pos < len(buf) && buf[pos] != '\n' {
				buf = append(buf[:pos], buf[pos+1:]...)
				redraw()
			}

		case ch == 0x01: // Ctrl+A — beginning of current line
			pos = currentLineStart()
			viewOffset = 0
			redraw()

		case ch == 0x05: // Ctrl+E — end of current line
			lineStart := currentLineStart()
			lineEnd := len(buf)
			for i := lineStart; i < len(buf); i++ {
				if buf[i] == '\n' {
					lineEnd = i
					break
				}
			}
			pos = lineEnd
			redraw()

		case ch == 0x0b: // Ctrl+K — kill to end of current line (not past '\n')
			killEnd := len(buf)
			for i := pos; i < len(buf); i++ {
				if buf[i] == '\n' {
					killEnd = i
					break
				}
			}
			buf = append(buf[:pos], buf[killEnd:]...)
			redraw()

		case ch == 0x18: // Ctrl+X — first key of a two-key chord
			ctrlXPending = true

		case ch == 0x7f || ch == 0x08: // Backspace / Ctrl+H
			if pos > 0 {
				if buf[pos-1] == '\n' {
					// Backspace across a newline: clear the current visual line,
					// move up to the previous line, and merge the two lines.
					io.WriteString(le.out, "\r\033[K") // erase current visual line
					fmt.Fprintf(le.out, "\033[1A")     // cursor up one line
					buf = append(buf[:pos-1], buf[pos:]...)
					pos--
					lineCount--
					viewOffset = 0
					redraw()
				} else {
					buf = append(buf[:pos-1], buf[pos:]...)
					pos--
					redraw()
				}
			}

		case ch == 0x1b: // Escape — consume the rest of the sequence
			switch seq := le.readEscSeq(); seq {
			case "[A", "OA": // Up arrow — history previous (disabled in multi-line mode)
				if lineCount == 0 && histIdx > 0 {
					if histIdx == len(le.history) {
						le.histBuf = string(buf) // save current draft
					}
					histIdx--
					buf = []rune(le.history[histIdx])
					pos = len(buf)
					viewOffset = 0
					// Recompute lineCount from the restored entry.
					lineCount = strings.Count(string(buf), "\n")
					redraw()
				}
			case "[B", "OB": // Down arrow — history next (disabled in multi-line mode)
				if lineCount == 0 && histIdx < len(le.history) {
					histIdx++
					if histIdx == len(le.history) {
						buf = []rune(le.histBuf)
					} else {
						buf = []rune(le.history[histIdx])
					}
					pos = len(buf)
					viewOffset = 0
					lineCount = strings.Count(string(buf), "\n")
					redraw()
				}
			case "[C", "OC": // Right arrow — stay within current line
				if pos < len(buf) && buf[pos] != '\n' {
					pos++
					redraw()
				}
			case "[D", "OD": // Left arrow — stay within current line
				if pos > 0 && buf[pos-1] != '\n' {
					pos--
					redraw()
				}
			case "[H", "OH", "[1~": // Home — beginning of current line
				pos = currentLineStart()
				viewOffset = 0
				redraw()
			case "[F", "OF", "[4~": // End — end of current line
				lineStart := currentLineStart()
				lineEnd := len(buf)
				for i := lineStart; i < len(buf); i++ {
					if buf[i] == '\n' {
						lineEnd = i
						break
					}
				}
				pos = lineEnd
				redraw()
			}

		case ch >= 0x20 && ch < 0x7f: // Printable ASCII
			buf = leInsertRune(buf, pos, rune(ch))
			pos++
			redraw()

		case ch >= 0xc0: // UTF-8 multi-byte lead byte
			if r := le.readUTF8Tail(ch); r != utf8.RuneError {
				buf = leInsertRune(buf, pos, r)
				pos++
				redraw()
			}
		}
	}
}

// openEditor writes buf to a temp file, opens it in $EDITOR (falling back to
// $VISUAL, then vi), and returns the saved content with trailing newlines trimmed.
// The caller must have already restored the terminal before calling this.
func (le *LineEditor) openEditor(buf []rune) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}

	tmp, err := os.CreateTemp("", "termlib-edit-*.txt")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if len(buf) > 0 {
		if _, err := tmp.WriteString(string(buf)); err != nil {
			tmp.Close()
			return "", err
		}
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}

	outW := os.Stdout
	if f, ok := le.out.(*os.File); ok {
		outW = f
	}
	cmd := exec.Command(editor, tmpPath)
	cmd.Stdin = le.in
	cmd.Stdout = outW
	cmd.Stderr = outW
	if err := cmd.Run(); err != nil {
		return "", err
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(data), "\r\n"), nil
}

// readEscSeq reads the bytes that follow an ESC character and returns a
// short string identifying the sequence, e.g. "[A" for up-arrow.
// It handles both CSI (\x1b[…) and SS3 (\x1bO…) forms.
func (le *LineEditor) readEscSeq() string {
	b := make([]byte, 1)
	if _, err := le.in.Read(b); err != nil {
		return ""
	}
	switch b[0] {
	case '[': // CSI sequence
		var seq []byte
		for {
			if _, err := le.in.Read(b); err != nil {
				break
			}
			seq = append(seq, b[0])
			// Sequences terminate on a letter or '~'
			if (b[0] >= 'A' && b[0] <= 'Z') || (b[0] >= 'a' && b[0] <= 'z') || b[0] == '~' {
				break
			}
		}
		return "[" + string(seq)
	case 'O': // SS3 sequence (application cursor keys)
		if _, err := le.in.Read(b); err != nil {
			return ""
		}
		return "O" + string(b[:1])
	default:
		return ""
	}
}

// readUTF8Tail reads the continuation bytes for a multi-byte UTF-8
// sequence whose leading byte is lead, and returns the decoded rune.
func (le *LineEditor) readUTF8Tail(lead byte) rune {
	var tailLen int
	switch {
	case lead >= 0xf0:
		tailLen = 3
	case lead >= 0xe0:
		tailLen = 2
	default:
		tailLen = 1
	}
	tail := make([]byte, tailLen)
	le.in.Read(tail)
	full := append([]byte{lead}, tail...)
	r, _ := utf8.DecodeRune(full)
	return r
}

// fallback reads a plain line without raw-mode manipulation. Used when
// stdin is not a TTY (e.g. during tests with piped input).
func (le *LineEditor) fallback(prompt string) (string, error) {
	io.WriteString(le.out, prompt)
	var sb strings.Builder
	b := make([]byte, 1)
	for {
		_, err := le.in.Read(b)
		if err != nil {
			return sb.String(), err
		}
		if b[0] == '\n' || b[0] == '\r' {
			return sb.String(), nil
		}
		sb.WriteByte(b[0])
	}
}

// leInsertRune inserts r into buf at position pos and returns the new slice.
func leInsertRune(buf []rune, pos int, r rune) []rune {
	buf = append(buf, 0)
	copy(buf[pos+1:], buf[pos:])
	buf[pos] = r
	return buf
}

// leCommonPrefix returns the longest string that is a prefix of every element
// of strs. Returns "" when strs is empty.
func leCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	prefix := []rune(strs[0])
	for _, s := range strs[1:] {
		sr := []rune(s)
		max := len(prefix)
		if len(sr) < max {
			max = len(sr)
		}
		i := 0
		for i < max && prefix[i] == sr[i] {
			i++
		}
		prefix = prefix[:i]
		if len(prefix) == 0 {
			return ""
		}
	}
	return string(prefix)
}
