// Package config loads and saves jrnl-md configuration from a TOML file.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config holds all jrnl-md configuration values.
type Config struct {
	General  General            `toml:"general"`
	Format   Format             `toml:"format"`
	Colors   Colors             `toml:"colors"`
	Journals map[string]Journal `toml:"journals"`
}

// General holds general behavioral settings.
type General struct {
	Editor           string `toml:"editor"`
	Timestamps       bool   `toml:"timestamps"`
	Linewrap         int    `toml:"linewrap"`
	DefaultListCount int    `toml:"default_list_count"`
}

// Format holds display format settings.
type Format struct {
	Time       string `toml:"time"`
	Date       string `toml:"date"`
	TagSymbols string `toml:"tag_symbols"`
}

// Colors holds color settings for display elements.
type Colors struct {
	Date string `toml:"date"`
	Body string `toml:"body"`
	Tags string `toml:"tags"`
}

// Journal holds per-journal settings.
type Journal struct {
	Path string `toml:"path"`
}

// defaults returns a Config with all default values filled in.
func defaults() Config {
	home, _ := os.UserHomeDir()
	return Config{
		General: General{
			Editor:           "",
			Timestamps:       true,
			Linewrap:         79,
			DefaultListCount: 10,
		},
		Format: Format{
			Time:       "03:04 PM",
			Date:       "2006-01-02",
			TagSymbols: "@",
		},
		Colors: Colors{
			Date: "none",
			Body: "none",
			Tags: "none",
		},
		Journals: map[string]Journal{
			"default": {Path: filepath.Join(home, "Documents", "Journal")},
		},
	}
}

// Load reads the config file at path. If the file does not exist, it
// is created with defaults. Parent directories are created as needed.
func Load(path string) (Config, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return Config{}, fmt.Errorf("creating config dir: %w", err)
	}

	cfg := defaults()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := save(path, cfg); err != nil {
			return Config{}, fmt.Errorf("auto-creating config: %w", err)
		}
		return cfg, nil
	}

	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Config{}, fmt.Errorf("decoding config %s: %w", path, err)
	}
	return cfg, nil
}

// save writes cfg to path using atomic write.
func save(path string, cfg Config) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".config-*.toml")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	if err := toml.NewEncoder(tmp).Encode(cfg); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("encoding config: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("syncing config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming config file: %w", err)
	}
	return nil
}

// JournalPath returns the resolved (tilde-expanded) path for the default journal.
func (c Config) JournalPath() string {
	j, ok := c.Journals["default"]
	if !ok {
		return ""
	}
	return expandTilde(j.Path)
}

// Editor returns the editor to use: config → $VISUAL → $EDITOR.
func (c Config) Editor() string {
	if c.General.Editor != "" {
		return c.General.Editor
	}
	if v := os.Getenv("VISUAL"); v != "" {
		return v
	}
	return os.Getenv("EDITOR")
}

// expandTilde replaces a leading ~ with the user's home directory.
func expandTilde(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}

// DefaultPath returns the default config file path (~/.config/jrnl-md/config.toml).
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "jrnl-md", "config.toml")
}
