package export

import (
	"sort"
	"strings"

	"github.com/glw907/jrnl-md/internal/journal"
)

// TagCounts returns a map of tag name to occurrence count across entries.
func TagCounts(entries []journal.Entry) map[string]int {
	counts := make(map[string]int)
	for _, e := range entries {
		for _, tag := range e.Tags {
			counts[tag]++
		}
	}
	return counts
}

func trimBody(body string) string {
	return strings.TrimRight(body, "\n ")
}

func sortedTagKeys(counts map[string]int) []string {
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
