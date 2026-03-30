package dateparse

import (
	"fmt"
	"testing"
	"time"
)

func TestParseExplicitDate(t *testing.T) {
	tests := []struct {
		input         string
		defaultHour   int
		defaultMinute int
		expectedHour  int
		expectedMin   int
	}{
		{"2026-03-29", 0, 0, 0, 0},
		{"2026-03-29", 9, 30, 9, 30},
		{"2026-01-01", 14, 15, 14, 15},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_h%d_m%d", tt.input, tt.defaultHour, tt.defaultMinute), func(t *testing.T) {
			got, err := Parse(tt.input, tt.defaultHour, tt.defaultMinute)
			if err != nil {
				t.Fatalf("Parse(%q) failed: %v", tt.input, err)
			}
			if got.Hour() != tt.expectedHour || got.Minute() != tt.expectedMin {
				t.Errorf("Parse(%q) hour=%d min=%d, want hour=%d min=%d",
					tt.input, got.Hour(), got.Minute(), tt.expectedHour, tt.expectedMin)
			}
		})
	}
}

func TestParseFullDatetimeIgnoresDefaults(t *testing.T) {
	got, err := Parse("2026-03-29 14:30", 9, 0)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if got.Hour() != 14 || got.Minute() != 30 {
		t.Errorf("expected 14:30, got %02d:%02d", got.Hour(), got.Minute())
	}
}

func TestParseNaturalLanguage(t *testing.T) {
	for _, input := range []string{"yesterday", "3 days ago"} {
		t.Run(input, func(t *testing.T) {
			got, err := Parse(input, 9, 0)
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
	got, err := Parse("yesterday", 9, 0)
	if err != nil {
		t.Fatalf("Parse(yesterday) failed: %v", err)
	}
	expected := time.Now().AddDate(0, 0, -1)
	if !sameDay(got, expected) {
		t.Errorf("yesterday = %v, want %v", got, expected)
	}
}

func TestParseInvalid(t *testing.T) {
	_, err := Parse("not a date at all xyzzy", 9, 0)
	if err == nil {
		t.Error("expected error for invalid date")
	}
}

func TestParseInclusive(t *testing.T) {
	got, err := ParseInclusive("2026-03-29", 9, 0)
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
