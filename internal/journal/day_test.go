package journal

import (
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
