package dateparse

import (
	"testing"
	"time"
)

func TestParseExplicitDate(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		{"2026-03-29", time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local)},
		{"2026-01-01", time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) failed: %v", tt.input, err)
			}
			if !sameDay(got, tt.expected) {
				t.Errorf("Parse(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseNaturalLanguage(t *testing.T) {
	for _, input := range []string{"yesterday", "3 days ago"} {
		t.Run(input, func(t *testing.T) {
			got, err := Parse(input)
			if err != nil {
				t.Fatalf("Parse(%q) failed: %v", input, err)
			}
			if got.IsZero() {
				t.Errorf("Parse(%q) returned zero time", input)
			}
		})
	}
}

func TestParseYesterday(t *testing.T) {
	got, err := Parse("yesterday")
	if err != nil {
		t.Fatalf("Parse(yesterday) failed: %v", err)
	}

	expected := time.Now().AddDate(0, 0, -1)
	if !sameDay(got, expected) {
		t.Errorf("yesterday = %v, want %v", got, expected)
	}
}

func TestParseInvalid(t *testing.T) {
	_, err := Parse("not a date at all xyzzy")
	if err == nil {
		t.Error("expected error for invalid date")
	}
}

func TestParseInclusive(t *testing.T) {
	got, err := ParseInclusive("2026-03-29")
	if err != nil {
		t.Fatalf("ParseInclusive failed: %v", err)
	}

	if got.Hour() != 23 || got.Minute() != 59 || got.Second() != 59 {
		t.Errorf("expected end of day, got %v", got)
	}
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
