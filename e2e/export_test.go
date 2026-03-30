package e2e

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestExportJSON(t *testing.T) {
	env := newTestEnv(t)
	seedJournal(t, env)

	stdout, _ := run(t, env, "--export", "json", "--num", "99")

	var result struct {
		Entries []map[string]any `json:"entries"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %q", err, stdout)
	}

	if len(result.Entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(result.Entries))
	}

	for i, entry := range result.Entries {
		for _, key := range []string{"date", "time", "body", "starred"} {
			if _, ok := entry[key]; !ok {
				t.Errorf("entry %d missing key %q", i, key)
			}
		}
	}
}

func TestExportMarkdown(t *testing.T) {
	env := newTestEnv(t)
	seedJournal(t, env)

	stdout, _ := run(t, env, "--export", "md", "--num", "99")

	if !strings.Contains(stdout, "#") {
		t.Errorf("expected markdown output to contain '#', got: %q", stdout)
	}
	if !strings.Contains(stdout, "Morning entry") {
		t.Errorf("expected markdown output to contain 'Morning entry', got: %q", stdout)
	}
}

func TestExportText(t *testing.T) {
	env := newTestEnv(t)
	seedJournal(t, env)

	stdout, _ := run(t, env, "--export", "txt", "--num", "99")

	if !strings.Contains(stdout, "Morning entry") {
		t.Errorf("expected text output to contain 'Morning entry', got: %q", stdout)
	}
}

func TestExportXML(t *testing.T) {
	env := newTestEnv(t)
	seedJournal(t, env)

	stdout, _ := run(t, env, "--export", "xml", "--num", "99")

	if !strings.Contains(stdout, "<") || !strings.Contains(stdout, ">") {
		t.Errorf("expected XML output to contain XML markers, got: %q", stdout)
	}
	if !strings.Contains(stdout, "Morning entry") {
		t.Errorf("expected XML output to contain 'Morning entry', got: %q", stdout)
	}
}

func TestExportYAML(t *testing.T) {
	env := newTestEnv(t)
	seedJournal(t, env)

	stdout, _ := run(t, env, "--export", "yaml", "--num", "99")

	if !strings.Contains(stdout, "entries:") {
		t.Errorf("expected YAML output to contain 'entries:', got: %q", stdout)
	}
	if !strings.Contains(stdout, "Morning entry") {
		t.Errorf("expected YAML output to contain 'Morning entry', got: %q", stdout)
	}
}

func TestExportUnknownFormat(t *testing.T) {
	env := newTestEnv(t)
	seedJournal(t, env)

	_, _, err := runErr(t, env, "--export", "csv", "--num", "99")

	if err == nil {
		t.Error("expected error for unknown export format 'csv', got nil")
	}
}

func TestExportFormats(t *testing.T) {
	formats := []string{"json", "md", "markdown", "txt", "text", "xml", "yaml"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			env := newTestEnv(t)
			seedJournal(t, env)

			stdout, _, err := runErr(t, env, "--export", format, "--num", "99")

			if err != nil {
				t.Errorf("expected no error for format %q, got: %v", format, err)
			}
			if strings.TrimSpace(stdout) == "" {
				t.Errorf("expected non-empty output for format %q", format)
			}
		})
	}
}
