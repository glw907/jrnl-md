package e2e

import (
	"strings"
	"testing"
	"time"
)

func setupTagsJournal(t *testing.T, env testEnv) {
	t.Helper()
	entries := []struct {
		date time.Time
		body string
	}{
		{time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC), "# 2026-04-06 Monday\n\nMet @alice and @bob. @work meeting.\n"},
		{time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC), "# 2026-04-05 Sunday\n\n@alice called. @reading.\n"},
		{time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC), "# 2026-04-03 Friday\n\n@work presentation. @alice again.\n"},
		{time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC), "# 2026-03-15 Sunday\n\n@home @reading.\n"},
	}
	for _, e := range entries {
		env.writeDayFile(t, e.date, e.body)
	}
}

func TestTagsAll(t *testing.T) {
	env := newTestEnv(t)
	setupTagsJournal(t, env)

	out, err := env.run(t, "tags")
	if err != nil {
		t.Fatalf("tags: %v", err)
	}
	mustContain(t, out, "@alice")
	mustContain(t, out, "@work")
	mustContain(t, out, "@reading")
}

func TestTagsFrequencySort(t *testing.T) {
	env := newTestEnv(t)
	setupTagsJournal(t, env)

	out, err := env.run(t, "tags")
	if err != nil {
		t.Fatalf("tags: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 {
		t.Fatal("tags: no output")
	}
	// @alice appears 3 times — should be first
	if !strings.HasPrefix(lines[0], "@alice") {
		t.Errorf("tags: @alice (3 occurrences) should be first, got: %q", lines[0])
	}
}

func TestTagsFormat(t *testing.T) {
	env := newTestEnv(t)
	setupTagsJournal(t, env)

	out, err := env.run(t, "tags")
	if err != nil {
		t.Fatalf("tags: %v", err)
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		if !strings.Contains(line, ": ") {
			t.Errorf("tags: line missing ': ' separator: %q", line)
		}
	}
}

func TestTagsWithDateFilter(t *testing.T) {
	env := newTestEnv(t)
	setupTagsJournal(t, env)

	out, err := env.run(t, "tags", "--from", "2026-04-01", "--to", "2026-04-06")
	if err != nil {
		t.Fatalf("tags --from --to: %v", err)
	}
	mustContain(t, out, "@alice")
	mustNotContain(t, out, "@home")
}

func TestTagsEmpty(t *testing.T) {
	env := newTestEnv(t)
	out, err := env.run(t, "tags")
	if err != nil {
		t.Fatalf("tags empty: %v", err)
	}
	if strings.TrimSpace(out) != "" {
		t.Errorf("tags on empty journal should produce no output: %q", out)
	}
}
