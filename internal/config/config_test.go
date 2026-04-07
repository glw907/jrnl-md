package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := defaults()
	if cfg.General.Timestamps != true {
		t.Errorf("Timestamps default: got %v, want true", cfg.General.Timestamps)
	}
	if cfg.General.Linewrap != 79 {
		t.Errorf("Linewrap default: got %d, want 79", cfg.General.Linewrap)
	}
	if cfg.General.DefaultListCount != 10 {
		t.Errorf("DefaultListCount default: got %d, want 10", cfg.General.DefaultListCount)
	}
	if cfg.Format.Time != "03:04 PM" {
		t.Errorf("Format.Time default: got %q, want %q", cfg.Format.Time, "03:04 PM")
	}
	if cfg.Format.Date != "2006-01-02" {
		t.Errorf("Format.Date default: got %q, want %q", cfg.Format.Date, "2006-01-02")
	}
	if cfg.Format.TagSymbols != "@" {
		t.Errorf("Format.TagSymbols default: got %q, want %q", cfg.Format.TagSymbols, "@")
	}
	if cfg.Colors.Date != "none" {
		t.Errorf("Colors.Date default: got %q, want %q", cfg.Colors.Date, "none")
	}
	if cfg.Colors.Body != "none" {
		t.Errorf("Colors.Body default: got %q, want %q", cfg.Colors.Body, "none")
	}
	if cfg.Colors.Tags != "none" {
		t.Errorf("Colors.Tags default: got %q, want %q", cfg.Colors.Tags, "none")
	}
}

func TestLoadAutoCreate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("config file not created: %v", err)
	}
	if cfg.General.DefaultListCount != 10 {
		t.Errorf("DefaultListCount: got %d, want 10", cfg.General.DefaultListCount)
	}
}

func TestLoadExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `[general]
timestamps = false
linewrap = 100
default_list_count = 5

[format]
time = "15:04"
date = "2006-01-02"
tag_symbols = "@#"

[colors]
date = "blue"
body = "none"
tags = "cyan"

[journals.default]
path = "/tmp/myjournal"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.General.Timestamps != false {
		t.Errorf("Timestamps: got %v, want false", cfg.General.Timestamps)
	}
	if cfg.General.Linewrap != 100 {
		t.Errorf("Linewrap: got %d, want 100", cfg.General.Linewrap)
	}
	if cfg.General.DefaultListCount != 5 {
		t.Errorf("DefaultListCount: got %d, want 5", cfg.General.DefaultListCount)
	}
	if cfg.Format.Time != "15:04" {
		t.Errorf("Format.Time: got %q, want %q", cfg.Format.Time, "15:04")
	}
	if cfg.Format.TagSymbols != "@#" {
		t.Errorf("TagSymbols: got %q, want %q", cfg.Format.TagSymbols, "@#")
	}
	if cfg.Colors.Date != "blue" {
		t.Errorf("Colors.Date: got %q, want blue", cfg.Colors.Date)
	}
	if cfg.Colors.Tags != "cyan" {
		t.Errorf("Colors.Tags: got %q, want cyan", cfg.Colors.Tags)
	}
	if cfg.Journals["default"].Path != "/tmp/myjournal" {
		t.Errorf("Journal path: got %q, want /tmp/myjournal", cfg.Journals["default"].Path)
	}
}

func TestDefaultJournalPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	jp := cfg.JournalPath()
	if jp == "" {
		t.Error("JournalPath: got empty string")
	}
}

func TestJournalPathExpansion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := "[journals.default]\npath = \"~/Documents/Journal\"\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	jp := cfg.JournalPath()
	if len(jp) == 0 || jp[0] == '~' {
		t.Errorf("JournalPath not expanded: %q", jp)
	}
}

func TestEditorResolution(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	os.Unsetenv("VISUAL")
	os.Unsetenv("EDITOR")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	_ = cfg.Editor()

	os.Setenv("VISUAL", "vim")
	os.Setenv("EDITOR", "nano")
	cfg.General.Editor = ""
	if e := cfg.Editor(); e != "vim" {
		t.Errorf("Editor with VISUAL set: got %q, want vim", e)
	}

	cfg.General.Editor = "micro"
	if e := cfg.Editor(); e != "micro" {
		t.Errorf("Editor with config set: got %q, want micro", e)
	}

	os.Unsetenv("VISUAL")
	os.Unsetenv("EDITOR")
}

func TestLoadMissingParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.toml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load with missing parent dir: %v", err)
	}
	if cfg.General.DefaultListCount != 10 {
		t.Errorf("DefaultListCount: got %d, want 10", cfg.General.DefaultListCount)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("config file not created: %v", err)
	}
}
