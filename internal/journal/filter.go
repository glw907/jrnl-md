package journal

import (
	"strings"
	"time"
)

// Filter specifies criteria for selecting entries.
type Filter struct {
	Tags      []string
	StartDate *time.Time
	EndDate   *time.Time
	Starred   bool
	Contains  string
	N         int
}

// Apply returns entries matching all filter criteria. If N is set,
// only the last N matching entries are returned.
func (f Filter) Apply(entries []Entry) []Entry {
	if f.isEmpty() {
		if f.N > 0 && f.N < len(entries) {
			return entries[len(entries)-f.N:]
		}
		return entries
	}

	tagSet := make(map[string]bool, len(f.Tags))
	for _, t := range f.Tags {
		tagSet[strings.ToLower(t)] = true
	}

	var result []Entry
	for _, e := range entries {
		if f.matches(e, tagSet) {
			result = append(result, e)
		}
	}

	if f.N > 0 && f.N < len(result) {
		result = result[len(result)-f.N:]
	}

	return result
}

func (f Filter) isEmpty() bool {
	return len(f.Tags) == 0 &&
		f.StartDate == nil &&
		f.EndDate == nil &&
		!f.Starred &&
		f.Contains == ""
}

func (f Filter) matches(e Entry, tagSet map[string]bool) bool {
	if f.Starred && !e.Starred {
		return false
	}
	if f.StartDate != nil && e.Date.Before(*f.StartDate) {
		return false
	}
	if f.EndDate != nil && e.Date.After(*f.EndDate) {
		return false
	}
	if len(tagSet) > 0 {
		found := false
		for _, t := range e.Tags {
			if tagSet[t] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if f.Contains != "" {
		if !strings.Contains(strings.ToLower(e.Body), strings.ToLower(f.Contains)) {
			return false
		}
	}
	return true
}
