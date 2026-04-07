package e2e

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigAutoCreate(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")

	journalDir := filepath.Join(dir, "journal")
	if err := os.MkdirAll(journalDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	env := testEnv{journalDir: journalDir, configPath: configPath}
	// list will auto-create config, but default journal path won't match our temp dir
	// Just verify config file is created
	_, _ = env.run(t, "list", "--all")

	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("config not auto-created: %v", err)
	}
}

func TestConfigFileFlag(t *testing.T) {
	env := newTestEnv(t)
	out, err := env.run(t, "list", "--all")
	if err != nil {
		t.Fatalf("list with --config-file: %v", err)
	}
	_ = out
}
