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

const msgEmptyAbort = "No entries found after editing. Were you trying to delete all entries? Aborting — no changes made."

func editEntry(fj *journal.FolderJournal, cfg config.Config, configPath string, passphrase string) error {
	return editDayFile(fj, cfg, configPath, passphrase, time.Now(), true)
}

func editDayFile(fj *journal.FolderJournal, cfg config.Config, configPath string, passphrase string, date time.Time, appendHeading bool) error {
	cfg.General.Editor = config.ResolveEditor(cfg)
	if cfg.General.Editor == "" {
		return fmt.Errorf("no editor configured. Set editor in %s or $VISUAL/$EDITOR", configPath)
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
	backup, _ := os.ReadFile(path)

	var startLine int
	if appendHeading {
		var err error
		startLine, err = editor.PrepareDayFile(path, date, ecfg)
		if err != nil {
			return fmt.Errorf("preparing day file: %w", err)
		}
	} else {
		data, err := os.ReadFile(path)
		if err != nil {
			startLine = 1
		} else {
			original := string(data)
			content := editor.EnsureBlankLineAfterLastHeading(original)
			if content != original {
				if err := atomicfile.WriteFile(path, []byte(content), 0644); err != nil {
					return fmt.Errorf("writing day file: %w", err)
				}
			}
			startLine = editor.EndOfContent(content)
		}
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

		if editor.IsEmptyContent(content) {
			fmt.Fprintln(os.Stderr, msgEmptyAbort)
			if backup != nil {
				if err := atomicfile.WriteFile(path, backup, 0644); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to restore backup for %s: %v\n", path, err)
				}
			}
			return nil
		}

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
		content = editor.EnsureBlankLineAfterLastHeading(content)
		startLine = editor.EndOfContent(content)
	}

	for {
		editedBytes, err := editor.WriteTempAndEdit(ecfg.Command, content, startLine)
		if err != nil {
			return err
		}

		content = string(editedBytes)

		if editor.IsEmptyContent(content) {
			fmt.Fprintln(os.Stderr, msgEmptyAbort)
			return nil
		}

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

		content = journal.CleanupDayContent(content)

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

	cfg.General.Editor = config.ResolveEditor(cfg)
	if cfg.General.Editor == "" {
		return fmt.Errorf("no editor configured. Set editor in %s or $VISUAL/$EDITOR", configPath)
	}

	// Single-day redirect: if all entries are from one day and that day has no
	// other entries, open the day file directly instead of the temp-file round-trip.
	if isSingleDayFullMatch(fj, entries) {
		date := entries[0].Date
		return editDayFile(fj, cfg, configPath, passphrase, date, false)
	}

	// Multi-day or partial-day: temp file round-trip
	content := journal.FormatEntries(entries, cfg.Format.Date, cfg.Format.Time)
	var tmpPath string

	for {
		var edited []byte
		var err error
		edited, tmpPath, err = editor.WriteTempAndEditKeep(cfg.General.Editor, content, 1)
		if err != nil {
			return err
		}

		editedStr := string(edited)

		if editor.IsEmptyContent(editedStr) {
			os.Remove(tmpPath)
			fmt.Fprintln(os.Stderr, msgEmptyAbort)
			return nil
		}

		newEntries, parseErr := journal.ParseMultiDay(editedStr, cfg.Format.Date, cfg.Format.Time)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "Error in edited content:\n  %s\n", parseErr)
			if !prompt.YesNo(os.Stdin, os.Stderr, "Re-open editor?") {
				fmt.Fprintf(os.Stderr, "Edits saved to %s\n", tmpPath)
				fmt.Fprintln(os.Stderr, "Journal unchanged.")
				return nil
			}
			content = editedStr
			os.Remove(tmpPath)
			continue
		}
		if err := fj.DeleteEntries(entries); err != nil {
			return fmt.Errorf("removing old entries (edits saved to %s): %w", tmpPath, err)
		}

		if err := fj.AddEntries(newEntries); err != nil {
			return fmt.Errorf("adding edited entries (edits saved to %s): %w", tmpPath, err)
		}

		os.Remove(tmpPath)

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
}

// isSingleDayFullMatch returns true if all entries are from the same calendar
// day and that day file has no other entries (full match).
func isSingleDayFullMatch(fj *journal.FolderJournal, entries []journal.Entry) bool {
	if len(entries) == 0 {
		return false
	}
	firstDay := entries[0].Date
	for _, e := range entries[1:] {
		if e.Date.Year() != firstDay.Year() || e.Date.Month() != firstDay.Month() || e.Date.Day() != firstDay.Day() {
			return false
		}
	}
	dayEntries, err := fj.DayEntries(firstDay)
	if err != nil {
		return false
	}
	return len(dayEntries) == len(entries)
}
