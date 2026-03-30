package export

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
)

func TestJSON(t *testing.T) {
	cfg := config.Default()

	tests := []struct {
		name    string
		entries []journal.Entry
		check   func(t *testing.T, output string)
	}{
		{
			name:    "empty entries",
			entries: nil,
			check: func(t *testing.T, output string) {
				var result map[string]any
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				entries := result["entries"].([]any)
				if len(entries) != 0 {
					t.Errorf("expected 0 entries, got %d", len(entries))
				}
			},
		},
		{
			name: "single entry with tags",
			entries: []journal.Entry{
				{
					Date:    time.Date(2026, 3, 29, 17, 13, 0, 0, time.Local),
					Body:    "Working on @project today.",
					Tags:    []string{"@project"},
					Starred: false,
				},
			},
			check: func(t *testing.T, output string) {
				var result map[string]any
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				entries := result["entries"].([]any)
				if len(entries) != 1 {
					t.Fatalf("expected 1 entry, got %d", len(entries))
				}
				entry := entries[0].(map[string]any)
				if entry["date"] != "2026-03-29" {
					t.Errorf("date = %q, want %q", entry["date"], "2026-03-29")
				}
				if entry["time"] != "05:13 PM" {
					t.Errorf("time = %q, want %q", entry["time"], "05:13 PM")
				}
				if entry["body"] != "Working on @project today." {
					t.Errorf("body = %q", entry["body"])
				}
				tags := result["tags"].(map[string]any)
				if tags["@project"] != float64(1) {
					t.Errorf("tags[@project] = %v, want 1", tags["@project"])
				}
			},
		},
		{
			name: "starred entry",
			entries: []journal.Entry{
				{
					Date:    time.Date(2026, 1, 1, 9, 0, 0, 0, time.Local),
					Body:    "New year!",
					Tags:    nil,
					Starred: true,
				},
			},
			check: func(t *testing.T, output string) {
				var result map[string]any
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				entry := result["entries"].([]any)[0].(map[string]any)
				if entry["starred"] != true {
					t.Errorf("starred = %v, want true", entry["starred"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := JSON(tt.entries, cfg)
			if err != nil {
				t.Fatalf("JSON() error: %v", err)
			}
			tt.check(t, output)
		})
	}
}
