package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ezerfernandes/repomni/internal/brancher"
	"github.com/ezerfernandes/repomni/internal/diffutil"
	"github.com/ezerfernandes/repomni/internal/gitutil"
	"github.com/spf13/cobra"
)

var (
	execDiffNoSync   bool
	execDiffNameOnly bool
	execDiffMainDir  string
)

var execDiffCmd = &cobra.Command{
	Use:   "diff [flags] -- <command> [args...]",
	Short: "Diff command output between main and branch repos",
	Long: `Run a command in both the main (parent) repo and the current branch repo,
then display a unified diff of their outputs.

Especially useful for linters and test commands to see how many errors the
branch is adding or removing compared to main.

Use -- to separate repomni flags from the command to run:

  repomni exec diff -- make lint
  repomni exec diff --no-sync -- go vet ./...
  repomni exec diff --name-only -- npm test`,
	Args:         cobra.ArbitraryArgs,
	SilenceUsage: true,
	RunE:         runExecDiff,
}

func init() {
	execCmd.AddCommand(execDiffCmd)
	execDiffCmd.Flags().BoolVar(&execDiffNoSync, "no-sync", false, "skip fetch+pull on the main repo")
	execDiffCmd.Flags().BoolVar(&execDiffNameOnly, "name-only", false, "only show whether outputs differ")
	execDiffCmd.Flags().StringVar(&execDiffMainDir, "main-dir", "", "explicit path to the main repo")
}

func runExecDiff(cmd *cobra.Command, args []string) error {
	// Parse the user command from after --.
	userCmd, err := parseUserCommand(cmd, args)
	if err != nil {
		return err
	}

	// Find current (branch) repo root.
	branchDir, err := gitutil.RunGit(".", "rev-parse", "--show-toplevel")
	if err != nil {
		return fmt.Errorf("not inside a git repository")
	}

	// Find main repo directory.
	mainDir, err := resolveMainDir(branchDir)
	if err != nil {
		return err
	}

	// Guard: must not be running from the main repo itself.
	absBranch, _ := filepath.Abs(branchDir)
	absMain, _ := filepath.Abs(mainDir)
	if absBranch == absMain {
		return fmt.Errorf("already in the main repo; run from a branch repo instead")
	}

	// Sync main repo.
	if !execDiffNoSync {
		syncMainRepo(mainDir)
	}

	// Run command in both repos.
	fmt.Fprintf(os.Stderr, "Running command in main repo (%s)...\n", filepath.Base(mainDir))
	mainOut := captureCommand(mainDir, userCmd[0], userCmd[1:]...)

	fmt.Fprintf(os.Stderr, "Running command in branch repo (%s)...\n", filepath.Base(branchDir))
	branchOut := captureCommand(branchDir, userCmd[0], userCmd[1:]...)

	// Display results.
	if execDiffNameOnly {
		fmt.Println(diffutil.SummaryLine(mainOut, branchOut))
		if mainOut != branchOut {
			os.Exit(1)
		}
		return nil
	}

	diff := diffutil.UnifiedDiff("main", "branch", mainOut, branchOut)
	if diff == "" {
		fmt.Println("Outputs are identical")
		return nil
	}

	fmt.Print(diffutil.ColorDiff(diff))
	os.Exit(1)
	return nil
}

// parseUserCommand extracts the command to run from args after the -- separator.
func parseUserCommand(cmd *cobra.Command, args []string) ([]string, error) {
	dashIdx := cmd.ArgsLenAtDash()
	if dashIdx == -1 {
		return nil, fmt.Errorf("usage: repomni exec diff [flags] -- <command> [args...]\n\nUse -- to separate flags from the command to run")
	}
	userCmd := args[dashIdx:]
	if len(userCmd) == 0 {
		return nil, fmt.Errorf("no command specified after --")
	}
	return userCmd, nil
}

// resolveMainDir determines the main repo directory.
func resolveMainDir(branchDir string) (string, error) {
	if execDiffMainDir != "" {
		if !gitutil.IsGitRepo(execDiffMainDir) {
			return "", fmt.Errorf("--main-dir %q is not a git repository", execDiffMainDir)
		}
		return execDiffMainDir, nil
	}

	parentDir := filepath.Dir(branchDir)
	mainDir, err := brancher.FindParentGitRepo(parentDir)
	if err != nil {
		return "", fmt.Errorf("could not find main repo: %w\nUse --main-dir to specify it explicitly", err)
	}
	return mainDir, nil
}

// syncMainRepo fetches and pulls the main repo, printing warnings on failure.
func syncMainRepo(mainDir string) {
	fmt.Fprintf(os.Stderr, "Syncing main repo...\n")
	if err := gitutil.Fetch(mainDir); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: fetch failed: %v\n", err)
		return
	}
	if _, err := gitutil.Pull(mainDir, "ff-only", false); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: pull failed: %v (continuing with current state)\n", err)
	}
}

// captureCommand runs a command in dir and returns its combined stdout+stderr output.
// The command's exit code is ignored since tools like linters return non-zero on findings.
func captureCommand(dir string, name string, args ...string) string {
	c := exec.Command(name, args...)
	c.Dir = dir
	out, _ := c.CombinedOutput()
	return string(out)
}
