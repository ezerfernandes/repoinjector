package cmd

import "github.com/spf13/cobra"

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage repoinjector configuration",
	Long: `Commands for configuring repoinjector globally, per-repository,
and managing setup scripts.`,
}

func init() {
	rootCmd.AddCommand(configCmd)
}
