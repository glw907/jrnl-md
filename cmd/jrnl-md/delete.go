package main

import (
	"fmt"
	"os"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
	"github.com/glw907/jrnl-md/internal/prompt"
)

func deleteEntries(fj *journal.FolderJournal, cfg config.Config, f *flags, tagArgs []string) error {
	entries := fj.AllEntries()

	flt, err := buildFilter(f, tagArgs)
	if err != nil {
		return fmt.Errorf("building filter: %w", err)
	}
	entries = flt.Apply(entries)

	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No entries to delete, because the search returned no results.")
		return nil
	}

	if len(entries) == 1 {
		fmt.Fprintf(os.Stderr, "1 entry found.\n")
	} else {
		fmt.Fprintf(os.Stderr, "%d entries found.\n", len(entries))
	}

	var toDelete []journal.Entry
	for _, e := range entries {
		msg := fmt.Sprintf("Delete entry '%s'?", e.FormatShort(cfg.Format.Date, cfg.Format.Time))
		if prompt.YesNo(os.Stdin, os.Stderr, msg) {
			toDelete = append(toDelete, e)
		}
	}

	if len(toDelete) == 0 {
		return nil
	}

	fj.DeleteEntries(toDelete)

	if err := fj.Save(); err != nil {
		return fmt.Errorf("saving journal: %w", err)
	}

	if len(toDelete) == 1 {
		fmt.Fprintf(os.Stderr, "1 entry deleted.\n")
	} else {
		fmt.Fprintf(os.Stderr, "%d entries deleted.\n", len(toDelete))
	}

	return nil
}
