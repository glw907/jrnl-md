package journal

import (
	"regexp"
	"strings"
)

var blankBeforeHeadingRe = regexp.MustCompile(`\n{3,}(## \[)`)

// CleanupDayContent applies light normalization to a day file's content:
//   - strips trailing empty entry headings (## [time] with no body)
//   - normalizes to exactly one blank line before ## headings
//   - ensures a single trailing newline
func CleanupDayContent(text string) string {
	lines := strings.Split(text, "\n")

	// Strip trailing empty entry headings:
	// A ## [time] heading is "empty" if everything after it is blank.
	for {
		// Find the last ## heading
		lastHeading := -1
		for i := len(lines) - 1; i >= 0; i-- {
			if entryRe.MatchString(lines[i]) {
				lastHeading = i
				break
			}
		}
		if lastHeading == -1 {
			break
		}
		// Check if everything after it is blank
		allBlank := true
		for i := lastHeading + 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) != "" {
				allBlank = false
				break
			}
		}
		if !allBlank {
			break
		}
		// Remove the heading and everything after it
		lines = lines[:lastHeading]
		// Also remove any trailing blank lines before where the heading was
		for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
			lines = lines[:len(lines)-1]
		}
	}

	result := strings.Join(lines, "\n")
	result = blankBeforeHeadingRe.ReplaceAllString(result, "\n\n$1")
	result = strings.TrimRight(result, "\n") + "\n"

	return result
}
