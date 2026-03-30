package main

import (
	"fmt"
	"os"
	"strings"
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

// serializeEntries renders a slice of entries as a multi-day markdown blob
// suitable for editing in a temp file. Entries must be sorted by date.
func serializeEntries(entries []journal.Entry, dateFmt, timeFmt string) string {
	var b strings.Builder
	var lastDayStr string
	for _, e := range entries {
		dayStr := e.Date.Format(dateFmt)
		if dayStr != lastDayStr {
			b.WriteString(journal.DayHeading(e.Date, dateFmt))
			b.WriteString("\n")
			lastDayStr = dayStr
		}
		b.WriteString("\n")
		b.WriteString(e.Format(timeFmt))
	}
	return b.String()
}

// editFiltered serializes entries to a temp file, opens the editor, parses
// the result, replaces the original entries in the journal, and saves.
func editFiltered(fj *journal.FolderJournal, cfg config.Config, configPath string, passphrase string, entries []journal.Entry) error {
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No entries to edit.")
		return nil
	}

	if cfg.General.Editor == "" {
		return fmt.Errorf("no editor configured. Set editor in %s", configPath)
	}

	content := serializeEntries(entries, cfg.Format.Date, cfg.Format.Time)

	tmpFile, err := os.CreateTemp("", "jrnl-md-edit-*.md")
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

	if err := editor.Launch(cfg.General.Editor, tmpPath, 1); err != nil {
		return fmt.Errorf("launching editor: %w", err)
	}

	edited, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("reading edited file: %w", err)
	}

	newEntries, err := journal.ParseMultiDay(string(edited), cfg.Format.Date, cfg.Format.Time)
	if err != nil {
		return fmt.Errorf("parsing edited entries: %w", err)
	}

	fj.ReplaceEntries(entries, newEntries)

	if err := fj.Save(); err != nil {
		return fmt.Errorf("saving journal: %w", err)
	}

	orig := len(entries)
	n := len(newEntries)
	switch {
	case n == 0 && orig > 0:
		fmt.Fprintf(os.Stderr, "%d entries deleted.\n", orig)
	case n == 1:
		fmt.Fprintf(os.Stderr, "1 entry edited.\n")
	default:
		fmt.Fprintf(os.Stderr, "%d entries edited.\n", n)
	}

	return nil
}
