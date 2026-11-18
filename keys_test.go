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

package termlib

import (
	"os"
	"testing"
)

// feed returns a read end holding data, already closed at the write end so
// that reads past the data see EOF.
func feed(t *testing.T, data string) *os.File {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.WriteString(data); err != nil {
		t.Fatal(err)
	}
	w.Close()
	t.Cleanup(func() { r.Close() })
	return r
}

/** TestReadKey checks each keystroke class ReadKey decodes: ASCII, control
 * characters, multi-byte UTF-8, the CSI and SS3 arrow and navigation
 * sequences, and the escape sequences that are consumed but not recognised.
 */
func TestReadKey(t *testing.T) {
	for _, tc := range []struct {
		name, in string
		want     Key
	}{
		{"letter", "q", Key('q')},
		{"space", " ", Key(' ')},
		{"ctrl-c", "\x03", Key(0x03)},
		{"enter", "\r", Key('\r')},
		{"two byte rune", "é", Key('é')},
		{"three byte rune", "日", Key('日')},
		{"four byte rune", "😀", Key('😀')},
		{"up", "\x1b[A", KeyUp},
		{"down", "\x1b[B", KeyDown},
		{"right", "\x1b[C", KeyRight},
		{"left", "\x1b[D", KeyLeft},
		{"ss3 up", "\x1bOA", KeyUp},
		{"ss3 left", "\x1bOD", KeyLeft},
		{"home csi", "\x1b[H", KeyHome},
		{"home tilde", "\x1b[1~", KeyHome},
		{"end csi", "\x1b[F", KeyEnd},
		{"end tilde", "\x1b[4~", KeyEnd},
		{"page up", "\x1b[5~", KeyPageUp},
		{"page down", "\x1b[6~", KeyPageDown},
		{"delete is unrecognised", "\x1b[3~", Key(0x1b)},
		{"modified arrow is unrecognised", "\x1b[1;5C", Key(0x1b)},
		{"invalid utf-8 tail", "\xc3(", KeyUnknown},
	} {
		got, err := ReadKey(feed(t, tc.in))
		if err != nil || got != tc.want {
			t.Errorf("%s: ReadKey(%q) = %d, %v; want %d", tc.name, tc.in, got, err, tc.want)
		}
	}
}

/** TestReadKeyConsumesWholeSequence checks that an escape sequence leaves
 * nothing behind: the key after it is read intact.
 */
func TestReadKeyConsumesWholeSequence(t *testing.T) {
	in := feed(t, "\x1b[3~x\x1b[1;5Cy")
	for i, want := range []Key{Key(0x1b), Key('x'), Key(0x1b), Key('y')} {
		got, err := ReadKey(in)
		if err != nil || got != want {
			t.Fatalf("key %d = %d, %v; want %d", i, got, err, want)
		}
	}
}

/** TestReadKeyEOF checks that an empty input returns KeyUnknown with an error. */
func TestReadKeyEOF(t *testing.T) {
	k, err := ReadKey(feed(t, ""))
	if err == nil || k != KeyUnknown {
		t.Errorf("ReadKey on empty input = %d, %v; want KeyUnknown and an error", k, err)
	}
}

/** TestKeyReader checks that KeyReader delivers keys in order and closes the
 * channel when the input ends.
 */
func TestKeyReader(t *testing.T) {
	var got []Key
	for k := range KeyReader(feed(t, "a\x1b[Ab")) {
		got = append(got, k)
	}
	want := []Key{Key('a'), KeyUp, Key('b')}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("key %d = %d, want %d", i, got[i], want[i])
		}
	}
}

/** TestEnterRawModeOnAPipe checks that a file that is not a terminal gives an
 * error and a restore function that is safe to call.
 */
func TestEnterRawModeOnAPipe(t *testing.T) {
	restore, err := EnterRawMode(feed(t, ""))
	if err == nil {
		t.Error("want an error for a pipe")
	}
	if restore == nil {
		t.Fatal("restore is nil, want a no-op")
	}
	restore()
}
