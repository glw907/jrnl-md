package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var binaryPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "jrnl-md-e2e-*")
	if err != nil {
		panic("MkdirTemp: " + err.Error())
	}
	defer os.RemoveAll(dir)

	binaryPath = filepath.Join(dir, "jrnl-md")
	build := exec.Command("go", "build", "-o", binaryPath, "../cmd/jrnl-md")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		panic("build failed: " + err.Error())
	}

	os.Exit(m.Run())
}

type testEnv struct {
	journalDir string
	configPath string
}

func newTestEnv(t *testing.T) testEnv {
	t.Helper()
	root := t.TempDir()
	journalDir := filepath.Join(root, "journal")
	if err := os.MkdirAll(journalDir, 0755); err != nil {
		t.Fatalf("MkdirAll journal: %v", err)
	}
	configDir := filepath.Join(root, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("MkdirAll config: %v", err)
	}
	configPath := filepath.Join(configDir, "config.toml")
	cfg := `[general]
timestamps = true
linewrap = 79
default_list_count = 10

[format]
time = "03:04 PM"
date = "2006-01-02"
tag_symbols = "@"

[colors]
date = "none"
body = "none"
tags = "none"

[journals.default]
path = "` + journalDir + `"
`
	if err := os.WriteFile(configPath, []byte(cfg), 0644); err != nil {
		t.Fatalf("WriteFile config: %v", err)
	}
	return testEnv{journalDir: journalDir, configPath: configPath}
}

func (e testEnv) run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	allArgs := append([]string{"--config-file", e.configPath}, args...)
	cmd := exec.Command(binaryPath, allArgs...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func (e testEnv) writeDayFile(t *testing.T, date time.Time, content string) {
	t.Helper()
	path := e.dayFilePath(date)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func (e testEnv) dayFilePath(date time.Time) string {
	return filepath.Join(e.journalDir,
		date.Format("2006"), date.Format("01"), date.Format("2006-01-02")+".md")
}

func (e testEnv) readDayFile(t *testing.T, date time.Time) string {
	t.Helper()
	data, err := os.ReadFile(e.dayFilePath(date))
	if err != nil {
		t.Fatalf("readDayFile %v: %v", date, err)
	}
	return string(data)
}

func mustContain(t *testing.T, output, substr string) {
	t.Helper()
	if !strings.Contains(output, substr) {
		t.Errorf("output does not contain %q:\n%s", substr, output)
	}
}

func mustNotContain(t *testing.T, output, substr string) {
	t.Helper()
	if strings.Contains(output, substr) {
		t.Errorf("output should not contain %q:\n%s", substr, output)
	}
}
