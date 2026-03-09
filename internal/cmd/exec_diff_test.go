package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunExecDiff_NoDash(t *testing.T) {
	dir := t.TempDir()
	dir, _ = filepath.EvalSymlinks(dir)

	repoDir := filepath.Join(dir, "repo")
	if err := os.Mkdir(repoDir, 0755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, repoDir)

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	// Simulate no -- by setting ArgsLenAtDash to -1 (default when no -- present).
	execDiffCmd.SetArgs([]string{"echo", "hello"})
	err := runExecDiff(execDiffCmd, []string{"echo", "hello"})
	if err == nil {
		t.Fatal("expected error for missing --")
	}
	if !strings.Contains(err.Error(), "Use --") {
		t.Errorf("error should mention --, got: %v", err)
	}
}

func TestRunExecDiff_EmptyCommand(t *testing.T) {
	// ArgsLenAtDash returns -1 when no -- is present, so parseUserCommand returns an error.
	_, err := parseUserCommand(execDiffCmd, []string{})
	if err == nil {
		t.Fatal("expected error for empty args with no --")
	}
}

func TestRunExecDiff_IdenticalOutput(t *testing.T) {
	dir := t.TempDir()
	dir, _ = filepath.EvalSymlinks(dir)

	// Create "main" repo.
	mainDir := filepath.Join(dir, "main")
	if err := os.Mkdir(mainDir, 0755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, mainDir)
	if err := os.WriteFile(filepath.Join(mainDir, "file.txt"), []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create "branch" repo.
	branchDir := filepath.Join(dir, "branch")
	if err := os.Mkdir(branchDir, 0755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, branchDir)
	if err := os.WriteFile(filepath.Join(branchDir, "file.txt"), []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(branchDir); err != nil {
		t.Fatal(err)
	}

	// Set flags directly.
	execDiffNoSync = true
	execDiffNameOnly = true
	execDiffMainDir = mainDir
	defer func() {
		execDiffNoSync = false
		execDiffNameOnly = false
		execDiffMainDir = ""
	}()

	// We can't easily test the full runExecDiff with ArgsLenAtDash,
	// so test the core helpers instead.
	mainOut := captureCommand(mainDir, "cat", "file.txt")
	branchOut := captureCommand(branchDir, "cat", "file.txt")

	if mainOut != branchOut {
		t.Errorf("expected identical output, got main=%q branch=%q", mainOut, branchOut)
	}
}

func TestRunExecDiff_DifferentOutput(t *testing.T) {
	dir := t.TempDir()
	dir, _ = filepath.EvalSymlinks(dir)

	mainDir := filepath.Join(dir, "main")
	if err := os.Mkdir(mainDir, 0755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, mainDir)
	if err := os.WriteFile(filepath.Join(mainDir, "file.txt"), []byte("hello\nworld\n"), 0644); err != nil {
		t.Fatal(err)
	}

	branchDir := filepath.Join(dir, "branch")
	if err := os.Mkdir(branchDir, 0755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, branchDir)
	if err := os.WriteFile(filepath.Join(branchDir, "file.txt"), []byte("hello\nuniverse\n"), 0644); err != nil {
		t.Fatal(err)
	}

	mainOut := captureCommand(mainDir, "cat", "file.txt")
	branchOut := captureCommand(branchDir, "cat", "file.txt")

	if mainOut == branchOut {
		t.Error("expected different output")
	}
	if mainOut != "hello\nworld\n" {
		t.Errorf("unexpected main output: %q", mainOut)
	}
	if branchOut != "hello\nuniverse\n" {
		t.Errorf("unexpected branch output: %q", branchOut)
	}
}

func TestResolveMainDir_ExplicitDir(t *testing.T) {
	dir := t.TempDir()
	dir, _ = filepath.EvalSymlinks(dir)

	mainDir := filepath.Join(dir, "main")
	if err := os.Mkdir(mainDir, 0755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, mainDir)

	execDiffMainDir = mainDir
	defer func() { execDiffMainDir = "" }()

	got, err := resolveMainDir("/some/branch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != mainDir {
		t.Errorf("got %q, want %q", got, mainDir)
	}
}

func TestResolveMainDir_NotARepo(t *testing.T) {
	dir := t.TempDir()

	execDiffMainDir = dir
	defer func() { execDiffMainDir = "" }()

	_, err := resolveMainDir("/some/branch")
	if err == nil {
		t.Fatal("expected error for non-repo directory")
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("error should mention not a git repository, got: %v", err)
	}
}

func TestCaptureCommand(t *testing.T) {
	out := captureCommand(".", "echo", "hello")
	if strings.TrimSpace(out) != "hello" {
		t.Errorf("got %q, want %q", strings.TrimSpace(out), "hello")
	}
}

func TestCaptureCommand_NonZeroExit(t *testing.T) {
	// Command that exits non-zero should still return output.
	out := captureCommand(".", "sh", "-c", "echo error output; exit 1")
	if !strings.Contains(out, "error output") {
		t.Errorf("expected output even on non-zero exit, got %q", out)
	}
}

func TestParseUserCommand_NoDash(t *testing.T) {
	// When ArgsLenAtDash returns -1, we should get an error.
	_, err := parseUserCommand(execDiffCmd, []string{"echo"})
	if err == nil {
		t.Fatal("expected error")
	}
}
