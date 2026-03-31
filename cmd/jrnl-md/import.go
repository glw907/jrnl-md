package main

import (
	"fmt"
	"io"
	"os"

	"github.com/glw907/jrnl-md/internal/journal"
)

// importEntries reads journal entries from source (file path or "-" for stdin),
// merges them into fj skipping duplicates, saves, and reports counts to stderr.
func importEntries(fj *journal.FolderJournal, source, dateFmt, timeFmt string) error {
	var data []byte
	var err error
	if source == "-" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(source)
	}
	if err != nil {
		return fmt.Errorf("reading import source: %w", err)
	}

	entries, err := journal.ParseMultiDay(string(data), dateFmt, timeFmt)
	if err != nil {
		return fmt.Errorf("parsing import file: %w", err)
	}

	var added, skipped int
	for _, e := range entries {
		ok, err := fj.ImportEntry(e)
		if err != nil {
			return fmt.Errorf("importing entry at %s: %w", e.Date.Format("2006-01-02 15:04"), err)
		}
		if ok {
			added++
		} else {
			skipped++
		}
	}

	fmt.Fprintf(os.Stderr, "Imported %d entries. Skipped %d duplicates.\n", added, skipped)
	return nil
}
