package e2e

import (
	"strings"
	"testing"
	"time"
)

func TestChangeTime(t *testing.T) {
	env := newTestEnv(t)
	seedFilterJournal(t, env)

	_, stderr := runWithStdin(t, env, "y\n", "--change-time=2026-03-20", "--on", "2026-03-15")

	if !strings.Contains(stderr, "1 entry") {
		t.Errorf("expected stderr to contain '1 entry', got: %q", stderr)
	}
	if !strings.Contains(stderr, "1 entry modified") {
		t.Errorf("expected stderr to contain '1 entry modified', got: %q", stderr)
	}

	target := time.Date(2026, 3, 20, 0, 0, 0, 0, time.Local)
	if !dayFileExists(t, env.journalDir, target) {
		t.Fatalf("expected day file for 2026-03-20 to exist")
	}

	content := dayFileContent(t, env.journalDir, target)
	if !strings.Contains(content, "Mid-month") {
		t.Errorf("expected day file for 2026-03-20 to contain 'Mid-month', got:\n%s", content)
	}
}

func TestChangeTimeDecline(t *testing.T) {
	env := newTestEnv(t)
	seedFilterJournal(t, env)

	_, stderr := runWithStdin(t, env, "n\n", "--change-time=2026-03-20", "--on", "2026-03-15")

	if strings.Contains(stderr, "modified") {
		t.Errorf("expected stderr NOT to contain 'modified', got: %q", stderr)
	}

	original := time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local)
	content := dayFileContent(t, env.journalDir, original)
	if !strings.Contains(content, "Mid-month") {
		t.Errorf("expected March 15 day file to still contain 'Mid-month', got:\n%s", content)
	}
}

func TestChangeTimeNoResults(t *testing.T) {
	env := newTestEnv(t)
	seedFilterJournal(t, env)

	_, stderr := run(t, env, "--change-time=2026-04-01", "--on", "2020-01-01")

	if !strings.Contains(stderr, "No entries to modify") {
		t.Errorf("expected stderr to contain 'No entries to modify', got: %q", stderr)
	}
}
