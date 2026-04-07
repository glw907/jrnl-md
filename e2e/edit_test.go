package e2e

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func editConfig(journalDir string, timestamps bool) string {
	ts := "false"
	if timestamps {
		ts = "true"
	}
	return `[general]
editor = "true"
timestamps = ` + ts + `
linewrap = 79
default_list_count = 10

[format]
time = "03:04 PM"
date = "2006-01-02"
tag_symbols = "@"

[colors]
date = "none"
body = "none"
tags = "none"

[journals.default]
path = "` + journalDir + `"
`
}

func TestEditCreatesFileWithTimestamp(t *testing.T) {
	env := newTestEnv(t)
	if err := os.WriteFile(env.configPath, []byte(editConfig(env.journalDir, true)), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := env.run(t, "edit"); err != nil {
		t.Fatalf("edit: %v", err)
	}

	today := time.Now()
	date := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	content := env.readDayFile(t, date)

	if !strings.HasPrefix(content, "# ") {
		t.Errorf("edit: file should start with day heading:\n%s", content)
	}
	if !strings.Contains(content, "## ") {
		t.Errorf("edit: file should contain timestamp heading:\n%s", content)
	}
	// Blank line between headings, plus trailing blank line for cursor entry
	lines := strings.Split(content, "\n")
	if len(lines) < 5 {
		t.Fatalf("edit: expected at least 5 lines, got %d:\n%s", len(lines), content)
	}
	if lines[1] != "" {
		t.Errorf("edit: expected blank line after day heading, got %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "## ") {
		t.Errorf("edit: expected timestamp heading on line 3, got %q", lines[2])
	}
	if lines[3] != "" {
		t.Errorf("edit: expected blank separator after timestamp heading, got %q", lines[3])
	}
	if lines[4] != "" {
		t.Errorf("edit: expected blank cursor entry line, got %q", lines[4])
	}
}

func TestEditCreatesFileWithoutTimestamp(t *testing.T) {
	env := newTestEnv(t)
	if err := os.WriteFile(env.configPath, []byte(editConfig(env.journalDir, false)), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := env.run(t, "edit"); err != nil {
		t.Fatalf("edit: %v", err)
	}

	today := time.Now()
	date := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	content := env.readDayFile(t, date)

	if !strings.HasPrefix(content, "# ") {
		t.Errorf("edit: file should start with day heading:\n%s", content)
	}
	if strings.Contains(content, "## ") {
		t.Errorf("edit: file should not contain timestamp heading:\n%s", content)
	}
	// Blank line after heading, plus trailing blank line for cursor entry
	lines := strings.Split(content, "\n")
	if len(lines) < 3 {
		t.Fatalf("edit: expected at least 3 lines, got %d:\n%s", len(lines), content)
	}
	if lines[1] != "" {
		t.Errorf("edit: expected blank separator after day heading, got %q", lines[1])
	}
	if lines[2] != "" {
		t.Errorf("edit: expected blank cursor entry line, got %q", lines[2])
	}
}

func TestEditOnDate(t *testing.T) {
	env := newTestEnv(t)
	if err := os.WriteFile(env.configPath, []byte(editConfig(env.journalDir, true)), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := env.run(t, "edit", "--on", "2026-04-06"); err != nil {
		t.Fatalf("edit --on: %v", err)
	}

	date := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	content := env.readDayFile(t, date)
	if !strings.Contains(content, "2026-04-06") {
		t.Errorf("edit --on: file missing date heading:\n%s", content)
	}
}

func TestEditExistingFilePreservesContent(t *testing.T) {
	env := newTestEnv(t)
	cfg := editConfig(env.journalDir, true)
	if err := os.WriteFile(env.configPath, []byte(cfg), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	date := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	env.writeDayFile(t, date, "# 2026-04-06 Monday\n\n## 09:00 AM\n\nMorning entry.\n")

	if _, err := env.run(t, "edit", "--on", "2026-04-06"); err != nil {
		t.Fatalf("edit: %v", err)
	}

	content := env.readDayFile(t, date)
	if !strings.Contains(content, "Morning entry.") {
		t.Errorf("edit should preserve existing content:\n%s", content)
	}
}

// TestEditCursorPosition verifies the invariant for every edit scenario:
// after edit prepares the file, the cursor lands on a blank line with a
// blank paragraph separator above it.
func TestEditCursorPosition(t *testing.T) {
	cases := []struct {
		name       string
		timestamps bool
		existing   string // pre-create file with this content if non-empty
		editArgs   []string
		date       time.Time
	}{
		{
			name:       "new file with timestamps",
			timestamps: true,
			editArgs:   []string{"edit"},
			date:       time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "new file without timestamps",
			timestamps: false,
			editArgs:   []string{"edit"},
			date:       time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "existing file with content",
			timestamps: true,
			existing:   "# 2026-03-15 Sunday\n\n## 09:00 AM\n\nMorning entry.\n",
			editArgs:   []string{"edit", "--on", "2026-03-15"},
			date:       time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "existing file multiple entries",
			timestamps: true,
			existing:   "# 2026-03-15 Sunday\n\n## 09:00 AM\n\nMorning.\n\n## 02:00 PM\n\nAfternoon.\n",
			editArgs:   []string{"edit", "--on", "2026-03-15"},
			date:       time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "existing file no timestamps",
			timestamps: false,
			existing:   "# 2026-03-15 Sunday\n\nSome thoughts.\n",
			editArgs:   []string{"edit", "--on", "2026-03-15"},
			date:       time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "existing file already has trailing blank lines",
			timestamps: true,
			existing:   "# 2026-03-15 Sunday\n\n## 09:00 AM\n\nContent.\n\n",
			editArgs:   []string{"edit", "--on", "2026-03-15"},
			date:       time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			env := newTestEnv(t)
			if err := os.WriteFile(env.configPath, []byte(editConfig(env.journalDir, tt.timestamps)), 0644); err != nil {
				t.Fatalf("WriteFile config: %v", err)
			}
			if tt.existing != "" {
				env.writeDayFile(t, tt.date, tt.existing)
			}

			if _, err := env.run(t, tt.editArgs...); err != nil {
				t.Fatalf("edit: %v", err)
			}

			content := env.readDayFile(t, tt.date)
			lines := strings.Split(content, "\n")

			// Replicate cursor line calculation from editor.Open
			cursorLine := strings.Count(content, "\n")
			if cursorLine == 0 {
				cursorLine = 1
			}

			idx := cursorLine - 1
			if idx >= len(lines) {
				t.Fatalf("cursor line %d is past end of file (%d lines)\ncontent:\n%s",
					cursorLine, len(lines), content)
			}
			if lines[idx] != "" {
				t.Errorf("cursor line %d should be empty, got %q\ncontent:\n%s",
					cursorLine, lines[idx], content)
			}
			if idx < 1 {
				t.Fatalf("cursor line %d has no line above it", cursorLine)
			}
			if lines[idx-1] != "" {
				t.Errorf("line %d above cursor should be blank paragraph separator, got %q\ncontent:\n%s",
					cursorLine-1, lines[idx-1], content)
			}
		})
	}
}

func TestEditNoEditor(t *testing.T) {
	env := newTestEnv(t)
	cfg := strings.Replace(editConfig(env.journalDir, true), `editor = "true"`, `editor = ""`, 1)
	if err := os.WriteFile(env.configPath, []byte(cfg), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cmd := exec.Command(binaryPath, "--config-file", env.configPath, "edit")
	cmd.Env = []string{"HOME=" + t.TempDir(), "PATH=" + os.Getenv("PATH")}
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("edit with no editor should fail")
	}
	if !strings.Contains(string(out), "no editor configured") {
		t.Errorf("expected 'no editor configured' error, got: %s", out)
	}
}
