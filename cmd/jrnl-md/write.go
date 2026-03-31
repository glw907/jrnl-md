package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/dateparse"
	"github.com/glw907/jrnl-md/internal/journal"
)

func writeInline(fj *journal.FolderJournal, text []string, cfg config.Config, now time.Time) error {
	body := strings.Join(text, " ")

	entryTime := now
	if idx := strings.Index(body, ": "); idx > 0 {
		candidate := body[:idx]
		if t, err := dateparse.Parse(candidate, cfg.General.DefaultHour, cfg.General.DefaultMinute); err == nil {
			entryTime = t
			body = body[idx+len(": "):]
		}
	}

	starred := strings.HasSuffix(body, "*") || strings.HasPrefix(body, "*")
	if starred {
		body = strings.Trim(body, "* ")
	}

	if err := fj.AddEntry(entryTime, body, starred); err != nil {
		return fmt.Errorf("adding entry: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Entry added.")
	return nil
}
