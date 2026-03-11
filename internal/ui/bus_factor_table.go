package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ezerfernandes/repomni/internal/busfactor"
)

// PrintBusFactorTable renders bus factor results as a formatted table.
func PrintBusFactorTable(results []busfactor.FileResult) {
	if len(results) == 0 {
		return
	}

	// Compute column widths.
	fileW := len("File")
	authorW := len("Top Author")
	for _, r := range results {
		if len(r.File) > fileW {
			fileW = len(r.File)
		}
		if len(r.TopAuthor) > authorW {
			authorW = len(r.TopAuthor)
		}
	}

	// Cap file width to avoid overly wide tables.
	if fileW > 60 {
		fileW = 60
	}

	hdrFmt := fmt.Sprintf("  %%-%ds  %%-%ds  %%-%ds  %%-%ds  %%s\n", fileW, 10, authorW, 12)

	fmt.Println()
	fmt.Printf(hdrFmt, "File", "Bus Factor", "Top Author", "Top Author %", "Total Authors")
	fmt.Printf(hdrFmt,
		strings.Repeat("─", fileW),
		strings.Repeat("─", 10),
		strings.Repeat("─", authorW),
		strings.Repeat("─", 12),
		strings.Repeat("─", 13))

	rowFmt := fmt.Sprintf("  %%-%ds  %%-%ds  %%-%ds  %%-%ds  %%s\n", fileW, 10, authorW, 12)

	red := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("#2e8b57"))

	for _, r := range results {
		file := r.File
		if len(file) > fileW {
			file = "…" + file[len(file)-fileW+1:]
		}

		bfStr := fmt.Sprintf("%d", r.BusFactor)
		switch {
		case r.BusFactor <= 1:
			bfStr = red.Render(bfStr)
		case r.BusFactor == 2:
			bfStr = yellow.Render(bfStr)
		default:
			bfStr = green.Render(bfStr)
		}

		pctStr := fmt.Sprintf("%.1f%%", r.TopAuthorPct)
		authorsStr := fmt.Sprintf("%d", r.TotalAuthors)

		fmt.Printf(rowFmt, file, bfStr, r.TopAuthor, pctStr, authorsStr)
	}
	fmt.Println()
}
