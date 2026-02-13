package cmd

import (
	"fmt"
	"os"

	"github.com/ezer/repoinjector/internal/config"
	"github.com/ezer/repoinjector/internal/ui"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Interactively configure repoinjector",
	Long: `Run an interactive wizard to set up repoinjector. Configures the source
directory, injection mode, and which items to inject.

The configuration is saved to ~/.config/repoinjector/config.yaml.`,
	RunE: runConfigure,
}

var (
	configureSource         string
	configureNonInteractive bool
)

func init() {
	rootCmd.AddCommand(configureCmd)
	configureCmd.Flags().StringVar(&configureSource, "source", "", "source directory (skip interactive prompt)")
	configureCmd.Flags().BoolVar(&configureNonInteractive, "non-interactive", false, "use defaults without prompting")
}

func runConfigure(cmd *cobra.Command, args []string) error {
	// Load existing config as defaults, or start fresh
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	if configureNonInteractive {
		if configureSource == "" {
			return fmt.Errorf("--source is required in non-interactive mode")
		}
		cfg.SourceDir = configureSource
	} else {
		// Override source if flag provided
		if configureSource != "" {
			cfg.SourceDir = configureSource
		}

		cfg, err = ui.RunConfigureForm(cfg)
		if err != nil {
			return fmt.Errorf("configuration cancelled: %w", err)
		}
	}

	// Validate source directory
	if cfg.SourceDir == "" {
		return fmt.Errorf("source directory cannot be empty")
	}
	info, err := os.Stat(cfg.SourceDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("source directory does not exist or is not a directory: %s", cfg.SourceDir)
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	path, _ := config.ConfigPath()
	fmt.Printf("\nConfiguration saved to %s\n", path)
	return nil
}
