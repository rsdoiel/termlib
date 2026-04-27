// widgets.go — reusable TUI drawing helpers built on Terminal.
// Copyright (C) 2025 R. S. Doiel
package termlib

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// Box-drawing and bar characters.
const (
	boxTopLeft     = "┌"
	boxTopRight    = "┐"
	boxBottomLeft  = "└"
	boxBottomRight = "┘"
	boxHoriz       = "─"
	boxVert        = "│"
	barFull        = "█"
	barEmpty       = "░"
)

/** DrawBox draws a Unicode box at the given terminal position. The box
 * occupies width columns and height rows. title is embedded in the top
 * border; pass an empty string for a plain border.
 *
 * Parameters:
 *   t      (*Terminal) — terminal to draw on.
 *   row    (int)       — top row of the box (1-based).
 *   col    (int)       — left column of the box (1-based).
 *   width  (int)       — total width including borders.
 *   height (int)       — total height including borders.
 *   title  (string)    — optional label in the top border.
 *
 * Example:
 *   termlib.DrawBox(term, 1, 1, 40, 10, "Now Playing")
 */
func DrawBox(t *Terminal, row, col, width, height int, title string) {
	// Top border
	t.Move(row, col)
	if title != "" {
		label := "─ " + title + " "
		labelW := 2 + utf8.RuneCountInString(title) + 1
		remaining := width - 2 - labelW
		if remaining < 0 {
			remaining = 0
			label = Truncate(label, width-2)
		}
		t.Print(boxTopLeft + label + strings.Repeat(boxHoriz, remaining) + boxTopRight)
	} else {
		t.Print(boxTopLeft + strings.Repeat(boxHoriz, width-2) + boxTopRight)
	}

	// Side borders (interior rows only)
	for r := row + 1; r < row+height-1; r++ {
		t.Move(r, col)
		t.Print(boxVert)
		t.Move(r, col+width-1)
		t.Print(boxVert)
	}

	// Bottom border
	t.Move(row+height-1, col)
	t.Print(boxBottomLeft + strings.Repeat(boxHoriz, width-2) + boxBottomRight)
}

/** DrawProgressBar draws a horizontal progress bar at the given position.
 * The bar renders as [████░░░░] where filled cells represent value/total.
 * width is the total width of the bar including the surrounding brackets.
 *
 * Parameters:
 *   t      (*Terminal) — terminal to draw on.
 *   row    (int)       — row (1-based).
 *   col    (int)       — starting column (1-based).
 *   width  (int)       — total width including brackets.
 *   value  (float64)   — current value.
 *   total  (float64)   — maximum value; bar is empty when total <= 0.
 *
 * Example:
 *   termlib.DrawProgressBar(term, 5, 3, 30, elapsed.Seconds(), total.Seconds())
 */
func DrawProgressBar(t *Terminal, row, col, width int, value, total float64) {
	inner := width - 2
	if inner < 1 {
		inner = 1
	}
	filled := 0
	if total > 0 {
		filled = int(float64(inner) * value / total)
		if filled > inner {
			filled = inner
		}
	}
	bar := "[" + strings.Repeat(barFull, filled) + strings.Repeat(barEmpty, inner-filled) + "]"
	t.Move(row, col)
	t.Print(bar)
}

/** Truncate shortens s to at most maxW Unicode code points. If truncation
 * occurs, the last code point is replaced with "…". Returns s unchanged
 * when it fits within maxW.
 *
 * Parameters:
 *   s    (string) — the string to truncate.
 *   maxW (int)    — maximum number of Unicode code points to return.
 *
 * Returns:
 *   string — s, possibly truncated with a trailing "…".
 *
 * Example:
 *   label := termlib.Truncate("Goldberg Variations", 12) // "Goldberg Va…"
 */
func Truncate(s string, maxW int) string {
	runes := []rune(s)
	if len(runes) <= maxW {
		return s
	}
	if maxW <= 1 {
		return "…"
	}
	return string(runes[:maxW-1]) + "…"
}

/** PadRight pads s with trailing spaces to exactly w Unicode code points.
 * If s is already longer than w it is truncated with Truncate.
 *
 * Parameters:
 *   s (string) — the string to pad or truncate.
 *   w (int)    — desired display width in Unicode code points.
 *
 * Returns:
 *   string — s padded or truncated to exactly w code points.
 *
 * Example:
 *   cell := termlib.PadRight(trackName, columnWidth)
 */
func PadRight(s string, w int) string {
	runes := []rune(s)
	if len(runes) >= w {
		return Truncate(s, w)
	}
	return s + strings.Repeat(" ", w-len(runes))
}

/** FormatDuration formats a time.Duration as "m:ss" or "h:mm:ss" for
 * durations of one hour or more. The result is always rounded to the
 * nearest second.
 *
 * Parameters:
 *   d (time.Duration) — the duration to format.
 *
 * Returns:
 *   string — formatted duration, e.g. "3:07" or "1:02:05".
 *
 * Example:
 *   label := termlib.FormatDuration(elapsed) // "2:34"
 */
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}
