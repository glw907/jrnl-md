package export

import (
	"strings"
	"testing"
	"time"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
)

func TestMarkdown(t *testing.T) {
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
				if output != "" {
					t.Errorf("expected empty output, got %q", output)
				}
			},
		},
		{
			name: "single entry",
			entries: []journal.Entry{
				{
					Date: time.Date(2026, 3, 29, 17, 13, 0, 0, time.Local),
					Body: "Hello world.",
				},
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "# 2026") {
					t.Error("missing year heading")
				}
				if !strings.Contains(output, "## March") {
					t.Error("missing month heading")
				}
				if !strings.Contains(output, "### 2026-03-29 05:13 PM") {
					t.Error("missing entry heading")
				}
				if !strings.Contains(output, "Hello world.") {
					t.Error("missing body")
				}
			},
		},
		{
			name: "entries across months",
			entries: []journal.Entry{
				{
					Date: time.Date(2026, 1, 15, 9, 0, 0, 0, time.Local),
					Body: "January entry.",
				},
				{
					Date: time.Date(2026, 3, 29, 17, 0, 0, 0, time.Local),
					Body: "March entry.",
				},
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "## January") {
					t.Error("missing January heading")
				}
				if !strings.Contains(output, "## March") {
					t.Error("missing March heading")
				}
				janIdx := strings.Index(output, "## January")
				marIdx := strings.Index(output, "## March")
				if janIdx > marIdx {
					t.Error("January should come before March")
				}
			},
		},
		{
			name: "entries across years",
			entries: []journal.Entry{
				{
					Date: time.Date(2025, 12, 31, 23, 0, 0, 0, time.Local),
					Body: "Old year.",
				},
				{
					Date: time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local),
					Body: "New year.",
				},
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "# 2025") {
					t.Error("missing 2025 heading")
				}
				if !strings.Contains(output, "# 2026") {
					t.Error("missing 2026 heading")
				}
			},
		},
		{
			name: "starred entry",
			entries: []journal.Entry{
				{
					Date:    time.Date(2026, 3, 29, 17, 0, 0, 0, time.Local),
					Body:    "Important!",
					Starred: true,
				},
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "### 2026-03-29 05:00 PM *") {
					t.Error("missing starred marker in heading")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := Markdown(tt.entries, cfg)
			if err != nil {
				t.Fatalf("Markdown() error: %v", err)
			}
			tt.check(t, output)
		})
	}
}
