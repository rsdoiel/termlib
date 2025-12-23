// termlib is a light weight terminal interface library.
// Copyright (C) 2025 R. S. Doiel
package termlib

import (
	"fmt"
	"io"
	"os"
	"sync"

	"golang.org/x/term"
)

// ANSI escape codes for colors and styles
const (
	Reset      = "\033[0m"
	Bold       = "\033[1m"
	Italic     = "\033[3m"
	Black      = "\033[30m"
	Red        = "\033[31m"
	Green      = "\033[32m"
	Yellow     = "\033[33m"
	Blue       = "\033[34m"
	Magenta    = "\033[35m"
	Cyan       = "\033[36m"
	White      = "\033[37m"
	BlackBg    = "\033[40m"
	RedBg      = "\033[41m"
	GreenBg    = "\033[42m"
	YellowBg   = "\033[43m"
	BlueBg     = "\033[44m"
	MagentaBg  = "\033[45m"
	CyanBg     = "\033[46m"
	WhiteBg    = "\033[47m"
)

// Terminal represents a terminal controller.
type Terminal struct {
	mu             sync.Mutex
	cursorRow      int
	cursorCol      int
	terminalWidth  int
	terminalHeight int
	fgColor        string
	bgColor        string
	isBold         bool
	isItalic       bool
	writer         io.Writer
	styleApplied   bool // Track if any style is applied
}

// New creates a new Terminal instance with the specified writer and default styles.
func New(writer io.Writer) *Terminal {
	return &Terminal{
		writer:        writer,
		cursorRow:     1,
		cursorCol:     1,
		terminalWidth: 80,
		terminalHeight: 24,
		fgColor:       Reset,
		bgColor:       Reset,
	}
}

// updateTerminalSize updates the terminal width and height.
func (t *Terminal) updateTerminalSize() {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err == nil {
		t.terminalWidth = width
		t.terminalHeight = height
	}
}

// GetTerminalWidth returns the terminal width.
func (t *Terminal) GetTerminalWidth() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.terminalWidth
}

// GetTerminalHeight returns the terminal height.
func (t *Terminal) GetTerminalHeight() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.terminalHeight
}

// Move moves the cursor to the specified row and column (1-based).
func (t *Terminal) Move(row, col int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cursorRow, t.cursorCol = row, col
	fmt.Fprintf(t.writer, "\033[%d;%dH", row, col)
}

// Clear clears the screen and moves the cursor to the top-left corner.
func (t *Terminal) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cursorRow, t.cursorCol = 1, 1
	fmt.Fprint(t.writer, "\033[2J\033[H")
}

// ClrToEOL clears from the current cursor position to the end of the line.
func (t *Terminal) ClrToEOL() {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Fprint(t.writer, "\033[0K")
}

// ClrToBOL clears from the current cursor position to the start of the line.
func (t *Terminal) ClrToBOL() {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Fprint(t.writer, "\033[1K")
}

// Print writes a string to the terminal with current style and updates the cursor position.
func (t *Terminal) Print(s string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.styleApplied {
		t.applyStyle()
		fmt.Fprint(t.writer, s)
		fmt.Fprint(t.writer, Reset)
		t.resetStyleState()
	} else {
		fmt.Fprint(t.writer, s)
	}
	for _, c := range s {
		if c == '\n' {
			t.cursorRow++
			t.cursorCol = 1
		} else {
			t.cursorCol++
		}
	}
}

// Printf writes a format string to the terminal with the current style and updates
// the cursor position.
func (t *Terminal) Printf(format string, a ...interface{}) {
	if len(a) > 0 {
		t.Print(fmt.Sprintf(format, a...))
	} else {
		t.Print(format)
	}
}

// Println writes parameters as a string with a trailing new line
func (t *Terminal) Println(a ...interface{}) {
	t.Print(fmt.Sprintln(a...))
}

// GetCurPos returns the current cursor position.
func (t *Terminal) GetCurPos() (int, int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.cursorRow, t.cursorCol
}

// Refresh ensures all buffered output is written to the terminal.
func (t *Terminal) Refresh() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if w, ok := t.writer.(*os.File); ok {
		w.Sync()
	}
}

// GetFgColor retrives the foreground color code.
func (t *Terminal) GetFgColor() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.fgColor
}

// GetBgColor retrieves the background color code.
func (t *Terminal) GetBgColor() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.bgColor
}

// SetFgColor sets the foreground color.
func (t *Terminal) SetFgColor(color string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.fgColor = color
	t.styleApplied = true
}

// SetBgColor sets the background color.
func (t *Terminal) SetBgColor(color string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.bgColor = color
	t.styleApplied = true
}

// SetBold enables bold text.
func (t *Terminal) SetBold() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.isBold = true
	t.styleApplied = true
}

// SetItalic enables italic text.
func (t *Terminal) SetItalic() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.isItalic = true
	t.styleApplied = true
}

// ResetStyle resets all styles to default.
func (t *Terminal) ResetStyle() {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Fprint(t.writer, Reset)
	t.resetStyleState()
}

// applyStyle applies the current style settings.
func (t *Terminal) applyStyle() {
	if t.fgColor != Reset {
		fmt.Fprint(t.writer, t.fgColor)
	}
	if t.bgColor != Reset {
		fmt.Fprint(t.writer, t.bgColor)
	}
	if t.isBold {
		fmt.Fprint(t.writer, Bold)
	}
	if t.isItalic {
		fmt.Fprint(t.writer, Italic)
	}
}

// resetStyleState resets the internal style state.
func (t *Terminal) resetStyleState() {
	t.fgColor = Reset
	t.bgColor = Reset
	t.isBold = false
	t.isItalic = false
	t.styleApplied = false
}

