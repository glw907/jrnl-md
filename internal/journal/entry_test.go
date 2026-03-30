package journal

import (
	"testing"
	"time"
)

func TestEntryFormat(t *testing.T) {
	tests := []struct {
		name     string
		entry    Entry
		timeFmt  string
		expected string
	}{
		{
			name: "basic entry",
			entry: Entry{
				Date: time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local),
				Body: "Went for a walk today.",
			},
			timeFmt:  "03:04 PM",
			expected: "## [09:00 AM]\n\nWent for a walk today.\n",
		},
		{
			name: "starred entry",
			entry: Entry{
				Date:    time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local),
				Body:    "Great day.",
				Starred: true,
			},
			timeFmt:  "03:04 PM",
			expected: "## [09:00 AM] *\n\nGreat day.\n",
		},
		{
			name: "multiline body",
			entry: Entry{
				Date: time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local),
				Body: "First paragraph.\n\nSecond paragraph.",
			},
			timeFmt:  "03:04 PM",
			expected: "## [02:30 PM]\n\nFirst paragraph.\n\nSecond paragraph.\n",
		},
		{
			name: "24h time format",
			entry: Entry{
				Date: time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local),
				Body: "Afternoon.",
			},
			timeFmt:  "15:04",
			expected: "## [14:30]\n\nAfternoon.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.entry.Format(tt.timeFmt)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestEntryFormatShort(t *testing.T) {
	entry := Entry{
		Date: time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local),
		Body: "Went for a walk this morning and stopped by the coffee shop on the way back home for a latte.",
	}

	got := entry.FormatShort("2006-01-02", "03:04 PM")
	if len(got) > 100 {
		t.Errorf("short string too long: %d chars", len(got))
	}
	if got[len(got)-3:] != "..." {
		t.Errorf("expected truncation with ..., got %q", got)
	}
	if got[:10] != "2026-03-29" {
		t.Errorf("expected date prefix, got %q", got[:10])
	}
}

func TestEntryFormatShortNoTruncation(t *testing.T) {
	entry := Entry{
		Date: time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local),
		Body: "Short note.",
	}

	got := entry.FormatShort("2006-01-02", "03:04 PM")
	if got[len(got)-3:] == "..." {
		t.Error("should not truncate short body")
	}
}

func TestParseTags(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		tagSymbols string
		expected   []string
	}{
		{
			name:       "single tag",
			body:       "Had a great day @mood",
			tagSymbols: "@",
			expected:   []string{"@mood"},
		},
		{
			name:       "multiple tags",
			body:       "Meeting @work about @project",
			tagSymbols: "@",
			expected:   []string{"@work", "@project"},
		},
		{
			name:       "hash tags",
			body:       "Reading #books and #learning",
			tagSymbols: "#",
			expected:   []string{"#books", "#learning"},
		},
		{
			name:       "mixed symbols",
			body:       "Day @mood #topic",
			tagSymbols: "@#",
			expected:   []string{"@mood", "#topic"},
		},
		{
			name:       "no tags",
			body:       "Just a plain entry",
			tagSymbols: "@",
			expected:   nil,
		},
		{
			name:       "tag in email ignored",
			body:       "Email user@example.com about it",
			tagSymbols: "@",
			expected:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTags(tt.body, tt.tagSymbols)
			if len(got) != len(tt.expected) {
				t.Errorf("got %d tags %v, want %d tags %v", len(got), got, len(tt.expected), tt.expected)
				return
			}
			for i, tag := range got {
				if tag != tt.expected[i] {
					t.Errorf("tag[%d] = %q, want %q", i, tag, tt.expected[i])
				}
			}
		})
	}
}
