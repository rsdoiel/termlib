// lineeditor.go — readline-style line editing for terminal prompts.
// Copyright (C) 2025 R. S. Doiel
package termlib

import (
	"errors"
	"fmt"
	"io"
	"os"
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
 *   Up / Down arrows     — cycle through command history
 *   Backspace            — delete the character before the cursor
 *   Ctrl+A               — move to beginning of line
 *   Ctrl+E               — move to end of line
 *   Ctrl+K               — kill (delete) from cursor to end of line
 *   Ctrl+C               — cancel input; returns ErrInterrupted
 *   Ctrl+D               — EOF on an empty line; delete-under-cursor otherwise
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
	in      *os.File
	out     io.Writer
	history []string
	histBuf string // draft saved while navigating history
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

	io.WriteString(le.out, prompt)

	buf := []rune{}
	pos := 0
	histIdx := len(le.history) // past the end = current (new) line

	// redraw reprints the prompt and buffer from the start of the current
	// terminal line, then positions the cursor at pos.
	redraw := func() {
		io.WriteString(le.out, "\r")
		io.WriteString(le.out, prompt)
		io.WriteString(le.out, string(buf))
		io.WriteString(le.out, "\033[K") // clear to end of line
		if back := len(buf) - pos; back > 0 {
			fmt.Fprintf(le.out, "\033[%dD", back)
		}
	}

	b := make([]byte, 1)
	for {
		if _, err := le.in.Read(b); err != nil {
			return string(buf), err
		}

		switch ch := b[0]; {

		case ch == '\r' || ch == '\n':
			io.WriteString(le.out, "\r\n")
			return string(buf), nil

		case ch == 0x03: // Ctrl+C
			io.WriteString(le.out, "\r\n")
			return "", ErrInterrupted

		case ch == 0x04: // Ctrl+D — EOF on empty line, delete-under-cursor otherwise
			if len(buf) == 0 {
				io.WriteString(le.out, "\r\n")
				return "", io.EOF
			}
			if pos < len(buf) {
				buf = append(buf[:pos], buf[pos+1:]...)
				redraw()
			}

		case ch == 0x01: // Ctrl+A — beginning of line
			pos = 0
			redraw()

		case ch == 0x05: // Ctrl+E — end of line
			pos = len(buf)
			redraw()

		case ch == 0x0b: // Ctrl+K — kill to end of line
			buf = buf[:pos]
			redraw()

		case ch == 0x7f || ch == 0x08: // Backspace / Ctrl+H
			if pos > 0 {
				buf = append(buf[:pos-1], buf[pos:]...)
				pos--
				redraw()
			}

		case ch == 0x1b: // Escape — consume the rest of the sequence
			switch seq := le.readEscSeq(); seq {
			case "[A", "OA": // Up arrow — history previous
				if histIdx > 0 {
					if histIdx == len(le.history) {
						le.histBuf = string(buf) // save current draft
					}
					histIdx--
					buf = []rune(le.history[histIdx])
					pos = len(buf)
					redraw()
				}
			case "[B", "OB": // Down arrow — history next
				if histIdx < len(le.history) {
					histIdx++
					if histIdx == len(le.history) {
						buf = []rune(le.histBuf)
					} else {
						buf = []rune(le.history[histIdx])
					}
					pos = len(buf)
					redraw()
				}
			case "[C", "OC": // Right arrow
				if pos < len(buf) {
					pos++
					redraw()
				}
			case "[D", "OD": // Left arrow
				if pos > 0 {
					pos--
					redraw()
				}
			case "[H", "OH", "[1~": // Home
				pos = 0
				redraw()
			case "[F", "OF", "[4~": // End
				pos = len(buf)
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
