package e2e

import (
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
