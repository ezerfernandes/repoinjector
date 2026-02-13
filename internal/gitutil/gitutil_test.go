package gitutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init", dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}
}

func TestFindGitDir(t *testing.T) {
	repo := t.TempDir()
	initGitRepo(t, repo)

	gitDir, err := FindGitDir(repo)
	if err != nil {
		t.Fatalf("FindGitDir failed: %v", err)
	}

	expected := filepath.Join(repo, ".git")
	if gitDir != expected {
		t.Errorf("expected %q, got %q", expected, gitDir)
	}
}

func TestFindGitDirNotARepo(t *testing.T) {
	dir := t.TempDir()
	_, err := FindGitDir(dir)
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestIsGitRepo(t *testing.T) {
	repo := t.TempDir()
	initGitRepo(t, repo)

	if !IsGitRepo(repo) {
		t.Error("expected true for git repo")
	}

	notRepo := t.TempDir()
	if IsGitRepo(notRepo) {
		t.Error("expected false for non-git directory")
	}
}

func TestFindGitRepos(t *testing.T) {
	parent := t.TempDir()

	// Create 2 git repos and 1 regular dir
	initGitRepo(t, filepath.Join(parent, "repo-a"))
	initGitRepo(t, filepath.Join(parent, "repo-b"))
	os.MkdirAll(filepath.Join(parent, "not-a-repo"), 0755)

	repos, err := FindGitRepos(parent)
	if err != nil {
		t.Fatalf("FindGitRepos failed: %v", err)
	}

	if len(repos) != 2 {
		t.Errorf("expected 2 repos, got %d: %v", len(repos), repos)
	}
}

func TestFindGitDirWorktree(t *testing.T) {
	// Create a main repo
	mainRepo := t.TempDir()
	initGitRepo(t, mainRepo)

	// Create an initial commit so we can create worktrees
	cmd := exec.Command("git", "-C", mainRepo, "commit", "--allow-empty", "-m", "init")
	if err := cmd.Run(); err != nil {
		t.Fatalf("git commit failed: %v", err)
	}

	// Create a worktree
	wtDir := filepath.Join(t.TempDir(), "worktree")
	cmd = exec.Command("git", "-C", mainRepo, "worktree", "add", wtDir, "-b", "test-branch")
	if err := cmd.Run(); err != nil {
		t.Fatalf("git worktree add failed: %v", err)
	}

	gitDir, err := FindGitDir(wtDir)
	if err != nil {
		t.Fatalf("FindGitDir on worktree failed: %v", err)
	}

	// gitDir should point inside the main repo's .git/worktrees/
	if gitDir == filepath.Join(wtDir, ".git") {
		t.Errorf("gitDir should not be the worktree .git file, got %q", gitDir)
	}
}
