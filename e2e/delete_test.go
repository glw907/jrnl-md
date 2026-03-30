package e2e

import (
	"strings"
	"testing"
	"time"
)

func TestDeleteConfirmYes(t *testing.T) {
	env := newTestEnv(t)
	seedFilterJournal(t, env)

	_, stderr := runWithStdin(t, env, "y\n", "--delete", "--on", "2026-03-15")

	if !strings.Contains(stderr, "1 entry") {
		t.Errorf("expected stderr to contain '1 entry', got: %q", stderr)
	}
	if !strings.Contains(stderr, "1 entry deleted") {
		t.Errorf("expected stderr to contain '1 entry deleted', got: %q", stderr)
	}

	march15 := time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local)
	if dayFileExists(t, env.journalDir, march15) {
		content := dayFileContent(t, env.journalDir, march15)
		if strings.Contains(content, "Mid-month") {
			t.Errorf("expected March 15 day file NOT to contain 'Mid-month' after deletion, got:\n%s", content)
		}
	}
}

func TestDeleteConfirmNo(t *testing.T) {
	env := newTestEnv(t)
	seedFilterJournal(t, env)

	_, stderr := runWithStdin(t, env, "n\n", "--delete", "--on", "2026-03-15")

	if !strings.Contains(stderr, "1 entry") {
		t.Errorf("expected stderr to contain '1 entry', got: %q", stderr)
	}
	if strings.Contains(stderr, "deleted") {
		t.Errorf("expected stderr NOT to contain 'deleted', got: %q", stderr)
	}

	march15 := time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local)
	content := dayFileContent(t, env.journalDir, march15)
	if !strings.Contains(content, "Mid-month") {
		t.Errorf("expected March 15 day file to still contain 'Mid-month', got:\n%s", content)
	}
}

func TestDeleteNoResults(t *testing.T) {
	env := newTestEnv(t)
	seedFilterJournal(t, env)

	_, stderr := run(t, env, "--delete", "--on", "2020-01-01")

	if !strings.Contains(stderr, "No entries to delete") {
		t.Errorf("expected stderr to contain 'No entries to delete', got: %q", stderr)
	}
}
