package cmd

import "github.com/spf13/cobra"

var branchCmd = &cobra.Command{
	Use:   "branch",
	Short: "Manage branch repos",
	Long: `Commands for creating, cloning, listing, and managing the workflow
state of branch repos.`,
}

func init() {
	rootCmd.AddCommand(branchCmd)
}
