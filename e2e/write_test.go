package e2e

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestWriteCreatesFile(t *testing.T) {
	env := newTestEnv(t)
	_, err := env.run(t, "write", "Hello, journal!")
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	today := time.Now()
	date := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	if _, err := os.Stat(env.dayFilePath(date)); err != nil {
		t.Errorf("day file not created: %v", err)
	}
}

func TestWriteContainsText(t *testing.T) {
	env := newTestEnv(t)
	_, err := env.run(t, "write", "Morning run completed.")
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	today := time.Now()
	date := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	content := env.readDayFile(t, date)
	if !strings.Contains(content, "Morning run completed.") {
		t.Errorf("day file missing text: %q", content)
	}
}

func TestWriteHeading(t *testing.T) {
	env := newTestEnv(t)
	_, err := env.run(t, "write", "Test entry.")
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	today := time.Now()
	date := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	content := env.readDayFile(t, date)
	expectedHeading := "# " + date.Format("2006-01-02") + " " + date.Format("Monday")
	if !strings.HasPrefix(content, expectedHeading) {
		t.Errorf("day file missing heading %q:\n%s", expectedHeading, content)
	}
}

func TestWriteTimestampHeading(t *testing.T) {
	env := newTestEnv(t)
	_, err := env.run(t, "write", "Entry with timestamp.")
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	today := time.Now()
	date := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	content := env.readDayFile(t, date)
	if !strings.Contains(content, "## ") {
		t.Errorf("timestamps=true but no ## heading found:\n%s", content)
	}
}

func TestWriteTimestampsOff(t *testing.T) {
	env := newTestEnv(t)
	cfg := `[general]
timestamps = false
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
path = "` + env.journalDir + `"
`
	if err := os.WriteFile(env.configPath, []byte(cfg), 0644); err != nil {
		t.Fatalf("WriteFile config: %v", err)
	}

	_, err := env.run(t, "write", "No timestamp entry.")
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	today := time.Now()
	date := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	content := env.readDayFile(t, date)
	if strings.Contains(content, "## ") {
		t.Errorf("timestamps=false but ## heading found:\n%s", content)
	}
}

func TestWriteSecondAppend(t *testing.T) {
	env := newTestEnv(t)

	if _, err := env.run(t, "write", "First entry."); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if _, err := env.run(t, "write", "Second entry."); err != nil {
		t.Fatalf("second write: %v", err)
	}

	today := time.Now()
	date := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	content := env.readDayFile(t, date)
	if !strings.Contains(content, "First entry.") || !strings.Contains(content, "Second entry.") {
		t.Errorf("both entries should be present:\n%s", content)
	}
}

func TestWriteRequiresArgs(t *testing.T) {
	env := newTestEnv(t)
	_, err := env.run(t, "write")
	if err == nil {
		t.Error("write with no args should fail")
	}
}
