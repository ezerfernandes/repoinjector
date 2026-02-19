package cmd

import (
	"fmt"

	"github.com/ezer/repoinjector/internal/gitutil"
	"github.com/ezer/repoinjector/internal/repoconfig"
	"github.com/spf13/cobra"
)

var setStateCmd = &cobra.Command{
	Use:   "set-state <state>",
	Short: "Set the workflow state for the current branch repo",
	Long: `Set a workflow state label for the current repository. The state is stored
in .git/repoinjector/config.yaml and displayed by the "branches" command.

Predefined states: active, review, done, paused.
Custom states are also accepted (lowercase letters, digits, hyphens).

Use "set-state --clear" to remove the state.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSetState,
}

var setStateClear bool

func init() {
	rootCmd.AddCommand(setStateCmd)
	setStateCmd.Flags().BoolVar(&setStateClear, "clear", false, "remove the workflow state")
}

func runSetState(cmd *cobra.Command, args []string) error {
	if !setStateClear && len(args) == 0 {
		return fmt.Errorf("provide a state name, or use --clear to remove")
	}

	repoRoot, err := gitutil.RunGit(".", "rev-parse", "--show-toplevel")
	if err != nil {
		return fmt.Errorf("not inside a git repository")
	}

	gitDir, err := gitutil.FindGitDir(repoRoot)
	if err != nil {
		return err
	}

	cfg, err := repoconfig.Load(gitDir)
	if err != nil {
		return err
	}
	if cfg == nil {
		cfg = &repoconfig.RepoConfig{Version: 1}
	}

	if setStateClear {
		cfg.State = ""
		if err := repoconfig.Save(gitDir, cfg); err != nil {
			return err
		}
		fmt.Println("State cleared.")
		return nil
	}

	state := args[0]
	if err := repoconfig.ValidateState(state); err != nil {
		return err
	}

	cfg.State = state
	if err := repoconfig.Save(gitDir, cfg); err != nil {
		return err
	}

	fmt.Printf("State set to: %s\n", state)
	return nil
}
