package journal

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	titleRe = regexp.MustCompile(`^#\s+(\S+)\s+\S+`)
	entryRe = regexp.MustCompile(`(?m)^##\s+\[([^\]]+)\](\s+\*)?\s*$`)
)

type day struct {
	date     time.Time
	entries  []Entry
	modified bool
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
		return d, fmt.Errorf("no day title found")
	}

	dayDate, err := time.ParseInLocation(dateFmt, titleMatch[1], time.Local)
	if err != nil {
		return d, fmt.Errorf("parsing day date %q: %w", titleMatch[1], err)
	}
	d.date = dayDate

	matches := entryRe.FindAllStringSubmatchIndex(text, -1)

	for i, match := range matches {
		timeStr := text[match[2]:match[3]]
		starred := match[4] != -1

		entryTime, err := time.ParseInLocation(timeFmt, timeStr, time.Local)
		if err != nil {
			return d, fmt.Errorf("parsing entry time %q: %w", timeStr, err)
		}

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

func (d *day) addEntry(body string, starred bool, date time.Time) {
	d.entries = append(d.entries, Entry{
		Date:    date,
		Body:    body,
		Starred: starred,
	})
	d.modified = true
}
