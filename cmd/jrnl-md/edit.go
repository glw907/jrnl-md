package main

import (
	"fmt"
	"os"
	"time"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/editor"
	"github.com/glw907/jrnl-md/internal/journal"
)

func editEntry(fj *journal.FolderJournal, cfg config.Config, configPath string, encrypted bool, passphrase string) error {
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

	if encrypted {
		return editor.LaunchEncrypted(cfg.General.Editor, fj.DayFilePath(now), now,
			cfg.Format.Date, cfg.Format.Time, passphrase, tmpl)
	}

	path := fj.DayFilePath(now)
	lineCount, err := editor.PrepareDayFile(path, now, cfg.Format.Date, cfg.Format.Time, tmpl)
	if err != nil {
		return fmt.Errorf("preparing day file: %w", err)
	}

	return editor.Launch(cfg.General.Editor, path, lineCount)
}
