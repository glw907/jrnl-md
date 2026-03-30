package export

import (
	"encoding/json"
	"fmt"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
)

type jsonOutput struct {
	Tags    map[string]int `json:"tags"`
	Entries []jsonEntry    `json:"entries"`
}

type jsonEntry struct {
	Date    string   `json:"date"`
	Time    string   `json:"time"`
	Body    string   `json:"body"`
	Tags    []string `json:"tags"`
	Starred bool     `json:"starred"`
}

// JSON formats entries as an indented JSON document.
func JSON(entries []journal.Entry, cfg config.Config) (string, error) {
	out := jsonOutput{
		Tags:    TagCounts(entries),
		Entries: make([]jsonEntry, 0, len(entries)),
	}

	for _, e := range entries {
		tags := e.Tags
		if tags == nil {
			tags = []string{}
		}
		out.Entries = append(out.Entries, jsonEntry{
			Date:    e.Date.Format(cfg.Format.Date),
			Time:    e.Date.Format(cfg.Format.Time),
			Body:    e.Body,
			Tags:    tags,
			Starred: e.Starred,
		})
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling JSON: %w", err)
	}
	return string(data) + "\n", nil
}
