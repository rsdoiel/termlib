// Copyright (C) 2025-2026 R. S. Doiel
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License

//go:build linux

package termlib

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// These tests drive LineEditor.Prompt through a real pseudo-terminal, so the
// raw-mode path (not the plain-line fallback) runs. They are Linux only: the
// ioctls that open a pty differ on macOS and Windows has none.

// pty is a pseudo-terminal pair. The slave is the editor's terminal; the
// master is what a person would see and type on.
type pty struct {
	master, slave *os.File
	mu            sync.Mutex
	seen          strings.Builder
}

// newPty opens a pty with the given width and starts collecting everything the
// editor writes to it.
func newPty(t *testing.T, cols uint16) *pty {
	t.Helper()
	mfd, err := unix.Open("/dev/ptmx", unix.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		t.Skipf("no /dev/ptmx: %v", err)
	}
	if err := unix.IoctlSetPointerInt(mfd, unix.TIOCSPTLCK, 0); err != nil {
		t.Fatal(err)
	}
	n, err := unix.IoctlGetInt(mfd, unix.TIOCGPTN)
	if err != nil {
		t.Fatal(err)
	}
	sfd, err := unix.Open("/dev/pts/"+itoa(n), unix.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		t.Skipf("cannot open the pty slave: %v", err)
	}
	if err := unix.IoctlSetWinsize(sfd, unix.TIOCSWINSZ, &unix.Winsize{Row: 24, Col: cols}); err != nil {
		t.Fatal(err)
	}
	p := &pty{master: os.NewFile(uintptr(mfd), "ptmx"), slave: os.NewFile(uintptr(sfd), "pts")}
	t.Cleanup(func() { p.slave.Close(); p.master.Close() })
	go func() {
		b := make([]byte, 4096)
		for {
			n, err := p.master.Read(b)
			p.mu.Lock()
			p.seen.Write(b[:n])
			p.mu.Unlock()
			if err != nil {
				return
			}
		}
	}()
	return p
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for ; n > 0; n /= 10 {
		s = string(rune('0'+n%10)) + s
	}
	return s
}

// output returns what the editor has written so far.
func (p *pty) output() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.seen.String()
}

// waitFor blocks until the output contains s, or fails the test.
func (p *pty) waitFor(t *testing.T, s string) {
	t.Helper()
	for end := time.Now().Add(5 * time.Second); time.Now().Before(end); time.Sleep(2 * time.Millisecond) {
		if strings.Contains(p.output(), s) {
			return
		}
	}
	t.Fatalf("timed out waiting for %q in output %q", s, p.output())
}

type promptResult struct {
	line string
	err  error
}

// start runs le.Prompt in a goroutine and waits until it has printed the
// prompt, which it does only after it has put the terminal in raw mode; typing
// earlier would be echoed and line-buffered by the terminal driver instead.
func (p *pty) start(t *testing.T, le *LineEditor, prompt string) <-chan promptResult {
	t.Helper()
	done := make(chan promptResult, 1)
	go func() {
		line, err := le.Prompt(prompt)
		done <- promptResult{line, err}
	}()
	p.waitFor(t, prompt[strings.LastIndex(prompt, "\n")+1:])
	return done
}

// typeKeys sends keys as if typed.
func (p *pty) typeKeys(t *testing.T, keys string) {
	t.Helper()
	if _, err := p.master.WriteString(keys); err != nil {
		t.Fatal(err)
	}
}

// finish waits for Prompt to return.
func finish(t *testing.T, done <-chan promptResult) promptResult {
	t.Helper()
	select {
	case r := <-done:
		return r
	case <-time.After(5 * time.Second):
		t.Fatal("Prompt did not return")
		return promptResult{}
	}
}

// newEditor makes a pty of the given width and an editor on it.
func newEditor(t *testing.T, cols uint16) (*LineEditor, *pty) {
	t.Helper()
	p := newPty(t, cols)
	return NewLineEditor(p.slave, p.slave), p
}

// edit types keys into a fresh Prompt and returns its result.
func edit(t *testing.T, le *LineEditor, p *pty, keys string) promptResult {
	t.Helper()
	done := p.start(t, le, "> ")
	p.typeKeys(t, keys)
	return finish(t, done)
}

const (
	keyUp    = "\x1b[A"
	keyDown  = "\x1b[B"
	keyRight = "\x1b[C"
	keyLeft  = "\x1b[D"
	keyHome  = "\x1b[H"
	keyEnd   = "\x1b[F"
)

/** TestPromptEditingKeys types key sequences into Prompt on a real terminal
 * and checks the line each returns: insertion, cursor movement, deletion, the
 * kill and delete keys, and multi-byte characters.
 */
func TestPromptEditingKeys(t *testing.T) {
	for _, tc := range []struct{ name, keys, want string }{
		{"plain", "hello\r", "hello"},
		{"empty", "\r", ""},
		{"insert at cursor", "ac" + keyLeft + "b\r", "abc"},
		{"ss3 arrows", "ac\x1bOD" + "b\r", "abc"},
		{"backspace", "abc\x7f\r", "ab"},
		{"ctrl-h", "abc\x08\r", "ab"},
		{"backspace at start", "\x7fabc\r", "abc"},
		{"left stops at start", "b" + keyLeft + keyLeft + keyLeft + "a\r", "ab"},
		{"right stops at end", "a" + keyRight + keyRight + "b\r", "ab"},
		{"ctrl-a ctrl-e", "bc\x01a\x05d\r", "abcd"},
		{"home end csi", "bc" + keyHome + "a" + keyEnd + "d\r", "abcd"},
		{"home end tilde", "bc\x1b[1~a\x1b[4~d\r", "abcd"},
		{"ctrl-k kills to end", "hello world\x01" + strings.Repeat(keyRight, 5) + "\x0b\r", "hello"},
		{"ctrl-d deletes under cursor", "abc" + keyLeft + "\x04\r", "ab"},
		{"ctrl-d at end of line does nothing", "ab\x04\r", "ab"},
		{"multi-byte insert", "é日😀\r", "é日😀"},
		{"backspace removes a whole rune", "é日\x7f\r", "é"},
		{"unknown escape sequence is ignored", "a\x1b[3~b\r", "ab"},
		{"unassigned control keys are ignored", "a\x02\x06b\r", "ab"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			le, p := newEditor(t, 80)
			if r := edit(t, le, p, tc.keys); r.err != nil || r.line != tc.want {
				t.Errorf("got %q, %v; want %q", r.line, r.err, tc.want)
			}
		})
	}
}

/** TestPromptCancelAndEOF checks Ctrl+C and Ctrl+D on an empty line. */
func TestPromptCancelAndEOF(t *testing.T) {
	le, p := newEditor(t, 80)
	if r := edit(t, le, p, "abc\x03"); !errors.Is(r.err, ErrInterrupted) || r.line != "" {
		t.Errorf("Ctrl+C: got %q, %v; want empty and ErrInterrupted", r.line, r.err)
	}
	le, p = newEditor(t, 80)
	if r := edit(t, le, p, "\x04"); r.err != io.EOF || r.line != "" {
		t.Errorf("Ctrl+D on an empty line: got %q, %v; want empty and io.EOF", r.line, r.err)
	}
}

/** TestPromptRestoresTheTerminal checks that the terminal is back in its
 * original mode (echo on, canonical) after Prompt returns, on every exit.
 */
func TestPromptRestoresTheTerminal(t *testing.T) {
	for _, keys := range []string{"x\r", "x\x03", "\x04"} {
		le, p := newEditor(t, 80)
		before, err := unix.IoctlGetTermios(int(p.slave.Fd()), unix.TCGETS)
		if err != nil {
			t.Fatal(err)
		}
		edit(t, le, p, keys)
		after, err := unix.IoctlGetTermios(int(p.slave.Fd()), unix.TCGETS)
		if err != nil {
			t.Fatal(err)
		}
		if *before != *after {
			t.Errorf("keys %q: terminal settings changed: before %+v, after %+v", keys, *before, *after)
		}
	}
}

/** TestPromptDoesNotEchoTypedKeys checks that the terminal driver's echo is
 * off while editing: the prompt and the editor's own redraws are the only
 * output, so a typed secret is shown exactly once, by the editor.
 */
func TestPromptDoesNotEchoTypedKeys(t *testing.T) {
	le, p := newEditor(t, 80)
	edit(t, le, p, "zq\r")
	if strings.Contains(p.output(), "\r\nzq") || strings.Count(p.output(), "zq") > 1 {
		t.Errorf("typed text appears to be echoed twice: %q", p.output())
	}
}

/** TestPromptHistory checks Up and Down: walking back, returning to the draft
 * typed before the first Up, stopping at either end, and a restored entry
 * being editable. */
func TestPromptHistory(t *testing.T) {
	for _, tc := range []struct{ name, keys, want string }{
		{"up once", keyUp + "\r", "two"},
		{"up twice", keyUp + keyUp + "\r", "one"},
		{"up past the oldest stays on it", keyUp + keyUp + keyUp + keyUp + "\r", "one"},
		{"up then down is the empty draft", keyUp + keyDown + "\r", ""},
		{"down with no history ahead does nothing", keyDown + "ab\r", "ab"},
		{"draft is kept", "dr" + keyUp + keyUp + keyDown + keyDown + "\r", "dr"},
		{"restored entry is editable", keyUp + "\x7fX\r", "twX"},
		{"ss3 up", "\x1bOA\r", "two"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			le, p := newEditor(t, 80)
			le.SetHistory([]string{"one", "two"})
			if r := edit(t, le, p, tc.keys); r.err != nil || r.line != tc.want {
				t.Errorf("got %q, %v; want %q", r.line, r.err, tc.want)
			}
		})
	}
}

/** TestPromptMultiLine checks Ctrl+J: the buffer keeps its newlines, Enter
 * submits the whole of it, history is off while the buffer has a newline,
 * Left and Ctrl+K stop at a line boundary, and Backspace joins two lines.
 */
func TestPromptMultiLine(t *testing.T) {
	for _, tc := range []struct{ name, keys, want string }{
		{"two lines", "ab\ncd\r", "ab\ncd"},
		{"three lines", "a\nb\nc\r", "a\nb\nc"},
		{"history is off in multi-line mode", "a\nb" + keyUp + "\r", "a\nb"},
		{"backspace joins lines", "a\n\x7fb\r", "ab"},
		{"left does not cross a newline", "a\nb" + keyLeft + keyLeft + "X\r", "a\nXb"},
		{"ctrl-k stops at the newline", "ab\ncd\x01\x0b\r", "ab\n"},
		{"ctrl-a is start of the current line", "ab\ncd\x01X\r", "ab\nXcd"},
		{"ctrl-e is end of the current line", "ab\ncd" + keyUp + "\x05X\r", "ab\ncdX"},
		{"ctrl-d on line two deletes under the cursor, not the newline", "a\nb\x01" + keyLeft + "\x04\r", "a\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			le, p := newEditor(t, 80)
			le.SetHistory([]string{"old"})
			if r := edit(t, le, p, tc.keys); r.err != nil || r.line != tc.want {
				t.Errorf("got %q, %v; want %q", r.line, r.err, tc.want)
			}
		})
	}
}

/** TestPromptTabCompletion checks the Completer hook: a single candidate is
 * inserted, several candidates are listed with their common prefix filled in
 * and then cycled with further Tabs, only the word before the cursor is
 * replaced, and Tab does nothing without a Completer or without candidates. */
func TestPromptTabCompletion(t *testing.T) {
	commands := func(line string) []string {
		word := line[strings.LastIndex(line, " ")+1:]
		var out []string
		for _, c := range []string{"/help", "/history", "/quit"} {
			if strings.HasPrefix(c, word) {
				out = append(out, c)
			}
		}
		return out
	}
	for _, tc := range []struct{ name, keys, want string }{
		{"single match", "/q\t\r", "/quit"},
		{"common prefix only", "/\t\r", "/"},
		{"common prefix of several", "/h\t\r", "/h"},
		{"first cycle", "/h\t\t\r", "/help"},
		{"second cycle", "/h\t\t\t\r", "/history"},
		{"cycle wraps", "/h\t\t\t\t\r", "/help"},
		{"only the current word", "say /q\t\r", "say /quit"},
		{"text after the cursor is kept", "/q tail" + strings.Repeat(keyLeft, 5) + "\t\r", "/quit tail"},
		{"no candidates", "zzz\t\r", "zzz"},
		{"typing after a tab starts a new completion", "/h\t\t" + "\x7f" + "\x7f" + "\x7f" + "\x7f" + "\x7f" + "/q\t\r", "/quit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			le, p := newEditor(t, 80)
			le.Completer = commands
			if r := edit(t, le, p, tc.keys); r.err != nil || r.line != tc.want {
				t.Errorf("got %q, %v; want %q", r.line, r.err, tc.want)
			}
		})
	}
	t.Run("candidates are listed", func(t *testing.T) {
		le, p := newEditor(t, 80)
		le.Completer = commands
		edit(t, le, p, "/h\t\r")
		for _, want := range []string{"  /help", "  /history"} {
			if !strings.Contains(p.output(), want) {
				t.Errorf("output lacks %q: %q", want, p.output())
			}
		}
	})
	t.Run("no completer", func(t *testing.T) {
		le, p := newEditor(t, 80)
		if r := edit(t, le, p, "ab\tc\r"); r.line != "abc" {
			t.Errorf("got %q, want Tab ignored", r.line)
		}
	})
}

// editorScript writes an executable script that stands in for $EDITOR. The
// script receives the temporary file as $1; body is shell.
func editorScript(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "editor.sh")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

/** TestPromptExternalEditor checks Ctrl+X Ctrl+E: the editor is started on a
 * file holding the buffer, what it saves is returned with trailing newlines
 * trimmed, $EDITOR is preferred to $VISUAL, a failing editor leaves the
 * session editable, and an unknown chord is dropped. */
func TestPromptExternalEditor(t *testing.T) {
	t.Run("saved content is returned", func(t *testing.T) {
		t.Setenv("EDITOR", editorScript(t, `printf 'from editor\n\n' > "$1"`))
		le, p := newEditor(t, 80)
		if r := edit(t, le, p, "draft\x18\x05"); r.err != nil || r.line != "from editor" {
			t.Errorf("got %q, %v; want %q", r.line, r.err, "from editor")
		}
	})
	t.Run("editor sees the buffer", func(t *testing.T) {
		seen := filepath.Join(t.TempDir(), "seen")
		t.Setenv("EDITOR", editorScript(t, `cp "$1" `+seen))
		le, p := newEditor(t, 80)
		r := edit(t, le, p, "one\ntwo\x18\x05")
		got, err := os.ReadFile(seen)
		if err != nil || string(got) != "one\ntwo" {
			t.Errorf("editor saw %q, %v; want %q", got, err, "one\ntwo")
		}
		if r.line != "one\ntwo" {
			t.Errorf("an untouched file returned %q, want the buffer", r.line)
		}
	})
	t.Run("EDITOR is preferred to VISUAL", func(t *testing.T) {
		t.Setenv("EDITOR", editorScript(t, `printf editor > "$1"`))
		t.Setenv("VISUAL", editorScript(t, `printf visual > "$1"`))
		le, p := newEditor(t, 80)
		if r := edit(t, le, p, "\x18\x05"); r.line != "editor" {
			t.Errorf("got %q, want editor", r.line)
		}
	})
	t.Run("VISUAL is the fallback", func(t *testing.T) {
		t.Setenv("EDITOR", "")
		t.Setenv("VISUAL", editorScript(t, `printf visual > "$1"`))
		le, p := newEditor(t, 80)
		if r := edit(t, le, p, "\x18\x05"); r.line != "visual" {
			t.Errorf("got %q, want visual", r.line)
		}
	})
	t.Run("a failing editor leaves the session editable", func(t *testing.T) {
		t.Setenv("EDITOR", editorScript(t, `exit 3`))
		le, p := newEditor(t, 80)
		r := edit(t, le, p, "ab\x18\x05cd\r")
		if r.err != nil || r.line != "abcd" {
			t.Errorf("got %q, %v; want abcd", r.line, r.err)
		}
		if !strings.Contains(p.output(), "(editor: exit status 3)") {
			t.Errorf("the failure was not reported: %q", p.output())
		}
	})
	t.Run("an unknown chord is dropped", func(t *testing.T) {
		le, p := newEditor(t, 80)
		if r := edit(t, le, p, "a\x18qb\r"); r.line != "ab" {
			t.Errorf("got %q, want ab", r.line)
		}
	})
}

/** TestPromptNarrowTerminal checks lines that do not fit the terminal: a long
 * line scrolls and is still returned whole, editing in the scrolled part
 * works, and the screen is never written wider than the terminal. */
func TestPromptNarrowTerminal(t *testing.T) {
	long := strings.Repeat("0123456789", 5)
	t.Run("long line is returned whole", func(t *testing.T) {
		le, p := newEditor(t, 20)
		if r := edit(t, le, p, long+"\r"); r.err != nil || r.line != long {
			t.Errorf("got %q, %v", r.line, r.err)
		}
	})
	t.Run("edit in the middle of a scrolled line", func(t *testing.T) {
		le, p := newEditor(t, 20)
		keys := long + strings.Repeat(keyLeft, 30) + "X\r"
		want := long[:20] + "X" + long[20:]
		if r := edit(t, le, p, keys); r.line != want {
			t.Errorf("got %q, want %q", r.line, want)
		}
	})
	t.Run("home on a scrolled line", func(t *testing.T) {
		le, p := newEditor(t, 20)
		if r := edit(t, le, p, long+keyHome+"X\r"); r.line != "X"+long {
			t.Errorf("got %q", r.line)
		}
	})
}

/** TestPromptWiderThanTheTerminal is the v0.0.10 bug through the real path: a
 * prompt that is wider than the terminal, or has a newline in it, must leave
 * the typed input usable. The head is printed once and input still arrives. */
func TestPromptWiderThanTheTerminal(t *testing.T) {
	for _, prompt := range []string{
		strings.Repeat("p", 30) + "> ",
		"first line\nsecond> ",
		"banner\n" + strings.Repeat("p", 30) + "> ",
	} {
		le, p := newEditor(t, 20)
		done := p.start(t, le, prompt)
		p.typeKeys(t, "typed text"+keyLeft+keyLeft+"X\r")
		r := finish(t, done)
		if r.err != nil || r.line != "typed teXxt" {
			t.Errorf("prompt %q: got %q, %v; want %q", prompt, r.line, r.err, "typed teXxt")
		}
	}
}
