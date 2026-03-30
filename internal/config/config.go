package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/glw907/jrnl-md/internal/atomicfile"
)

type Config struct {
	General  GeneralConfig            `toml:"general"`
	Format   FormatConfig             `toml:"format"`
	Colors   ColorConfig              `toml:"colors"`
	Journals map[string]JournalConfig `toml:"journals"`
}

type GeneralConfig struct {
	Editor          string `toml:"editor"`
	Encrypt         bool   `toml:"encrypt"`
	DefaultHour     int    `toml:"default_hour"`
	DefaultMinute   int    `toml:"default_minute"`
	Highlight       bool   `toml:"highlight"`
	Linewrap        int    `toml:"linewrap"`
	IndentCharacter string `toml:"indent_character"`
	Template        string `toml:"template"`
}

type FormatConfig struct {
	Time          string `toml:"time"`
	Date          string `toml:"date"`
	TagSymbols    string `toml:"tag_symbols"`
	FileExtension string `toml:"file_extension"`
}

type ColorConfig struct {
	Date string `toml:"date"`
	Body string `toml:"body"`
	Tags string `toml:"tags"`
}

type JournalConfig struct {
	Path       string `toml:"path"`
	Encrypt    *bool  `toml:"encrypt,omitempty"`
	Editor     string `toml:"editor,omitempty"`
	Template   string `toml:"template,omitempty"`
	TagSymbols string `toml:"tag_symbols,omitempty"`
}

func Default() Config {
	return Config{
		General: GeneralConfig{
			Editor:          "",
			Encrypt:         false,
			DefaultHour:     9,
			DefaultMinute:   0,
			Highlight:       true,
			Linewrap:        79,
			IndentCharacter: "|",
			Template:        "",
		},
		Format: FormatConfig{
			Time:          "03:04 PM",
			Date:          "2006-01-02",
			TagSymbols:    "@",
			FileExtension: "md",
		},
		Colors: ColorConfig{
			Date: "none",
			Body: "none",
			Tags: "none",
		},
		Journals: map[string]JournalConfig{},
	}
}

// ResolvedJournalConfig returns a copy of global with journal-specific overrides applied.
// Non-empty journal fields take precedence over global values.
func ResolvedJournalConfig(global Config, j JournalConfig) Config {
	result := global
	if j.Editor != "" {
		result.General.Editor = j.Editor
	}
	if j.Template != "" {
		result.General.Template = j.Template
	}
	if j.TagSymbols != "" {
		result.Format.TagSymbols = j.TagSymbols
	}
	return result
}

func Load(path string) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("reading config: %w", err)
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

func Save(cfg Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	if err := atomicfile.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	return nil
}

func DefaultPath() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("determining config directory: %w", err)
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "jrnl-md", "config.toml"), nil
}
