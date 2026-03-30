package journal

import (
	"testing"
	"time"
)

func makeTestEntries() []Entry {
	return []Entry{
		{
			Date:    time.Date(2026, 3, 27, 9, 0, 0, 0, time.Local),
			Body:    "Thursday morning @work",
			Tags:    []string{"@work"},
			Starred: false,
		},
		{
			Date:    time.Date(2026, 3, 28, 14, 0, 0, 0, time.Local),
			Body:    "Friday afternoon @personal @mood",
			Tags:    []string{"@personal", "@mood"},
			Starred: true,
		},
		{
			Date: time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local),
			Body: "Saturday morning",
			Tags: nil,
		},
	}
}

func TestFilterByTag(t *testing.T) {
	entries := makeTestEntries()
	f := Filter{Tags: []string{"@work"}}
	result := f.Apply(entries)

	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if result[0].Body != "Thursday morning @work" {
		t.Errorf("wrong entry: %q", result[0].Body)
	}
}

func TestFilterByMultipleTags(t *testing.T) {
	entries := makeTestEntries()
	f := Filter{Tags: []string{"@personal", "@mood"}}
	result := f.Apply(entries)

	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
}

func TestFilterByStarred(t *testing.T) {
	entries := makeTestEntries()
	f := Filter{Starred: true}
	result := f.Apply(entries)

	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if !result[0].Starred {
		t.Error("expected starred entry")
	}
}

func TestFilterByDateRange(t *testing.T) {
	entries := makeTestEntries()
	start := time.Date(2026, 3, 28, 0, 0, 0, 0, time.Local)
	end := time.Date(2026, 3, 28, 23, 59, 59, 0, time.Local)
	f := Filter{StartDate: &start, EndDate: &end}
	result := f.Apply(entries)

	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if result[0].Date.Day() != 28 {
		t.Errorf("expected day 28, got %d", result[0].Date.Day())
	}
}

func TestFilterByContains(t *testing.T) {
	entries := makeTestEntries()
	f := Filter{Contains: "morning"}
	result := f.Apply(entries)

	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
}

func TestFilterByN(t *testing.T) {
	entries := makeTestEntries()
	f := Filter{N: 2}
	result := f.Apply(entries)

	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result[0].Date.Day() != 28 {
		t.Errorf("expected day 28 first, got %d", result[0].Date.Day())
	}
}

func TestFilterCombined(t *testing.T) {
	entries := makeTestEntries()
	start := time.Date(2026, 3, 27, 0, 0, 0, 0, time.Local)
	f := Filter{
		StartDate: &start,
		Contains:  "morning",
	}
	result := f.Apply(entries)

	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
}

func TestFilterEmpty(t *testing.T) {
	entries := makeTestEntries()
	f := Filter{}
	result := f.Apply(entries)

	if len(result) != 3 {
		t.Fatalf("expected 3 entries (no filter), got %d", len(result))
	}
}
