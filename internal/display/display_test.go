package display

import (
	"strings"
	"testing"
)

func TestWrapText(t *testing.T) {
	text := "This is a long line that should be wrapped at a reasonable column width for display"
	wrapped := WrapText(text, 40)

	for i, line := range strings.Split(wrapped, "\n") {
		if len(line) > 40 {
			t.Errorf("line %d too long (%d chars): %q", i, len(line), line)
		}
	}
}

func TestWrapTextShortLine(t *testing.T) {
	text := "Short line."
	wrapped := WrapText(text, 80)

	if wrapped != text {
		t.Errorf("short line should not be wrapped: %q", wrapped)
	}
}

func TestIndentBody(t *testing.T) {
	body := "First line.\nSecond line."
	got := IndentBody(body, "| ")

	want := "| First line.\n| Second line."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIndentBodyBlankLines(t *testing.T) {
	body := "First paragraph.\n\nSecond paragraph."
	got := IndentBody(body, "| ")

	want := "| First paragraph.\n\n| Second paragraph."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
