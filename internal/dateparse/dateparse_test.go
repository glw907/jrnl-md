package dateparse

import (
	"testing"
	"time"
)

// anchor is a fixed reference time for deterministic natural-language tests.
// 2026-04-06 Monday 12:00:00 UTC
var anchor = time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)

func TestParseExplicitFull(t *testing.T) {
	cases := []struct {
		input string
		year  int
		month time.Month
		day   int
	}{
		{"2026-04-06", 2026, time.April, 6},
		{"2024-01-01", 2024, time.January, 1},
		{"1999-12-31", 1999, time.December, 31},
	}
	for _, tt := range cases {
		t.Run(tt.input, func(t *testing.T) {
			d, err := Parse(tt.input, anchor)
			if err != nil {
				t.Fatalf("Parse(%q): %v", tt.input, err)
			}
			if d.Year() != tt.year || d.Month() != tt.month || d.Day() != tt.day {
				t.Errorf("Parse(%q): got %v, want %d-%02d-%02d",
					tt.input, d, tt.year, tt.month, tt.day)
			}
		})
	}
}

func TestParseExplicitYearMonth(t *testing.T) {
	d, err := Parse("2026-03", anchor)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if d.Year() != 2026 || d.Month() != time.March || d.Day() != 1 {
		t.Errorf("got %v, want 2026-03-01", d)
	}
}

func TestParseExplicitYear(t *testing.T) {
	d, err := Parse("2025", anchor)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if d.Year() != 2025 || d.Month() != time.January || d.Day() != 1 {
		t.Errorf("got %v, want 2025-01-01", d)
	}
}

func TestParseExplicitMonthDay(t *testing.T) {
	d, err := Parse("03-15", anchor)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if d.Year() != anchor.Year() || d.Month() != time.March || d.Day() != 15 {
		t.Errorf("got %v, want %d-03-15", d, anchor.Year())
	}
}

func TestParseYesterday(t *testing.T) {
	d, err := Parse("yesterday", anchor)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := anchor.AddDate(0, 0, -1)
	if d.Year() != want.Year() || d.Month() != want.Month() || d.Day() != want.Day() {
		t.Errorf("yesterday: got %v, want %v", d, want)
	}
}

func TestParseToday(t *testing.T) {
	d, err := Parse("today", anchor)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if d.Year() != anchor.Year() || d.Month() != anchor.Month() || d.Day() != anchor.Day() {
		t.Errorf("today: got %v, want %v", d, anchor)
	}
}

func TestParseNDaysAgo(t *testing.T) {
	d, err := Parse("3 days ago", anchor)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := anchor.AddDate(0, 0, -3)
	if d.Year() != want.Year() || d.Month() != want.Month() || d.Day() != want.Day() {
		t.Errorf("3 days ago: got %v, want %v", d, want)
	}
}

func TestParseLastWeekday(t *testing.T) {
	// anchor is Monday 2026-04-06; "last monday" should be 2026-03-30
	d, err := Parse("last monday", anchor)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if d.Weekday() != time.Monday {
		t.Errorf("last monday weekday: got %v, want Monday", d.Weekday())
	}
	if !d.Before(anchor) {
		t.Errorf("last monday should be before anchor: got %v", d)
	}
}

func TestParseInvalid(t *testing.T) {
	cases := []string{
		"not-a-date",
		"",
		"9999-99-99",
		"foo bar baz",
	}
	for _, tt := range cases {
		t.Run(tt, func(t *testing.T) {
			_, err := Parse(tt, anchor)
			if err == nil {
				t.Errorf("Parse(%q): expected error, got nil", tt)
			}
		})
	}
}
