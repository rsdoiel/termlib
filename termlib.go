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
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Package termlib is a light weight terminal interface library, sort of
// an ncurses light. It explores the space between the fmt package and a rich
// library like tcell.
package termlib

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	"golang.org/x/term"
)

/** ANSI escape codes for styles and colors. Reset, Bold and Italic are styles;
 * Black through White set the foreground and BlackBg through WhiteBg the
 * background. Pass a color to Terminal.SetFgColor or Terminal.SetBgColor, or
 * write one to a terminal directly.
 *
 * Example:
 *   term.SetFgColor(termlib.Red)
 *   term.Print("error") // prints in red, then the style resets
 */
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Italic    = "\033[3m"
	Black     = "\033[30m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Magenta   = "\033[35m"
	Cyan      = "\033[36m"
	White     = "\033[37m"
	BlackBg   = "\033[40m"
	RedBg     = "\033[41m"
	GreenBg   = "\033[42m"
	YellowBg  = "\033[43m"
	BlueBg    = "\033[44m"
	MagentaBg = "\033[45m"
	CyanBg    = "\033[46m"
	WhiteBg   = "\033[47m"
)

// terminalSize reads the size of standard output. It is a variable so that a
// test can supply sizes where standard output is not a terminal.
var terminalSize = func() (width, height int, err error) {
	return term.GetSize(int(os.Stdout.Fd()))
}

/** Terminal is a buffered terminal controller. Every method adds to an internal
 * buffer; nothing reaches the underlying writer until Refresh, so a frame is
 * written in one piece and does not flicker. It tracks the cursor position it
 * has been told about (Move, Print, Clear), not the real one. Its methods are
 * safe to call from several goroutines.
 *
 * Styles are one-shot: SetFgColor, SetBgColor, SetBold and SetItalic apply to
 * the next Print (or Printf, Println) only, and the style is reset after it.
 *
 * Example:
 *   term := termlib.New(os.Stdout)
 *   term.Clear()
 *   term.Move(2, 4)
 *   term.Print("hello")
 *   term.Refresh()
 */
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
	out            io.Writer // underlying destination (e.g. os.Stdout)
	buf            bytes.Buffer
	styleApplied   bool
}

/** New returns a Terminal that writes to writer when Refresh is called. The
 * cursor starts at row 1, column 1 and the size is read from standard output,
 * falling back to 80 columns by 24 rows when it is not a terminal.
 *
 * Parameters:
 *   writer (io.Writer) — the destination, typically os.Stdout
 *
 * Returns:
 *   *Terminal — a terminal with default styles and an empty buffer
 *
 * Example:
 *   term := termlib.New(os.Stdout)
 */
func New(writer io.Writer) *Terminal {
	t := &Terminal{
		out:            writer,
		cursorRow:      1,
		cursorCol:      1,
		terminalWidth:  80,
		terminalHeight: 24,
		fgColor:        Reset,
		bgColor:        Reset,
	}
	t.UpdateTerminalSize()
	return t
}

/** UpdateTerminalSize re-reads the size of standard output and stores it. Call it
 * after a window resize. When standard output is not a terminal the stored
 * size is left as it was.
 *
 * Example:
 *   term.UpdateTerminalSize()
 *   w, h := term.GetTerminalWidth(), term.GetTerminalHeight()
 */
func (t *Terminal) UpdateTerminalSize() {
	width, height, err := terminalSize()
	if err == nil {
		t.mu.Lock()
		defer t.mu.Unlock()
		t.terminalWidth = width
		t.terminalHeight = height
	}
}

/** GetTerminalWidth returns the stored width in columns (80 by default).
 *
 * Returns:
 *   int — columns, as of New or the last UpdateTerminalSize
 *
 * Example:
 *   w := term.GetTerminalWidth()
 */
func (t *Terminal) GetTerminalWidth() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.terminalWidth
}

/** GetTerminalHeight returns the stored height in rows (24 by default).
 *
 * Returns:
 *   int — rows, as of New or the last UpdateTerminalSize
 *
 * Example:
 *   h := term.GetTerminalHeight()
 */
func (t *Terminal) GetTerminalHeight() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.terminalHeight
}

/** Move buffers a cursor move to the given position and records it as the
 * current position. Nothing is clamped to the terminal size.
 *
 * Parameters:
 *   row (int) — the row, 1 is the top
 *   col (int) — the column, 1 is the left edge
 *
 * Example:
 *   term.Move(term.GetTerminalHeight(), 1) // bottom left
 */
func (t *Terminal) Move(row, col int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cursorRow, t.cursorCol = row, col
	fmt.Fprintf(&t.buf, "\033[%d;%dH", row, col)
}

/** Clear buffers a clear-screen and a move to row 1, column 1, and records
 * that position.
 *
 * Example:
 *   term.Clear()
 *   term.Refresh()
 */
func (t *Terminal) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cursorRow, t.cursorCol = 1, 1
	fmt.Fprint(&t.buf, "\033[2J\033[H")
}

/** ClrToEOL buffers a clear from the cursor to the end of the line. The cursor
 * does not move.
 *
 * Example:
 *   term.Move(5, 1)
 *   term.ClrToEOL()
 */
func (t *Terminal) ClrToEOL() {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Fprint(&t.buf, "\033[0K")
}

/** ClrToBOL buffers a clear from the start of the line to the cursor. The
 * cursor does not move.
 *
 * Example:
 *   term.ClrToBOL()
 */
func (t *Terminal) ClrToBOL() {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Fprint(&t.buf, "\033[1K")
}

/** HideCursor buffers the escape sequence that hides the cursor during a redraw.
 * Pair it with ShowCursor, or the cursor stays hidden after the program exits.
 *
 * Example:
 *   term.HideCursor()
 *   // ... draw ...
 *   term.ShowCursor()
 *   term.Refresh()
 */
func (t *Terminal) HideCursor() {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Fprint(&t.buf, "\033[?25l")
}

/** ShowCursor buffers the escape sequence that shows the cursor again.
 *
 * Example:
 *   term.ShowCursor()
 */
func (t *Terminal) ShowCursor() {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Fprint(&t.buf, "\033[?25h")
}

/** Print buffers s, wrapped in the current style when one is set, and then
 * resets the style. The recorded cursor moves one column for each rune and to
 * column 1 of the next row for each newline; wide runes and wrapping at the
 * right edge are not accounted for.
 *
 * Parameters:
 *   s (string) — the text to write
 *
 * Example:
 *   term.SetBold()
 *   term.Print("Title") // bold; the next Print is plain
 */
func (t *Terminal) Print(s string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.styleApplied {
		t.applyStyle()
		fmt.Fprint(&t.buf, s)
		fmt.Fprint(&t.buf, Reset)
		t.resetStyleState()
	} else {
		fmt.Fprint(&t.buf, s)
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

/** Printf formats as fmt.Printf does (a doubled percent sign prints one, with or
 * without arguments) and writes the result with Print.
 *
 * Parameters:
 *   format (string) — a fmt format string
 *   a (...interface{}) — the values to format
 *
 * Example:
 *   term.Printf("%d%% done", 40)
 */
func (t *Terminal) Printf(format string, a ...interface{}) {
	t.Print(fmt.Sprintf(format, a...))
}

/** Println formats as fmt.Sprintln does, spaces between operands and a trailing
 * newline, and writes the result with Print.
 *
 * Parameters:
 *   a (...interface{}) — the values to write
 *
 * Example:
 *   term.Println("rows:", 24)
 */
func (t *Terminal) Println(a ...interface{}) {
	t.Print(fmt.Sprintln(a...))
}

/** GetCurPos returns the cursor position Terminal has recorded from Move, Clear
 * and Print. It is not read back from the terminal.
 *
 * Returns:
 *   int — the row, 1-based
 *   int — the column, 1-based
 *
 * Example:
 *   row, col := term.GetCurPos()
 */
func (t *Terminal) GetCurPos() (int, int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.cursorRow, t.cursorCol
}

/** Refresh writes everything buffered to the underlying writer in one call and
 * empties the buffer. It writes nothing when the buffer is empty. Write errors
 * are ignored.
 *
 * Example:
 *   term.Print("frame")
 *   term.Refresh()
 */
func (t *Terminal) Refresh() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.buf.Len() == 0 {
		return
	}
	t.out.Write(t.buf.Bytes()) //nolint:errcheck
	t.buf.Reset()
}

/** GetFgColor returns the pending foreground color, or Reset when none is set.
 * It is Reset again after the next Print.
 *
 * Returns:
 *   string — an ANSI color constant such as Red, or Reset
 *
 * Example:
 *   term.SetFgColor(termlib.Red)
 *   c := term.GetFgColor() // termlib.Red
 */
func (t *Terminal) GetFgColor() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.fgColor
}

/** GetBgColor returns the pending background color, or Reset when none is set.
 * It is Reset again after the next Print.
 *
 * Returns:
 *   string — an ANSI color constant such as RedBg, or Reset
 *
 * Example:
 *   c := term.GetBgColor()
 */
func (t *Terminal) GetBgColor() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.bgColor
}

/** SetFgColor sets the foreground color for the next Print only.
 *
 * Parameters:
 *   color (string) — an ANSI color constant such as Red
 *
 * Example:
 *   term.SetFgColor(termlib.Green)
 *   term.Print("ok")
 */
func (t *Terminal) SetFgColor(color string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.fgColor = color
	t.styleApplied = true
}

/** SetBgColor sets the background color for the next Print only.
 *
 * Parameters:
 *   color (string) — an ANSI color constant such as BlueBg
 *
 * Example:
 *   term.SetBgColor(termlib.BlueBg)
 *   term.Print(" selected ")
 */
func (t *Terminal) SetBgColor(color string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.bgColor = color
	t.styleApplied = true
}

/** SetBold makes the next Print bold.
 *
 * Example:
 *   term.SetBold()
 *   term.Print("Title")
 */
func (t *Terminal) SetBold() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.isBold = true
	t.styleApplied = true
}

/** SetItalic makes the next Print italic.
 *
 * Example:
 *   term.SetItalic()
 *   term.Print("note")
 */
func (t *Terminal) SetItalic() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.isItalic = true
	t.styleApplied = true
}

/** ResetStyle buffers a reset sequence and clears any pending color, bold and
 * italic. Use it to cancel a style set but not yet printed.
 *
 * Example:
 *   term.SetBold()
 *   term.ResetStyle() // the next Print is plain
 */
func (t *Terminal) ResetStyle() {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Fprint(&t.buf, Reset)
	t.resetStyleState()
}

// applyStyle applies the current style settings.
func (t *Terminal) applyStyle() {
	if t.fgColor != Reset {
		fmt.Fprint(&t.buf, t.fgColor)
	}
	if t.bgColor != Reset {
		fmt.Fprint(&t.buf, t.bgColor)
	}
	if t.isBold {
		fmt.Fprint(&t.buf, Bold)
	}
	if t.isItalic {
		fmt.Fprint(&t.buf, Italic)
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
