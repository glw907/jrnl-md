package export

import (
	"testing"
	"time"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
)

func TestText(t *testing.T) {
	cfg := config.Default()

	tests := []struct {
		name    string
		entries []journal.Entry
		want    string
	}{
		{
			name:    "empty entries",
			entries: nil,
			want:    "",
		},
		{
			name: "single entry",
			entries: []journal.Entry{
				{
					Date: time.Date(2026, 3, 29, 17, 13, 0, 0, time.Local),
					Body: "Hello world.",
				},
			},
			want: "[2026-03-29 05:13 PM] Hello world.\n",
		},
		{
			name: "multiple entries",
			entries: []journal.Entry{
				{
					Date: time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local),
					Body: "Morning.",
				},
				{
					Date: time.Date(2026, 3, 29, 17, 0, 0, 0, time.Local),
					Body: "Evening.",
				},
			},
			want: "[2026-03-29 09:00 AM] Morning.\n\n[2026-03-29 05:00 PM] Evening.\n",
		},
		{
			name: "multiline body",
			entries: []journal.Entry{
				{
					Date: time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local),
					Body: "Line one.\nLine two.",
				},
			},
			want: "[2026-03-29 09:00 AM] Line one.\nLine two.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Text(tt.entries, cfg)
			if err != nil {
				t.Fatalf("Text() error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Text() =\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}
