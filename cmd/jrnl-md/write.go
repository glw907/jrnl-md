package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
)

func writeInline(fj *journal.FolderJournal, text []string, cfg config.Config, now time.Time) error {
	body := strings.Join(text, " ")
	starred := strings.HasSuffix(body, "*") || strings.HasPrefix(body, "*")
	if starred {
		body = strings.Trim(body, "* ")
	}

	fj.AddEntry(now, body, starred)

	if err := fj.Save(); err != nil {
		return fmt.Errorf("saving journal: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Entry added.")
	return nil
}
