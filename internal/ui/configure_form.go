package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/ezer/repoinjector/internal/config"
)

func RunConfigureForm(cfg *config.Config) (*config.Config, error) {
	var sourceDir string
	if cfg.SourceDir != "" {
		sourceDir = cfg.SourceDir
	}

	var mode string
	if cfg.Mode != "" {
		mode = string(cfg.Mode)
	} else {
		mode = string(config.ModeSymlink)
	}

	// Build item options from defaults
	allItems := config.DefaultItems()
	enabledSet := make(map[string]bool)
	for _, item := range cfg.Items {
		if item.Enabled {
			enabledSet[item.SourcePath] = true
		}
	}

	type itemOption struct {
		label      string
		sourcePath string
	}
	options := []itemOption{
		{label: ".claude/skills/ (directory)", sourcePath: ".claude/skills"},
		{label: ".claude/hooks.json (file)", sourcePath: ".claude/hooks.json"},
		{label: ".envrc (file)", sourcePath: ".envrc"},
		{label: ".env (file)", sourcePath: ".env"},
	}

	var selectedItems []string
	// Pre-select currently enabled items
	for _, opt := range options {
		if enabledSet[opt.sourcePath] || len(cfg.Items) == 0 {
			selectedItems = append(selectedItems, opt.sourcePath)
		}
	}

	var selectOptions []huh.Option[string]
	for _, opt := range options {
		selectOptions = append(selectOptions, huh.NewOption(opt.label, opt.sourcePath))
	}

	var confirm bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Source directory").
				Description("Path to the repo containing files to inject").
				Value(&sourceDir).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("source directory is required")
					}
					abs, err := filepath.Abs(s)
					if err != nil {
						return err
					}
					info, err := os.Stat(abs)
					if err != nil {
						return fmt.Errorf("directory not found: %s", abs)
					}
					if !info.IsDir() {
						return fmt.Errorf("not a directory: %s", abs)
					}
					return nil
				}),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Injection mode").
				Description("Symlinks reflect source changes instantly. Copy creates independent snapshots.").
				Options(
					huh.NewOption("Symlink (recommended)", "symlink"),
					huh.NewOption("Copy", "copy"),
				).
				Value(&mode),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Items to inject").
				Description("Select which files/directories to inject into target repos").
				Options(selectOptions...).
				Value(&selectedItems),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Save this configuration?").
				Value(&confirm),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	if !confirm {
		return nil, fmt.Errorf("cancelled by user")
	}

	// Resolve source dir to absolute
	absSource, err := filepath.Abs(sourceDir)
	if err != nil {
		return nil, err
	}

	// Build enabled set from selection
	selectedSet := make(map[string]bool)
	for _, s := range selectedItems {
		selectedSet[s] = true
	}

	var items []config.Item
	for _, def := range allItems {
		items = append(items, config.Item{
			Type:       def.Type,
			SourcePath: def.SourcePath,
			TargetPath: def.TargetPath,
			Enabled:    selectedSet[def.SourcePath],
		})
	}

	return &config.Config{
		Version:   1,
		SourceDir: absSource,
		Mode:      config.InjectionMode(mode),
		Items:     items,
	}, nil
}
