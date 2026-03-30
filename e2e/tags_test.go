package e2e

import (
	"strings"
	"testing"
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
