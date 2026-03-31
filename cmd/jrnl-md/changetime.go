package main

import (
	"fmt"
	"os"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/dateparse"
	"github.com/glw907/jrnl-md/internal/journal"
	"github.com/glw907/jrnl-md/internal/prompt"
)

func changeTime(fj *journal.FolderJournal, cfg config.Config, f *flags, tagArgs []string) error {
	newTime, err := dateparse.Parse(f.changeTime, cfg.General.DefaultHour, cfg.General.DefaultMinute)
	if err != nil {
		return fmt.Errorf("parsing --change-time date: %w", err)
	}

	flt, err := buildFilter(f, tagArgs, cfg)
	if err != nil {
		return fmt.Errorf("building filter: %w", err)
	}
	entries, err := fj.Entries(&flt)
	if err != nil {
		return fmt.Errorf("loading journal: %w", err)
	}

	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No entries to modify.")
		return nil
	}

	if len(entries) == 1 {
		fmt.Fprintf(os.Stderr, "1 entry found.\n")
	} else {
		fmt.Fprintf(os.Stderr, "%d entries found.\n", len(entries))
	}

	var changed int
	for _, e := range entries {
		msg := fmt.Sprintf("Change time for '%s'?", e.FormatShort(cfg.Format.Date, cfg.Format.Time))
		if prompt.YesNo(os.Stdin, os.Stderr, msg) {
			updated := e
			updated.Date = newTime
			if err := fj.UpdateEntry(e, updated); err != nil {
				return fmt.Errorf("changing time: %w", err)
			}
			changed++
		}
	}

	if changed == 0 {
		return nil
	}

	if changed == 1 {
		fmt.Fprintf(os.Stderr, "1 entry modified.\n")
	} else {
		fmt.Fprintf(os.Stderr, "%d entries modified.\n", changed)
	}

	return nil
}
