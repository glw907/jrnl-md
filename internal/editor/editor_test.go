package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIsEmptyContent(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"", true},
		{"   \n\n  ", true},
		{"\t\n", true},
		{"# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nEntry.\n", false},
		{"some text", false},
	}
	for _, tt := range tests {
		got := IsEmptyContent(tt.input)
		if got != tt.want {
			t.Errorf("IsEmptyContent(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestPrepareEncryptedNew(t *testing.T) {
	date := time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local)
	cfg := Config{DateFmt: "2006-01-02", TimeFmt: "03:04 PM"}
	content, cursorLine := PrepareEncryptedContent("", date, cfg)

	if !strings.HasPrefix(content, "# 2026-03-29 Sunday") {
		t.Errorf("missing day heading, got: %q", content[:40])
	}
	if !strings.Contains(content, "## [02:30 PM]") {
		t.Errorf("missing entry heading")
	}
	// Cursor should land on the blank line after the entry heading.
	lines := strings.Split(content, "\n")
	if cursorLine < 1 || cursorLine > len(lines) {
		t.Fatalf("cursorLine %d out of range (content has %d lines)", cursorLine, len(lines))
	}
	if lines[cursorLine-1] != "" {
		t.Errorf("cursor line %d should be blank, got %q", cursorLine, lines[cursorLine-1])
	}
	if cursorLine < 2 || !strings.HasPrefix(lines[cursorLine-2], "## [") {
		t.Errorf("line before cursor should be entry heading, got %q", lines[cursorLine-2])
	}
}

func TestPrepareEncryptedExisting(t *testing.T) {
	existing := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nMorning entry.\n"
	date := time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local)
	cfg := Config{DateFmt: "2006-01-02", TimeFmt: "03:04 PM"}
	content, _ := PrepareEncryptedContent(existing, date, cfg)

	if !strings.Contains(content, "Morning entry.") {
		t.Error("lost existing content")
	}
	if !strings.Contains(content, "## [02:30 PM]") {
		t.Error("missing new entry heading")
	}
}

func TestPrepareEncryptedWithTemplate(t *testing.T) {
	date := time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local)
	cfg := Config{DateFmt: "2006-01-02", TimeFmt: "03:04 PM", Template: "## Mood\n"}
	content, _ := PrepareEncryptedContent("", date, cfg)

	if !strings.Contains(content, "## Mood") {
		t.Error("missing template content")
	}
}

func TestPrepareDayFileNew(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "29.md")

	date := time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local)
	cfg := Config{DateFmt: "2006-01-02", TimeFmt: "03:04 PM"}
	cursorLine, err := PrepareDayFile(path, date, cfg)
	if err != nil {
		t.Fatalf("PrepareDayFile failed: %v", err)
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

	// Cursor should land on the blank line after the entry heading.
	lines := strings.Split(content, "\n")
	if cursorLine < 1 || cursorLine > len(lines) {
		t.Fatalf("cursorLine %d out of range (content has %d lines)", cursorLine, len(lines))
	}
	if lines[cursorLine-1] != "" {
		t.Errorf("cursor line %d should be blank, got %q", cursorLine, lines[cursorLine-1])
	}
	if cursorLine < 2 || !strings.HasPrefix(lines[cursorLine-2], "## [") {
		t.Errorf("line before cursor should be entry heading, got %q", lines[cursorLine-2])
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
	cfg := Config{DateFmt: "2006-01-02", TimeFmt: "03:04 PM"}
	cursorLine, err := PrepareDayFile(path, date, cfg)
	if err != nil {
		t.Fatalf("PrepareDayFile failed: %v", err)
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

	// Cursor should land on the blank line after the new entry heading.
	lines := strings.Split(content, "\n")
	if cursorLine < 1 || cursorLine > len(lines) {
		t.Fatalf("cursorLine %d out of range (content has %d lines)", cursorLine, len(lines))
	}
	if lines[cursorLine-1] != "" {
		t.Errorf("cursor line %d should be blank, got %q", cursorLine, lines[cursorLine-1])
	}
	if cursorLine < 2 || !strings.HasPrefix(lines[cursorLine-2], "## [") {
		t.Errorf("line before cursor should be entry heading, got %q", lines[cursorLine-2])
	}
}

func TestEndOfContent(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "empty body with blank line after heading",
			input: "# 2026-04-01 Wednesday\n\n## [09:00 AM]\n\n",
			want:  5, // line after blank separator, where body content goes
		},
		{
			name:  "empty body with two blank lines after heading",
			input: "# 2026-04-01 Wednesday\n\n## [09:00 AM]\n\n\n",
			want:  5, // line after blank separator
		},
		{
			name:  "entry with body",
			input: "# 2026-04-01 Wednesday\n\n## [09:00 AM]\n\nSome text.\n",
			want:  countLines("# 2026-04-01 Wednesday\n\n## [09:00 AM]\n\nSome text.\n"),
		},
		{
			name:  "two entries, second empty",
			input: "# 2026-04-01 Wednesday\n\n## [09:00 AM]\n\nMorning.\n\n## [02:00 PM]\n\n",
			want:  9, // line after blank separator following second heading
		},
		{
			name:  "two entries both with body",
			input: "# 2026-04-01 Wednesday\n\n## [09:00 AM]\n\nMorning.\n\n## [02:00 PM]\n\nAfternoon.\n",
			want:  countLines("# 2026-04-01 Wednesday\n\n## [09:00 AM]\n\nMorning.\n\n## [02:00 PM]\n\nAfternoon.\n"),
		},
		{
			name:  "empty string",
			input: "",
			want:  1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EndOfContent(tt.input)
			if got != tt.want {
				t.Errorf("EndOfContent() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestEnsureBlankLineAfterLastHeading(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "heading without trailing blank line",
			input: "# 2026-04-01 Wednesday\n\n## [09:00 AM]\n",
			want:  "# 2026-04-01 Wednesday\n\n## [09:00 AM]\n\n\n",
		},
		{
			name:  "heading with one trailing blank line",
			input: "# 2026-04-01 Wednesday\n\n## [09:00 AM]\n\n",
			want:  "# 2026-04-01 Wednesday\n\n## [09:00 AM]\n\n\n",
		},
		{
			name:  "entry with body unchanged",
			input: "# 2026-04-01 Wednesday\n\n## [09:00 AM]\n\nSome text.\n",
			want:  "# 2026-04-01 Wednesday\n\n## [09:00 AM]\n\nSome text.\n",
		},
		{
			name:  "empty string unchanged",
			input: "",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EnsureBlankLineAfterLastHeading(tt.input)
			if got != tt.want {
				t.Errorf("EnsureBlankLineAfterLastHeading() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestEndOfContentRequiresBlankLine verifies that EndOfContent positions the
// cursor correctly only after EnsureBlankLineAfterLastHeading has been called.
// This is the exact scenario from the desktop launcher: jrnl-md "" creates a
// file ending with "## [time]\n" (no blank line), then --edit opens it.
func TestEndOfContentRequiresBlankLine(t *testing.T) {
	// Actual file content after `jrnl-md ""` — no blank line after heading
	raw := "# 2026-04-01 Wednesday\n\n## [11:18 AM]\n"

	// Without EnsureBlankLineAfterLastHeading, cursor lands beyond EOF
	rawLines := len(strings.Split(raw, "\n"))
	cursorRaw := EndOfContent(raw)
	if cursorRaw <= rawLines {
		t.Errorf("expected cursor %d to exceed file lines %d on raw file (this is the bug)", cursorRaw, rawLines)
	}

	// After ensuring blank lines, cursor lands on a valid line for body content
	fixed := EnsureBlankLineAfterLastHeading(raw)
	fixedLines := strings.Split(fixed, "\n")
	cursorFixed := EndOfContent(fixed)
	if cursorFixed < 1 || cursorFixed > len(fixedLines) {
		t.Fatalf("cursor %d out of range for %d lines", cursorFixed, len(fixedLines))
	}
	// Cursor line should be blank (ready for typing)
	if fixedLines[cursorFixed-1] != "" {
		t.Errorf("cursor line %d should be blank, got %q", cursorFixed, fixedLines[cursorFixed-1])
	}
	// Line before cursor should be blank separator
	if cursorFixed < 3 || fixedLines[cursorFixed-2] != "" {
		t.Errorf("line %d (before cursor) should be blank separator, got %q", cursorFixed-1, fixedLines[cursorFixed-2])
	}
	// Two lines before cursor should be the heading
	if !strings.HasPrefix(fixedLines[cursorFixed-3], "## [") {
		t.Errorf("line %d should be entry heading, got %q", cursorFixed-2, fixedLines[cursorFixed-3])
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

			cfg := Config{DateFmt: "2006-01-02", TimeFmt: "03:04 PM", Template: tt.template}
			_, err := PrepareDayFile(path, date, cfg)
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
