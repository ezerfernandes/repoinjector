package ui

import (
	"fmt"

	"github.com/ezer/repoinjector/internal/injector"
)

func PrintResults(results []injector.Result) {
	created, skipped, errors := 0, 0, 0

	for _, r := range results {
		icon := actionIcon(r.Action)
		fmt.Printf("  %s %s — %s\n", icon, r.Item.TargetPath, r.Detail)

		switch r.Action {
		case "created", "updated", "removed":
			created++
		case "skipped", "dry-run":
			skipped++
		case "error":
			errors++
		}
	}

	fmt.Printf("\nDone. %d changed, %d skipped, %d errors.\n", created, skipped, errors)
}

func PrintStatusTable(repoPath string, statuses []injector.ItemStatus) {
	fmt.Printf("\nRepository: %s\n", repoPath)
	fmt.Println("  Item                   Present   Current   Excluded")
	fmt.Println("  ─────────────────────  ────────  ────────  ────────")

	for _, s := range statuses {
		present := boolIcon(s.Present)
		current := "-"
		if s.Present {
			current = boolIcon(s.Current)
		}
		excluded := boolIcon(s.Excluded)

		fmt.Printf("  %-21s  %-8s  %-8s  %s\n",
			s.Item.TargetPath, present, current, excluded)
	}
	fmt.Println()
}

func actionIcon(action string) string {
	switch action {
	case "created", "updated", "removed":
		return "[ok]"
	case "skipped", "dry-run":
		return "[--]"
	case "error":
		return "[!!]"
	default:
		return "[??]"
	}
}

func boolIcon(v bool) string {
	if v {
		return "Yes"
	}
	return "No"
}
