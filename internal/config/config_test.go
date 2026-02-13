package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Version != 1 {
		t.Errorf("expected version 1, got %d", cfg.Version)
	}
	if cfg.Mode != ModeSymlink {
		t.Errorf("expected mode symlink, got %s", cfg.Mode)
	}
	if len(cfg.Items) != 4 {
		t.Errorf("expected 4 default items, got %d", len(cfg.Items))
	}
}

func TestEnabledItems(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Items[1].Enabled = false
	cfg.Items[3].Enabled = false

	enabled := cfg.EnabledItems()
	if len(enabled) != 2 {
		t.Errorf("expected 2 enabled items, got %d", len(enabled))
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Override config dir for testing
	tmpDir := t.TempDir()
	origFn := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", origFn)

	cfg := DefaultConfig()
	cfg.SourceDir = "/some/test/path"

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	path := filepath.Join(tmpDir, "repoinjector", "config.yaml")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file not created at %s", path)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.SourceDir != cfg.SourceDir {
		t.Errorf("expected source_dir %q, got %q", cfg.SourceDir, loaded.SourceDir)
	}
	if loaded.Mode != cfg.Mode {
		t.Errorf("expected mode %q, got %q", cfg.Mode, loaded.Mode)
	}
	if len(loaded.Items) != len(cfg.Items) {
		t.Errorf("expected %d items, got %d", len(cfg.Items), len(loaded.Items))
	}
}
