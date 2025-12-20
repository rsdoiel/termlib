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
	"fmt"
	"os"
	"sync"

	"golang.org/x/term"
)

// TermLib represents a terminal controller.
type TermLib struct {
	mu             sync.Mutex
	cursorRow      int
	cursorCol      int
	terminalWidth  int
	terminalHeight int
}

// New creates a new TermLib instance.
func New() *TermLib {
	t := &TermLib{
		cursorRow:      1,
		cursorCol:      1,
		terminalWidth:  80, // Default width
		terminalHeight: 24, // Default height
	}
	t.updateTerminalSize()
	return t
}

// updateTerminalSize updates the terminal width and height.
func (t *TermLib) updateTerminalSize() {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err == nil {
		t.terminalWidth = width
		t.terminalHeight = height
	}
}

// GetTerminalWidth returns the terminal width.
func (t *TermLib) GetTerminalWidth() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.terminalWidth
}

// GetTerminalHeight returns the terminal height.
func (t *TermLib) GetTerminalHeight() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.terminalHeight
}

// Move moves the cursor to the specified row and column (1-based).
func (t *TermLib) Move(row, col int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cursorRow, t.cursorCol = row, col
	fmt.Printf("\033[%d;%dH", row, col)
}

// Clear clears the screen and moves the cursor to the top-left corner.
func (t *TermLib) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cursorRow, t.cursorCol = 1, 1
	fmt.Print("\033[2J\033[H")
}

// ClrToEOL clears from the current cursor position to the end of the line.
func (t *TermLib) ClrToEOL() {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Print("\033[0K")
}

// ClrToBOL clears from the current cursor position to the start of the line.
func (t *TermLib) ClrToBOL() {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Print("\033[1K")
}

// Print writes a string to the terminal and updates the cursor position.
func (t *TermLib) Print(s string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Print(s)
	// Update cursor position based on the string written
	for _, c := range s {
		if c == '\n' {
			t.cursorRow++
			t.cursorCol = 1
		} else {
			t.cursorCol++
		}
	}
}

// GetCurPos returns the current cursor position.
func (t *TermLib) GetCurPos() (int, int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.cursorRow, t.cursorCol
}

// Refresh ensures all buffered output is written to the terminal.
func (t *TermLib) Refresh() {
	fmt.Print("")
}
