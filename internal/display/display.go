// Package display formats journal days for terminal output.
package display

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/glw907/jrnl-md/internal/journal"
	"golang.org/x/term"
)

// FormatOpts controls how days are displayed.
type FormatOpts struct {
	Linewrap   int
	DateColor  string
	BodyColor  string
	TagsColor  string
	TagSymbols string
}

// TerminalWidth returns the current terminal width, or 80 if it cannot
// be determined (e.g. output is not a terminal).
func TerminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

// FormatDay formats a single Day for full display output.
func FormatDay(day journal.Day, opts FormatOpts) string {
	var sb strings.Builder

	heading := fmt.Sprintf("# %s %s", day.Date.Format("2006-01-02"), day.Date.Format("Monday"))
	sb.WriteString(colorize(heading, opts.DateColor))
	sb.WriteString("\n")

	body := wrapText(day.Body, opts.Linewrap)
	body = highlightTags(body, opts.TagSymbols, opts.TagsColor)
	if opts.BodyColor != "none" && opts.BodyColor != "" {
		body = colorize(body, opts.BodyColor)
	}
	sb.WriteString(body)

	return sb.String()
}

// FormatDays formats multiple days separated by blank lines.
func FormatDays(days []journal.Day, opts FormatOpts) string {
	var parts []string
	for _, day := range days {
		parts = append(parts, FormatDay(day, opts))
	}
	return strings.Join(parts, "\n")
}

// ShortLine produces a single-line summary: "date  first content line…"
// truncated to termWidth. timeFmt is used to detect and skip timestamp
// headings; an empty timeFmt means timestamps are disabled.
func ShortLine(day journal.Day, termWidth int, timeFmt string) string {
	date := day.Date.Format("2006-01-02")
	content := firstContentLine(day.Body, timeFmt)
	prefix := date + "  "
	full := prefix + content
	return truncate(full, termWidth)
}

// firstContentLine returns the first non-blank content line from body,
// skipping the blank line after the heading and any timestamp headings.
func firstContentLine(body, timeFmt string) string {
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			candidate := strings.TrimPrefix(trimmed, "## ")
			if timeFmt != "" {
				if looksLikeTimestamp(candidate, timeFmt) {
					continue
				}
			}
			return candidate
		}
		return trimmed
	}
	return ""
}

// looksLikeTimestamp reports whether s can be parsed with the given
// Go time format.
func looksLikeTimestamp(s, timeFmt string) bool {
	_, err := time.Parse(timeFmt, s)
	return err == nil
}

// wrapText wraps text at width, preserving blank lines (paragraph breaks).
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		if utf8.RuneCountInString(line) <= width {
			result = append(result, line)
			continue
		}
		result = append(result, wrapLine(line, width)...)
	}
	return strings.Join(result, "\n")
}

// wrapLine wraps a single long line at word boundaries.
func wrapLine(line string, width int) []string {
	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{line}
	}
	var lines []string
	current := words[0]
	for _, word := range words[1:] {
		if utf8.RuneCountInString(current)+1+utf8.RuneCountInString(word) <= width {
			current += " " + word
		} else {
			lines = append(lines, current)
			current = word
		}
	}
	lines = append(lines, current)
	return lines
}

// truncate shortens s to at most width runes, appending "…" if truncated.
func truncate(s string, width int) string {
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	return string(runes[:width-1]) + "…"
}

// highlightTags applies color to @tag patterns in text.
func highlightTags(text, tagSyms, colorName string) string {
	if colorName == "none" || colorName == "" || tagSyms == "" {
		return text
	}
	c := colorForName(colorName)
	if c == nil {
		return text
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = replaceTagsInLine(line, tagSyms, c)
	}
	return strings.Join(lines, "\n")
}

// replaceTagsInLine colorizes tag tokens in a single line.
func replaceTagsInLine(line, tagSyms string, c *color.Color) string {
	runes := []rune(line)
	var sb strings.Builder
	i := 0
	for i < len(runes) {
		if strings.ContainsRune(tagSyms, runes[i]) && i+1 < len(runes) {
			j := i + 1
			for j < len(runes) && isTagChar(runes[j]) {
				j++
			}
			if j > i+1 {
				sb.WriteString(c.Sprint(string(runes[i:j])))
				i = j
				continue
			}
		}
		sb.WriteRune(runes[i])
		i++
	}
	return sb.String()
}

func isTagChar(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

// colorize wraps text with the given named color.
func colorize(text, colorName string) string {
	c := colorForName(colorName)
	if c == nil {
		return text
	}
	return c.Sprint(text)
}

// colorForName maps config color name to a fatih/color attribute.
func colorForName(name string) *color.Color {
	switch strings.ToLower(name) {
	case "black":
		return color.New(color.FgBlack)
	case "red":
		return color.New(color.FgRed)
	case "green":
		return color.New(color.FgGreen)
	case "yellow":
		return color.New(color.FgYellow)
	case "blue":
		return color.New(color.FgBlue)
	case "magenta":
		return color.New(color.FgMagenta)
	case "cyan":
		return color.New(color.FgCyan)
	case "white":
		return color.New(color.FgWhite)
	default:
		return nil
	}
}
