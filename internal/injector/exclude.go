package injector

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	excludeMarkerStart = "# BEGIN repomni managed block"
	excludeMarkerEnd   = "# END repomni managed block"
)

// UpdateExclude writes the given paths into a managed block in .git/info/exclude.
// Existing managed block content is replaced. Other content is preserved.
func UpdateExclude(gitDir string, paths []string) error {
	excludePath := filepath.Join(gitDir, "info", "exclude")

	// Ensure the info directory exists
	if err := os.MkdirAll(filepath.Join(gitDir, "info"), 0755); err != nil {
		return err
	}

	content, _ := os.ReadFile(excludePath)

	cleaned := removeManagedBlock(string(content))

	block := excludeMarkerStart + "\n"
	for _, p := range paths {
		block += p + "\n"
	}
	block += excludeMarkerEnd + "\n"

	newContent := strings.TrimRight(cleaned, "\n")
	if newContent != "" {
		newContent += "\n\n"
	}
	newContent += block

	return os.WriteFile(excludePath, []byte(newContent), 0644)
}

// CleanExclude removes the managed block from .git/info/exclude.
func CleanExclude(gitDir string) error {
	excludePath := filepath.Join(gitDir, "info", "exclude")

	content, err := os.ReadFile(excludePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	cleaned := removeManagedBlock(string(content))
	return os.WriteFile(excludePath, []byte(strings.TrimRight(cleaned, "\n")+"\n"), 0644)
}

// HasManagedBlock checks if .git/info/exclude contains a repomni managed block.
func HasManagedBlock(gitDir string) bool {
	excludePath := filepath.Join(gitDir, "info", "exclude")
	content, err := os.ReadFile(excludePath)
	if err != nil {
		return false
	}
	return strings.Contains(string(content), excludeMarkerStart)
}

// GetExcludedPaths returns the paths in the managed block.
func GetExcludedPaths(gitDir string) []string {
	excludePath := filepath.Join(gitDir, "info", "exclude")
	content, err := os.ReadFile(excludePath)
	if err != nil {
		return nil
	}

	var paths []string
	inBlock := false
	for _, line := range strings.Split(string(content), "\n") {
		if line == excludeMarkerStart {
			inBlock = true
			continue
		}
		if line == excludeMarkerEnd {
			break
		}
		if inBlock && strings.TrimSpace(line) != "" {
			paths = append(paths, line)
		}
	}
	return paths
}

func removeManagedBlock(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inBlock := false

	for _, line := range lines {
		if line == excludeMarkerStart {
			inBlock = true
			continue
		}
		if line == excludeMarkerEnd {
			inBlock = false
			continue
		}
		if !inBlock {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
