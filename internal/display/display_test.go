package display

import (
	"strings"
	"testing"
	"time"

	"github.com/glw907/jrnl-md/internal/journal"
)

func TestWrapText(t *testing.T) {
	long := strings.Repeat("word ", 20) // 100+ chars
	wrapped := wrapText(long, 40)
	for _, line := range strings.Split(wrapped, "\n") {
		if len(line) > 40 {
			t.Errorf("line exceeds width 40: %q (len=%d)", line, len(line))
		}
	}
}

func TestWrapTextPreservesBlankLines(t *testing.T) {
	input := "First paragraph.\n\nSecond paragraph.\n"
	wrapped := wrapText(input, 79)
	if !strings.Contains(wrapped, "\n\n") {
		t.Error("wrapText: blank lines should be preserved")
	}
}

func TestWrapTextShortLinesUnchanged(t *testing.T) {
	input := "Short line.\nAnother short.\n"
	wrapped := wrapText(input, 79)
	if wrapped != input {
		t.Errorf("short lines changed:\ngot:  %q\nwant: %q", wrapped, input)
	}
}

func TestTruncateWithEllipsis(t *testing.T) {
	cases := []struct {
		input string
		width int
		want  string
	}{
		{"Hello world", 20, "Hello world"},
		{"Hello world", 8, "Hello w…"},
		{"Hi", 5, "Hi"},
		{"", 10, ""},
	}
	for _, tt := range cases {
		t.Run(tt.input, func(t *testing.T) {
			got := truncate(tt.input, tt.width)
			if got != tt.want {
				t.Errorf("truncate(%q, %d): got %q, want %q", tt.input, tt.width, got, tt.want)
			}
		})
	}
}

func TestShortLine(t *testing.T) {
	date := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	day := journal.Day{
		Date: date,
		Body: "\nWent for a morning run. Feeling good about the week ahead.\n",
	}

	line := ShortLine(day, 79, "")
	if !strings.HasPrefix(line, "2026-04-06") {
		t.Errorf("ShortLine: should start with date: %q", line)
	}
	if !strings.Contains(line, "Went for a morning run") {
		t.Errorf("ShortLine: should contain first content line: %q", line)
	}
}

func TestShortLineSkipsTimestampHeading(t *testing.T) {
	date := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	day := journal.Day{
		Date: date,
		Body: "\n## 09:00 AM\n\nActual content here.\n",
	}

	line := ShortLine(day, 79, "03:04 PM")
	if strings.Contains(line, "09:00 AM") {
		t.Errorf("ShortLine: should skip timestamp heading: %q", line)
	}
	if !strings.Contains(line, "Actual content") {
		t.Errorf("ShortLine: should use content after timestamp: %q", line)
	}
}

func TestShortLineHeadingMarkersStripped(t *testing.T) {
	date := time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC)
	day := journal.Day{
		Date: date,
		Body: "\n## Quarterly Review\n\nDetails here.\n",
	}

	line := ShortLine(day, 79, "")
	if strings.Contains(line, "##") {
		t.Errorf("ShortLine: ## markers should be stripped: %q", line)
	}
	if !strings.Contains(line, "Quarterly Review") {
		t.Errorf("ShortLine: should contain heading text: %q", line)
	}
}

func TestShortLineTruncated(t *testing.T) {
	date := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	longBody := "\n" + strings.Repeat("a very long word sequence ", 10) + "\n"
	day := journal.Day{Date: date, Body: longBody}

	line := ShortLine(day, 50, "")
	if len([]rune(line)) > 50 {
		t.Errorf("ShortLine: should be truncated to width 50, got len=%d: %q",
			len([]rune(line)), line)
	}
	if !strings.HasSuffix(line, "…") {
		t.Errorf("ShortLine: truncated line should end with ellipsis: %q", line)
	}
}

func TestFormatFullDay(t *testing.T) {
	date := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	day := journal.Day{
		Date: date,
		Body: "\nWent for a morning run. Met with @sarah.\n",
	}

	opts := FormatOpts{
		Linewrap:   79,
		DateColor:  "none",
		BodyColor:  "none",
		TagsColor:  "none",
		TagSymbols: "@",
	}
	out := FormatDay(day, opts)
	if !strings.Contains(out, "2026-04-06") {
		t.Errorf("FormatDay: missing date: %q", out)
	}
	if !strings.Contains(out, "morning run") {
		t.Errorf("FormatDay: missing body: %q", out)
	}
}

func TestFormatDays(t *testing.T) {
	days := []journal.Day{
		{
			Date: time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC),
			Body: "\nFirst day.\n",
		},
		{
			Date: time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC),
			Body: "\nSecond day.\n",
		},
	}
	opts := FormatOpts{Linewrap: 79, DateColor: "none", BodyColor: "none", TagsColor: "none", TagSymbols: "@"}
	out := FormatDays(days, opts)
	if !strings.Contains(out, "2026-04-06") || !strings.Contains(out, "2026-04-05") {
		t.Errorf("FormatDays: missing dates: %q", out)
	}
	if !strings.Contains(out, "\n\n") {
		t.Error("FormatDays: should have blank line between days")
	}
}
