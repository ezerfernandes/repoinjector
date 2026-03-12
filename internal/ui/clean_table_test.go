package ui

import (
	"strings"
	"testing"
)

func TestPrintCleanCandidates(t *testing.T) {
	candidates := []CleanCandidate{
		{
			Info:      BranchInfo{Name: "feature-a", State: "merged"},
			Size:      1024,
			SizeHuman: "1.0 KB",
		},
		{
			Info:      BranchInfo{Name: "feature-b", State: "open"},
			Size:      2048,
			SizeHuman: "2.0 KB",
			Skipped:   true,
			Reason:    "has open PR",
		},
		{
			Info:      BranchInfo{Name: "feature-c", State: ""},
			Size:      512,
			SizeHuman: "512 B",
		},
	}

	output := captureStdout(t, func() {
		PrintCleanCandidates(candidates)
	})

	// Verify header
	if !strings.Contains(output, "Name") || !strings.Contains(output, "State") || !strings.Contains(output, "Size") {
		t.Errorf("missing table headers in output:\n%s", output)
	}

	// Verify branch names appear
	if !strings.Contains(output, "feature-a") || !strings.Contains(output, "feature-b") || !strings.Contains(output, "feature-c") {
		t.Errorf("missing branch names in output:\n%s", output)
	}

	// Verify deletable count: 2 of 3
	if !strings.Contains(output, "2 of 3 branch(es) will be deleted") {
		t.Errorf("missing or wrong deletable count in output:\n%s", output)
	}

	// Verify skip reason appears
	if !strings.Contains(output, "skip: has open PR") {
		t.Errorf("missing skip reason in output:\n%s", output)
	}
}

func TestPrintCleanCandidates_Empty(t *testing.T) {
	output := captureStdout(t, func() {
		PrintCleanCandidates(nil)
	})

	if !strings.Contains(output, "0 of 0 branch(es) will be deleted") {
		t.Errorf("unexpected output for empty candidates:\n%s", output)
	}
}
