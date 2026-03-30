package export

import (
	"fmt"
	"strings"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
)

// Text formats entries as bracketed-timestamp plain text lines.
func Text(entries []journal.Entry, cfg config.Config) (string, error) {
	if len(entries) == 0 {
		return "", nil
	}

	var b strings.Builder
	for i, e := range entries {
		if i > 0 {
			b.WriteString("\n")
		}
		timestamp := fmt.Sprintf("%s %s",
			e.Date.Format(cfg.Format.Date),
			e.Date.Format(cfg.Format.Time))
		body := trimBody(e.Body)
		fmt.Fprintf(&b, "[%s] %s\n", timestamp, body)
	}
	return b.String(), nil
}
