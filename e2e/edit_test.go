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
