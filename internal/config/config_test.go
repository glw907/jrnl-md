package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()

	if cfg.General.Editor != "" {
		t.Errorf("expected empty editor, got %q", cfg.General.Editor)
	}
	if cfg.General.DefaultHour != 9 {
		t.Errorf("expected default_hour 9, got %d", cfg.General.DefaultHour)
	}
	if cfg.General.DefaultMinute != 0 {
		t.Errorf("expected default_minute 0, got %d", cfg.General.DefaultMinute)
	}
	if cfg.General.Linewrap != 79 {
		t.Errorf("expected linewrap 79, got %d", cfg.General.Linewrap)
	}
	if cfg.General.IndentCharacter != "|" {
		t.Errorf("expected indent_character |, got %q", cfg.General.IndentCharacter)
	}
	if cfg.Format.Time != "03:04 PM" {
		t.Errorf("expected time '03:04 PM', got %q", cfg.Format.Time)
	}
	if cfg.Format.Date != "2006-01-02" {
		t.Errorf("expected date '2006-01-02', got %q", cfg.Format.Date)
	}
	if cfg.Format.TagSymbols != "@" {
		t.Errorf("expected tag_symbols @, got %q", cfg.Format.TagSymbols)
	}
	if cfg.Format.FileExtension != "md" {
		t.Errorf("expected file_extension md, got %q", cfg.Format.FileExtension)
	}
}

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[general]
editor = "hx"
linewrap = 72
indent_character = ">"

[format]
time = "15:04"
tag_symbols = "#@"
file_extension = "txt"

[journals.default]
path = "/tmp/journal"

[journals.work]
path = "/tmp/work"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.General.Editor != "hx" {
		t.Errorf("expected editor hx, got %q", cfg.General.Editor)
	}
	if cfg.General.Linewrap != 72 {
		t.Errorf("expected linewrap 72, got %d", cfg.General.Linewrap)
	}
	if cfg.General.IndentCharacter != ">" {
		t.Errorf("expected indent_character >, got %q", cfg.General.IndentCharacter)
	}
	if cfg.Format.Time != "15:04" {
		t.Errorf("expected time 15:04, got %q", cfg.Format.Time)
	}
	if cfg.Format.TagSymbols != "#@" {
		t.Errorf("expected tag_symbols #@, got %q", cfg.Format.TagSymbols)
	}
	if cfg.Format.FileExtension != "txt" {
		t.Errorf("expected file_extension txt, got %q", cfg.Format.FileExtension)
	}

	if j, ok := cfg.Journals["default"]; !ok || j.Path != "/tmp/journal" {
		t.Errorf("expected default journal at /tmp/journal, got %+v", j)
	}
	if j, ok := cfg.Journals["work"]; !ok || j.Path != "/tmp/work" {
		t.Errorf("expected work journal at /tmp/work, got %+v", j)
	}
}

func TestDefaultPath(t *testing.T) {
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath failed: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty config path")
	}
	if filepath.Base(path) != "config.toml" {
		t.Errorf("expected config.toml, got %q", filepath.Base(path))
	}
}

func TestResolvedJournalConfig(t *testing.T) {
	global := Default()
	global.General.Editor = "nano"
	global.General.Template = ""
	global.Format.TagSymbols = "@"

	t.Run("override all three fields", func(t *testing.T) {
		j := JournalConfig{
			Path:       "/tmp/j",
			Editor:     "vim",
			Template:   "/tmp/template.md",
			TagSymbols: "#",
		}
		resolved := ResolvedJournalConfig(global, j)
		if resolved.General.Editor != "vim" {
			t.Errorf("expected editor vim, got %q", resolved.General.Editor)
		}
		if resolved.General.Template != "/tmp/template.md" {
			t.Errorf("expected template /tmp/template.md, got %q", resolved.General.Template)
		}
		if resolved.Format.TagSymbols != "#" {
			t.Errorf("expected tag_symbols #, got %q", resolved.Format.TagSymbols)
		}
	})

	t.Run("override none — global preserved", func(t *testing.T) {
		j := JournalConfig{Path: "/tmp/j"}
		resolved := ResolvedJournalConfig(global, j)
		if resolved.General.Editor != "nano" {
			t.Errorf("expected editor nano, got %q", resolved.General.Editor)
		}
		if resolved.Format.TagSymbols != "@" {
			t.Errorf("expected tag_symbols @, got %q", resolved.Format.TagSymbols)
		}
	})

	t.Run("override editor only", func(t *testing.T) {
		j := JournalConfig{Path: "/tmp/j", Editor: "hx"}
		resolved := ResolvedJournalConfig(global, j)
		if resolved.General.Editor != "hx" {
			t.Errorf("expected editor hx, got %q", resolved.General.Editor)
		}
		if resolved.Format.TagSymbols != "@" {
			t.Errorf("tag_symbols should not change, got %q", resolved.Format.TagSymbols)
		}
	})

	t.Run("global is not mutated", func(t *testing.T) {
		j := JournalConfig{Path: "/tmp/j", Editor: "vim", TagSymbols: "#"}
		_ = ResolvedJournalConfig(global, j)
		if global.General.Editor != "nano" {
			t.Errorf("global should not be mutated, editor is now %q", global.General.Editor)
		}
	})
}
