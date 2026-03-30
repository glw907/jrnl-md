package dateparse

import (
	"fmt"
	"strings"
	"time"

	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	"github.com/olebedev/when/rules/en"
)

var parser *when.Parser

func init() {
	parser = when.New(nil)
	parser.Add(en.All...)
	parser.Add(common.All...)
}

// Parse interprets a date string, trying explicit formats first then
// falling back to natural language parsing.
func Parse(input string) (time.Time, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	layouts := []string{
		"2006-01-02",
		"2006-01-02 15:04",
		"2006-01-02 03:04 PM",
		"01/02/2006",
		"Jan 2, 2006",
		"January 2, 2006",
	}

	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, input, time.Local); err == nil {
			return t, nil
		}
	}

	result, err := parser.Parse(input, time.Now())
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing date %q: %w", input, err)
	}
	if result == nil {
		return time.Time{}, fmt.Errorf("could not parse date %q", input)
	}

	return result.Time, nil
}

// ParseInclusive parses a date and returns end-of-day (23:59:59) so
// range filters include the full day.
func ParseInclusive(input string) (time.Time, error) {
	t, err := Parse(input)
	if err != nil {
		return t, err
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.Local), nil
}
