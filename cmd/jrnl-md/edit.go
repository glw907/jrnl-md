package main

import (
	"fmt"
	"os"
	"time"

	"github.com/glw907/jrnl-md/internal/dateparse"
	"github.com/glw907/jrnl-md/internal/editor"
	"github.com/glw907/jrnl-md/internal/journal"
	"github.com/spf13/cobra"
)

type editFlags struct {
	on string
}

func newEditCmd(rf *rootFlags) *cobra.Command {
	var f editFlags

	cmd := &cobra.Command{
		Use:   "edit [--on <date>]",
		Short: "Open a day file in your editor",
		Long: `Open a day file in your editor.

Defaults to today. --on selects a specific date. If the day file does not
exist, it is created with a day heading (and a timestamp heading for today,
if timestamps are enabled) before the editor opens. The cursor is positioned
at the end of the file, ready for a new paragraph.

The editor is resolved from the config file, then $VISUAL, then $EDITOR.`,
		Example: `  jrnl-md edit
  jrnl-md edit --on yesterday
  jrnl-md edit --on 2026-01-15`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEdit(cmd, args, rf, &f)
		},
	}

	cmd.Flags().StringVar(&f.on, "on", "", "edit a specific date (default: today)")
	return cmd
}

func runEdit(cmd *cobra.Command, args []string, rf *rootFlags, f *editFlags) error {
	cfg, err := loadConfig(rf)
	if err != nil {
		return err
	}

	editorName := cfg.Editor()
	if editorName == "" {
		return fmt.Errorf("no editor configured: set editor in config or $VISUAL/$EDITOR")
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	date := today
	if f.on != "" {
		parsed, err := dateparse.Parse(f.on, now)
		if err != nil {
			return fmt.Errorf("parsing --on date: %w", err)
		}
		date = parsed
	}

	s := journal.NewStore(cfg.JournalPath(), cfg.Format.Date, "", cfg.Format.TagSymbols)

	if _, err := s.Load(date); os.IsNotExist(err) {
		body := "\n"
		if cfg.General.Timestamps && date.Equal(today) {
			body = "\n## " + now.Format(cfg.Format.Time) + "\n"
		}
		if err := s.Save(journal.Day{Date: date, Body: body}); err != nil {
			return fmt.Errorf("creating day file: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("loading day file: %w", err)
	}

	return editor.Open(editorName, s.DayPath(date))
}
