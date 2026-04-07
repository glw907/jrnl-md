// Package editor launches an external text editor for day file editing.
package editor

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/glw907/jrnl-md/internal/fsutil"
)

// Open launches editorName to edit the file at path. It ensures the file
// ends with two blank lines (a paragraph separator and a cursor entry line),
// then positions the cursor on the last line. Blocks until the editor exits.
func Open(editorName, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := ensureEntryPoint(string(data))
	if content != string(data) {
		if err := fsutil.AtomicWrite(path, []byte(content), 0644); err != nil {
			return err
		}
	}
	line := strings.Count(content, "\n")
	if line == 0 {
		line = 1
	}
	args := buildArgs(editorName, path, line)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ensureEntryPoint returns content ending with \n\n\n: a blank paragraph
// separator followed by a blank cursor entry line.
func ensureEntryPoint(content string) string {
	trimmed := strings.TrimRight(content, "\n")
	return trimmed + "\n\n\n"
}

func buildArgs(editorName, path string, line int) []string {
	base := filepath.Base(editorName)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	if strings.HasPrefix(name, "nvim-") {
		return []string{editorName, "+" + strconv.Itoa(line), path}
	}

	switch name {
	case "vi", "vim", "nvim", "gvim", "neovim", "emacs", "emacsclient", "nano":
		return []string{editorName, "+" + strconv.Itoa(line), path}
	case "micro":
		return []string{editorName, path + ":" + strconv.Itoa(line)}
	case "code", "code-insiders":
		return []string{editorName, "--goto", path + ":" + strconv.Itoa(line)}
	default:
		return []string{editorName, path}
	}
}
