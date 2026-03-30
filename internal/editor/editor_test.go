package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPrepareDayFileNew(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "29.md")

	date := time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local)
	lineCount, err := PrepareDayFile(path, date, "2006-01-02", "03:04 PM", "")
	if err != nil {
		t.Fatalf("PrepareDayFile failed: %v", err)
	}

	if lineCount < 4 {
		t.Errorf("expected at least 4 lines, got %d", lineCount)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.HasPrefix(content, "# 2026-03-29 Sunday") {
		t.Errorf("missing day title, got: %q", content[:40])
	}
	if !strings.Contains(content, "## [02:30 PM]") {
		t.Errorf("missing entry heading, got: %q", content)
	}
}

func TestPrepareDayFileExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "29.md")

	existing := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nMorning entry.\n"
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	date := time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local)
	lineCount, err := PrepareDayFile(path, date, "2006-01-02", "03:04 PM", "")
	if err != nil {
		t.Fatalf("PrepareDayFile failed: %v", err)
	}

	if lineCount < 6 {
		t.Errorf("expected at least 6 lines, got %d", lineCount)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "Morning entry.") {
		t.Error("lost existing content")
	}
	if !strings.Contains(content, "## [02:30 PM]") {
		t.Error("missing new entry heading")
	}
	if strings.Count(content, "# 2026-03-29") != 1 {
		t.Error("should have exactly one day title")
	}
}

func TestEditorCommand(t *testing.T) {
	cmd, args := editorArgs("hx", "/tmp/test.md", 10)
	if cmd != "hx" {
		t.Errorf("cmd = %q", cmd)
	}
	if len(args) != 2 || args[0] != "+10" || args[1] != "/tmp/test.md" {
		t.Errorf("args = %v", args)
	}
}

func TestEditorCommandVim(t *testing.T) {
	cmd, args := editorArgs("vim", "/tmp/test.md", 10)
	if cmd != "vim" {
		t.Errorf("cmd = %q", cmd)
	}
	if len(args) != 2 || args[0] != "+10" || args[1] != "/tmp/test.md" {
		t.Errorf("args = %v", args)
	}
}

func TestEditorCommandVSCode(t *testing.T) {
	cmd, args := editorArgs("code", "/tmp/test.md", 10)
	if cmd != "code" {
		t.Errorf("cmd = %q", cmd)
	}
	found := false
	for _, a := range args {
		if a == "--goto" || strings.Contains(a, ":10") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected --goto with line number, got args = %v", args)
	}
}

func TestPrepareDayFileWithTemplate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "day.md")
	date := time.Date(2026, 3, 29, 17, 13, 0, 0, time.Local)

	tests := []struct {
		name     string
		template string
		check    func(t *testing.T, content string)
	}{
		{
			name:     "no template",
			template: "",
			check: func(t *testing.T, content string) {
				if strings.Count(content, "##") != 1 {
					t.Errorf("expected 1 entry heading, content:\n%s", content)
				}
			},
		},
		{
			name:     "simple template",
			template: "## Mood\n\n## Gratitude\n",
			check: func(t *testing.T, content string) {
				if !strings.Contains(content, "## Mood") {
					t.Error("missing template Mood heading")
				}
				if !strings.Contains(content, "## Gratitude") {
					t.Error("missing template Gratitude heading")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Remove(path)

			_, err := PrepareDayFile(path, date, "2006-01-02", "03:04 PM", tt.template)
			if err != nil {
				t.Fatalf("PrepareDayFile() error: %v", err)
			}

			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading file: %v", err)
			}
			tt.check(t, string(data))
		})
	}
}
