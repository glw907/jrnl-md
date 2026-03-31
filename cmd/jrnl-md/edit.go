package main

import (
	"fmt"
	"os"
	"time"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/editor"
	"github.com/glw907/jrnl-md/internal/journal"
)

func editEntry(fj *journal.FolderJournal, cfg config.Config, configPath string, passphrase string) error {
	if cfg.General.Editor == "" {
		return fmt.Errorf("no editor configured. Set editor in %s", configPath)
	}

	now := time.Now()

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
		return editor.LaunchEncrypted(fj.DayFilePath(now), now, ecfg)
	}

	path := fj.DayFilePath(now)
	lineCount, err := editor.PrepareDayFile(path, now, ecfg)
	if err != nil {
		return fmt.Errorf("preparing day file: %w", err)
	}

	return editor.Launch(cfg.General.Editor, path, lineCount)
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

	for _, e := range entries {
		if err := fj.DeleteEntry(e); err != nil {
			return fmt.Errorf("removing old entry: %w", err)
		}
	}

	for _, e := range newEntries {
		if err := fj.AddEntry(e.Date, e.Body, e.Starred); err != nil {
			return fmt.Errorf("adding edited entry: %w", err)
		}
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
