// Package editor launches an external text editor for day file editing.
package editor

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Resolve returns the editor to use: configEditor → $VISUAL → $EDITOR.
func Resolve(configEditor string) string {
	if configEditor != "" {
		return configEditor
	}
	if v := os.Getenv("VISUAL"); v != "" {
		return v
	}
	return os.Getenv("EDITOR")
}

// Open launches editorName to edit the file at path, positioning the cursor
// at the first content line. Blocks until the editor exits.
func Open(editorName, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	line := cursorLine(string(data))
	args := buildArgs(editorName, path, line)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// cursorLine returns the 1-based line number of the first non-blank body
// line after the day heading.
func cursorLine(content string) int {
	lines := strings.Split(content, "\n")
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "" {
			return i + 1
		}
	}
	if len(lines) > 1 {
		return len(lines)
	}
	return 2
}

// buildArgs constructs the command arguments for the editor.
func buildArgs(editorName, path string, line int) []string {
	base := filepath.Base(editorName)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	switch name {
	case "vi", "vim", "nvim", "gvim", "neovim":
		return []string{editorName, "+" + strconv.Itoa(line), path}
	case "micro":
		return []string{editorName, path + ":" + strconv.Itoa(line)}
	case "emacs", "emacsclient":
		return []string{editorName, "+" + strconv.Itoa(line), path}
	case "nano":
		return []string{editorName, "+" + strconv.Itoa(line), path}
	case "code", "code-insiders":
		return []string{editorName, "--goto", path + ":" + strconv.Itoa(line)}
	default:
		return []string{editorName, path}
	}
}
