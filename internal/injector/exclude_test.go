package injector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupGitDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "info"), 0755)
	return dir
}

func TestUpdateExclude(t *testing.T) {
	gitDir := setupGitDir(t)
	paths := []string{".claude/skills", ".envrc", ".env"}

	if err := UpdateExclude(gitDir, paths); err != nil {
		t.Fatalf("UpdateExclude failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(gitDir, "info", "exclude"))
	if err != nil {
		t.Fatalf("cannot read exclude file: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, excludeMarkerStart) {
		t.Error("missing start marker")
	}
	if !strings.Contains(s, excludeMarkerEnd) {
		t.Error("missing end marker")
	}
	if !strings.Contains(s, ".claude/skills") {
		t.Error("missing .claude/skills path")
	}
	if !strings.Contains(s, ".envrc") {
		t.Error("missing .envrc path")
	}
}

func TestUpdateExcludeIdempotent(t *testing.T) {
	gitDir := setupGitDir(t)
	paths := []string{".envrc"}

	UpdateExclude(gitDir, paths)
	UpdateExclude(gitDir, paths)

	content, _ := os.ReadFile(filepath.Join(gitDir, "info", "exclude"))
	count := strings.Count(string(content), excludeMarkerStart)
	if count != 1 {
		t.Errorf("expected 1 managed block, found %d", count)
	}
}

func TestUpdateExcludePreservesExisting(t *testing.T) {
	gitDir := setupGitDir(t)

	existing := "# existing patterns\n*.log\n"
	os.WriteFile(filepath.Join(gitDir, "info", "exclude"), []byte(existing), 0644)

	UpdateExclude(gitDir, []string{".env"})

	content, _ := os.ReadFile(filepath.Join(gitDir, "info", "exclude"))
	s := string(content)
	if !strings.Contains(s, "*.log") {
		t.Error("existing content was not preserved")
	}
	if !strings.Contains(s, ".env") {
		t.Error("new path was not added")
	}
}

func TestCleanExclude(t *testing.T) {
	gitDir := setupGitDir(t)

	existing := "# existing\n*.log\n"
	os.WriteFile(filepath.Join(gitDir, "info", "exclude"), []byte(existing), 0644)

	UpdateExclude(gitDir, []string{".env"})
	CleanExclude(gitDir)

	content, _ := os.ReadFile(filepath.Join(gitDir, "info", "exclude"))
	s := string(content)
	if strings.Contains(s, excludeMarkerStart) {
		t.Error("managed block was not removed")
	}
	if !strings.Contains(s, "*.log") {
		t.Error("existing content was removed")
	}
}

func TestGetExcludedPaths(t *testing.T) {
	gitDir := setupGitDir(t)
	UpdateExclude(gitDir, []string{".envrc", ".env"})

	paths := GetExcludedPaths(gitDir)
	if len(paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(paths))
	}
}

func TestHasManagedBlock(t *testing.T) {
	gitDir := setupGitDir(t)

	if HasManagedBlock(gitDir) {
		t.Error("should not have managed block initially")
	}

	UpdateExclude(gitDir, []string{".env"})

	if !HasManagedBlock(gitDir) {
		t.Error("should have managed block after update")
	}
}
