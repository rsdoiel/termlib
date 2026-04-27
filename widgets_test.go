// widgets_test.go — tests for TUI drawing helpers.
// Copyright (C) 2025 R. S. Doiel
package termlib

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestTruncate(t *testing.T) {
	cases := []struct {
		in   string
		maxW int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 8, "hello w…"},
		{"hello", 1, "…"},
		{"", 5, ""},
		{"日本語テスト", 4, "日本語…"},
	}
	for _, c := range cases {
		got := Truncate(c.in, c.maxW)
		if got != c.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", c.in, c.maxW, got, c.want)
		}
	}
}

func TestPadRight(t *testing.T) {
	cases := []struct {
		in   string
		w    int
		want string
	}{
		{"hi", 5, "hi   "},
		{"hello", 5, "hello"},
		{"hello world", 5, "hell…"},
		{"", 3, "   "},
	}
	for _, c := range cases {
		got := PadRight(c.in, c.w)
		if got != c.want {
			t.Errorf("PadRight(%q, %d) = %q, want %q", c.in, c.w, got, c.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{3*time.Minute + 7*time.Second, "3:07"},
		{0, "0:00"},
		{59 * time.Second, "0:59"},
		{time.Hour + 2*time.Minute + 5*time.Second, "1:02:05"},
		{90*time.Minute + 30*time.Second, "1:30:30"},
		// rounding: 500ms rounds up to 1s
		{500 * time.Millisecond, "0:01"},
	}
	for _, c := range cases {
		got := FormatDuration(c.d)
		if got != c.want {
			t.Errorf("FormatDuration(%v) = %q, want %q", c.d, got, c.want)
		}
	}
}

func TestDrawProgressBar(t *testing.T) {
	var buf bytes.Buffer
	term := New(&buf)

	DrawProgressBar(term, 1, 1, 12, 5, 10) // 50% of width=12 → inner=10 → 5 filled
	got := buf.String()

	// Output starts with a Move sequence then the bar
	if !strings.Contains(got, "[█████░░░░░]") {
		t.Errorf("DrawProgressBar 50%% got %q, expected bar [█████░░░░░]", got)
	}
}

func TestDrawProgressBarEmpty(t *testing.T) {
	var buf bytes.Buffer
	term := New(&buf)

	DrawProgressBar(term, 1, 1, 12, 0, 0) // total=0 → all empty
	got := buf.String()
	if !strings.Contains(got, "[░░░░░░░░░░]") {
		t.Errorf("DrawProgressBar empty got %q, expected all-empty bar", got)
	}
}

func TestDrawBoxNoTitle(t *testing.T) {
	var buf bytes.Buffer
	term := New(&buf)

	DrawBox(term, 1, 1, 10, 3, "")
	got := buf.String()

	if !strings.Contains(got, "┌────────┐") {
		t.Errorf("DrawBox top border missing, got %q", got)
	}
	if !strings.Contains(got, "└────────┘") {
		t.Errorf("DrawBox bottom border missing, got %q", got)
	}
}

func TestDrawBoxWithTitle(t *testing.T) {
	var buf bytes.Buffer
	term := New(&buf)

	DrawBox(term, 1, 1, 20, 3, "Test")
	got := buf.String()

	if !strings.Contains(got, "─ Test ─") {
		t.Errorf("DrawBox title missing, got %q", got)
	}
}
