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

// dateOnlyLayouts are layouts that match a date without a time component.
// When one of these matches, defaultHour and defaultMinute are applied.
var dateOnlyLayouts = []string{
	"2006-01-02",
	"01/02/2006",
	"Jan 2, 2006",
	"January 2, 2006",
}

var dateTimeLayouts = []string{
	"2006-01-02 15:04",
	"2006-01-02 03:04 PM",
}

// Parse interprets a date string, trying explicit formats first then
// falling back to natural language parsing. For date-only layouts,
// defaultHour and defaultMinute are applied instead of midnight.
func Parse(input string, defaultHour, defaultMinute int) (time.Time, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	for _, layout := range dateOnlyLayouts {
		if t, err := time.ParseInLocation(layout, input, time.Local); err == nil {
			return time.Date(t.Year(), t.Month(), t.Day(), defaultHour, defaultMinute, 0, 0, time.Local), nil
		}
	}

	// Date+time layouts — time component is explicit, do not apply defaults.
	for _, layout := range dateTimeLayouts {
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
func ParseInclusive(input string, defaultHour, defaultMinute int) (time.Time, error) {
	t, err := Parse(input, defaultHour, defaultMinute)
	if err != nil {
		return t, err
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.Local), nil
}
