package gitutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindGitDir returns the path to the .git directory for a repository.
// Handles both regular repos (.git is a directory) and worktrees (.git is a file).
func FindGitDir(repoRoot string) (string, error) {
	gitPath := filepath.Join(repoRoot, ".git")
	info, err := os.Lstat(gitPath)
	if err != nil {
		return "", fmt.Errorf("not a git repository (no .git found): %w", err)
	}

	if info.IsDir() {
		return gitPath, nil
	}

	// .git is a file — worktree pointer
	content, err := os.ReadFile(gitPath)
	if err != nil {
		return "", fmt.Errorf("cannot read .git file: %w", err)
	}

	line := strings.TrimSpace(string(content))
	if !strings.HasPrefix(line, "gitdir: ") {
		return "", fmt.Errorf("unexpected .git file format: %s", line)
	}

	gitdir := strings.TrimPrefix(line, "gitdir: ")
	if !filepath.IsAbs(gitdir) {
		gitdir = filepath.Join(repoRoot, gitdir)
	}

	return filepath.Clean(gitdir), nil
}

// IsGitRepo checks if the given directory is a git repository.
func IsGitRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

// FindGitRepos returns all immediate subdirectories of parentDir that are git repos.
func FindGitRepos(parentDir string) ([]string, error) {
	entries, err := os.ReadDir(parentDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read directory %s: %w", parentDir, err)
	}

	var repos []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		candidate := filepath.Join(parentDir, e.Name())
		if IsGitRepo(candidate) {
			repos = append(repos, candidate)
		}
	}
	return repos, nil
}
