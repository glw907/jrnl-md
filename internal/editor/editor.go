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

	content += "\n" + journal.EntryHeading(date, cfg.TimeFmt, false) + "\n"
	cursorLine := countLines(content) // body line after blank separator
	content += "\n"

	if cfg.Template != "" {
		content += cfg.Template
		if !strings.HasSuffix(cfg.Template, "\n") {
			content += "\n"
		}
	}

	if err := atomicfile.WriteFile(path, []byte(content), 0644); err != nil {
		return 0, fmt.Errorf("writing day file: %w", err)
	}

	return cursorLine, nil
}

// WriteTempAndEdit writes content to a temp file, opens the editor at startLine,
// reads back the edited bytes, and removes the temp file on return.
func WriteTempAndEdit(editorCmd, content string, startLine int) ([]byte, error) {
	edited, _, err := writeTempAndEdit(editorCmd, content, startLine, true)
	return edited, err
}

// WriteTempAndEditKeep is like WriteTempAndEdit but does not remove the temp
// file. The caller is responsible for cleanup via the returned path.
func WriteTempAndEditKeep(editorCmd, content string, startLine int) ([]byte, string, error) {
	return writeTempAndEdit(editorCmd, content, startLine, false)
}

func writeTempAndEdit(editorCmd, content string, startLine int, cleanup bool) ([]byte, string, error) {
	tmpFile, err := os.CreateTemp("", "jrnl-md-*.md")
	if err != nil {
		return nil, "", fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	if cleanup {
		defer os.Remove(tmpPath)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return nil, "", fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return nil, "", fmt.Errorf("closing temp file: %w", err)
	}

	if err := Launch(editorCmd, tmpPath, startLine); err != nil {
		return nil, "", fmt.Errorf("launching editor: %w", err)
	}

	edited, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, "", fmt.Errorf("reading edited file: %w", err)
	}
	return edited, tmpPath, nil
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

// EndOfContent returns the line number where new content should be typed.
// If the last entry has an empty body (just a heading with no text after it),
// returns the body line (after the heading and blank separator).
// Otherwise returns the end of file.
func EndOfContent(text string) int {
	lines := strings.Split(text, "\n")
	// Find the last entry heading (## [...])
	lastHeading := -1
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "## [") {
			lastHeading = i
		}
	}
	if lastHeading >= 0 {
		// Check if this entry has body content (non-blank lines after the heading)
		hasBody := false
		for _, line := range lines[lastHeading+1:] {
			if strings.TrimSpace(line) != "" {
				hasBody = true
				break
			}
		}
		if !hasBody {
			// Cursor on the line after the blank separator (1-indexed):
			// ## [time]    ← lastHeading
			// (blank)      ← separator for good markdown
			// (cursor)     ← where the user types
			return lastHeading + 3
		}
	}
	n := countLines(text)
	if n < 1 {
		return 1
	}
	return n
}

// EnsureBlankLineAfterLastHeading ensures the file has a blank separator line
// after the last entry heading so the cursor can land on the body line.
func EnsureBlankLineAfterLastHeading(text string) string {
	lines := strings.Split(text, "\n")
	// Find last non-empty line
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		if !strings.HasPrefix(strings.TrimSpace(lines[i]), "## [") {
			return text // last content isn't a heading, nothing to do
		}
		// Count trailing newlines after the heading
		trailing := len(lines) - 1 - i // number of empty lines after heading
		need := 2                       // blank separator + cursor line
		if trailing < need {
			return text + strings.Repeat("\n", need-trailing)
		}
		return text
	}
	return text
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
	existing += "\n" + journal.EntryHeading(date, cfg.TimeFmt, false) + "\n"
	cursorLine := countLines(existing) // blank line after heading
	existing += "\n"
	if cfg.Template != "" {
		existing += cfg.Template
		if !strings.HasSuffix(cfg.Template, "\n") {
			existing += "\n"
		}
	}
	return existing, cursorLine
}
