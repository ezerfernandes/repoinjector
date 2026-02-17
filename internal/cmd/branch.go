package cmd

import (
	"fmt"
	"os"

	"github.com/ezer/repoinjector/internal/brancher"
	"github.com/ezer/repoinjector/internal/gitutil"
	"github.com/ezer/repoinjector/internal/scripter"
	"github.com/spf13/cobra"
)

var branchCmd = &cobra.Command{
	Use:   "branch <branch-name>",
	Short: "Clone the parent repo and create a new branch",
	Long: `Finds the closest parent directory that is a git repository, clones it
into the current directory using the branch name, and checks out a new branch
with that name.

This is useful for creating isolated working copies for feature branches.`,
	Args: cobra.ExactArgs(1),
	RunE: runBranch,
}

func init() {
	rootCmd.AddCommand(branchCmd)
}

func runBranch(cmd *cobra.Command, args []string) error {
	result, err := brancher.Branch(".", args[0])
	if err != nil {
		return err
	}
	fmt.Printf("Cloned %s into %s\n", result.RemoteURL, result.TargetDir)
	fmt.Printf("Checked out new branch: %s\n", result.Branch)

	// Run setup script if configured in the parent repo.
	gitDir, err := gitutil.FindGitDir(result.ParentRepo)
	if err == nil {
		if _, exists := scripter.GetScript(gitDir, scripter.ScriptSetup); exists {
			fmt.Println("Running setup script...")
			if err := scripter.RunScript(gitDir, scripter.ScriptSetup, result.TargetDir); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: setup script failed: %v\n", err)
			}
		}
	}

	return nil
}
