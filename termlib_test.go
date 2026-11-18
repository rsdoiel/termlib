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

package termlib

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestClrToBOL(t *testing.T) {
	var buf bytes.Buffer
	term := New(&buf)

	term.Move(3, 5)
	term.ClrToBOL()
	term.Refresh()
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
	term.Refresh()

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

func TestTerminal(t *testing.T) {
	var buf bytes.Buffer
	term := New(&buf)

	term.Move(5, 10)
	term.Clear()
	term.Print("Hello")
	term.ClrToEOL()
	term.Refresh()
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
	term.Refresh()

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
	term.Refresh()

	expected := "\033[34mFirst \033[0m\033[32mSecond\033[0m"
	got := buf.String()
	if got != expected {
		t.Errorf("Multiple prints output mismatch: got %q, want %q", got, expected)
	}
}

func TestPrintf(t *testing.T) {
	var buf bytes.Buffer
	format := "Hello %q %d,\n"
	expected := fmt.Sprintf(format, "Robert", 63)
	term := New(&buf)
	term.Printf(format, "Robert", 63)
	term.Refresh()
	got := buf.String()
	if got != expected {
		t.Errorf("Printf output mismatch: got %q, want %q", got, expected)
	}
}

func TestPrintln(t *testing.T) {
	const name, age = "Kim", 22
	var buf bytes.Buffer
	expected := fmt.Sprintln(name, "is", age, "years old.")
	term := New(&buf)
	term.Println(name, "is", age, "years old.")
	term.Refresh()
	got := buf.String()
	if got != expected {
		t.Errorf("Println output mismatch: got %q, want %q", got, expected)
	}
}

/** TestCursorPositionTracking checks that Move, Print, Clear and a newline in
 * printed text keep the reported cursor position in step with what was written.
 */
func TestCursorPositionTracking(t *testing.T) {
	var out bytes.Buffer
	term := New(&out)
	if r, c := term.GetCurPos(); r != 1 || c != 1 {
		t.Errorf("new terminal at %d,%d, want 1,1", r, c)
	}
	term.Move(5, 10)
	if r, c := term.GetCurPos(); r != 5 || c != 10 {
		t.Errorf("after Move(5,10) at %d,%d", r, c)
	}
	term.Print("abc")
	if r, c := term.GetCurPos(); r != 5 || c != 13 {
		t.Errorf("after Print(abc) at %d,%d, want 5,13", r, c)
	}
	term.Print("é日\nxy")
	if r, c := term.GetCurPos(); r != 6 || c != 3 {
		t.Errorf("after a newline at %d,%d, want 6,3 (runes count as one column)", r, c)
	}
	term.Clear()
	if r, c := term.GetCurPos(); r != 1 || c != 1 {
		t.Errorf("after Clear at %d,%d, want 1,1", r, c)
	}
}

/** TestRefreshWritesOnceAndDrains checks that output is held until Refresh,
 * written in one piece, and not written a second time.
 */
func TestRefreshWritesOnceAndDrains(t *testing.T) {
	var out bytes.Buffer
	term := New(&out)
	term.Refresh()
	if out.Len() != 0 {
		t.Fatalf("Refresh of an empty buffer wrote %q", out.String())
	}
	term.Move(2, 3)
	term.Print("hi")
	if out.Len() != 0 {
		t.Fatalf("output reached the writer before Refresh: %q", out.String())
	}
	term.Refresh()
	if got := out.String(); got != "\033[2;3Hhi" {
		t.Errorf("Refresh wrote %q, want %q", got, "\033[2;3Hhi")
	}
	term.Refresh()
	if got := out.String(); got != "\033[2;3Hhi" {
		t.Errorf("a second Refresh wrote again: %q", got)
	}
}

/** TestHideAndShowCursor checks the escape sequences for cursor visibility. */
func TestHideAndShowCursor(t *testing.T) {
	var out bytes.Buffer
	term := New(&out)
	term.HideCursor()
	term.ShowCursor()
	term.Refresh()
	if got := out.String(); got != "\033[?25l\033[?25h" {
		t.Errorf("got %q", got)
	}
}

/** TestColorGettersAndSize checks the defaults and the setters' effect on the
 * getters, and that the size is usable whether or not stdout is a terminal.
 */
func TestColorGettersAndSize(t *testing.T) {
	term := New(&bytes.Buffer{})
	if term.GetFgColor() != Reset || term.GetBgColor() != Reset {
		t.Errorf("defaults %q, %q, want Reset for both", term.GetFgColor(), term.GetBgColor())
	}
	term.SetFgColor(Red)
	term.SetBgColor(BlueBg)
	if term.GetFgColor() != Red || term.GetBgColor() != BlueBg {
		t.Errorf("got %q, %q", term.GetFgColor(), term.GetBgColor())
	}
	if term.GetTerminalWidth() <= 0 || term.GetTerminalHeight() <= 0 {
		t.Errorf("size %dx%d, want positive", term.GetTerminalWidth(), term.GetTerminalHeight())
	}
}

/** TestFmtHelp checks that every placeholder is replaced, all occurrences of
 * it, and that unknown braces are left alone.
 */
func TestFmtHelp(t *testing.T) {
	src := "{app_name} {version} {release_date} {release_hash} {app_name} {other}"
	got := FmtHelp(src, "tl", "1.2.3", "2026-11-18", "abc1234")
	want := "tl 1.2.3 2026-11-18 abc1234 tl {other}"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

/** TestPrintfFormatsLikeFmt checks that Printf is fmt.Printf: a doubled percent
 * sign prints one, with or without arguments. Printf used to skip formatting when
 * it had no arguments, so "100%% done" printed as "100%% done".
 */
func TestPrintfFormatsLikeFmt(t *testing.T) {
	var out bytes.Buffer
	term := New(&out)
	term.Printf("100%% done;")
	term.Printf(" %d%% of %s;", 50, "x")
	term.Printf(" plain")
	term.Refresh()
	if got, want := out.String(), "100% done; 50% of x; plain"; got != want {
		t.Errorf("Printf wrote %q, want %q", got, want)
	}
}

/** TestResizeWhileDrawing is the data race that go test -race reports when a
 * resize handler calls UpdateTerminalSize while another goroutine draws and reads
 * the size: UpdateTerminalSize wrote the stored size without the lock every other
 * method takes. The size source is replaced because standard output is not a
 * terminal under go test. The race detector is what fails; without -race this test
 * only checks that nothing panics.
 */
func TestResizeWhileDrawing(t *testing.T) {
	saved := terminalSize
	defer func() { terminalSize = saved }()
	n := 0
	terminalSize = func() (int, int, error) { n++; return 80 + n%7, 24 + n%5, nil }
	var out bytes.Buffer
	term := New(&out)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 500; i++ {
			term.UpdateTerminalSize()
		}
	}()
	for i := 0; i < 500; i++ {
		term.Move(1, 1)
		term.Print("x")
		_ = term.GetTerminalWidth() + term.GetTerminalHeight()
		term.Refresh()
	}
	<-done
}
