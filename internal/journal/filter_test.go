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

func TestFilterAndTags(t *testing.T) {
	entries := makeTestEntries()
	// "Friday afternoon @personal @mood" has both tags; others don't
	f := Filter{Tags: []string{"@personal", "@mood"}, AndTags: true}
	result := f.Apply(entries)

	if len(result) != 1 {
		t.Fatalf("expected 1 entry with both tags, got %d", len(result))
	}
	if result[0].Body != "Friday afternoon @personal @mood" {
		t.Errorf("wrong entry: %q", result[0].Body)
	}
}

func TestFilterAndTagsNoMatch(t *testing.T) {
	entries := makeTestEntries()
	// No entry has both @work and @personal
	f := Filter{Tags: []string{"@work", "@personal"}, AndTags: true}
	result := f.Apply(entries)

	if len(result) != 0 {
		t.Fatalf("expected 0 entries for AND with no match, got %d", len(result))
	}
}

func TestFilterNotTags(t *testing.T) {
	entries := makeTestEntries()
	// Exclude entries containing @work
	f := Filter{NotTags: []string{"@work"}}
	result := f.Apply(entries)

	if len(result) != 2 {
		t.Fatalf("expected 2 entries (without @work), got %d", len(result))
	}
	for _, e := range result {
		for _, tag := range e.Tags {
			if tag == "@work" {
				t.Errorf("entry with @work should be excluded: %q", e.Body)
			}
		}
	}
}

func TestFilterNotTagsMultiple(t *testing.T) {
	entries := makeTestEntries()
	// Exclude both @work and @personal — only "Saturday morning" remains
	f := Filter{NotTags: []string{"@work", "@personal"}}
	result := f.Apply(entries)

	if len(result) != 1 {
		t.Fatalf("expected 1 entry after excluding @work and @personal, got %d", len(result))
	}
	if result[0].Body != "Saturday morning" {
		t.Errorf("wrong remaining entry: %q", result[0].Body)
	}
}

func TestFilterNotStarred(t *testing.T) {
	entries := makeTestEntries()
	// Only one entry is starred; exclude it
	f := Filter{NotStarred: true}
	result := f.Apply(entries)

	if len(result) != 2 {
		t.Fatalf("expected 2 non-starred entries, got %d", len(result))
	}
	for _, e := range result {
		if e.Starred {
			t.Errorf("starred entry should be excluded: %q", e.Body)
		}
	}
}

func TestFilterNotTagged(t *testing.T) {
	entries := makeTestEntries()
	// Only "Saturday morning" has no tags
	f := Filter{NotTagged: true}
	result := f.Apply(entries)

	if len(result) != 1 {
		t.Fatalf("expected 1 untagged entry, got %d", len(result))
	}
	if result[0].Body != "Saturday morning" {
		t.Errorf("wrong entry: %q", result[0].Body)
	}
}

func TestFilterDateRange(t *testing.T) {
	t.Run("no dates", func(t *testing.T) {
		f := Filter{}
		start, end := f.DateRange()
		if start != nil || end != nil {
			t.Errorf("expected nil/nil, got %v/%v", start, end)
		}
	})

	t.Run("start only", func(t *testing.T) {
		st := time.Date(2026, 3, 15, 14, 30, 0, 0, time.Local)
		f := Filter{StartDate: &st}
		start, end := f.DateRange()
		if start == nil {
			t.Fatal("expected non-nil start")
		}
		if start.Hour() != 0 || start.Minute() != 0 || start.Day() != 15 {
			t.Errorf("start not truncated to start of day: %v", start)
		}
		if end != nil {
			t.Errorf("expected nil end, got %v", end)
		}
	})

	t.Run("end only", func(t *testing.T) {
		en := time.Date(2026, 3, 20, 10, 0, 0, 0, time.Local)
		f := Filter{EndDate: &en}
		start, end := f.DateRange()
		if start != nil {
			t.Errorf("expected nil start, got %v", start)
		}
		if end == nil {
			t.Fatal("expected non-nil end")
		}
		if end.Hour() != 23 || end.Minute() != 59 || end.Day() != 20 {
			t.Errorf("end not expanded to end of day: %v", end)
		}
	})

	t.Run("both dates", func(t *testing.T) {
		st := time.Date(2026, 3, 15, 14, 0, 0, 0, time.Local)
		en := time.Date(2026, 3, 20, 10, 0, 0, 0, time.Local)
		f := Filter{StartDate: &st, EndDate: &en}
		start, end := f.DateRange()
		if start == nil || end == nil {
			t.Fatal("expected non-nil start and end")
		}
		if start.Day() != 15 || start.Hour() != 0 {
			t.Errorf("start wrong: %v", start)
		}
		if end.Day() != 20 || end.Hour() != 23 {
			t.Errorf("end wrong: %v", end)
		}
	})
}
