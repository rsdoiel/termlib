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


func TestClrToBOL(t *testing.T) {
	var buf bytes.Buffer
	term := New(&buf)

	term.Move(3, 5)
	term.ClrToBOL()
	expected := "\033[3;5H\033[1K"
	got := buf.String()
	if got != expected {
		t.Errorf("ClrToBOL() = %q, want %q", got, expected)
	}
}

func TestStyleAndColor(t *testing.T) {
	var buf bytes.Buffer
	term := New(&buf)

	term.SetFgColor(Red)
	term.SetBgColor(YellowBg)
	term.SetBold()
	term.SetItalic()
	term.Print("Styled")

	expected := "\033[31m\033[43m\033[1m\033[3mStyled\033[0m"
	got := buf.String()
	if got != expected {
		t.Errorf("Style and color output mismatch: got %q, want %q", got, expected)
	}
}


func TestNewWithStdout(t *testing.T) {
	term := New(os.Stdout)
	// Just verify that it can be created with os.Stdout
	if term == nil {
		t.Fatal("New(os.Stdout) returned nil")
	}
}


func TestTermLib(t *testing.T) {
	var buf bytes.Buffer
	term := New(&buf)

	term.Move(5, 10)
	term.Clear()
	term.Print("Hello")
	term.ClrToEOL()
	expected := "\033[5;10H\033[2J\033[HHello\033[0K"
	got := buf.String()
	if got != expected {
		t.Errorf("ClrToEOL() = %q, want %q", got, expected)
	}
}

func TestResetStyle(t *testing.T) {
	var buf bytes.Buffer
	term := New(&buf)

	term.SetFgColor(Red)
	term.SetBgColor(YellowBg)
	term.SetBold()
	term.SetItalic()
	term.ResetStyle()
	term.Print("Normal")

	expected := "\033[0mNormal"
	got := buf.String()
	if got != expected {
		t.Errorf("ResetStyle() output mismatch: got %q, want %q", got, expected)
	}
}

func TestMultiplePrints(t *testing.T) {
	var buf bytes.Buffer
	term := New(&buf)

	term.SetFgColor(Blue)
	term.Print("First ")
	term.SetFgColor(Green)
	term.Print("Second")

	expected := "\033[34mFirst \033[0m\033[32mSecond\033[0m"
	got := buf.String()
	if got != expected {
		t.Errorf("Multiple prints output mismatch: got %q, want %q", got, expected)
	}
}

