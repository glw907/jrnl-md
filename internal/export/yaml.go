package export

import (
	"fmt"
	"strings"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
)

// YAML formats entries as a YAML document. Hand-templated since the
// schema is flat enough to not need a YAML library.
func YAML(entries []journal.Entry, cfg config.Config) (string, error) {
	var b strings.Builder

	counts := TagCounts(entries)
	if len(counts) == 0 {
		b.WriteString("tags: {}\n")
	} else {
		b.WriteString("tags:\n")
		for _, tag := range sortedTagKeys(counts) {
			fmt.Fprintf(&b, "  %q: %d\n", tag, counts[tag])
		}
	}

	if len(entries) == 0 {
		b.WriteString("entries: []\n")
		return b.String(), nil
	}

	b.WriteString("entries:\n")
	for _, e := range entries {
		fmt.Fprintf(&b, "  - date: %q\n", e.Date.Format(cfg.Format.Date))
		fmt.Fprintf(&b, "    time: %q\n", e.Date.Format(cfg.Format.Time))

		body := trimBody(e.Body)
		if strings.Contains(body, "\n") || strings.Contains(body, `"`) {
			b.WriteString("    body: |\n")
			for _, line := range strings.Split(body, "\n") {
				fmt.Fprintf(&b, "      %s\n", line)
			}
		} else {
			fmt.Fprintf(&b, "    body: %q\n", body)
		}

		if len(e.Tags) == 0 {
			b.WriteString("    tags: []\n")
		} else {
			b.WriteString("    tags:\n")
			for _, tag := range e.Tags {
				fmt.Fprintf(&b, "      - %q\n", tag)
			}
		}

		fmt.Fprintf(&b, "    starred: %t\n", e.Starred)
	}

	return b.String(), nil
}
