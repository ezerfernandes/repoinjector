package injector

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ezer/repoinjector/internal/config"
	"github.com/ezer/repoinjector/internal/gitutil"
)

type Result struct {
	Item   config.Item
	Action string // "created", "updated", "skipped", "error"
	Detail string
}

type Options struct {
	DryRun bool
	Force  bool
	Mode   config.InjectionMode
}

func Inject(cfg *config.Config, targetDir string, opts Options) ([]Result, error) {
	targetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve target path: %w", err)
	}

	if !gitutil.IsGitRepo(targetDir) {
		return nil, fmt.Errorf("%s is not a git repository", targetDir)
	}

	gitDir, err := gitutil.FindGitDir(targetDir)
	if err != nil {
		return nil, err
	}

	sourceDir, err := filepath.Abs(cfg.SourceDir)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve source path: %w", err)
	}

	if sourceDir == targetDir {
		return nil, fmt.Errorf("source and target are the same directory: %s", targetDir)
	}

	mode := opts.Mode
	if mode == "" {
		mode = cfg.Mode
	}

	var results []Result
	var excludePaths []string

	for _, item := range cfg.EnabledItems() {
		src := filepath.Join(sourceDir, item.SourcePath)
		dst := filepath.Join(targetDir, item.TargetPath)

		// Check source exists
		srcInfo, err := os.Stat(src)
		if err != nil {
			results = append(results, Result{Item: item, Action: "skipped", Detail: fmt.Sprintf("source not found: %s", src)})
			continue
		}

		if opts.DryRun {
			action := "symlink"
			if mode == config.ModeCopy {
				action = "copy"
			}
			results = append(results, Result{Item: item, Action: "dry-run", Detail: fmt.Sprintf("would %s %s -> %s", action, src, dst)})
			excludePaths = append(excludePaths, item.TargetPath)
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			results = append(results, Result{Item: item, Action: "error", Detail: fmt.Sprintf("cannot create parent dir: %v", err)})
			continue
		}

		var result Result
		if mode == config.ModeSymlink {
			if srcInfo.IsDir() {
				result = symlinkDir(item, src, dst, opts.Force)
			} else {
				result = symlinkFile(item, src, dst, opts.Force)
			}
		} else {
			if srcInfo.IsDir() {
				result = copyDir(item, src, dst, opts.Force)
			} else {
				result = copyFile(item, src, dst, opts.Force)
			}
		}

		results = append(results, result)
		excludePaths = append(excludePaths, item.TargetPath)
	}

	// Update .git/info/exclude
	if !opts.DryRun && len(excludePaths) > 0 {
		if err := UpdateExclude(gitDir, excludePaths); err != nil {
			return results, fmt.Errorf("failed to update .git/info/exclude: %w", err)
		}
	}

	return results, nil
}

func Eject(cfg *config.Config, targetDir string) ([]Result, error) {
	targetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve target path: %w", err)
	}

	gitDir, err := gitutil.FindGitDir(targetDir)
	if err != nil {
		return nil, err
	}

	var results []Result

	for _, item := range cfg.EnabledItems() {
		dst := filepath.Join(targetDir, item.TargetPath)

		info, err := os.Lstat(dst)
		if err != nil {
			results = append(results, Result{Item: item, Action: "skipped", Detail: "not present"})
			continue
		}

		if info.Mode()&os.ModeSymlink != 0 {
			if err := os.Remove(dst); err != nil {
				results = append(results, Result{Item: item, Action: "error", Detail: fmt.Sprintf("cannot remove symlink: %v", err)})
			} else {
				results = append(results, Result{Item: item, Action: "removed", Detail: "symlink removed"})
			}
		} else if info.IsDir() {
			// Only remove if it's a symlinked dir (already handled above via Lstat)
			// For copied directories, remove them
			if err := os.RemoveAll(dst); err != nil {
				results = append(results, Result{Item: item, Action: "error", Detail: fmt.Sprintf("cannot remove directory: %v", err)})
			} else {
				results = append(results, Result{Item: item, Action: "removed", Detail: "directory removed"})
			}
		} else {
			if err := os.Remove(dst); err != nil {
				results = append(results, Result{Item: item, Action: "error", Detail: fmt.Sprintf("cannot remove file: %v", err)})
			} else {
				results = append(results, Result{Item: item, Action: "removed", Detail: "file removed"})
			}
		}
	}

	// Clean up empty parent directories (.claude/ if empty)
	for _, item := range cfg.EnabledItems() {
		dst := filepath.Join(targetDir, item.TargetPath)
		parent := filepath.Dir(dst)
		if parent != targetDir {
			removeIfEmptyDir(parent)
		}
	}

	if err := CleanExclude(gitDir); err != nil {
		return results, fmt.Errorf("failed to clean .git/info/exclude: %w", err)
	}

	return results, nil
}

type ItemStatus struct {
	Item     config.Item
	Present  bool
	Current  bool   // symlink points to correct source, or copy matches
	Excluded bool
	Detail   string
}

func Status(cfg *config.Config, targetDir string) ([]ItemStatus, error) {
	targetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve target path: %w", err)
	}

	gitDir, err := gitutil.FindGitDir(targetDir)
	if err != nil {
		return nil, err
	}

	sourceDir, err := filepath.Abs(cfg.SourceDir)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve source path: %w", err)
	}

	excludedPaths := GetExcludedPaths(gitDir)
	excludeSet := make(map[string]bool)
	for _, p := range excludedPaths {
		excludeSet[p] = true
	}

	var statuses []ItemStatus

	for _, item := range cfg.EnabledItems() {
		src := filepath.Join(sourceDir, item.SourcePath)
		dst := filepath.Join(targetDir, item.TargetPath)

		status := ItemStatus{
			Item:     item,
			Excluded: excludeSet[item.TargetPath],
		}

		info, err := os.Lstat(dst)
		if err != nil {
			status.Detail = "not present"
			statuses = append(statuses, status)
			continue
		}

		status.Present = true

		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(dst)
			if err == nil && target == src {
				status.Current = true
				status.Detail = "symlink ok"
			} else if err == nil {
				status.Detail = fmt.Sprintf("symlink points to %s (expected %s)", target, src)
			} else {
				status.Detail = "cannot read symlink"
			}
		} else {
			status.Detail = "regular file/dir (not a symlink)"
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

func symlinkFile(item config.Item, src, dst string, force bool) Result {
	return createSymlink(item, src, dst, force)
}

func symlinkDir(item config.Item, src, dst string, force bool) Result {
	return createSymlink(item, src, dst, force)
}

func createSymlink(item config.Item, src, dst string, force bool) Result {
	existing, err := os.Readlink(dst)
	if err == nil {
		if existing == src {
			return Result{Item: item, Action: "skipped", Detail: "already up to date"}
		}
		// Symlink exists but points elsewhere — remove and recreate
		os.Remove(dst)
	} else {
		// Check if a regular file/dir exists
		if _, statErr := os.Lstat(dst); statErr == nil {
			if !force {
				return Result{Item: item, Action: "skipped", Detail: "regular file exists (use --force to overwrite)"}
			}
			os.RemoveAll(dst)
		}
	}

	if err := os.Symlink(src, dst); err != nil {
		return Result{Item: item, Action: "error", Detail: fmt.Sprintf("cannot create symlink: %v", err)}
	}

	return Result{Item: item, Action: "created", Detail: fmt.Sprintf("symlinked -> %s", src)}
}

func copyFile(item config.Item, src, dst string, force bool) Result {
	if _, err := os.Lstat(dst); err == nil {
		if !force {
			// Check if content matches
			if filesEqual(src, dst) {
				return Result{Item: item, Action: "skipped", Detail: "already up to date"}
			}
			return Result{Item: item, Action: "skipped", Detail: "file exists with different content (use --force to overwrite)"}
		}
		os.Remove(dst)
	}

	if err := copyFileContent(src, dst); err != nil {
		return Result{Item: item, Action: "error", Detail: fmt.Sprintf("cannot copy: %v", err)}
	}

	return Result{Item: item, Action: "created", Detail: "copied"}
}

func copyDir(item config.Item, src, dst string, force bool) Result {
	if _, err := os.Lstat(dst); err == nil {
		if !force {
			return Result{Item: item, Action: "skipped", Detail: "directory exists (use --force to overwrite)"}
		}
		os.RemoveAll(dst)
	}

	if err := copyDirRecursive(src, dst); err != nil {
		return Result{Item: item, Action: "error", Detail: fmt.Sprintf("cannot copy directory: %v", err)}
	}

	return Result{Item: item, Action: "created", Detail: "copied"}
}

func copyFileContent(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDirRecursive(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		return copyFileContent(path, target)
	})
}

func filesEqual(a, b string) bool {
	dataA, errA := os.ReadFile(a)
	dataB, errB := os.ReadFile(b)
	if errA != nil || errB != nil {
		return false
	}
	return strings.TrimSpace(string(dataA)) == strings.TrimSpace(string(dataB))
}

func removeIfEmptyDir(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	if len(entries) == 0 {
		os.Remove(dir)
	}
}
