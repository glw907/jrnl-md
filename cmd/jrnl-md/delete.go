package main

import (
	"fmt"
	"os"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
	"github.com/glw907/jrnl-md/internal/prompt"
)

func deleteEntries(fj *journal.FolderJournal, cfg config.Config, f *flags, tagArgs []string) error {
	flt, err := buildFilter(f, tagArgs, cfg)
	if err != nil {
		return fmt.Errorf("building filter: %w", err)
	}
	entries, err := fj.Entries(&flt)
	if err != nil {
		return fmt.Errorf("loading journal: %w", err)
	}

	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No entries to delete, because the search returned no results.")
		return nil
	}

	if len(entries) == 1 {
		fmt.Fprintf(os.Stderr, "1 entry found.\n")
	} else {
		fmt.Fprintf(os.Stderr, "%d entries found.\n", len(entries))
	}

	var deleted int
	for _, e := range entries {
		msg := fmt.Sprintf("Delete entry '%s'?", e.FormatShort(cfg.Format.Date, cfg.Format.Time))
		if prompt.YesNo(os.Stdin, os.Stderr, msg) {
			if err := fj.DeleteEntry(e); err != nil {
				return fmt.Errorf("deleting entry: %w", err)
			}
			deleted++
		}
	}

	if deleted == 0 {
		return nil
	}

	if deleted == 1 {
		fmt.Fprintf(os.Stderr, "1 entry deleted.\n")
	} else {
		fmt.Fprintf(os.Stderr, "%d entries deleted.\n", deleted)
	}

	return nil
}
