package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ezerfernandes/repomni/internal/diffutil"
)

// PrintDiffSummary renders a human-readable table of semantic diff info.
func PrintDiffSummary(summary *diffutil.DiffSummary) {
	if len(summary.Files) == 0 {
		fmt.Println("No changes.")
		return
	}

	// Compute column widths.
	fileW := len("File")
	typeW := len("Change")
	deltaW := len("Delta")
	for _, f := range summary.Files {
		name := displayName(f)
		if len(name) > fileW {
			fileW = len(name)
		}
		if len(f.ChangeType) > typeW {
			typeW = len(f.ChangeType)
		}
		d := formatDelta(f)
		if len(d) > deltaW {
			deltaW = len(d)
		}
	}

	hdrFmt := fmt.Sprintf("  %%-%ds  %%-%ds  %%-%ds  %%s\n", fileW, typeW, deltaW)
	fmt.Println()
	fmt.Printf(hdrFmt, "File", "Change", "Delta", "Semantics")
	fmt.Printf(hdrFmt,
		strings.Repeat("─", fileW),
		strings.Repeat("─", typeW),
		strings.Repeat("─", deltaW),
		strings.Repeat("─", 9))

	rowFmt := fmt.Sprintf("  %%-%ds  %%-%ds  %%-%ds  %%s\n", fileW, typeW, deltaW)
	for _, f := range summary.Files {
		fmt.Printf(rowFmt, displayName(f), f.ChangeType, formatDelta(f), semantics(f))
	}

	fmt.Printf("\n%d file(s) changed, %d insertions(+), %d deletions(-)\n",
		len(summary.Files), summary.TotalAdded, summary.TotalRemoved)
}

// PrintDiffJSON outputs the diff summary as JSON.
func PrintDiffJSON(summary *diffutil.DiffSummary) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(summary)
}

func displayName(f diffutil.FileDiff) string {
	if f.OldName != f.NewName {
		return filepath.Base(f.OldName) + " → " + filepath.Base(f.NewName)
	}
	return f.NewName
}

func formatDelta(f diffutil.FileDiff) string {
	if f.Binary {
		return "binary"
	}
	return fmt.Sprintf("+%d/-%d", f.Added, f.Removed)
}

func semantics(f diffutil.FileDiff) string {
	var parts []string
	if f.Imports {
		parts = append(parts, "imports")
	}
	if len(f.Functions) > 0 {
		parts = append(parts, "func: "+strings.Join(f.Functions, ", "))
	}
	if len(f.Types) > 0 {
		parts = append(parts, "type: "+strings.Join(f.Types, ", "))
	}
	return strings.Join(parts, "; ")
}
