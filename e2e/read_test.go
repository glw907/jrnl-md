package e2e

import (
	"strings"
	"testing"
	"time"
)

func seedJournal(t *testing.T, env testEnv) {
	t.Helper()
	day1 := time.Date(2026, 3, 28, 0, 0, 0, 0, time.Local)
	day2 := time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local)

	writeDayFile(t, env.journalDir, day1, "# 2026-03-28 Saturday\n\n## [09:00 AM]\n\nMorning entry with @work tag.\n\n## [02:00 PM] *\n\nStarred afternoon entry.\n")
	writeDayFile(t, env.journalDir, day2, "# 2026-03-29 Sunday\n\n## [10:00 AM]\n\nSunday @personal reflection about @life.\n")
}

func TestReadAllEntries(t *testing.T) {
	env := newTestEnv(t)
	seedJournal(t, env)

	stdout, stderr := run(t, env, "--num", "99")

	if !strings.Contains(stderr, "3 entries found") {
		t.Errorf("expected stderr to contain '3 entries found', got: %q", stderr)
	}
	for _, want := range []string{"Morning entry with @work tag", "Starred afternoon entry", "Sunday @personal reflection"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("expected stdout to contain %q, got: %q", want, stdout)
		}
	}
}

func TestNumFlag(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		count   int
		has     []string
		missing []string
	}{
		{
			name:    "--num 1",
			args:    []string{"--num", "1"},
			has:     []string{"Sunday"},
			missing: []string{"Morning entry"},
		},
		{
			name:    "-n 2",
			args:    []string{"-n", "2"},
			has:     []string{"Starred afternoon"},
			missing: []string{"Morning entry"},
		},
		{
			name:    "-1 shorthand",
			args:    []string{"-1"},
			has:     []string{"Sunday"},
			missing: []string{"Morning entry"},
		},
		{
			name:    "-2 shorthand",
			args:    []string{"-2"},
			has:     []string{"Starred afternoon"},
			missing: []string{"Morning entry"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestEnv(t)
			seedJournal(t, env)

			stdout, _ := run(t, env, tc.args...)

			for _, want := range tc.has {
				if !strings.Contains(stdout, want) {
					t.Errorf("expected stdout to contain %q, got: %q", want, stdout)
				}
			}
			for _, notWant := range tc.missing {
				if strings.Contains(stdout, notWant) {
					t.Errorf("expected stdout NOT to contain %q, got: %q", notWant, stdout)
				}
			}
		})
	}
}

func TestShortFlag(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"--short --num 99", []string{"--short", "--num", "99"}},
		{"-s --num 99", []string{"-s", "--num", "99"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestEnv(t)
			seedJournal(t, env)

			stdout, _ := run(t, env, tc.args...)

			lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
			if len(lines) != 3 {
				t.Errorf("expected exactly 3 output lines, got %d: %q", len(lines), stdout)
			}
			for _, line := range lines {
				if !strings.Contains(line, "2026") {
					t.Errorf("expected each line to contain a date, got: %q", line)
				}
			}
		})
	}
}

func TestStarredFilter(t *testing.T) {
	env := newTestEnv(t)
	seedJournal(t, env)

	stdout, stderr := run(t, env, "--starred", "--num", "99")

	if !strings.Contains(stderr, "1 entries found") {
		t.Errorf("expected stderr to contain '1 entries found', got: %q", stderr)
	}
	if !strings.Contains(stdout, "Starred afternoon") {
		t.Errorf("expected stdout to contain 'Starred afternoon', got: %q", stdout)
	}
	if strings.Contains(stdout, "Morning entry") {
		t.Errorf("expected stdout NOT to contain 'Morning entry', got: %q", stdout)
	}
}
