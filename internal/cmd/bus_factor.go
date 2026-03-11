package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ezerfernandes/repomni/internal/busfactor"
	"github.com/ezerfernandes/repomni/internal/ui"
	"github.com/spf13/cobra"
)

var busFactorCmd = &cobra.Command{
	Use:   "bus-factor [directory]",
	Short: "Analyze code ownership concentration",
	Long: `Analyze code ownership concentration in a git repository.

Computes the bus factor for each file: the minimum number of authors
whose combined commits cover more than 50% of the file's history.

Files with a bus factor of 1 are highest risk — a single person
is responsible for most of the file's development.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBusFactor,
}

var (
	busFactorTop       int
	busFactorThreshold float64
	busFactorJSON      bool
)

func init() {
	rootCmd.AddCommand(busFactorCmd)
	busFactorCmd.Flags().IntVar(&busFactorTop, "top", 20, "show N riskiest files")
	busFactorCmd.Flags().Float64Var(&busFactorThreshold, "threshold", 0.5, "ownership concentration threshold")
	busFactorCmd.Flags().BoolVar(&busFactorJSON, "json", false, "output as JSON")
}

func runBusFactor(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	results, err := busfactor.Analyze(absDir, busFactorThreshold)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Println("No files with commit history found.")
		return nil
	}

	// Limit to top N.
	if busFactorTop > 0 && len(results) > busFactorTop {
		results = results[:busFactorTop]
	}

	if busFactorJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	ui.PrintBusFactorTable(results)
	return nil
}
