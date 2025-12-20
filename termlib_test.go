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

import (
	"bytes"
	"os"
	"testing"
)

func TestTermLib(t *testing.T) {
	// Redirect stdout for testing
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	term := New()

	// Test Move
	term.Move(5, 10)
	row, col := term.GetCurPos()
	if row != 5 || col != 10 {
		t.Errorf("Move(5, 10) failed, got (%d, %d)", row, col)
	}

	// Test Clear
	term.Clear()
	row, col = term.GetCurPos()
	if row != 1 || col != 1 {
		t.Errorf("Clear() failed, got (%d, %d)", row, col)
	}

	// Test Print
	term.Print("Hello")
	row, col = term.GetCurPos()
	if row != 1 || col != 6 {
		t.Errorf("Print(\"Hello\") failed, got (%d, %d)", row, col)
	}

	// Test ClrToEOL
	term.ClrToEOL()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	got := buf.String()
	expected := "\033[5;10H\033[2J\033[HHello\033[0K"
	if got != expected {
		t.Errorf("Output mismatch: got %q, want %q", got, expected)
	}
}

func TestClrToBOL(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	term := New()
	term.Move(3, 5)
	term.ClrToBOL()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	got := buf.String()
	expected := "\033[3;5H\033[1K"
	if got != expected {
		t.Errorf("ClrToBOL() = %q, want %q", got, expected)
	}
}
