package journal

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	titleRe = regexp.MustCompile(`^#\s+(\S+)\s+\S+`)
	entryRe = regexp.MustCompile(`(?m)^##\s+\[([^\]]+)\](\s+\*)?\s*$`)
)

// ParseError reports a parse failure with location and context.
type ParseError struct {
	File     string // file path or description (set by caller)
	Line     int    // 1-based line number in the source text
	Value    string // the bad value found
	Expected string // what was expected
}

func (e *ParseError) Error() string {
	loc := e.File
	if loc == "" {
		loc = "input"
	}
	if e.Value != "" {
		return fmt.Sprintf("%s: line %d: can't parse %q (expected %s)", loc, e.Line, e.Value, e.Expected)
	}
	return fmt.Sprintf("%s: line %d: missing (expected %s)", loc, e.Line, e.Expected)
}

type day struct {
	date    time.Time
	entries []Entry
}

// Format serializes a day as markdown: day heading followed by entry sections.
func (d day) Format(dateFmt, timeFmt string) string {
	var b strings.Builder

	b.WriteString(DayHeading(d.date, dateFmt))
	b.WriteString("\n")

	for _, e := range d.entries {
		b.WriteString("\n")
		b.WriteString(e.Format(timeFmt))
	}

	return b.String()
}

// parseDay parses markdown text into a day with its entries.
func parseDay(text, dateFmt, timeFmt string) (day, error) {
	var d day

	titleMatch := titleRe.FindStringSubmatch(text)
	if titleMatch == nil {
		return d, &ParseError{
			Line:     1,
			Expected: fmt.Sprintf("day heading like \"# %s %s\"", time.Now().Format(dateFmt), time.Now().Format("Monday")),
		}
	}

	dayDate, err := time.ParseInLocation(dateFmt, titleMatch[1], time.Local)
	if err != nil {
		return d, &ParseError{
			Line:     1,
			Value:    titleMatch[1],
			Expected: fmt.Sprintf("date in format %q", dateFmt),
		}
	}
	d.date = dayDate

	matches := entryRe.FindAllStringSubmatchIndex(text, -1)

	for i, match := range matches {
		timeStr := text[match[2]:match[3]]

		// Compute 1-based line number for this match
		lineNum := 1 + strings.Count(text[:match[0]], "\n")

		entryTime, err := time.ParseInLocation(timeFmt, timeStr, time.Local)
		if err != nil {
			return d, &ParseError{
				Line:     lineNum,
				Value:    timeStr,
				Expected: fmt.Sprintf("time in format %q, e.g. \"## [%s]\"", timeFmt, time.Now().Format(timeFmt)),
			}
		}

		starred := match[4] != -1

		entryDate := time.Date(
			dayDate.Year(), dayDate.Month(), dayDate.Day(),
			entryTime.Hour(), entryTime.Minute(), entryTime.Second(),
			0, time.Local,
		)

		bodyStart := match[1]
		var bodyEnd int
		if i+1 < len(matches) {
			bodyEnd = matches[i+1][0]
		} else {
			bodyEnd = len(text)
		}

		body := strings.TrimSpace(text[bodyStart:bodyEnd])

		d.entries = append(d.entries, Entry{
			Date:    entryDate,
			Body:    body,
			Starred: starred,
		})
	}

	return d, nil
}

// ParseDayContent validates day file content by parsing it. Returns any
// parse error. Exported for use by the editor validation loop.
func ParseDayContent(text, dateFmt, timeFmt string) error {
	_, err := parseDay(text, dateFmt, timeFmt)
	return err
}

func (d *day) addEntry(body string, starred bool, date time.Time) {
	d.entries = append(d.entries, Entry{
		Date:    date,
		Body:    body,
		Starred: starred,
	})
}

// FormatEntries serializes a flat slice of entries as a multi-day markdown blob.
// Entries must be sorted by date; unsorted input produces malformed day headings.
func FormatEntries(entries []Entry, dateFmt, timeFmt string) string {
	var b strings.Builder
	var lastDayStr string
	for _, e := range entries {
		dayStr := e.Date.Format(dateFmt)
		if dayStr != lastDayStr {
			b.WriteString(DayHeading(e.Date, dateFmt))
			b.WriteString("\n")
			lastDayStr = dayStr
		}
		b.WriteString("\n")
		b.WriteString(e.Format(timeFmt))
	}
	return b.String()
}

// ParseMultiDay parses a multi-day markdown blob into a flat slice of entries.
// Each day section begins with "# YYYY-MM-DD Weekday". Lines starting with
// "# " followed by a non-digit are not treated as day headings.
func ParseMultiDay(text, dateFmt, timeFmt string) ([]Entry, error) {
	lines := strings.Split(text, "\n")
	var sectionStarts []int
	for i, line := range lines {
		if strings.HasPrefix(line, "# ") && len(line) > 2 && line[2] >= '0' && line[2] <= '9' {
			sectionStarts = append(sectionStarts, i)
		}
	}

	if len(sectionStarts) == 0 {
		return nil, nil
	}

	var entries []Entry
	for i, start := range sectionStarts {
		end := len(lines)
		if i+1 < len(sectionStarts) {
			end = sectionStarts[i+1]
		}
		section := strings.Join(lines[start:end], "\n")
		d, err := parseDay(section, dateFmt, timeFmt)
		if err != nil {
			var pe *ParseError
			if errors.As(err, &pe) {
				pe.Line += start
				return nil, pe
			}
			return nil, fmt.Errorf("parsing day section at line %d: %w", start+1, err)
		}
		entries = append(entries, d.entries...)
	}

	return entries, nil
}
