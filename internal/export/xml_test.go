package export

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
)

func TestXML(t *testing.T) {
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
				if !strings.Contains(output, "<?xml") {
					t.Error("missing XML declaration")
				}
				if !strings.Contains(output, "<entries") {
					t.Error("missing entries element")
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
				if !strings.Contains(output, `date="2026-03-29T17:13:00"`) {
					t.Error("missing or wrong date attribute")
				}
				if !strings.Contains(output, `starred="false"`) {
					t.Error("missing starred attribute")
				}
				if !strings.Contains(output, "Working on @project today.") {
					t.Error("missing body text")
				}
				body := output[strings.Index(output, "<journal"):]
				if err := xml.Unmarshal([]byte(body), &struct{}{}); err != nil {
					t.Errorf("invalid XML: %v", err)
				}
			},
		},
		{
			name: "starred entry",
			entries: []journal.Entry{
				{
					Date:    time.Date(2026, 1, 1, 9, 0, 0, 0, time.Local),
					Body:    "New year!",
					Starred: true,
				},
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, `starred="true"`) {
					t.Error("missing starred=true")
				}
			},
		},
		{
			name: "multiple tags sorted",
			entries: []journal.Entry{
				{
					Date: time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local),
					Body: "Tagged @zebra and @alpha.",
					Tags: []string{"@zebra", "@alpha"},
				},
			},
			check: func(t *testing.T, output string) {
				tagsBlock := output[strings.Index(output, "<tags>"):]
				alphaIdx := strings.Index(tagsBlock, `<tag name="@alpha">`)
				zebraIdx := strings.Index(tagsBlock, `<tag name="@zebra">`)
				if alphaIdx == -1 || zebraIdx == -1 {
					t.Fatalf("missing tag elements in <tags> block:\n%s", tagsBlock)
				}
				if alphaIdx > zebraIdx {
					t.Error("tags in <tags> block should be sorted alphabetically")
				}
			},
		},
		{
			name: "body with angle brackets",
			entries: []journal.Entry{
				{
					Date: time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local),
					Body: "Use <b>bold</b> & \"quotes\".",
				},
			},
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "&lt;b&gt;bold&lt;/b&gt;") {
					t.Error("angle brackets not escaped")
				}
				if !strings.Contains(output, "&amp;") {
					t.Error("ampersand not escaped")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := XML(tt.entries, cfg)
			if err != nil {
				t.Fatalf("XML() error: %v", err)
			}
			tt.check(t, output)
		})
	}
}
