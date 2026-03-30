package journal

import (
	"strings"
	"time"
)

// Filter specifies criteria for selecting entries.
type Filter struct {
	Tags       []string
	AndTags    bool // if true, entry must match ALL Tags (default: any)
	NotTags    []string
	NotStarred bool
	NotTagged  bool
	StartDate  *time.Time
	EndDate    *time.Time
	Starred    bool
	Contains   string
	N          int
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

	notTagSet := make(map[string]bool, len(f.NotTags))
	for _, t := range f.NotTags {
		notTagSet[strings.ToLower(t)] = true
	}

	containsLower := strings.ToLower(f.Contains)

	var result []Entry
	for _, e := range entries {
		if f.matches(e, tagSet, notTagSet, containsLower) {
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
		len(f.NotTags) == 0 &&
		!f.NotStarred &&
		!f.NotTagged &&
		f.StartDate == nil &&
		f.EndDate == nil &&
		!f.Starred &&
		f.Contains == ""
}

func (f Filter) matches(e Entry, tagSet, notTagSet map[string]bool, containsLower string) bool {
	if f.Starred && !e.Starred {
		return false
	}
	if f.NotStarred && e.Starred {
		return false
	}
	if f.NotTagged && len(e.Tags) > 0 {
		return false
	}
	if f.StartDate != nil && e.Date.Before(*f.StartDate) {
		return false
	}
	if f.EndDate != nil && e.Date.After(*f.EndDate) {
		return false
	}
	if len(tagSet) > 0 {
		if f.AndTags {
			for tag := range tagSet {
				found := false
				for _, et := range e.Tags {
					if strings.ToLower(et) == tag {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		} else {
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
	}
	for tag := range notTagSet {
		for _, et := range e.Tags {
			if strings.ToLower(et) == tag {
				return false
			}
		}
	}
	if containsLower != "" {
		if !strings.Contains(strings.ToLower(e.Body), containsLower) {
			return false
		}
	}
	return true
}
