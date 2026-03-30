package journal

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Entry represents a single journal entry with its timestamp, body,
// star status, and extracted tags.
type Entry struct {
	Date    time.Time
	Body    string
	Starred bool
	Tags    []string
}

// Format renders the entry as a markdown section with ## [time] heading and body.
func (e Entry) Format(timeFmt string) string {
	heading := fmt.Sprintf("## [%s]", e.Date.Format(timeFmt))
	if e.Starred {
		heading += " *"
	}

	body := strings.TrimRight(e.Body, "\n ")
	if body != "" {
		return heading + "\n\n" + body + "\n"
	}
	return heading + "\n"
}

// FormatShort returns a compact one-line representation with date, time,
// and a truncated body preview.
func (e Entry) FormatShort(dateFmt, timeFmt string) string {
	dateStr := e.Date.Format(dateFmt)
	timeStr := e.Date.Format(timeFmt)

	preview := strings.ReplaceAll(e.Body, "\n", " ")
	preview = strings.TrimSpace(preview)
	if len(preview) > 60 {
		preview = preview[:57] + "..."
	}

	if preview != "" {
		return dateStr + " " + timeStr + " " + preview
	}
	return dateStr + " " + timeStr
}

// TagParser extracts tags from body text using a precompiled regex for the
// configured tag symbols.
type TagParser struct {
	re *regexp.Regexp
}

// NewTagParser compiles a tag regex for the given symbols. Returns nil if
// tagSymbols is empty.
func NewTagParser(tagSymbols string) *TagParser {
	if tagSymbols == "" {
		return nil
	}
	pattern := fmt.Sprintf(`(?:^|\s)([%s][\w][\w-]*)`, regexp.QuoteMeta(tagSymbols))
	return &TagParser{re: regexp.MustCompile(pattern)}
}

// Parse extracts tags from body text.
func (tp *TagParser) Parse(body string) []string {
	if tp == nil {
		return nil
	}
	matches := tp.re.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		return nil
	}
	var tags []string
	for _, m := range matches {
		tags = append(tags, strings.ToLower(m[1]))
	}
	return tags
}

// ParseTags extracts tags from body text using the configured tag symbols.
func ParseTags(body, tagSymbols string) []string {
	return NewTagParser(tagSymbols).Parse(body)
}
