// Package journal manages day-level markdown journal files.
package journal

import (
	"fmt"
	"strings"
	"time"
)

// Day represents a single calendar day's journal entry.
// Body contains everything after the "# date" heading line,
// including timestamp headings if timestamps are enabled.
type Day struct {
	Date time.Time
	Body string
}

// headingLine returns the "# YYYY-MM-DD Weekday" line for a date.
func headingLine(date time.Time) string {
	return fmt.Sprintf("# %s %s", date.Format("2006-01-02"), date.Format("Monday"))
}

// formatDay serializes a Day to its markdown file content.
func formatDay(day Day) string {
	var b strings.Builder
	b.WriteString(headingLine(day.Date))
	b.WriteString("\n")
	b.WriteString(day.Body)
	return b.String()
}

// parseDay parses day file content into a Day. The date is provided
// from the file path rather than re-parsed from the heading, since
// the file path is the authoritative source.
func parseDay(date time.Time, content string) (Day, error) {
	idx := strings.Index(content, "\n")
	body := ""
	if idx >= 0 {
		body = content[idx+1:]
	}
	return Day{Date: date, Body: body}, nil
}
