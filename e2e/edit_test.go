package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeMockEditor writes a bash script that replaces oldText with newText in
// the last argument (the temp file path). Returns the script path.
func writeMockEditor(t *testing.T, dir, oldText, newText string) string {
	t.Helper()
	script := fmt.Sprintf("#!/bin/bash\nFILE=\"${@: -1}\"\nsed -i 's/%s/%s/g' \"$FILE\"\n", oldText, newText)
	path := filepath.Join(dir, "mock-editor.sh")
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write mock editor: %v", err)
	}
	return path
}

// patchConfigEditor replaces the editor field in a config file.
func patchConfigEditor(t *testing.T, configPath, editorPath string) {
	t.Helper()
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	patched := strings.Replace(string(data), `editor = ""`, fmt.Sprintf(`editor = %q`, editorPath), 1)
	if !strings.Contains(patched, editorPath) {
		t.Fatalf("editor patch did not apply — check testConfigHeader")
	}
	if err := os.WriteFile(configPath, []byte(patched), 0644); err != nil {
		t.Fatalf("failed to write patched config: %v", err)
	}
}

func TestEditNoFilter(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	editorPath := writeMockEditor(t, env.dir, "First @work entry", "Edited work entry")
	patchConfigEditor(t, env.configPath, editorPath)

	_, stderr := run(t, env, "--edit", "--num", "99")

	if strings.Contains(stderr, "error") || strings.Contains(stderr, "Error") {
		t.Fatalf("unexpected error in stderr: %q", stderr)
	}

	// Edited entry must appear in the journal
	march1 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local)
	content := dayFileContent(t, env.journalDir, march1)
	if !strings.Contains(content, "Edited work entry") {
		t.Errorf("expected 'Edited work entry' in day file, got:\n%s", content)
	}
	if strings.Contains(content, "First @work entry") {
		t.Errorf("expected original text to be gone, got:\n%s", content)
	}
}

func TestEditWithTagFilter(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	editorPath := writeMockEditor(t, env.dir, "First @work entry", "Edited work entry")
	patchConfigEditor(t, env.configPath, editorPath)

	// Only edit @work entries; starred and @personal entries must be unchanged.
	_, stderr := run(t, env, "@work", "--edit")

	if strings.Contains(stderr, "error") || strings.Contains(stderr, "Error") {
		t.Fatalf("unexpected error in stderr: %q", stderr)
	}

	march1 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local)
	march15 := time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local)

	content1 := dayFileContent(t, env.journalDir, march1)
	if !strings.Contains(content1, "Edited work entry") {
		t.Errorf("expected @work entry to be edited, got:\n%s", content1)
	}
	if !strings.Contains(content1, "Starred afternoon entry") {
		t.Errorf("expected starred entry to be unchanged, got:\n%s", content1)
	}

	content15 := dayFileContent(t, env.journalDir, march15)
	if !strings.Contains(content15, "Mid-month @personal") {
		t.Errorf("expected @personal entry to be unchanged, got:\n%s", content15)
	}
}

func TestEditNoEntries(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)

	// No editor needed — should return early with a message
	_, stderr := run(t, env, "--edit", "--on", "2020-01-01")

	if !strings.Contains(stderr, "No entries to edit") {
		t.Errorf("expected 'No entries to edit' in stderr, got: %q", stderr)
	}
}

func TestEditNoEditor(t *testing.T) {
	env := newTestEnv(t)
	seedCompatJournal(t, env)
	// config already has editor = "" — do not patch it

	_, stderr, _ := runErr(t, env, "--edit", "--num", "99")

	if !strings.Contains(stderr, "no editor configured") {
		t.Errorf("expected 'no editor configured' in stderr, got: %q", stderr)
	}
}

func TestEditEmptyBufferAborts(t *testing.T) {
	env := newTestEnv(t)

	// Write a day file for today with a known entry
	today := time.Now()
	dayContent := fmt.Sprintf("# %s %s\n\n## [09:00 AM]\n\nFirst @work entry.\n",
		today.Format("2006-01-02"), today.Format("Monday"))
	writeDayFile(t, env.journalDir, today, dayContent)

	// Mock editor that empties the file
	script := "#!/bin/bash\nFILE=\"${@: -1}\"\necho -n '' > \"$FILE\"\n"
	editorPath := filepath.Join(env.dir, "empty-editor.sh")
	if err := os.WriteFile(editorPath, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	patchConfigEditor(t, env.configPath, editorPath)

	// Bare invocation (no args, no flags) opens today's day file directly (editDayFilePlain)
	_, stderr := run(t, env)

	if !strings.Contains(stderr, "no changes made") && !strings.Contains(stderr, "No entries found") {
		t.Errorf("expected abort message, got: %q", stderr)
	}

	// Verify entries are still intact (backup restored)
	content := dayFileContent(t, env.journalDir, today)
	if !strings.Contains(content, "First @work entry") {
		t.Errorf("entries should be preserved after empty buffer abort, got:\n%s", content)
	}
}

func TestEditDirectCleansUpEmptyHeading(t *testing.T) {
	env := newTestEnv(t)

	// Write a day file with one entry
	today := time.Now()
	dayContent := fmt.Sprintf("# %s %s\n\n## [09:00 AM]\n\nExisting entry.\n",
		today.Format("2006-01-02"), today.Format("Monday"))
	writeDayFile(t, env.journalDir, today, dayContent)

	// Mock editor that does nothing (leaves the appended empty heading as-is)
	script := "#!/bin/bash\n# no-op editor\n"
	editorPath := filepath.Join(env.dir, "noop-editor.sh")
	if err := os.WriteFile(editorPath, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	patchConfigEditor(t, env.configPath, editorPath)

	// Bare edit — should append a heading, then cleanup should strip it
	run(t, env)

	content := dayFileContent(t, env.journalDir, today)
	if !strings.Contains(content, "Existing entry.") {
		t.Errorf("existing entry should be preserved, got:\n%s", content)
	}

	// The appended empty ## heading should have been cleaned up
	headingCount := strings.Count(content, "## [")
	if headingCount != 1 {
		t.Errorf("expected 1 entry heading after cleanup, got %d in:\n%s", headingCount, content)
	}
}
