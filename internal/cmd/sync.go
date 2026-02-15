package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ezer/repoinjector/internal/gitutil"
	"github.com/ezer/repoinjector/internal/syncer"
	"github.com/ezer/repoinjector/internal/ui"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync [directory]",
	Short: "Pull updates for git repos in a directory",
	Long: `Fetch and pull updates for all git repositories that are immediate
subdirectories of the target directory.

If no directory is specified, the current directory is used. Each repo is
checked for upstream changes, and repos that are behind are pulled.

Repos with dirty working trees are skipped unless --autostash is used.
Diverged repos are always skipped (manual resolution required).`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSync,
}

var (
	syncDryRun    bool
	syncAutoStash bool
	syncJobs      int
	syncNoFetch   bool
	syncStrategy  string
	syncJSON      bool
)

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "show what would be done without pulling")
	syncCmd.Flags().BoolVar(&syncAutoStash, "autostash", false, "stash dirty working trees before pull")
	syncCmd.Flags().IntVarP(&syncJobs, "jobs", "j", 1, "number of parallel sync workers")
	syncCmd.Flags().BoolVar(&syncNoFetch, "no-fetch", false, "skip git fetch (local status only)")
	syncCmd.Flags().StringVar(&syncStrategy, "strategy", "ff-only", "pull strategy: ff-only, rebase, merge")
	syncCmd.Flags().BoolVar(&syncJSON, "json", false, "output as JSON")
}

func runSync(cmd *cobra.Command, args []string) error {
	target := "."
	if len(args) > 0 {
		target = args[0]
	}
	target, err := filepath.Abs(target)
	if err != nil {
		return err
	}

	repos, err := gitutil.FindGitRepos(target)
	if err != nil {
		return err
	}
	if len(repos) == 0 {
		return fmt.Errorf("no git repositories found under %s", target)
	}

	opts := syncer.SyncOptions{
		DryRun:    syncDryRun,
		AutoStash: syncAutoStash,
		Jobs:      syncJobs,
		NoFetch:   syncNoFetch,
		Strategy:  syncStrategy,
	}

	results, summary := syncer.SyncAll(repos, opts)

	if syncJSON {
		return ui.PrintSyncJSON(results, summary)
	}

	ui.PrintSyncResults(results, summary)

	if syncDryRun {
		fmt.Println("\nDry run \u2014 no changes were made.")
	}

	if summary.Errors > 0 {
		return fmt.Errorf("%d repos had errors", summary.Errors)
	}
	if summary.Conflicts > 0 {
		fmt.Fprintf(os.Stderr, "Warning: %d repo(s) have conflicts requiring manual resolution\n", summary.Conflicts)
	}

	return nil
}
