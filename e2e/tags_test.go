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
	// @common appears 3 times, @medium 2 times, @rare 1 time
	writeDayFile(t, env.journalDir, time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local),
		"# 2026-03-01 Sunday\n\n## [09:00 AM]\n\nEntry with @common and @rare.\n\n## [10:00 AM]\n\nEntry with @common and @medium.\n\n## [11:00 AM]\n\nEntry with @common only.\n\n## [12:00 PM]\n\nEntry with @medium only.\n")

	stdout, _ := run(t, env, "--tags", "--num", "99")

	posCommon := strings.Index(stdout, "@common")
	posMedium := strings.Index(stdout, "@medium")
	posRare := strings.Index(stdout, "@rare")

	if posCommon == -1 || posMedium == -1 || posRare == -1 {
		t.Fatalf("expected all three tags in output, got: %s", stdout)
	}
	if posCommon > posMedium {
		t.Errorf("@common (3 occurrences) should appear before @medium (2 occurrences)")
	}
	if posMedium > posRare {
		t.Errorf("@medium (2 occurrences) should appear before @rare (1 occurrence)")
	}
}
