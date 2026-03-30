package e2e

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var binary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "jrnl-md-e2e-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir for binary: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	binary = filepath.Join(dir, "jrnl-md")

	cmd := exec.Command("go", "build", "-o", binary, "./cmd/jrnl-md")
	cmd.Dir = filepath.Join("..")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build binary: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

type testEnv struct {
	dir        string
	configPath string
	journalDir string
}

const testConfigHeader = `[general]
editor = ""
highlight = false
linewrap = 0
indent_character = ""

[format]
time = "03:04 PM"
date = "2006-01-02"
tag_symbols = "@"
file_extension = "md"

[colors]
date = "none"
body = "none"
tags = "none"

`

func newTestEnv(t *testing.T) testEnv {
	t.Helper()
	dir := t.TempDir()
	journalDir := filepath.Join(dir, "journal")
	if err := os.MkdirAll(journalDir, 0755); err != nil {
		t.Fatalf("failed to create journal dir: %v", err)
	}

	configPath := filepath.Join(dir, "config.toml")
	config := testConfigHeader + fmt.Sprintf("[journals.default]\npath = %q\n", journalDir)
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	return testEnv{
		dir:        dir,
		configPath: configPath,
		journalDir: journalDir,
	}
}

func newMultiTestEnv(t *testing.T, journals map[string]string) testEnv {
	t.Helper()
	dir := t.TempDir()

	for _, path := range journals {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("failed to create journal dir: %v", err)
		}
	}

	var sb strings.Builder
	sb.WriteString(testConfigHeader)
	for name, path := range journals {
		fmt.Fprintf(&sb, "[journals.%s]\npath = %q\n\n", name, path)
	}

	configPath := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(configPath, []byte(sb.String()), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	return testEnv{
		dir:        dir,
		configPath: configPath,
	}
}

func writeDayFile(t *testing.T, journalDir string, date time.Time, content string) {
	t.Helper()
	path := filepath.Join(journalDir, date.Format("2006"), date.Format("01"), date.Format("02")+".md")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create dirs for day file: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write day file: %v", err)
	}
}

func run(t *testing.T, env testEnv, args ...string) (stdout, stderr string) {
	t.Helper()
	stdout, stderr, _ = runErr(t, env, args...)
	return stdout, stderr
}

func runWithStdin(t *testing.T, env testEnv, stdin string, args ...string) (stdout, stderr string) {
	t.Helper()
	cmd := newCmd(env, args...)
	cmd.Stdin = strings.NewReader(stdin)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	_ = cmd.Run()
	return outBuf.String(), errBuf.String()
}

func runErr(t *testing.T, env testEnv, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := newCmd(env, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

func newCmd(env testEnv, args ...string) *exec.Cmd {
	full := append([]string{"--config", env.configPath}, args...)
	return exec.Command(binary, full...)
}

func dayFileContent(t *testing.T, journalDir string, date time.Time) string {
	t.Helper()
	path := filepath.Join(journalDir, date.Format("2006"), date.Format("01"), date.Format("02")+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read day file %s: %v", path, err)
	}
	return string(data)
}

func dayFileExists(t *testing.T, journalDir string, date time.Time) bool {
	t.Helper()
	path := filepath.Join(journalDir, date.Format("2006"), date.Format("01"), date.Format("02")+".md")
	_, err := os.Stat(path)
	return err == nil
}
