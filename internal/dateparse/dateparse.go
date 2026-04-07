// Package dateparse parses explicit and natural-language date strings.
package dateparse

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	"github.com/olebedev/when/rules/en"
)

// w is the shared natural-language parser.
var w *when.Parser

func init() { //nolint -- package-level init is justified: when.Parser has no zero value
	w = when.New(nil)
	w.Add(en.All...)
	w.Add(common.All...)
}

// Parse parses a date string relative to now. Accepts:
//   - YYYY-MM-DD  → exact date
//   - YYYY-MM     → first of month
//   - YYYY        → January 1 of year
//   - MM-DD       → that month/day in now's year
//   - Natural language: "yesterday", "today", "last monday", "3 days ago"
//
// Returns the parsed date at midnight UTC.
func Parse(s string, now time.Time) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	if t, ok := parseExplicit(s, now); ok {
		return t, nil
	}

	r, err := w.Parse(s, now)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing date %q: %w", s, err)
	}
	if r == nil {
		return time.Time{}, fmt.Errorf("unrecognized date: %q", s)
	}
	t := r.Time
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
}

// parseExplicit tries the structured date formats.
func parseExplicit(s string, now time.Time) (time.Time, bool) {
	parts := strings.Split(s, "-")
	switch len(parts) {
	case 3:
		y, ey := strconv.Atoi(parts[0])
		m, em := strconv.Atoi(parts[1])
		d, ed := strconv.Atoi(parts[2])
		if ey != nil || em != nil || ed != nil {
			return time.Time{}, false
		}
		if y < 1000 || y > 9999 || m < 1 || m > 12 || d < 1 || d > 31 {
			return time.Time{}, false
		}
		t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
		if t.Month() != time.Month(m) || t.Day() != d {
			return time.Time{}, false
		}
		return t, true

	case 2:
		if len(parts[0]) == 4 {
			y, ey := strconv.Atoi(parts[0])
			m, em := strconv.Atoi(parts[1])
			if ey != nil || em != nil || y < 1000 || m < 1 || m > 12 {
				return time.Time{}, false
			}
			return time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.UTC), true
		}
		m, em := strconv.Atoi(parts[0])
		d, ed := strconv.Atoi(parts[1])
		if em != nil || ed != nil || m < 1 || m > 12 || d < 1 || d > 31 {
			return time.Time{}, false
		}
		t := time.Date(now.Year(), time.Month(m), d, 0, 0, 0, 0, time.UTC)
		if t.Month() != time.Month(m) || t.Day() != d {
			return time.Time{}, false
		}
		return t, true

	case 1:
		y, ey := strconv.Atoi(parts[0])
		if ey != nil || y < 1000 || y > 9999 {
			return time.Time{}, false
		}
		return time.Date(y, time.January, 1, 0, 0, 0, 0, time.UTC), true
	}

	return time.Time{}, false
}
