package e2e

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestInlineWrite(t *testing.T) {
	env := newTestEnv(t)
	today := time.Now()

	_, stderr := run(t, env, "Hello world.")

	if !strings.Contains(stderr, "Entry added") {
		t.Errorf("expected stderr to contain 'Entry added', got: %q", stderr)
	}

	if !dayFileExists(t, env.journalDir, today) {
		t.Fatal("expected day file to exist for today")
	}

	content := dayFileContent(t, env.journalDir, today)

	dateStr := today.Format("2006-01-02")
	if !strings.Contains(content, dateStr) {
		t.Errorf("expected day file to contain date %q, got:\n%s", dateStr, content)
	}

	weekday := today.Weekday().String()
	if !strings.Contains(content, weekday) {
		t.Errorf("expected day file to contain weekday %q, got:\n%s", weekday, content)
	}

	if !strings.Contains(content, "Hello world.") {
		t.Errorf("expected day file to contain 'Hello world.', got:\n%s", content)
	}

	if !strings.Contains(content, "## [") {
		t.Errorf("expected day file to contain '## [' (entry time heading), got:\n%s", content)
	}
}

func TestInlineWriteMultipleEntries(t *testing.T) {
	env := newTestEnv(t)
	today := time.Now()

	run(t, env, "First entry body.")
	run(t, env, "Second entry body.")

	content := dayFileContent(t, env.journalDir, today)

	if !strings.Contains(content, "First entry body.") {
		t.Errorf("expected day file to contain 'First entry body.', got:\n%s", content)
	}

	if !strings.Contains(content, "Second entry body.") {
		t.Errorf("expected day file to contain 'Second entry body.', got:\n%s", content)
	}

	count := strings.Count(content, "## [")
	if count < 2 {
		t.Errorf("expected at least 2 '## [' headings, found %d in:\n%s", count, content)
	}
}

func TestStarredEntryTrailing(t *testing.T) {
	env := newTestEnv(t)
	today := time.Now()

	run(t, env, "Important entry", "*")

	content := dayFileContent(t, env.journalDir, today)

	if !strings.Contains(content, "Important entry") {
		t.Errorf("expected day file to contain 'Important entry', got:\n%s", content)
	}

	lines := strings.Split(content, "\n")
	headingStarred := false
	for _, line := range lines {
		if strings.HasPrefix(line, "## [") && strings.Contains(line, "*") {
			headingStarred = true
			break
		}
	}
	if !headingStarred {
		t.Errorf("expected a '## [' heading line to contain '*' (starred entry), got:\n%s", content)
	}

	for _, line := range lines {
		if !strings.HasPrefix(line, "## [") && strings.TrimSpace(line) == "*" {
			t.Errorf("expected body NOT to contain a bare '*' line, got:\n%s", content)
			break
		}
	}
}

func TestStarredEntryLeading(t *testing.T) {
	env := newTestEnv(t)
	today := time.Now()

	run(t, env, "*", "Important entry")

	content := dayFileContent(t, env.journalDir, today)

	if !strings.Contains(content, "Important entry") {
		t.Errorf("expected day file to contain 'Important entry', got:\n%s", content)
	}

	lines := strings.Split(content, "\n")
	headingStarred := false
	for _, line := range lines {
		if strings.HasPrefix(line, "## [") && strings.Contains(line, "*") {
			headingStarred = true
			break
		}
	}
	if !headingStarred {
		t.Errorf("expected a '## [' heading line to contain '*' (starred entry), got:\n%s", content)
	}
}

func TestWriteDatePrefix(t *testing.T) {
	env := newTestEnv(t)
	yesterday := time.Now().AddDate(0, 0, -1)

	_, stderr := run(t, env, "yesterday: Prefixed entry text.")

	if !strings.Contains(stderr, "Entry added") {
		t.Errorf("expected 'Entry added' in stderr, got: %q", stderr)
	}
	if !dayFileExists(t, env.journalDir, yesterday) {
		t.Fatal("expected day file for yesterday")
	}
	content := dayFileContent(t, env.journalDir, yesterday)
	if !strings.Contains(content, "Prefixed entry text.") {
		t.Errorf("expected entry body in yesterday's day file, got:\n%s", content)
	}
	today := time.Now()
	if dayFileExists(t, env.journalDir, today) {
		t.Error("entry should be in yesterday's file, not today's")
	}
}

func TestWriteDatePrefixExplicit(t *testing.T) {
	env := newTestEnv(t)
	target := time.Date(2025, 1, 15, 0, 0, 0, 0, time.Local)

	_, stderr := run(t, env, "2025-01-15: Historical entry.")

	if !strings.Contains(stderr, "Entry added") {
		t.Errorf("expected 'Entry added' in stderr, got: %q", stderr)
	}
	if !dayFileExists(t, env.journalDir, target) {
		t.Fatal("expected day file for 2025-01-15")
	}
	content := dayFileContent(t, env.journalDir, target)
	if !strings.Contains(content, "Historical entry.") {
		t.Errorf("expected entry body in 2025-01-15 day file, got:\n%s", content)
	}
}

func TestWriteNoDatePrefix(t *testing.T) {
	// "foo: bar" where "foo" is not a parseable date should be treated as plain body text.
	env := newTestEnv(t)
	today := time.Now()

	_, stderr := run(t, env, "foo: bar")

	if !strings.Contains(stderr, "Entry added") {
		t.Errorf("expected 'Entry added' in stderr, got: %q", stderr)
	}
	content := dayFileContent(t, env.journalDir, today)
	if !strings.Contains(content, "foo: bar") {
		t.Errorf("expected full 'foo: bar' body in today's day file, got:\n%s", content)
	}
}

func TestWriteFromStdin(t *testing.T) {
	env := newTestEnv(t)
	today := time.Now()

	_, stderr := runWithStdin(t, env, "Stdin entry body.\n")

	if !strings.Contains(stderr, "Entry added") {
		t.Errorf("expected 'Entry added' in stderr, got: %q", stderr)
	}
	if !dayFileExists(t, env.journalDir, today) {
		t.Fatal("expected day file for today")
	}
	content := dayFileContent(t, env.journalDir, today)
	if !strings.Contains(content, "Stdin entry body.") {
		t.Errorf("expected stdin body in day file, got:\n%s", content)
	}
}

func TestConfigFileFlag(t *testing.T) {
	env := newTestEnv(t)
	today := time.Now()

	// Use --config-file explicitly (not the hidden --config alias)
	cmd := exec.Command(binary, "--config-file", env.configPath, "Config-file flag entry.")
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		t.Fatalf("binary exited non-zero: %v\nstderr: %s", err, errBuf.String())
	}

	if !strings.Contains(errBuf.String(), "Entry added") {
		t.Errorf("expected 'Entry added' in stderr, got: %q", errBuf.String())
	}
	if !dayFileExists(t, env.journalDir, today) {
		t.Fatal("expected day file for today")
	}
}
