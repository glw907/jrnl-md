package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glw907/jrnl-md/internal/atomicfile"
	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/crypto"
	"github.com/glw907/jrnl-md/internal/editor"
	"github.com/glw907/jrnl-md/internal/journal"
	"github.com/glw907/jrnl-md/internal/prompt"
)

func editEntry(fj *journal.FolderJournal, cfg config.Config, configPath string, passphrase string) error {
	return editDayFile(fj, cfg, configPath, passphrase, time.Now(), true)
}

func editDayFile(fj *journal.FolderJournal, cfg config.Config, configPath string, passphrase string, date time.Time, appendHeading bool) error {
	if cfg.General.Editor == "" {
		return fmt.Errorf("no editor configured. Set editor in %s", configPath)
	}

	path := fj.DayFilePath(date)

	var tmpl string
	if cfg.General.Template != "" {
		tmplPath, err := expandPath(cfg.General.Template)
		if err != nil {
			return fmt.Errorf("expanding template path: %w", err)
		}
		data, err := os.ReadFile(tmplPath)
		if err != nil {
			return fmt.Errorf("reading template %s: %w", tmplPath, err)
		}
		tmpl = string(data)
	}

	ecfg := editor.Config{
		Command:    cfg.General.Editor,
		DateFmt:    cfg.Format.Date,
		TimeFmt:    cfg.Format.Time,
		Passphrase: passphrase,
		Template:   tmpl,
	}

	if fj.Encrypted() {
		return editDayFileEncrypted(path, date, ecfg, appendHeading)
	}

	return editDayFilePlain(path, date, ecfg, appendHeading)
}

func editDayFilePlain(path string, date time.Time, ecfg editor.Config, appendHeading bool) error {
	// Read backup
	backup, _ := os.ReadFile(path)

	// Prepare file
	var startLine int
	if appendHeading {
		var err error
		startLine, err = editor.PrepareDayFile(path, date, ecfg)
		if err != nil {
			return fmt.Errorf("preparing day file: %w", err)
		}
	} else {
		startLine = 1
	}

	for {
		if err := editor.Launch(ecfg.Command, path, startLine); err != nil {
			return fmt.Errorf("launching editor: %w", err)
		}

		edited, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading edited file: %w", err)
		}

		content := string(edited)

		// Empty check
		if editor.IsEmptyContent(content) {
			fmt.Fprintln(os.Stderr, "No entries found after editing. Were you trying to delete all entries? Aborting — no changes made.")
			if backup != nil {
				if err := atomicfile.WriteFile(path, backup, 0644); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to restore backup for %s: %v\n", path, err)
				}
			}
			return nil
		}

		// Validate
		parseErr := journal.ParseDayContent(content, ecfg.DateFmt, ecfg.TimeFmt)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "Error in edited content:\n  %s\n", parseErr)
			if !prompt.YesNo(os.Stdin, os.Stderr, "Re-open editor?") {
				fmt.Fprintf(os.Stderr, "Warning: %s may contain invalid entries\n", path)
				return nil
			}
			startLine = 1
			continue
		}

		// Cleanup
		cleaned := journal.CleanupDayContent(content)
		if cleaned != content {
			if err := atomicfile.WriteFile(path, []byte(cleaned), 0644); err != nil {
				return fmt.Errorf("writing cleaned file: %w", err)
			}
		}

		return nil
	}
}

func editDayFileEncrypted(encPath string, date time.Time, ecfg editor.Config, appendHeading bool) error {
	// Read and decrypt existing content
	var existing string
	data, err := os.ReadFile(encPath)
	if err == nil {
		plain, err := crypto.Decrypt(data, ecfg.Passphrase)
		if err != nil {
			return fmt.Errorf("decrypting %s: %w", encPath, err)
		}
		existing = string(plain)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reading %s: %w", encPath, err)
	}

	content := existing
	var startLine int
	if appendHeading {
		content, startLine = editor.PrepareEncryptedContent(content, date, ecfg)
	} else {
		startLine = 1
	}

	for {
		editedBytes, err := editor.WriteTempAndEdit(ecfg.Command, content, startLine)
		if err != nil {
			return err
		}

		content = string(editedBytes)

		// Empty check
		if editor.IsEmptyContent(content) {
			fmt.Fprintln(os.Stderr, "No entries found after editing. Were you trying to delete all entries? Aborting — no changes made.")
			return nil
		}

		// Validate
		parseErr := journal.ParseDayContent(content, ecfg.DateFmt, ecfg.TimeFmt)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "Error in edited content:\n  %s\n", parseErr)
			if !prompt.YesNo(os.Stdin, os.Stderr, "Re-open editor?") {
				// Discard — original encrypted file untouched
				return nil
			}
			startLine = 1
			continue
		}

		// Cleanup
		content = journal.CleanupDayContent(content)

		// Re-encrypt and write
		dir := filepath.Dir(encPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}

		enc, err := crypto.Encrypt([]byte(content), ecfg.Passphrase)
		if err != nil {
			return fmt.Errorf("encrypting: %w", err)
		}

		if err := atomicfile.WriteFile(encPath, enc, 0600); err != nil {
			return fmt.Errorf("writing %s: %w", encPath, err)
		}

		return nil
	}
}

// editFiltered opens filtered entries in the editor, parses the result,
// replaces the original entries in the journal, and saves.
func editFiltered(fj *journal.FolderJournal, cfg config.Config, configPath string, passphrase string, entries []journal.Entry) error {
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No entries to edit.")
		return nil
	}

	if cfg.General.Editor == "" {
		return fmt.Errorf("no editor configured. Set editor in %s", configPath)
	}

	content := journal.FormatEntries(entries, cfg.Format.Date, cfg.Format.Time)

	edited, err := editor.WriteTempAndEdit(cfg.General.Editor, content, 1)
	if err != nil {
		return err
	}

	newEntries, err := journal.ParseMultiDay(string(edited), cfg.Format.Date, cfg.Format.Time)
	if err != nil {
		return fmt.Errorf("parsing edited entries: %w", err)
	}

	if err := fj.DeleteEntries(entries); err != nil {
		return fmt.Errorf("removing old entries: %w", err)
	}

	if err := fj.AddEntries(newEntries); err != nil {
		return fmt.Errorf("adding edited entries: %w", err)
	}

	n := len(newEntries)
	switch {
	case n == 0:
		fmt.Fprintf(os.Stderr, "%d entries deleted.\n", len(entries))
	case n == 1:
		fmt.Fprintf(os.Stderr, "1 entry edited.\n")
	default:
		fmt.Fprintf(os.Stderr, "%d entries edited.\n", n)
	}

	return nil
}
