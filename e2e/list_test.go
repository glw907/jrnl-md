package e2e

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func setupListJournal(t *testing.T, env testEnv) {
	t.Helper()
	entries := []struct {
		date time.Time
		body string
	}{
		{time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC), "# 2026-04-06 Monday\n\nMorning run. Met with @alice.\n"},
		{time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC), "# 2026-04-05 Sunday\n\nQuiet day at home. @reading.\n"},
		{time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC), "# 2026-04-03 Friday\n\nWork review with @alice and @bob.\n"},
		{time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC), "# 2026-03-15 Sunday\n\nSpring cleaning. @home.\n"},
		{time.Date(2025, 4, 6, 0, 0, 0, 0, time.UTC), "# 2025-04-06 Sunday\n\nLast year entry. @work.\n"},
	}
	for _, e := range entries {
		env.writeDayFile(t, e.date, e.body)
	}
}

func TestListDefault(t *testing.T) {
	env := newTestEnv(t)
	setupListJournal(t, env)

	out, err := env.run(t, "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	mustContain(t, out, "2026-04-06")
}

func TestListAll(t *testing.T) {
	env := newTestEnv(t)
	setupListJournal(t, env)

	out, err := env.run(t, "list", "--all")
	if err != nil {
		t.Fatalf("list --all: %v", err)
	}
	mustContain(t, out, "2026-04-06")
	mustContain(t, out, "2025-04-06")
}

func TestListN(t *testing.T) {
	env := newTestEnv(t)
	setupListJournal(t, env)

	out, err := env.run(t, "list", "-n", "2")
	if err != nil {
		t.Fatalf("list -n 2: %v", err)
	}
	mustContain(t, out, "2026-04-06")
	mustContain(t, out, "2026-04-05")
	mustNotContain(t, out, "2026-04-03")
}

func TestListShort(t *testing.T) {
	env := newTestEnv(t)
	setupListJournal(t, env)

	out, err := env.run(t, "list", "--short", "--all")
	if err != nil {
		t.Fatalf("list --short: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		if len(line) < 10 || line[4] != '-' {
			t.Errorf("short line doesn't start with date: %q", line)
		}
	}
}

func TestListFrom(t *testing.T) {
	env := newTestEnv(t)
	setupListJournal(t, env)

	out, err := env.run(t, "list", "--from", "2026-04-01", "--all")
	if err != nil {
		t.Fatalf("list --from: %v", err)
	}
	mustContain(t, out, "2026-04-06")
	mustNotContain(t, out, "2026-03-15")
	mustNotContain(t, out, "2025-04-06")
}

func TestListTo(t *testing.T) {
	env := newTestEnv(t)
	setupListJournal(t, env)

	out, err := env.run(t, "list", "--to", "2026-04-04", "--all")
	if err != nil {
		t.Fatalf("list --to: %v", err)
	}
	mustContain(t, out, "2026-04-03")
	mustNotContain(t, out, "2026-04-05")
	mustNotContain(t, out, "2026-04-06")
}

func TestListOn(t *testing.T) {
	env := newTestEnv(t)
	setupListJournal(t, env)

	out, err := env.run(t, "list", "--on", "2026-04-05")
	if err != nil {
		t.Fatalf("list --on: %v", err)
	}
	mustContain(t, out, "2026-04-05")
	mustNotContain(t, out, "2026-04-06")
	mustNotContain(t, out, "2026-04-03")
}

func TestListTagFilter(t *testing.T) {
	env := newTestEnv(t)
	setupListJournal(t, env)

	out, err := env.run(t, "list", "--all", "@alice")
	if err != nil {
		t.Fatalf("list @alice: %v", err)
	}
	mustContain(t, out, "2026-04-06")
	mustContain(t, out, "2026-04-03")
	mustNotContain(t, out, "2026-04-05")
}

func TestListTagAnd(t *testing.T) {
	env := newTestEnv(t)
	setupListJournal(t, env)

	out, err := env.run(t, "list", "--all", "--and", "@alice", "@bob")
	if err != nil {
		t.Fatalf("list --and @alice @bob: %v", err)
	}
	mustContain(t, out, "2026-04-03")
	mustNotContain(t, out, "2026-04-06")
}

func TestListTagNot(t *testing.T) {
	env := newTestEnv(t)
	setupListJournal(t, env)

	out, err := env.run(t, "list", "--all", "--not", "@alice")
	if err != nil {
		t.Fatalf("list --not @alice: %v", err)
	}
	mustNotContain(t, out, "2026-04-06")
	mustNotContain(t, out, "2026-04-03")
	mustContain(t, out, "2026-04-05")
}

func TestListContains(t *testing.T) {
	env := newTestEnv(t)
	setupListJournal(t, env)

	out, err := env.run(t, "list", "--all", "--contains", "spring cleaning")
	if err != nil {
		t.Fatalf("list --contains: %v", err)
	}
	mustContain(t, out, "2026-03-15")
	mustNotContain(t, out, "2026-04-06")
}

func TestListYearFilter(t *testing.T) {
	env := newTestEnv(t)
	setupListJournal(t, env)

	out, err := env.run(t, "list", "--year", "2025")
	if err != nil {
		t.Fatalf("list --year: %v", err)
	}
	mustContain(t, out, "2025-04-06")
	mustNotContain(t, out, "2026-04-06")
}

func TestListMonthFilter(t *testing.T) {
	env := newTestEnv(t)
	setupListJournal(t, env)

	out, err := env.run(t, "list", "--all", "--month", "3")
	if err != nil {
		t.Fatalf("list --month: %v", err)
	}
	mustContain(t, out, "2026-03-15")
	mustNotContain(t, out, "2026-04-06")
}

func TestListDayFilter(t *testing.T) {
	env := newTestEnv(t)
	setupListJournal(t, env)

	out, err := env.run(t, "list", "--all", "--day", "6")
	if err != nil {
		t.Fatalf("list --day: %v", err)
	}
	mustContain(t, out, "2026-04-06")
	mustContain(t, out, "2025-04-06")
	mustNotContain(t, out, "2026-04-05")
}

func TestListTodayInHistory(t *testing.T) {
	env := newTestEnv(t)
	today := time.Now()
	pastDate := time.Date(today.Year()-1, today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	heading := fmt.Sprintf("# %s %s\n", pastDate.Format("2006-01-02"), pastDate.Format("Monday"))
	env.writeDayFile(t, pastDate, heading+"\nHistory entry.\n")

	out, err := env.run(t, "list", "--today-in-history")
	if err != nil {
		t.Fatalf("list --today-in-history: %v", err)
	}
	mustContain(t, out, pastDate.Format("2006-01-02"))
}

func TestListEmpty(t *testing.T) {
	env := newTestEnv(t)
	out, err := env.run(t, "list", "--all")
	if err != nil {
		t.Fatalf("list empty journal: %v", err)
	}
	if strings.TrimSpace(out) != "" {
		t.Errorf("empty journal list should produce no output: %q", out)
	}
}
