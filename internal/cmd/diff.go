package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ezerfernandes/repomni/internal/diffutil"
	"github.com/ezerfernandes/repomni/internal/gitutil"
	"github.com/ezerfernandes/repomni/internal/ui"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff [target]",
	Short: "Show a semantic summary of code changes",
	Long: `Run git diff on a repository and display a semantic summary of changes,
including affected Go functions, types, and import modifications.

If no target is specified, the current directory is used.
Use --all to diff all git repos under the target directory.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDiff,
}

var (
	diffStaged bool
	diffBranch string
	diffAll    bool
	diffJSON   bool
)

func init() {
	rootCmd.AddCommand(diffCmd)
	diffCmd.Flags().BoolVar(&diffStaged, "staged", false, "show staged changes only")
	diffCmd.Flags().StringVar(&diffBranch, "branch", "", "diff against a branch or ref (e.g. main, origin/main)")
	diffCmd.Flags().BoolVar(&diffAll, "all", false, "diff all git repos under target directory")
	diffCmd.Flags().BoolVar(&diffJSON, "json", false, "output as JSON")
}

func runDiff(cmd *cobra.Command, args []string) error {
	target := "."
	if len(args) > 0 {
		target = args[0]
	}
	target, err := filepath.Abs(target)
	if err != nil {
		return err
	}

	var targets []string
	if diffAll {
		targets, err = gitutil.FindGitRepos(target)
		if err != nil {
			return err
		}
		if len(targets) == 0 {
			return fmt.Errorf("no git repositories found under %s", target)
		}
	} else {
		targets = []string{target}
	}

	for i, t := range targets {
		summary, err := diffRepo(t)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error diffing %s: %v\n", t, err)
			continue
		}

		if diffAll && len(targets) > 1 && !diffJSON {
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("── %s ──\n", filepath.Base(t))
		}

		if diffJSON {
			if err := ui.PrintDiffJSON(summary); err != nil {
				return err
			}
		} else {
			ui.PrintDiffSummary(summary)
		}
	}

	return nil
}

func diffRepo(dir string) (*diffutil.DiffSummary, error) {
	gitArgs := []string{"diff"}
	if diffStaged {
		gitArgs = append(gitArgs, "--staged")
	}
	if diffBranch != "" {
		gitArgs = append(gitArgs, diffBranch+"...HEAD")
	}

	output, err := gitutil.RunGit(dir, gitArgs...)
	if err != nil {
		return nil, err
	}

	return diffutil.Parse(strings.NewReader(output))
}
