package diffutil

import (
	"strings"
	"testing"
)

func TestUnifiedDiff_Identical(t *testing.T) {
	text := "line1\nline2\nline3\n"
	diff := UnifiedDiff("a", "b", text, text)
	if diff != "" {
		t.Errorf("expected empty diff for identical inputs, got:\n%s", diff)
	}
}

func TestUnifiedDiff_SingleLineDifference(t *testing.T) {
	old := "hello\n"
	new := "world\n"
	diff := UnifiedDiff("main", "branch", old, new)
	if !strings.Contains(diff, "-hello") {
		t.Errorf("diff should contain removed line, got:\n%s", diff)
	}
	if !strings.Contains(diff, "+world") {
		t.Errorf("diff should contain added line, got:\n%s", diff)
	}
	if !strings.Contains(diff, "--- main") {
		t.Errorf("diff should contain old label, got:\n%s", diff)
	}
	if !strings.Contains(diff, "+++ branch") {
		t.Errorf("diff should contain new label, got:\n%s", diff)
	}
}

func TestUnifiedDiff_MultiLine(t *testing.T) {
	old := "a\nb\nc\nd\n"
	new := "a\nc\nd\ne\n"
	diff := UnifiedDiff("old", "new", old, new)
	if !strings.Contains(diff, "-b") {
		t.Errorf("diff should show removed line 'b', got:\n%s", diff)
	}
	if !strings.Contains(diff, "+e") {
		t.Errorf("diff should show added line 'e', got:\n%s", diff)
	}
}

func TestColorDiff_PreservesContent(t *testing.T) {
	diff := "--- main\n+++ branch\n@@ -1 +1 @@\n-old\n+new\n context\n"
	colored := ColorDiff(diff)
	// Should preserve all text content regardless of whether ANSI codes are added
	// (lipgloss may skip ANSI in non-TTY environments).
	if !strings.Contains(colored, "old") || !strings.Contains(colored, "new") {
		t.Error("ColorDiff should preserve text content")
	}
	if !strings.Contains(colored, "context") {
		t.Error("ColorDiff should preserve context lines")
	}
}

func TestColorDiff_CharLevelHighlighting(t *testing.T) {
	// Simulate a linter error where only the line number changed.
	diff := "--- main\n+++ branch\n@@ -1 +1 @@\n-main.go:10: unused variable\n+main.go:25: unused variable\n"
	colored := ColorDiff(diff)
	// Both the old and new content must be preserved.
	if !strings.Contains(colored, "10") || !strings.Contains(colored, "25") {
		t.Error("ColorDiff should preserve the changed line numbers")
	}
	if !strings.Contains(colored, "unused variable") {
		t.Error("ColorDiff should preserve the unchanged text")
	}
}

func TestColorDiff_PureAddition(t *testing.T) {
	// + lines without preceding - lines should still render.
	diff := "--- main\n+++ branch\n@@ -0,0 +1 @@\n+new line\n"
	colored := ColorDiff(diff)
	if !strings.Contains(colored, "new line") {
		t.Error("ColorDiff should render pure additions")
	}
}

func TestHighlightLinePair(t *testing.T) {
	r, a := highlightLinePair("-main.go:10: error msg", "+main.go:25: error msg")
	// The unchanged parts ("main.go:", ": error msg") must be present.
	if !strings.Contains(r, "main.go:") || !strings.Contains(a, "main.go:") {
		t.Error("highlightLinePair should preserve common prefix")
	}
	if !strings.Contains(r, "error msg") || !strings.Contains(a, "error msg") {
		t.Error("highlightLinePair should preserve common suffix")
	}
	// The changed parts ("10" and "25") must be present.
	if !strings.Contains(r, "10") || !strings.Contains(a, "25") {
		t.Error("highlightLinePair should preserve changed characters")
	}
}

func TestSplitChars(t *testing.T) {
	got := splitChars("abc")
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Errorf("splitChars(\"abc\") = %v, want [a b c]", got)
	}
	// Empty string.
	got = splitChars("")
	if len(got) != 0 {
		t.Errorf("splitChars(\"\") = %v, want []", got)
	}
}

func TestSummaryLine_Identical(t *testing.T) {
	text := "same\n"
	got := SummaryLine(text, text)
	if got != "Outputs are identical" {
		t.Errorf("got %q, want %q", got, "Outputs are identical")
	}
}

func TestSummaryLine_Different(t *testing.T) {
	old := "a\nb\n"
	new := "a\nc\nd\n"
	got := SummaryLine(old, new)
	if !strings.Contains(got, "Outputs differ") {
		t.Errorf("expected 'Outputs differ' prefix, got %q", got)
	}
	if !strings.Contains(got, "added") || !strings.Contains(got, "removed") {
		t.Errorf("expected counts in summary, got %q", got)
	}
}
