package e2e

import (
	"strings"
	"testing"
	"time"
)

func TestTagsListing(t *testing.T) {
	env := newTestEnv(t)
	seedFilterJournal(t, env)

	stdout, _ := run(t, env, "--tags", "--num", "99")

	for _, tag := range []string{"@work", "@project", "@personal"} {
		if !strings.Contains(stdout, tag) {
			t.Errorf("expected stdout to contain %q, got: %q", tag, stdout)
		}
	}
}

func TestTagsWithCount(t *testing.T) {
	env := newTestEnv(t)
	seedFilterJournal(t, env)

	stdout, _ := run(t, env, "--tags", "--num", "99")

	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		if !strings.Contains(line, ": ") {
			t.Errorf("expected output line to contain ': ' (tag : count format), got: %q", line)
		}
	}
}

func TestTagsEmpty(t *testing.T) {
	env := newTestEnv(t)

	stdout, _ := run(t, env, "--tags", "--num", "99")

	if !strings.Contains(stdout, "No tags found") {
		t.Errorf("expected stdout to contain 'No tags found', got: %q", stdout)
	}
}

func TestTagsFrequencySort(t *testing.T) {
	env := newTestEnv(t)
	// @zebra appears 3 times (least alphabetical, most frequent)
	// @alpha appears 1 time (most alphabetical, least frequent)
	// This would fail with alphabetical sort but pass with frequency sort
	writeDayFile(t, env.journalDir, time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local),
		"# 2026-03-01 Sunday\n\n## [09:00 AM]\n\nEntry with @zebra.\n\n## [10:00 AM]\n\nEntry with @zebra and @alpha.\n\n## [11:00 AM]\n\nEntry with @zebra only.\n")

	stdout, _ := run(t, env, "--tags", "--num", "99")

	posZebra := strings.Index(stdout, "@zebra")
	posAlpha := strings.Index(stdout, "@alpha")

	if posZebra == -1 || posAlpha == -1 {
		t.Fatalf("expected both tags in output, got: %s", stdout)
	}
	// @zebra (3 occurrences) must appear before @alpha (1 occurrence)
	// This fails with alphabetical sort since 'a' < 'z'
	if posZebra > posAlpha {
		t.Errorf("@zebra (3 occurrences) should appear before @alpha (1 occurrence), but got:\n%s", stdout)
	}
}
