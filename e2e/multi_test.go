package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestListJournals(t *testing.T) {
	env := newTestEnv(t)
	stdout, _ := run(t, env, "--list")
	if !strings.Contains(stdout, "default") {
		t.Errorf("expected stdout to contain %q, got: %s", "default", stdout)
	}
}

func TestListMultipleJournals(t *testing.T) {
	dir := t.TempDir()
	env := newMultiTestEnv(t, map[string]string{
		"default":  filepath.Join(dir, "default"),
		"work":     filepath.Join(dir, "work"),
		"personal": filepath.Join(dir, "personal"),
	})
	stdout, _ := run(t, env, "--list")
	for _, name := range []string{"default", "work", "personal"} {
		if !strings.Contains(stdout, name) {
			t.Errorf("expected stdout to contain %q, got: %s", name, stdout)
		}
	}
}

func TestMultiJournalPrefix(t *testing.T) {
	dir := t.TempDir()
	defaultDir := filepath.Join(dir, "default")
	workDir := filepath.Join(dir, "work")
	env := newMultiTestEnv(t, map[string]string{
		"default": defaultDir,
		"work":    workDir,
	})

	_, stderr := run(t, env, "work:", "Meeting notes for today.")

	if !strings.Contains(stderr, "Entry added") {
		t.Errorf("expected stderr to contain %q, got: %s", "Entry added", stderr)
	}

	today := time.Now()
	if !dayFileExists(t, workDir, today) {
		t.Errorf("expected day file to exist in work dir")
	}
	if dayFileExists(t, defaultDir, today) {
		t.Errorf("expected day file to NOT exist in default dir")
	}

	content := dayFileContent(t, workDir, today)
	if !strings.Contains(content, "Meeting notes") {
		t.Errorf("expected work day file to contain %q, got: %s", "Meeting notes", content)
	}
}

func TestMultiJournalReadIsolation(t *testing.T) {
	dir := t.TempDir()
	defaultDir := filepath.Join(dir, "default")
	workDir := filepath.Join(dir, "work")
	env := newMultiTestEnv(t, map[string]string{
		"default": defaultDir,
		"work":    workDir,
	})

	run(t, env, "Default journal entry.")
	run(t, env, "work:", "Work journal entry.")

	defaultOut, _ := run(t, env, "--num", "99")
	if !strings.Contains(defaultOut, "Default journal entry") {
		t.Errorf("expected default read to contain default entry, got: %s", defaultOut)
	}
	if strings.Contains(defaultOut, "Work journal entry") {
		t.Errorf("expected default read to NOT contain work entry, got: %s", defaultOut)
	}

	workOut, _ := run(t, env, "work:", "--num", "99")
	if !strings.Contains(workOut, "Work journal entry") {
		t.Errorf("expected work read to contain work entry, got: %s", workOut)
	}
	if strings.Contains(workOut, "Default journal entry") {
		t.Errorf("expected work read to NOT contain default entry, got: %s", workOut)
	}
}

func TestUnknownJournalPrefix(t *testing.T) {
	env := newTestEnv(t)
	_, stderr := run(t, env, "nonexistent:", "some text")
	if !strings.Contains(stderr, "Entry added") {
		t.Errorf("unrecognized journal prefix should be treated as entry text, stderr: %q", stderr)
	}

	today := time.Now()
	content := dayFileContent(t, env.journalDir, today)
	if !strings.Contains(content, "nonexistent:") {
		t.Errorf("entry body should contain the prefix as text: %s", content)
	}
}

func TestCustomConfigPath(t *testing.T) {
	env := newTestEnv(t)
	_, stderr := run(t, env, "Custom config test entry.")
	if !strings.Contains(stderr, "Entry added") {
		t.Errorf("expected stderr to contain %q, got: %s", "Entry added", stderr)
	}
	today := time.Now()
	if !dayFileExists(t, env.journalDir, today) {
		t.Errorf("expected day file to exist in journal dir via custom config")
	}
	content := dayFileContent(t, env.journalDir, today)
	if !strings.Contains(content, "Custom config test entry") {
		t.Errorf("expected day file to contain entry text, got: %s", content)
	}
}

func TestConfigDefaultsCreated(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	if err := exec.Command(binary, "--config", configPath, "--list").Run(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("config should be created: %v", err)
	}
	content := string(data)

	checks := []struct {
		key  string
		want string
	}{
		{"default_hour", "9"},
		{"default_minute", "0"},
		{"highlight", "true"},
		{"linewrap", "79"},
		{"tag_symbols", `"@"`},
		{"file_extension", `"md"`},
		{"indent_character", `"|"`},
	}

	for _, c := range checks {
		if !strings.Contains(content, c.key) {
			t.Errorf("default config missing key %q", c.key)
		} else if !strings.Contains(content, c.want) {
			t.Errorf("default for %q should contain %q in config", c.key, c.want)
		}
	}
}
