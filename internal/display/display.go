package display

import (
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/term"
)

// WrapText wraps text to the given column width, preserving newlines
// between paragraphs.
func WrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		if len(paragraph) <= width {
			lines = append(lines, paragraph)
			continue
		}

		words := strings.Fields(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		current := words[0]
		for _, word := range words[1:] {
			if len(current)+1+len(word) > width {
				lines = append(lines, current)
				current = word
			} else {
				current += " " + word
			}
		}
		lines = append(lines, current)
	}

	return strings.Join(lines, "\n")
}

// IndentBody prepends indent to each non-empty line of body.
func IndentBody(body, indent string) string {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}

// TerminalWidth returns the width of stdout, defaulting to 79.
func TerminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 79
	}
	return w
}

// ColorFunc returns a function that wraps text in the named ANSI color.
// Returns nil for "none" or unrecognized names.
func ColorFunc(name string) func(a ...any) string {
	switch strings.ToLower(name) {
	case "black":
		return color.New(color.FgBlack).SprintFunc()
	case "red":
		return color.New(color.FgRed).SprintFunc()
	case "green":
		return color.New(color.FgGreen).SprintFunc()
	case "yellow":
		return color.New(color.FgYellow).SprintFunc()
	case "blue":
		return color.New(color.FgBlue).SprintFunc()
	case "magenta":
		return color.New(color.FgMagenta).SprintFunc()
	case "cyan":
		return color.New(color.FgCyan).SprintFunc()
	case "white":
		return color.New(color.FgWhite).SprintFunc()
	default:
		return nil
	}
}

// HighlightTags replaces tag occurrences in body with colorFn-wrapped
// versions. tagSymbols is the set of tag prefix characters (e.g. "@").
// If colorFn is nil or tagSymbols is empty, body is returned unchanged.
func HighlightTags(body, tagSymbols string, colorFn func(a ...any) string) string {
	if colorFn == nil || tagSymbols == "" {
		return body
	}
	escaped := regexp.QuoteMeta(tagSymbols)
	re := regexp.MustCompile(`[` + escaped + `]\w+`)
	return re.ReplaceAllStringFunc(body, func(match string) string {
		return colorFn(match)
	})
}
