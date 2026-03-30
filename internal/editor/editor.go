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
	"github.com/glw907/jrnl-md/internal/crypto"
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

// prepareEncryptedContent builds the editor content for an encrypted day file.
// If existing is empty, a new day heading is created. A new entry heading is
// always appended.
func prepareEncryptedContent(existing string, date time.Time, cfg Config) (string, int) {
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

// LaunchEncrypted decrypts the day file (if it exists), appends an entry
// heading, opens the editor, then re-encrypts and writes atomically.
func LaunchEncrypted(encPath string, date time.Time, cfg Config) error {
	var existing string
	data, err := os.ReadFile(encPath)
	if err == nil {
		plain, err := crypto.Decrypt(data, cfg.Passphrase)
		if err != nil {
			return fmt.Errorf("decrypting %s: %w", encPath, err)
		}
		existing = string(plain)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reading %s: %w", encPath, err)
	}

	content, lineCount := prepareEncryptedContent(existing, date, cfg)

	tmpFile, err := os.CreateTemp("", "jrnl-md-*.md")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := Launch(cfg.Command, tmpPath, lineCount); err != nil {
		return fmt.Errorf("launching editor: %w", err)
	}

	edited, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("reading edited file: %w", err)
	}

	dir := filepath.Dir(encPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	enc, err := crypto.Encrypt(edited, cfg.Passphrase)
	if err != nil {
		return fmt.Errorf("encrypting: %w", err)
	}

	if err := atomicfile.WriteFile(encPath, enc, 0600); err != nil {
		return fmt.Errorf("writing %s: %w", encPath, err)
	}

	return nil
}
