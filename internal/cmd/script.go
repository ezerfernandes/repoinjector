package cmd

import (
	"fmt"

	"github.com/ezer/repoinjector/internal/gitutil"
	"github.com/ezer/repoinjector/internal/scripter"
	"github.com/ezer/repoinjector/internal/ui"
	"github.com/spf13/cobra"
)

var scriptCmd = &cobra.Command{
	Use:   "script",
	Short: "Manage setup scripts for this repository",
	Long: `Interactively create or edit a setup script that runs automatically
when you create a new branch for this repository using "repoinjector branch".

The script is stored inside .git and is never committed.`,
	RunE: runScript,
}

func init() {
	rootCmd.AddCommand(scriptCmd)
}

func runScript(cmd *cobra.Command, args []string) error {
	repoRoot, err := gitutil.RunGit(".", "rev-parse", "--show-toplevel")
	if err != nil {
		return fmt.Errorf("not inside a git repository")
	}

	gitDir, err := gitutil.FindGitDir(repoRoot)
	if err != nil {
		return err
	}

	action, err := ui.RunScriptForm(gitDir)
	if err != nil {
		return err
	}

	switch action {
	case ui.ScriptSaved:
		fmt.Printf("\nScript saved to %s\n", scripter.ScriptPath(gitDir, scripter.ScriptSetup))
	case ui.ScriptDeleted:
		fmt.Println("\nSetup script deleted.")
	}

	return nil
}
