package e2e

import (
	"strings"
	"testing"
	"time"
)

func seedFilterJournal(t *testing.T, env testEnv) {
	t.Helper()
	writeDayFile(t, env.journalDir, time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local),
		"# 2026-03-01 Sunday\n\n## [09:00 AM]\n\nMarch first entry with @work tag.\n")
	writeDayFile(t, env.journalDir, time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local),
		"# 2026-03-15 Sunday\n\n## [09:00 AM]\n\nMid-month entry about @project planning.\n")
	writeDayFile(t, env.journalDir, time.Date(2026, 3, 28, 0, 0, 0, 0, time.Local),
		"# 2026-03-28 Saturday\n\n## [09:00 AM]\n\nLate month @personal entry.\n")
}

func TestFromFilter(t *testing.T) {
	env := newTestEnv(t)
	seedFilterJournal(t, env)

	stdout, stderr := run(t, env, "--from", "2026-03-15", "--num", "99")

	if !strings.Contains(stderr, "2 entries found") {
		t.Errorf("expected stderr to contain '2 entries found', got: %q", stderr)
	}
	if !strings.Contains(stdout, "Mid-month entry") {
		t.Errorf("expected stdout to contain 'Mid-month entry', got: %q", stdout)
	}
	if !strings.Contains(stdout, "Late month") {
		t.Errorf("expected stdout to contain 'Late month', got: %q", stdout)
	}
	if strings.Contains(stdout, "March first entry") {
		t.Errorf("expected stdout NOT to contain 'March first entry', got: %q", stdout)
	}
}

func TestToFilter(t *testing.T) {
	env := newTestEnv(t)
	seedFilterJournal(t, env)

	stdout, stderr := run(t, env, "--to", "2026-03-15", "--num", "99")

	if !strings.Contains(stderr, "2 entries found") {
		t.Errorf("expected stderr to contain '2 entries found', got: %q", stderr)
	}
	if !strings.Contains(stdout, "March first entry") {
		t.Errorf("expected stdout to contain 'March first entry', got: %q", stdout)
	}
	if !strings.Contains(stdout, "Mid-month entry") {
		t.Errorf("expected stdout to contain 'Mid-month entry', got: %q", stdout)
	}
	if strings.Contains(stdout, "Late month") {
		t.Errorf("expected stdout NOT to contain 'Late month', got: %q", stdout)
	}
}

func TestOnFilter(t *testing.T) {
	env := newTestEnv(t)
	seedFilterJournal(t, env)

	stdout, stderr := run(t, env, "--on", "2026-03-15", "--num", "99")

	if !strings.Contains(stderr, "1 entries found") {
		t.Errorf("expected stderr to contain '1 entries found', got: %q", stderr)
	}
	if !strings.Contains(stdout, "Mid-month entry") {
		t.Errorf("expected stdout to contain 'Mid-month entry', got: %q", stdout)
	}
	if strings.Contains(stdout, "March first entry") {
		t.Errorf("expected stdout NOT to contain 'March first entry', got: %q", stdout)
	}
	if strings.Contains(stdout, "Late month") {
		t.Errorf("expected stdout NOT to contain 'Late month', got: %q", stdout)
	}
}

func TestFromToRange(t *testing.T) {
	env := newTestEnv(t)
	seedFilterJournal(t, env)

	stdout, stderr := run(t, env, "--from", "2026-03-10", "--to", "2026-03-20", "--num", "99")

	if !strings.Contains(stderr, "1 entries found") {
		t.Errorf("expected stderr to contain '1 entries found', got: %q", stderr)
	}
	if !strings.Contains(stdout, "Mid-month entry") {
		t.Errorf("expected stdout to contain 'Mid-month entry', got: %q", stdout)
	}
	if strings.Contains(stdout, "March first entry") {
		t.Errorf("expected stdout NOT to contain 'March first entry', got: %q", stdout)
	}
	if strings.Contains(stdout, "Late month") {
		t.Errorf("expected stdout NOT to contain 'Late month', got: %q", stdout)
	}
}

func TestContainsFilter(t *testing.T) {
	env := newTestEnv(t)
	seedFilterJournal(t, env)

	stdout, stderr := run(t, env, "--contains", "planning", "--num", "99")

	if !strings.Contains(stderr, "1 entries found") {
		t.Errorf("expected stderr to contain '1 entries found', got: %q", stderr)
	}
	if !strings.Contains(stdout, "Mid-month entry") {
		t.Errorf("expected stdout to contain 'Mid-month entry', got: %q", stdout)
	}
}

func TestContainsCaseInsensitive(t *testing.T) {
	env := newTestEnv(t)
	seedFilterJournal(t, env)

	stdout, stderr := run(t, env, "--contains", "PLANNING", "--num", "99")

	if !strings.Contains(stderr, "1 entries found") {
		t.Errorf("expected stderr to contain '1 entries found', got: %q", stderr)
	}
	if !strings.Contains(stdout, "Mid-month entry") {
		t.Errorf("expected stdout to contain 'Mid-month entry', got: %q", stdout)
	}
}

func TestTagPositionalFilter(t *testing.T) {
	env := newTestEnv(t)
	seedFilterJournal(t, env)

	stdout, stderr := run(t, env, "@work", "--num", "99")

	if !strings.Contains(stderr, "1 entries found") {
		t.Errorf("expected stderr to contain '1 entries found', got: %q", stderr)
	}
	if !strings.Contains(stdout, "March first entry") {
		t.Errorf("expected stdout to contain 'March first entry', got: %q", stdout)
	}
	if strings.Contains(stdout, "Mid-month entry") {
		t.Errorf("expected stdout NOT to contain 'Mid-month entry', got: %q", stdout)
	}
	if strings.Contains(stdout, "Late month") {
		t.Errorf("expected stdout NOT to contain 'Late month', got: %q", stdout)
	}
}

func TestTagPositionalMultiple(t *testing.T) {
	env := newTestEnv(t)
	seedFilterJournal(t, env)

	stdout, stderr := run(t, env, "@work", "@personal", "--num", "99")

	if !strings.Contains(stderr, "2 entries found") {
		t.Errorf("expected stderr to contain '2 entries found', got: %q", stderr)
	}
	if !strings.Contains(stdout, "March first entry") {
		t.Errorf("expected stdout to contain 'March first entry', got: %q", stdout)
	}
	if !strings.Contains(stdout, "Late month") {
		t.Errorf("expected stdout to contain 'Late month', got: %q", stdout)
	}
	if strings.Contains(stdout, "Mid-month entry") {
		t.Errorf("expected stdout NOT to contain 'Mid-month entry', got: %q", stdout)
	}
}

func TestCombinedFilters(t *testing.T) {
	env := newTestEnv(t)
	seedFilterJournal(t, env)

	stdout, stderr := run(t, env, "--from", "2026-03-10", "--contains", "entry", "--num", "99")

	if !strings.Contains(stderr, "2 entries found") {
		t.Errorf("expected stderr to contain '2 entries found', got: %q", stderr)
	}
	if !strings.Contains(stdout, "Mid-month entry") {
		t.Errorf("expected stdout to contain 'Mid-month entry', got: %q", stdout)
	}
	if !strings.Contains(stdout, "Late month") {
		t.Errorf("expected stdout to contain 'Late month', got: %q", stdout)
	}
	if strings.Contains(stdout, "March first entry") {
		t.Errorf("expected stdout NOT to contain 'March first entry', got: %q", stdout)
	}
}
