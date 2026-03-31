package editor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/glw907/jrnl-md/internal/atomicfile"
	"github.com/glw907/jrnl-md/internal/journal"
)

// Config holds editor-related configuration.
type Config struct {
	Command    string
	DateFmt    string
	TimeFmt    string
	Passphrase string
	Template   string
}

// PrepareDayFile ensures a day file exists with a day heading and a new
// entry time heading appended. Returns the total line count.
func PrepareDayFile(path string, date time.Time, cfg Config) (int, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, fmt.Errorf("creating directory: %w", err)
	}

	var content string
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return 0, fmt.Errorf("reading existing day file: %w", err)
	}
	if err == nil {
		content = string(data)
	}

	if content == "" {
		content = journal.DayHeading(date, cfg.DateFmt) + "\n"
	}

	content += "\n" + journal.EntryHeading(date, cfg.TimeFmt, false) + "\n\n"

	if cfg.Template != "" {
		content += cfg.Template
		if !strings.HasSuffix(cfg.Template, "\n") {
			content += "\n"
		}
	}

	if err := atomicfile.WriteFile(path, []byte(content), 0644); err != nil {
		return 0, fmt.Errorf("writing day file: %w", err)
	}

	return countLines(content), nil
}

// WriteTempAndEdit writes content to a temp file, opens the editor at startLine,
// reads back the edited bytes, and removes the temp file on return.
func WriteTempAndEdit(editorCmd, content string, startLine int) ([]byte, error) {
	tmpFile, err := os.CreateTemp("", "jrnl-md-*.md")
	if err != nil {
		return nil, fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("closing temp file: %w", err)
	}

	if err := Launch(editorCmd, tmpPath, startLine); err != nil {
		return nil, fmt.Errorf("launching editor: %w", err)
	}

	edited, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("reading edited file: %w", err)
	}
	return edited, nil
}

// Launch opens the given file in the editor command, positioning the
// cursor at the specified line.
func Launch(editorCmd, path string, line int) error {
	cmd, args := editorArgs(editorCmd, path, line)

	proc := exec.Command(cmd, args...)
	proc.Stdin = os.Stdin
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr

	return proc.Run()
}

func editorArgs(editor, path string, line int) (string, []string) {
	base := strings.Fields(editor)[0]

	switch base {
	case "code", "codium", "vscodium":
		return base, []string{"--wait", "--goto", fmt.Sprintf("%s:%d", path, line)}
	case "subl", "sublime":
		return base, []string{"--wait", fmt.Sprintf("%s:%d", path, line)}
	case "nano":
		return base, []string{fmt.Sprintf("+%d", line), path}
	default:
		return base, []string{fmt.Sprintf("+%d", line), path}
	}
}

func countLines(text string) int {
	return strings.Count(text, "\n") + 1
}

// IsEmptyContent returns true if the text is empty or whitespace-only.
func IsEmptyContent(text string) bool {
	return strings.TrimSpace(text) == ""
}

// PrepareEncryptedContent builds the editor content for an encrypted day file.
// If existing is empty, a new day heading is created. A new entry heading is
// always appended.
func PrepareEncryptedContent(existing string, date time.Time, cfg Config) (string, int) {
	if existing == "" {
		existing = journal.DayHeading(date, cfg.DateFmt) + "\n"
	}
	existing += "\n" + journal.EntryHeading(date, cfg.TimeFmt, false) + "\n\n"
	if cfg.Template != "" {
		existing += cfg.Template
		if !strings.HasSuffix(cfg.Template, "\n") {
			existing += "\n"
		}
	}
	return existing, countLines(existing)
}
