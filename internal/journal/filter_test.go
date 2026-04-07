package journal

import (
	"testing"
	"time"
)

func makeDay(dateStr, body string) Day {
	d, _ := time.Parse("2006-01-02", dateStr)
	return Day{Date: d, Body: body}
}

func TestFilterMatchAll(t *testing.T) {
	day := makeDay("2026-04-06", "\nSome content.\n")
	f := Filter{}
	if !f.Match(day, "@") {
		t.Error("empty filter should match all days")
	}
}

func TestFilterStart(t *testing.T) {
	start := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	f := Filter{Start: &start}

	before := makeDay("2026-04-04", "\nContent.\n")
	on := makeDay("2026-04-05", "\nContent.\n")
	after := makeDay("2026-04-06", "\nContent.\n")

	if f.Match(before, "@") {
		t.Error("before start date should not match")
	}
	if !f.Match(on, "@") {
		t.Error("on start date should match")
	}
	if !f.Match(after, "@") {
		t.Error("after start date should match")
	}
}

func TestFilterEnd(t *testing.T) {
	end := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	f := Filter{End: &end}

	before := makeDay("2026-04-04", "\nContent.\n")
	on := makeDay("2026-04-05", "\nContent.\n")
	after := makeDay("2026-04-06", "\nContent.\n")

	if !f.Match(before, "@") {
		t.Error("before end date should match")
	}
	if !f.Match(on, "@") {
		t.Error("on end date should match")
	}
	if f.Match(after, "@") {
		t.Error("after end date should not match")
	}
}

func TestFilterTagOR(t *testing.T) {
	f := Filter{Tags: []string{"@alice", "@bob"}}
	withAlice := makeDay("2026-04-01", "\nMet @alice today.\n")
	withBob := makeDay("2026-04-02", "\nCalled @bob.\n")
	withNeither := makeDay("2026-04-03", "\nQuiet day.\n")
	withBoth := makeDay("2026-04-04", "\n@alice and @bob together.\n")

	if !f.Match(withAlice, "@") {
		t.Error("withAlice should match OR filter")
	}
	if !f.Match(withBob, "@") {
		t.Error("withBob should match OR filter")
	}
	if f.Match(withNeither, "@") {
		t.Error("withNeither should not match")
	}
	if !f.Match(withBoth, "@") {
		t.Error("withBoth should match OR filter")
	}
}

func TestFilterTagAND(t *testing.T) {
	f := Filter{Tags: []string{"@alice", "@bob"}, AndTags: true}
	withAlice := makeDay("2026-04-01", "\nMet @alice today.\n")
	withBoth := makeDay("2026-04-04", "\n@alice and @bob together.\n")

	if f.Match(withAlice, "@") {
		t.Error("AND filter: only one tag, should not match")
	}
	if !f.Match(withBoth, "@") {
		t.Error("AND filter: both tags present, should match")
	}
}

func TestFilterNotTags(t *testing.T) {
	f := Filter{NotTags: []string{"@work"}}
	withWork := makeDay("2026-04-01", "\nDid @work stuff.\n")
	withoutWork := makeDay("2026-04-02", "\nRelaxed today.\n")

	if f.Match(withWork, "@") {
		t.Error("day with excluded tag should not match")
	}
	if !f.Match(withoutWork, "@") {
		t.Error("day without excluded tag should match")
	}
}

func TestFilterContains(t *testing.T) {
	f := Filter{Contains: "morning run"}
	match := makeDay("2026-04-01", "\nWent for a morning run today.\n")
	noMatch := makeDay("2026-04-02", "\nStayed home.\n")
	caseMatch := makeDay("2026-04-03", "\nWent for a Morning Run.\n")

	if !f.Match(match, "@") {
		t.Error("should match contains")
	}
	if f.Match(noMatch, "@") {
		t.Error("should not match contains")
	}
	if !f.Match(caseMatch, "@") {
		t.Error("contains should be case-insensitive")
	}
}

func TestFilterYear(t *testing.T) {
	f := Filter{Year: 2025}
	in := makeDay("2025-06-15", "\nContent.\n")
	out := makeDay("2026-06-15", "\nContent.\n")

	if !f.Match(in, "@") {
		t.Error("should match year filter")
	}
	if f.Match(out, "@") {
		t.Error("should not match different year")
	}
}

func TestFilterMonth(t *testing.T) {
	f := Filter{Month: 3}
	march2025 := makeDay("2025-03-15", "\nContent.\n")
	march2026 := makeDay("2026-03-10", "\nContent.\n")
	april2026 := makeDay("2026-04-10", "\nContent.\n")

	if !f.Match(march2025, "@") {
		t.Error("march 2025 should match month=3")
	}
	if !f.Match(march2026, "@") {
		t.Error("march 2026 should match month=3")
	}
	if f.Match(april2026, "@") {
		t.Error("april should not match month=3")
	}
}

func TestFilterDayOfMonth(t *testing.T) {
	f := Filter{DayOfMonth: 15}
	match := makeDay("2026-03-15", "\nContent.\n")
	noMatch := makeDay("2026-03-14", "\nContent.\n")

	if !f.Match(match, "@") {
		t.Error("day 15 should match")
	}
	if f.Match(noMatch, "@") {
		t.Error("day 14 should not match")
	}
}

func TestFilterTodayInHistory(t *testing.T) {
	today := time.Now()
	f := Filter{TodayInHistory: true}

	thisMonthDay := makeDay(
		time.Date(today.Year()-1, today.Month(), today.Day(), 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
		"\nContent.\n",
	)
	thisYear := makeDay(today.Format("2006-01-02"), "\nContent.\n")

	if !f.Match(thisMonthDay, "@") {
		t.Error("same month/day in prior year should match TodayInHistory")
	}
	if f.Match(thisYear, "@") {
		t.Error("today (current year) should not match TodayInHistory")
	}
}

func TestFilterN(t *testing.T) {
	f := Filter{N: 5}
	_ = f.N
}

func TestExtractTagsCustomSymbol(t *testing.T) {
	body := "Met with @alice and #project today."
	tags := extractTags(body, "@#")

	found := map[string]bool{}
	for _, tag := range tags {
		found[tag] = true
	}
	if !found["@alice"] {
		t.Error("should find @alice")
	}
	if !found["#project"] {
		t.Error("should find #project with # symbol")
	}
}
