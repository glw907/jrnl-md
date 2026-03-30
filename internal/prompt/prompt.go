package prompt

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// YesNo prints "message [y/N] " to w and reads one line from r.
// Returns true only for "y" or "Y".
func YesNo(r io.Reader, w io.Writer, message string) bool {
	fmt.Fprintf(w, "%s [y/N] ", message)
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return false
	}
	line := strings.TrimSpace(scanner.Text())
	return line == "y" || line == "Y"
}
