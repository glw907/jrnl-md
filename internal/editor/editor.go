package editor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/glw907/jrnl-md/internal/atomicfile"
)

// PrepareDayFile ensures a day file exists with a day heading and a new
// entry time heading appended. Returns the total line count.
func PrepareDayFile(path string, date time.Time, dateFmt, timeFmt, template string) (int, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, fmt.Errorf("creating directory: %w", err)
	}

	var content string
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return 0, fmt.Errorf("reading existing day file: %w", err)
	}
	if err == nil {
		content = string(data)
	}

	if content == "" {
		content = fmt.Sprintf("# %s %s\n", date.Format(dateFmt), date.Format("Monday"))
	}

	content += fmt.Sprintf("\n## [%s]\n\n", date.Format(timeFmt))

	if template != "" {
		content += template
		if !strings.HasSuffix(template, "\n") {
			content += "\n"
		}
	}

	if err := atomicfile.WriteFile(path, []byte(content), 0644); err != nil {
		return 0, fmt.Errorf("writing day file: %w", err)
	}

	return countLines(content), nil
}

// Launch opens the given file in the editor command, positioning the
// cursor at the specified line.
func Launch(editorCmd, path string, line int) error {
	cmd, args := editorArgs(editorCmd, path, line)

	proc := exec.Command(cmd, args...)
	proc.Stdin = os.Stdin
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr

	return proc.Run()
}

func editorArgs(editor, path string, line int) (string, []string) {
	base := strings.Fields(editor)[0]

	switch base {
	case "code", "codium", "vscodium":
		return base, []string{"--wait", "--goto", fmt.Sprintf("%s:%d", path, line)}
	case "subl", "sublime":
		return base, []string{"--wait", fmt.Sprintf("%s:%d", path, line)}
	case "nano":
		return base, []string{fmt.Sprintf("+%d", line), path}
	default:
		return base, []string{fmt.Sprintf("+%d", line), path}
	}
}

func countLines(text string) int {
	return strings.Count(text, "\n") + 1
}
