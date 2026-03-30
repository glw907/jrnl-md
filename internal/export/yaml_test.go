package export

import (
	"strings"
	"testing"
	"time"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
)

func TestYAML(t *testing.T) {
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
				if !strings.Contains(output, "tags: {}") {
					t.Error("expected empty tags")
				}
				if !strings.Contains(output, "entries: []") {
					t.Error("expected empty entries")
				}
			},
		},
		{
			name: "single entry with tag",
			entries: []journal.Entry{
				{
					Date:    time.Date(2026, 3, 29, 17, 13, 0, 0, time.Local),
					Body:    "Working on @project today.",
					Tags:    []string{"@project"},
					Starred: false,
				},
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, `date: "2026-03-29"`) {
					t.Errorf("missing date, got:\n%s", output)
				}
				if !strings.Contains(output, `time: "05:13 PM"`) {
					t.Errorf("missing time, got:\n%s", output)
				}
				if !strings.Contains(output, `body: "Working on @project today."`) {
					t.Errorf("missing body, got:\n%s", output)
				}
				if !strings.Contains(output, "starred: false") {
					t.Errorf("missing starred, got:\n%s", output)
				}
				if !strings.Contains(output, `"@project": 1`) {
					t.Errorf("missing tag count, got:\n%s", output)
				}
			},
		},
		{
			name: "body with quotes",
			entries: []journal.Entry{
				{
					Date: time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local),
					Body: `She said "hello" today.`,
				},
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, `"hello"`) || !strings.Contains(output, `body:`) {
					t.Errorf("body with quotes not handled, got:\n%s", output)
				}
			},
		},
		{
			name: "multiline body",
			entries: []journal.Entry{
				{
					Date: time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local),
					Body: "Line one.\nLine two.",
				},
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "body: |") {
					t.Errorf("multiline body should use block scalar, got:\n%s", output)
				}
				if !strings.Contains(output, "      Line one.") {
					t.Errorf("missing indented first line, got:\n%s", output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := YAML(tt.entries, cfg)
			if err != nil {
				t.Fatalf("YAML() error: %v", err)
			}
			tt.check(t, output)
		})
	}
}
