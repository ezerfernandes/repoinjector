package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/ezerfernandes/repomni/internal/logger"
	"github.com/spf13/cobra"
)

var version = "dev"

var verbose bool

func init() {
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
			version = info.Main.Version
		}
	}
}

var rootCmd = &cobra.Command{
	Use:   "repomni",
	Short: "Inject shared config files into multiple repo clones",
	Long: `Repomni symlinks or copies shared configuration files (.claude/skills,
.claude/hooks, .envrc, .env) from a central source into one or more target
repository clones, keeping injected files invisible to git.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.Init(verbose)
	},
}

// Execute runs the root Cobra command and exits on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = version
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable debug logging")
}
