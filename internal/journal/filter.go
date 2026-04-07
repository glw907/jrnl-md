package journal

import (
	"strings"
	"time"
	"unicode"
)

// Filter specifies which days to include in a list or tags operation.
// All non-zero fields must match (AND semantics across different filter types).
type Filter struct {
	Tags           []string
	NotTags        []string
	AndTags        bool
	Start          *time.Time
	End            *time.Time
	Contains       string
	N              int
	Year           int
	Month          int
	DayOfMonth     int
	TodayInHistory bool
}

// Match reports whether day satisfies all filter criteria.
// tagSyms is the set of tag symbol characters (e.g. "@").
func (f Filter) Match(day Day, tagSyms string) bool {
	if f.Start != nil && day.Date.Before(*f.Start) {
		return false
	}
	if f.End != nil && day.Date.After(*f.End) {
		return false
	}
	if f.Year != 0 && day.Date.Year() != f.Year {
		return false
	}
	if f.Month != 0 && int(day.Date.Month()) != f.Month {
		return false
	}
	if f.DayOfMonth != 0 && day.Date.Day() != f.DayOfMonth {
		return false
	}
	if f.TodayInHistory {
		today := time.Now()
		if day.Date.Year() >= today.Year() {
			return false
		}
		if int(day.Date.Month()) != int(today.Month()) || day.Date.Day() != today.Day() {
			return false
		}
	}
	if f.Contains != "" {
		if !strings.Contains(strings.ToLower(day.Body), strings.ToLower(f.Contains)) {
			return false
		}
	}
	if len(f.Tags) > 0 || len(f.NotTags) > 0 {
		tags := extractTags(day.Body, tagSyms)
		tagSet := make(map[string]bool, len(tags))
		for _, tag := range tags {
			tagSet[tag] = true
		}
		if len(f.Tags) > 0 {
			if f.AndTags {
				for _, want := range f.Tags {
					if !tagSet[want] {
						return false
					}
				}
			} else {
				found := false
				for _, want := range f.Tags {
					if tagSet[want] {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		}
		for _, excl := range f.NotTags {
			if tagSet[excl] {
				return false
			}
		}
	}
	return true
}

func extractTags(body, tagSyms string) []string {
	if tagSyms == "" {
		return nil
	}
	var tags []string
	runes := []rune(body)
	for i := 0; i < len(runes); i++ {
		if strings.ContainsRune(tagSyms, runes[i]) {
			if i+1 < len(runes) && (unicode.IsLetter(runes[i+1]) || runes[i+1] == '_') {
				j := i + 1
				for j < len(runes) && (unicode.IsLetter(runes[j]) || unicode.IsDigit(runes[j]) || runes[j] == '_') {
					j++
				}
				tags = append(tags, string(runes[i:j]))
				i = j - 1
			}
		}
	}
	return tags
}
