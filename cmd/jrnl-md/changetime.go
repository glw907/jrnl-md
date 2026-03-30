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

	entries := fj.AllEntries()

	flt, err := buildFilter(f, tagArgs, cfg)
	if err != nil {
		return fmt.Errorf("building filter: %w", err)
	}
	entries = flt.Apply(entries)

	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No entries to modify.")
		return nil
	}

	if len(entries) == 1 {
		fmt.Fprintf(os.Stderr, "1 entry found.\n")
	} else {
		fmt.Fprintf(os.Stderr, "%d entries found.\n", len(entries))
	}

	var toChange []journal.Entry
	for _, e := range entries {
		msg := fmt.Sprintf("Change time for '%s'?", e.FormatShort(cfg.Format.Date, cfg.Format.Time))
		if prompt.YesNo(os.Stdin, os.Stderr, msg) {
			toChange = append(toChange, e)
		}
	}

	if len(toChange) == 0 {
		return nil
	}

	fj.ChangeEntryTimes(toChange, newTime)

	if err := fj.Save(); err != nil {
		return fmt.Errorf("saving journal: %w", err)
	}

	if len(toChange) == 1 {
		fmt.Fprintf(os.Stderr, "1 entry modified.\n")
	} else {
		fmt.Fprintf(os.Stderr, "%d entries modified.\n", len(toChange))
	}

	return nil
}
