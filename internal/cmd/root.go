package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var version = "dev"

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
}

// Execute runs the root Cobra command and exits on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() { rootCmd.Version = version }
