package main

import (
	"fmt"
	"os"
	"strings"
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
if timestamps are enabled) before the editor opens. Re-opening today's file
adds a new timestamp heading. If the previous timestamp heading has no content,
it is removed first. The cursor is positioned at the end of the file, ready
for a new paragraph.

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

// stripEmptyTimestamp removes a trailing ## heading with no content after it.
func stripEmptyTimestamp(body string) string {
	trimmed := strings.TrimRight(body, "\n")
	i := strings.LastIndex(trimmed, "\n")
	if i == -1 {
		return body
	}
	if strings.HasPrefix(trimmed[i+1:], "## ") {
		return strings.TrimRight(trimmed[:i], "\n") + "\n"
	}
	return body
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

	tsHeading := ""
	if cfg.General.Timestamps && date.Equal(today) {
		tsHeading = "## " + now.Format(cfg.Format.Time)
	}

	existing, err := s.Load(date)
	if os.IsNotExist(err) {
		body := "\n"
		if tsHeading != "" {
			body = "\n" + tsHeading + "\n"
		}
		if err := s.Save(journal.Day{Date: date, Body: body}); err != nil {
			return fmt.Errorf("creating day file: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("loading day file: %w", err)
	} else {
		body := stripEmptyTimestamp(existing.Body)
		if tsHeading != "" {
			body = strings.TrimRight(body, "\n") + "\n\n" + tsHeading + "\n"
		}
		if body != existing.Body {
			existing.Body = body
			if err := s.Save(existing); err != nil {
				return fmt.Errorf("updating day file: %w", err)
			}
		}
	}

	return editor.Open(editorName, s.DayPath(date))
}
