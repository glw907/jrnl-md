package display

import (
	"fmt"
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

func TestColorFuncNoneReturnsNil(t *testing.T) {
	if ColorFunc("none") != nil {
		t.Error(`ColorFunc("none") should return nil`)
	}
}

func TestColorFuncUnknownReturnsNil(t *testing.T) {
	if ColorFunc("not-a-color") != nil {
		t.Error(`ColorFunc("not-a-color") should return nil`)
	}
}

func TestHighlightTagsNilColorFn(t *testing.T) {
	body := "Entry with @work tag"
	result := HighlightTags(body, "@", nil)
	if result != body {
		t.Errorf("nil colorFn should return body unchanged, got %q", result)
	}
}

func TestHighlightTagsNoTagsInBody(t *testing.T) {
	body := "Entry with no tags here"
	colorFn := func(a ...any) string { return "X" }
	result := HighlightTags(body, "@", colorFn)
	if result != body {
		t.Errorf("body with no tags should be returned unchanged, got %q", result)
	}
}

func TestHighlightTagsSingleSymbol(t *testing.T) {
	body := "Entry with @work and @home tags"
	colorFn := func(a ...any) string {
		s := fmt.Sprint(a...)
		return "[" + s + "]"
	}
	result := HighlightTags(body, "@", colorFn)
	want := "Entry with [@work] and [@home] tags"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestHighlightTagsMultipleSymbols(t *testing.T) {
	body := "Entry with @work and #project tags"
	colorFn := func(a ...any) string {
		s := fmt.Sprint(a...)
		return "[" + s + "]"
	}
	result := HighlightTags(body, "@#", colorFn)
	want := "Entry with [@work] and [#project] tags"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestHighlightTagsEmptySymbols(t *testing.T) {
	body := "Entry with @work tag"
	colorFn := func(a ...any) string { return "X" }
	result := HighlightTags(body, "", colorFn)
	if result != body {
		t.Errorf("empty tagSymbols should return body unchanged, got %q", result)
	}
}
