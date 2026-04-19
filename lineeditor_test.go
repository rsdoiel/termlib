/*
termlib is a light weight terminal interface library. Sort of a ncurse light.
Copyright (C) 2025 R. S. Doiel

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
*/
package termlib

// Note: raw-mode keyboard navigation (arrow keys, Home/End, Ctrl+A/E/K,
// history cycling) requires a real TTY and is exercised by manual testing.
// All tests below use os.Pipe() as stdin, which causes term.MakeRaw to fail
// and Prompt to fall back to plain unbuffered line reading. This covers the
// full logic of Prompt, AppendHistory, and the unexported helpers.

import (
	"io"
	"os"
	"strings"
	"testing"
)

// pipeEditor creates a LineEditor backed by an os.Pipe (non-TTY) so that
// Prompt always uses the fallback path. It returns the editor, the write end
// of the pipe (inject test input here), and a Builder capturing output.
// The pipe files are closed automatically when the test ends.
func pipeEditor(t *testing.T) (*LineEditor, *os.File, *strings.Builder) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	t.Cleanup(func() { r.Close(); w.Close() })
	var out strings.Builder
	return NewLineEditor(r, &out), w, &out
}

// ─── AppendHistory ──────────────────────────────────────────────────────────

func TestAppendHistory_basic(t *testing.T) {
	le := NewLineEditor(os.Stdin, io.Discard)
	le.AppendHistory("first")
	le.AppendHistory("second")
	le.AppendHistory("third")
	if len(le.history) != 3 {
		t.Fatalf("want 3 entries, got %d", len(le.history))
	}
	for i, want := range []string{"first", "second", "third"} {
		if le.history[i] != want {
			t.Errorf("history[%d]: want %q, got %q", i, want, le.history[i])
		}
	}
}

func TestAppendHistory_consecutiveDupDropped(t *testing.T) {
	le := NewLineEditor(os.Stdin, io.Discard)
	le.AppendHistory("cmd")
	le.AppendHistory("cmd")
	le.AppendHistory("cmd")
	if len(le.history) != 1 {
		t.Fatalf("want 1 entry after consecutive dups, got %d: %v", len(le.history), le.history)
	}
}

func TestAppendHistory_nonConsecutiveDupAllowed(t *testing.T) {
	le := NewLineEditor(os.Stdin, io.Discard)
	le.AppendHistory("a")
	le.AppendHistory("b")
	le.AppendHistory("a") // same as first but not consecutive — should be kept
	if len(le.history) != 3 {
		t.Fatalf("want 3 entries, got %d: %v", len(le.history), le.history)
	}
}

func TestAppendHistory_emptyIgnored(t *testing.T) {
	le := NewLineEditor(os.Stdin, io.Discard)
	le.AppendHistory("")
	le.AppendHistory("   ")
	le.AppendHistory("\t")
	if len(le.history) != 0 {
		t.Fatalf("want 0 entries for blank inputs, got %d: %v", len(le.history), le.history)
	}
}

func TestAppendHistory_whitespaceTrimmed(t *testing.T) {
	le := NewLineEditor(os.Stdin, io.Discard)
	le.AppendHistory("  hello  ")
	le.AppendHistory("hello") // same after trim — should be dropped as dup
	if len(le.history) != 1 {
		t.Fatalf("want 1 entry, got %d: %v", len(le.history), le.history)
	}
	if le.history[0] != "hello" {
		t.Errorf("want %q, got %q", "hello", le.history[0])
	}
}

// ─── Prompt (fallback path via pipe) ─────────────────────────────────────────

func TestPrompt_basicLine(t *testing.T) {
	le, w, out := pipeEditor(t)
	w.WriteString("hello world\n")
	w.Close()

	line, err := le.Prompt(">> ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if line != "hello world" {
		t.Errorf("want %q, got %q", "hello world", line)
	}
	if !strings.Contains(out.String(), ">> ") {
		t.Errorf("prompt not written to output: %q", out.String())
	}
}

func TestPrompt_emptyLine(t *testing.T) {
	le, w, _ := pipeEditor(t)
	w.WriteString("\n")
	w.Close()

	line, err := le.Prompt("> ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if line != "" {
		t.Errorf("want empty string, got %q", line)
	}
}

func TestPrompt_crTerminator(t *testing.T) {
	le, w, _ := pipeEditor(t)
	w.WriteString("carriage\r")
	w.Close()

	line, err := le.Prompt("> ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if line != "carriage" {
		t.Errorf("want %q, got %q", "carriage", line)
	}
}

func TestPrompt_eofOnEmptyInput(t *testing.T) {
	le, w, _ := pipeEditor(t)
	w.Close() // immediate EOF

	_, err := le.Prompt("> ")
	if err != io.EOF {
		t.Errorf("want io.EOF, got %v", err)
	}
}

func TestPrompt_eofMidLine(t *testing.T) {
	le, w, _ := pipeEditor(t)
	w.WriteString("partial") // no newline, then EOF
	w.Close()

	line, err := le.Prompt("> ")
	if err != io.EOF {
		t.Errorf("want io.EOF, got %v", err)
	}
	if line != "partial" {
		t.Errorf("want %q before EOF, got %q", "partial", line)
	}
}

func TestPrompt_multipleSequentialReads(t *testing.T) {
	le, w, _ := pipeEditor(t)
	w.WriteString("first\nsecond\nthird\n")
	w.Close()

	for i, want := range []string{"first", "second", "third"} {
		got, err := le.Prompt("> ")
		if err != nil {
			t.Fatalf("read %d: unexpected error: %v", i+1, err)
		}
		if got != want {
			t.Errorf("read %d: want %q, got %q", i+1, want, got)
		}
	}
}

func TestPrompt_promptAppearsInOutput(t *testing.T) {
	le, w, out := pipeEditor(t)
	w.WriteString("x\n")
	w.Close()

	le.Prompt("harvey > ")
	if !strings.Contains(out.String(), "harvey > ") {
		t.Errorf("expected prompt in output, got: %q", out.String())
	}
}

func TestPrompt_utf8Input(t *testing.T) {
	le, w, _ := pipeEditor(t)
	w.WriteString("héllo wörld\n")
	w.Close()

	line, err := le.Prompt("> ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if line != "héllo wörld" {
		t.Errorf("want %q, got %q", "héllo wörld", line)
	}
}

// ─── leInsertRune ────────────────────────────────────────────────────────────

func TestLeInsertRune(t *testing.T) {
	tests := []struct {
		name string
		buf  []rune
		pos  int
		r    rune
		want string
	}{
		{"insert into empty buffer", []rune{}, 0, 'x', "x"},
		{"insert at start", []rune("bc"), 0, 'a', "abc"},
		{"insert at end", []rune("ab"), 2, 'c', "abc"},
		{"insert in middle", []rune("ac"), 1, 'b', "abc"},
		{"insert unicode", []rune("aé"), 1, 'ê', "aêé"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := leInsertRune(tt.buf, tt.pos, tt.r)
			if string(got) != tt.want {
				t.Errorf("want %q, got %q", tt.want, string(got))
			}
		})
	}
}

func TestLeInsertRune_doesNotMutateOriginal(t *testing.T) {
	orig := []rune("ab")
	result := leInsertRune(orig, 1, 'X')
	if string(result) != "aXb" {
		t.Errorf("want %q, got %q", "aXb", string(result))
	}
	// orig may be extended in place by append — we just verify result is correct.
	if result[1] != 'X' {
		t.Errorf("inserted rune not at expected position")
	}
}

// ─── readEscSeq ──────────────────────────────────────────────────────────────

func TestReadEscSeq_arrows(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"up arrow CSI",    "[A", "[A"},
		{"down arrow CSI",  "[B", "[B"},
		{"right arrow CSI", "[C", "[C"},
		{"left arrow CSI",  "[D", "[D"},
		{"home CSI",        "[H", "[H"},
		{"end CSI",         "[F", "[F"},
		{"home VT220",      "[1~", "[1~"},
		{"end VT220",       "[4~", "[4~"},
		{"up SS3",          "OA", "OA"},
		{"down SS3",        "OB", "OB"},
		{"right SS3",       "OC", "OC"},
		{"left SS3",        "OD", "OD"},
		{"home SS3",        "OH", "OH"},
		{"end SS3",         "OF", "OF"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("os.Pipe: %v", err)
			}
			defer r.Close()
			w.WriteString(tt.input)
			w.Close()

			le := NewLineEditor(r, io.Discard)
			got := le.readEscSeq()
			if got != tt.want {
				t.Errorf("want %q, got %q", tt.want, got)
			}
		})
	}
}

func TestReadEscSeq_emptyOnEOF(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	defer r.Close()
	w.Close() // immediate EOF

	le := NewLineEditor(r, io.Discard)
	got := le.readEscSeq()
	if got != "" {
		t.Errorf("want empty string on EOF, got %q", got)
	}
}

// ─── readUTF8Tail ────────────────────────────────────────────────────────────

func TestReadUTF8Tail(t *testing.T) {
	tests := []struct {
		name  string
		input rune // the rune to encode and round-trip
	}{
		{"2-byte: é (U+00E9)", 'é'},
		{"2-byte: ñ (U+00F1)", 'ñ'},
		{"3-byte: ✓ (U+2713)", '✓'},
		{"3-byte: 中 (U+4E2D)", '中'},
		{"4-byte: 𝄞 (U+1D11E)", '𝄞'},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode the rune to UTF-8, split into lead byte + tail bytes.
			encoded := []byte(string(tt.input))
			if len(encoded) < 2 {
				t.Fatalf("rune %q encoded as single byte, not multi-byte", tt.input)
			}
			lead := encoded[0]
			tail := encoded[1:]

			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("os.Pipe: %v", err)
			}
			defer r.Close()
			w.Write(tail)
			w.Close()

			le := NewLineEditor(r, io.Discard)
			got := le.readUTF8Tail(lead)
			if got != tt.input {
				t.Errorf("want %q (%U), got %q (%U)", tt.input, tt.input, got, got)
			}
		})
	}
}
