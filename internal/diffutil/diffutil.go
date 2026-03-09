package diffutil

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pmezard/go-difflib/difflib"
)

var (
	addedStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#2e8b57"))
	removedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#b22222"))
	addedEmphStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Background(lipgloss.Color("#2e8b57"))
	removedEmphStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Background(lipgloss.Color("#b22222"))
	hunkStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	headerStyle      = lipgloss.NewStyle().Bold(true)
)

// UnifiedDiff computes a unified diff between oldText and newText.
// Returns an empty string if the texts are identical.
func UnifiedDiff(oldLabel, newLabel, oldText, newText string) string {
	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(oldText),
		B:        difflib.SplitLines(newText),
		FromFile: oldLabel,
		ToFile:   newLabel,
		Context:  3,
	})
	return diff
}

// ColorDiff applies ANSI colors to a unified diff string.
// Paired -/+ lines get character-level highlighting so that small changes
// (like a line number in an error message) stand out immediately.
func ColorDiff(diff string) string {
	lines := strings.Split(diff, "\n")
	var b strings.Builder
	first := true

	writeLine := func(s string) {
		if !first {
			b.WriteByte('\n')
		}
		first = false
		b.WriteString(s)
	}

	i := 0
	for i < len(lines) {
		line := lines[i]

		// Collect a change hunk: consecutive - lines then consecutive + lines.
		if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			var removed []string
			for i < len(lines) && strings.HasPrefix(lines[i], "-") && !strings.HasPrefix(lines[i], "---") {
				removed = append(removed, lines[i])
				i++
			}
			var added []string
			for i < len(lines) && strings.HasPrefix(lines[i], "+") && !strings.HasPrefix(lines[i], "+++") {
				added = append(added, lines[i])
				i++
			}

			// Pair lines for character-level highlighting.
			pairs := min(len(removed), len(added))
			for j := 0; j < pairs; j++ {
				r, a := highlightLinePair(removed[j], added[j])
				writeLine(r)
				writeLine(a)
			}
			for j := pairs; j < len(removed); j++ {
				writeLine(removedStyle.Render(removed[j]))
			}
			for j := pairs; j < len(added); j++ {
				writeLine(addedStyle.Render(added[j]))
			}
			continue
		}

		switch {
		case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			writeLine(headerStyle.Render(line))
		case strings.HasPrefix(line, "+"):
			writeLine(addedStyle.Render(line))
		case strings.HasPrefix(line, "@@"):
			writeLine(hunkStyle.Render(line))
		default:
			writeLine(line)
		}
		i++
	}
	return b.String()
}

// highlightLinePair renders a paired -/+ line with character-level emphasis
// on the parts that actually changed.
func highlightLinePair(removedLine, addedLine string) (string, string) {
	oldText := removedLine[1:] // strip "-"
	newText := addedLine[1:]   // strip "+"

	oldChars := splitChars(oldText)
	newChars := splitChars(newText)

	matcher := difflib.NewMatcher(oldChars, newChars)
	opcodes := matcher.GetOpCodes()

	var oldBuf, newBuf strings.Builder
	oldBuf.WriteString(removedStyle.Render("-"))
	newBuf.WriteString(addedStyle.Render("+"))

	for _, op := range opcodes {
		switch op.Tag {
		case 'e': // equal
			oldBuf.WriteString(removedStyle.Render(strings.Join(oldChars[op.I1:op.I2], "")))
			newBuf.WriteString(addedStyle.Render(strings.Join(newChars[op.J1:op.J2], "")))
		case 'r': // replace
			oldBuf.WriteString(removedEmphStyle.Render(strings.Join(oldChars[op.I1:op.I2], "")))
			newBuf.WriteString(addedEmphStyle.Render(strings.Join(newChars[op.J1:op.J2], "")))
		case 'd': // delete (only in old)
			oldBuf.WriteString(removedEmphStyle.Render(strings.Join(oldChars[op.I1:op.I2], "")))
		case 'i': // insert (only in new)
			newBuf.WriteString(addedEmphStyle.Render(strings.Join(newChars[op.J1:op.J2], "")))
		}
	}

	return oldBuf.String(), newBuf.String()
}

// splitChars splits a string into individual character strings for use
// with SequenceMatcher.
func splitChars(s string) []string {
	chars := make([]string, 0, len(s))
	for _, r := range s {
		chars = append(chars, string(r))
	}
	return chars
}

// SummaryLine returns a one-line summary comparing oldText and newText.
func SummaryLine(oldText, newText string) string {
	if oldText == newText {
		return "Outputs are identical"
	}

	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:       difflib.SplitLines(oldText),
		B:       difflib.SplitLines(newText),
		Context: 0,
	})

	var added, removed int
	for _, line := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			// skip headers
		case strings.HasPrefix(line, "+"):
			added++
		case strings.HasPrefix(line, "-"):
			removed++
		}
	}

	return fmt.Sprintf("Outputs differ (%d lines added, %d lines removed)", added, removed)
}
