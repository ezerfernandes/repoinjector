package diffutil

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// FileDiff represents the parsed diff for a single file.
type FileDiff struct {
	OldName    string   `json:"old_name"`
	NewName    string   `json:"new_name"`
	ChangeType string   `json:"change_type"` // "modified", "added", "deleted", "renamed", "binary"
	Binary     bool     `json:"binary"`
	Added      int      `json:"added"`
	Removed    int      `json:"removed"`
	Hunks      []Hunk   `json:"hunks,omitempty"`
	Functions  []string `json:"functions,omitempty"`  // Go functions/methods affected
	Types      []string `json:"types,omitempty"`      // Go types affected
	Imports    bool     `json:"imports,omitempty"`     // whether imports changed
}

// Hunk represents a single diff hunk.
type Hunk struct {
	Header   string `json:"header"`
	Added    int    `json:"added"`
	Removed  int    `json:"removed"`
	FuncName string `json:"func_name,omitempty"` // extracted from @@ header
}

// DiffSummary is the top-level result of parsing a diff.
type DiffSummary struct {
	Files      []FileDiff `json:"files"`
	TotalAdded int        `json:"total_added"`
	TotalRemoved int      `json:"total_removed"`
}

var (
	hunkHeaderRe = regexp.MustCompile(`^@@ .+ @@\s*(.*)`)
	goFuncRe     = regexp.MustCompile(`^func\s+(\([^)]+\)\s+)?(\w+)`)
	goTypeRe     = regexp.MustCompile(`^type\s+(\w+)\s+`)
	diffFileRe   = regexp.MustCompile(`^diff --git a/(.+) b/(.+)`)
)

// Parse reads unified diff output and returns structured data.
func Parse(r io.Reader) (*DiffSummary, error) {
	scanner := bufio.NewScanner(r)
	summary := &DiffSummary{}

	var current *FileDiff
	var currentHunk *Hunk

	flush := func() {
		if current == nil {
			return
		}
		if currentHunk != nil {
			current.Hunks = append(current.Hunks, *currentHunk)
			currentHunk = nil
		}
		current.Functions = dedup(current.Functions)
		current.Types = dedup(current.Types)
		inferChangeType(current)
		summary.TotalAdded += current.Added
		summary.TotalRemoved += current.Removed
		summary.Files = append(summary.Files, *current)
		current = nil
	}

	for scanner.Scan() {
		line := scanner.Text()

		// New file diff header
		if m := diffFileRe.FindStringSubmatch(line); m != nil {
			flush()
			current = &FileDiff{
				OldName: m[1],
				NewName: m[2],
			}
			continue
		}

		if current == nil {
			continue
		}

		// Binary file
		if strings.HasPrefix(line, "Binary files") || strings.HasPrefix(line, "GIT binary patch") {
			current.Binary = true
			current.ChangeType = "binary"
			continue
		}

		// Renamed file indicator
		if strings.HasPrefix(line, "rename from ") || strings.HasPrefix(line, "similarity index") {
			if current.ChangeType == "" {
				current.ChangeType = "renamed"
			}
			continue
		}

		// New file mode
		if strings.HasPrefix(line, "new file mode") {
			current.ChangeType = "added"
			continue
		}

		// Deleted file mode
		if strings.HasPrefix(line, "deleted file mode") {
			current.ChangeType = "deleted"
			continue
		}

		// Hunk header
		if m := hunkHeaderRe.FindStringSubmatch(line); m != nil {
			if currentHunk != nil {
				current.Hunks = append(current.Hunks, *currentHunk)
			}
			currentHunk = &Hunk{Header: line}
			// Extract function context from @@ header (Go convention)
			ctx := strings.TrimSpace(m[1])
			if ctx != "" {
				currentHunk.FuncName = ctx
				extractSemantics(current, ctx)
				// If hunk is inside an import block, mark imports changed
				if strings.Contains(ctx, "import") {
					current.Imports = true
				}
			}
			continue
		}

		// Diff content lines
		if currentHunk != nil {
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				currentHunk.Added++
				current.Added++
				content := line[1:]
				extractLineSemantics(current, content)
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				currentHunk.Removed++
				current.Removed++
				content := line[1:]
				extractLineSemantics(current, content)
			}
		}
	}

	flush()

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading diff: %w", err)
	}

	return summary, nil
}

// extractSemantics pulls Go-specific info from hunk context strings.
func extractSemantics(f *FileDiff, ctx string) {
	if !isGoFile(f.NewName) && !isGoFile(f.OldName) {
		return
	}
	if m := goFuncRe.FindStringSubmatch(ctx); m != nil {
		f.Functions = append(f.Functions, m[2])
	}
	if m := goTypeRe.FindStringSubmatch(ctx); m != nil {
		f.Types = append(f.Types, m[1])
	}
}

// extractLineSemantics checks added/removed lines for Go semantics.
func extractLineSemantics(f *FileDiff, content string) {
	if !isGoFile(f.NewName) && !isGoFile(f.OldName) {
		return
	}
	trimmed := strings.TrimSpace(content)

	// Import changes
	if strings.HasPrefix(trimmed, "import ") || trimmed == "import (" || trimmed == ")" {
		f.Imports = true
		return
	}
	// Lines inside import block (quoted paths)
	if strings.HasPrefix(trimmed, "\"") || strings.HasPrefix(trimmed, "// ") {
		// Could be import line; mark imports if we're in an import context
		// We rely on the import keyword detection above for accuracy
		return
	}

	if m := goFuncRe.FindStringSubmatch(trimmed); m != nil {
		f.Functions = append(f.Functions, m[2])
	}
	if m := goTypeRe.FindStringSubmatch(trimmed); m != nil {
		f.Types = append(f.Types, m[1])
	}
}

func isGoFile(name string) bool {
	return strings.HasSuffix(name, ".go")
}

func inferChangeType(f *FileDiff) {
	if f.ChangeType != "" {
		return
	}
	if f.OldName != f.NewName {
		f.ChangeType = "renamed"
	} else {
		f.ChangeType = "modified"
	}
}

func dedup(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}
