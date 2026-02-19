package ui

import (
	"fmt"
	"strings"
)

// BranchInfo holds the collected information about one branch repo.
type BranchInfo struct {
	Path   string `json:"path"`
	Name   string `json:"name"`
	Branch string `json:"branch"`
	State  string `json:"state"`
	Dirty  bool   `json:"dirty"`
}

// PrintBranchesTable renders a colored table of branch repos.
func PrintBranchesTable(infos []BranchInfo) {
	nameW := len("Name")
	branchW := len("Branch")
	stateW := len("State")
	for _, info := range infos {
		if len(info.Name) > nameW {
			nameW = len(info.Name)
		}
		if len(info.Branch) > branchW {
			branchW = len(info.Branch)
		}
		display := info.State
		if display == "" {
			display = "--"
		}
		if len(display) > stateW {
			stateW = len(display)
		}
	}

	fmt.Println()
	hdrFmt := fmt.Sprintf("  %%-%ds  %%-%ds  %%-%ds  %%s\n", nameW, branchW, stateW)
	fmt.Printf(hdrFmt, "Name", "Branch", "State", "Dirty")
	fmt.Printf(hdrFmt,
		strings.Repeat("─", nameW),
		strings.Repeat("─", branchW),
		strings.Repeat("─", stateW),
		strings.Repeat("─", 5))

	for _, info := range infos {
		dirty := " "
		if info.Dirty {
			dirty = "*"
		}

		stateDisplay := RenderState(info.State)

		// Compute raw (uncolored) state width for manual padding,
		// since ANSI codes break %-*s formatting.
		rawState := info.State
		if rawState == "" {
			rawState = "--"
		}
		pad := stateW - len(rawState)
		if pad < 0 {
			pad = 0
		}

		fmt.Printf("  %-*s  %-*s  %s%s  %s\n",
			nameW, info.Name,
			branchW, info.Branch,
			stateDisplay, strings.Repeat(" ", pad),
			dirty)
	}
	fmt.Println()
}
