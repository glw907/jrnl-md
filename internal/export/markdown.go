package export

import (
	"fmt"
	"strings"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
)

// Markdown formats entries grouped by year and month with heading hierarchy.
func Markdown(entries []journal.Entry, cfg config.Config) (string, error) {
	if len(entries) == 0 {
		return "", nil
	}

	var b strings.Builder
	var curYear int
	var curMonth string

	for _, e := range entries {
		year := e.Date.Year()
		month := e.Date.Format("January")

		if year != curYear {
			if curYear != 0 {
				b.WriteString("\n")
			}
			fmt.Fprintf(&b, "# %d\n\n", year)
			curYear = year
			curMonth = ""
		}

		if month != curMonth {
			fmt.Fprintf(&b, "## %s\n\n", month)
			curMonth = month
		}

		heading := fmt.Sprintf("### %s %s",
			e.Date.Format(cfg.Format.Date),
			e.Date.Format(cfg.Format.Time))
		if e.Starred {
			heading += " *"
		}
		b.WriteString(heading + "\n\n")

		body := trimBody(e.Body)
		if body != "" {
			b.WriteString(body + "\n\n")
		}
	}

	return b.String(), nil
}
