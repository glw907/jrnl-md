package journal

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestDayFormat(t *testing.T) {
	d := day{
		date: time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local),
		entries: []Entry{
			{
				Date: time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local),
				Body: "Morning thoughts.",
			},
			{
				Date: time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local),
				Body: "Afternoon entry.",
			},
		},
	}

	got := d.Format("2006-01-02", "03:04 PM")
	expected := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nMorning thoughts.\n\n## [02:30 PM]\n\nAfternoon entry.\n"

	if got != expected {
		t.Errorf("got:\n%s\nwant:\n%s", got, expected)
	}
}

func TestDayFormatStarred(t *testing.T) {
	d := day{
		date: time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local),
		entries: []Entry{
			{
				Date:    time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local),
				Body:    "Great day.",
				Starred: true,
			},
		},
	}

	got := d.Format("2006-01-02", "03:04 PM")
	expected := "# 2026-03-29 Sunday\n\n## [09:00 AM] *\n\nGreat day.\n"

	if got != expected {
		t.Errorf("got:\n%s\nwant:\n%s", got, expected)
	}
}

func TestParseDay(t *testing.T) {
	text := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nMorning thoughts.\n\n## [02:30 PM]\n\nAfternoon entry.\n"

	d, err := parseDay(text, "2006-01-02", "03:04 PM")
	if err != nil {
		t.Fatalf("parseDay failed: %v", err)
	}

	if d.date.Day() != 29 || d.date.Month() != 3 || d.date.Year() != 2026 {
		t.Errorf("wrong date: %v", d.date)
	}

	if len(d.entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(d.entries))
	}

	if d.entries[0].Body != "Morning thoughts." {
		t.Errorf("entry 0 body = %q", d.entries[0].Body)
	}
	if d.entries[0].Date.Hour() != 9 || d.entries[0].Date.Minute() != 0 {
		t.Errorf("entry 0 time = %v", d.entries[0].Date)
	}

	if d.entries[1].Body != "Afternoon entry." {
		t.Errorf("entry 1 body = %q", d.entries[1].Body)
	}
	if d.entries[1].Date.Hour() != 14 || d.entries[1].Date.Minute() != 30 {
		t.Errorf("entry 1 time = %v", d.entries[1].Date)
	}
}

func TestParseDayStarred(t *testing.T) {
	text := "# 2026-03-29 Sunday\n\n## [09:00 AM] *\n\nStarred entry.\n"

	d, err := parseDay(text, "2006-01-02", "03:04 PM")
	if err != nil {
		t.Fatalf("parseDay failed: %v", err)
	}

	if len(d.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(d.entries))
	}

	if !d.entries[0].Starred {
		t.Error("expected starred entry")
	}
	if d.entries[0].Body != "Starred entry." {
		t.Errorf("body = %q", d.entries[0].Body)
	}
}

func TestParseDayMultiParagraph(t *testing.T) {
	text := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nFirst paragraph.\n\nSecond paragraph.\n"

	d, err := parseDay(text, "2006-01-02", "03:04 PM")
	if err != nil {
		t.Fatalf("parseDay failed: %v", err)
	}

	if len(d.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(d.entries))
	}

	if d.entries[0].Body != "First paragraph.\n\nSecond paragraph." {
		t.Errorf("body = %q", d.entries[0].Body)
	}
}

func TestRoundtrip(t *testing.T) {
	d := day{
		date: time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local),
		entries: []Entry{
			{
				Date: time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local),
				Body: "Morning thoughts.",
			},
			{
				Date:    time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local),
				Body:    "Afternoon. @mood",
				Starred: true,
			},
		},
	}

	dateFmt := "2006-01-02"
	timeFmt := "03:04 PM"

	serialized := d.Format(dateFmt, timeFmt)
	parsed, err := parseDay(serialized, dateFmt, timeFmt)
	if err != nil {
		t.Fatalf("roundtrip parse failed: %v", err)
	}

	if len(parsed.entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(parsed.entries))
	}

	if parsed.entries[0].Body != "Morning thoughts." {
		t.Errorf("entry 0 body = %q", parsed.entries[0].Body)
	}
	if parsed.entries[1].Body != "Afternoon. @mood" {
		t.Errorf("entry 1 body = %q", parsed.entries[1].Body)
	}
	if !parsed.entries[1].Starred {
		t.Error("entry 1 should be starred")
	}
}

func TestAddEntry(t *testing.T) {
	d := day{
		date: time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local),
	}

	d.addEntry("New entry body.", false, time.Date(2026, 3, 29, 10, 0, 0, 0, time.Local))

	if len(d.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(d.entries))
	}

	if d.entries[0].Body != "New entry body." {
		t.Errorf("body = %q", d.entries[0].Body)
	}
}

func TestParseMultiDaySingleDay(t *testing.T) {
	text := "# 2026-03-01 Sunday\n\n## [09:00 AM]\n\nFirst entry.\n\n## [02:00 PM]\n\nSecond entry.\n"

	entries, err := ParseMultiDay(text, "2006-01-02", "03:04 PM")
	if err != nil {
		t.Fatalf("ParseMultiDay failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Body != "First entry." {
		t.Errorf("entry 0 body = %q", entries[0].Body)
	}
	if entries[1].Body != "Second entry." {
		t.Errorf("entry 1 body = %q", entries[1].Body)
	}
}

func TestParseMultiDayMultipleDays(t *testing.T) {
	text := "# 2026-03-01 Sunday\n\n## [09:00 AM]\n\nDay one entry.\n\n# 2026-03-15 Sunday\n\n## [10:00 AM]\n\nDay two entry.\n"

	entries, err := ParseMultiDay(text, "2006-01-02", "03:04 PM")
	if err != nil {
		t.Fatalf("ParseMultiDay failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Body != "Day one entry." {
		t.Errorf("entry 0 body = %q", entries[0].Body)
	}
	if entries[0].Date.Day() != 1 {
		t.Errorf("entry 0 day = %d, want 1", entries[0].Date.Day())
	}
	if entries[1].Body != "Day two entry." {
		t.Errorf("entry 1 body = %q", entries[1].Body)
	}
	if entries[1].Date.Day() != 15 {
		t.Errorf("entry 1 day = %d, want 15", entries[1].Date.Day())
	}
}

func TestParseMultiDayEmpty(t *testing.T) {
	entries, err := ParseMultiDay("", "2006-01-02", "03:04 PM")
	if err != nil {
		t.Fatalf("ParseMultiDay on empty string failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries on empty input, got %d", len(entries))
	}
}

func TestParseDayErrorMissingTitle(t *testing.T) {
	_, err := parseDay("no heading here\n\n## [09:00 AM]\n\nBody.\n", "2006-01-02", "03:04 PM")
	if err == nil {
		t.Fatal("expected error for missing day heading")
	}
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *ParseError, got %T: %v", err, err)
	}
	if pe.Line != 1 {
		t.Errorf("line = %d, want 1", pe.Line)
	}
	if !strings.Contains(pe.Error(), "expected") {
		t.Errorf("error should contain expected format: %s", pe.Error())
	}
}

func TestParseDayErrorBadTime(t *testing.T) {
	_, err := parseDay("# 2026-03-29 Sunday\n\n## [3:59pm]\n\nBody.\n", "2006-01-02", "03:04 PM")
	if err == nil {
		t.Fatal("expected error for bad time")
	}
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *ParseError, got %T: %v", err, err)
	}
	if pe.Line == 0 {
		t.Error("expected non-zero line number")
	}
	if pe.Value != "3:59pm" {
		t.Errorf("value = %q, want %q", pe.Value, "3:59pm")
	}
}

func TestParseDayErrorBadDate(t *testing.T) {
	_, err := parseDay("# not-a-date Sunday\n\n## [09:00 AM]\n\nBody.\n", "2006-01-02", "03:04 PM")
	if err == nil {
		t.Fatal("expected error for bad date")
	}
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *ParseError, got %T: %v", err, err)
	}
	if pe.Line != 1 {
		t.Errorf("line = %d, want 1", pe.Line)
	}
}

func TestParseMultiDayErrorIncludesLineOffset(t *testing.T) {
	text := "# 2026-03-01 Sunday\n\n## [09:00 AM]\n\nGood entry.\n\n# 2026-03-15 Sunday\n\n## [bad-time]\n\nBad entry.\n"
	_, err := ParseMultiDay(text, "2006-01-02", "03:04 PM")
	if err == nil {
		t.Fatal("expected error")
	}
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *ParseError, got %T: %v", err, err)
	}
	// "## [bad-time]" is on line 9 of the full text (1-based).
	// parseDay sees it as line 3 within the section; section starts at line 6 (0-based index).
	// Offset: 3 + 6 = 9.
	if pe.Line != 9 {
		t.Errorf("line = %d, want 9", pe.Line)
	}
}

func TestParseMultiDayRoundtrip(t *testing.T) {
	original := "# 2026-03-01 Sunday\n\n## [09:00 AM]\n\nMorning. @work\n\n## [02:00 PM] *\n\nStarred entry.\n\n# 2026-03-15 Sunday\n\n## [10:00 AM]\n\nMid-month.\n"

	entries, err := ParseMultiDay(original, "2006-01-02", "03:04 PM")
	if err != nil {
		t.Fatalf("ParseMultiDay failed: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if !entries[1].Starred {
		t.Error("expected entry 1 to be starred")
	}
	if entries[1].Body != "Starred entry." {
		t.Errorf("entry 1 body = %q", entries[1].Body)
	}
}
