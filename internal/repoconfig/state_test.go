package repoconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateState_KnownStates(t *testing.T) {
	for _, s := range KnownStates() {
		if err := ValidateState(string(s)); err != nil {
			t.Errorf("ValidateState(%q) should pass: %v", s, err)
		}
	}
}

func TestValidateState_CustomState(t *testing.T) {
	valid := []string{"my-state", "wip", "ready-for-qa", "v2"}
	for _, s := range valid {
		if err := ValidateState(s); err != nil {
			t.Errorf("ValidateState(%q) should pass: %v", s, err)
		}
	}
}

func TestValidateState_Empty(t *testing.T) {
	if err := ValidateState(""); err == nil {
		t.Error("ValidateState(\"\") should fail")
	}
}

func TestValidateState_InvalidChars(t *testing.T) {
	invalid := []string{"Active", "DONE", "in progress", "state!", "review/done"}
	for _, s := range invalid {
		if err := ValidateState(s); err == nil {
			t.Errorf("ValidateState(%q) should fail", s)
		}
	}
}

func TestIsKnownState(t *testing.T) {
	if !IsKnownState("active") {
		t.Error("active should be a known state")
	}
	if !IsKnownState("review") {
		t.Error("review should be a known state")
	}
	if !IsKnownState("done") {
		t.Error("done should be a known state")
	}
	if !IsKnownState("paused") {
		t.Error("paused should be a known state")
	}
	if IsKnownState("custom") {
		t.Error("custom should NOT be a known state")
	}
	if IsKnownState("") {
		t.Error("empty should NOT be a known state")
	}
}

func TestSaveAndLoad_WithState(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	original := &RepoConfig{
		Version: 1,
		State:   "review",
		Items: []RepoItemConfig{
			{TargetPath: ".envrc", Enabled: true},
		},
	}

	if err := Save(gitDir, original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(gitDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded == nil {
		t.Fatal("Load returned nil")
	}
	if loaded.State != "review" {
		t.Errorf("State mismatch: got %q, want %q", loaded.State, "review")
	}
}

func TestSaveAndLoad_NoState(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	original := &RepoConfig{
		Version: 1,
		Items: []RepoItemConfig{
			{TargetPath: ".envrc", Enabled: true},
		},
	}

	if err := Save(gitDir, original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(gitDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.State != "" {
		t.Errorf("State should be empty, got %q", loaded.State)
	}
}

func TestLoad_LegacyConfig(t *testing.T) {
	// Simulate a config file written before the state field existed.
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	configDir := filepath.Join(gitDir, "repoinjector")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	legacy := []byte("version: 1\nitems:\n  - target_path: .envrc\n    enabled: true\n")
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), legacy, 0644); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(gitDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded == nil {
		t.Fatal("Load returned nil")
	}
	if loaded.State != "" {
		t.Errorf("State should be empty for legacy config, got %q", loaded.State)
	}
	if len(loaded.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(loaded.Items))
	}
}
