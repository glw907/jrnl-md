package export

import (
	"testing"

	"github.com/glw907/jrnl-md/internal/journal"
)

func TestTagCounts(t *testing.T) {
	tests := []struct {
		name    string
		entries []journal.Entry
		want    map[string]int
	}{
		{
			name:    "empty entries",
			entries: nil,
			want:    map[string]int{},
		},
		{
			name: "single tag",
			entries: []journal.Entry{
				{Tags: []string{"@work"}},
			},
			want: map[string]int{"@work": 1},
		},
		{
			name: "multiple entries same tag",
			entries: []journal.Entry{
				{Tags: []string{"@work"}},
				{Tags: []string{"@work", "@idea"}},
			},
			want: map[string]int{"@work": 2, "@idea": 1},
		},
		{
			name: "entry with no tags",
			entries: []journal.Entry{
				{Tags: nil},
				{Tags: []string{"@daily"}},
			},
			want: map[string]int{"@daily": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TagCounts(tt.entries)
			if len(got) != len(tt.want) {
				t.Fatalf("TagCounts() returned %d tags, want %d", len(got), len(tt.want))
			}
			for tag, wantCount := range tt.want {
				if got[tag] != wantCount {
					t.Errorf("TagCounts()[%q] = %d, want %d", tag, got[tag], wantCount)
				}
			}
		})
	}
}
