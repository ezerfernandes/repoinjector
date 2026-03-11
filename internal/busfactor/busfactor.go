package busfactor

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// FileResult holds the bus factor analysis for a single file.
type FileResult struct {
	File         string         `json:"file"`
	BusFactor    int            `json:"bus_factor"`
	TopAuthor    string         `json:"top_author"`
	TopAuthorPct float64        `json:"top_author_pct"`
	TotalAuthors int            `json:"total_authors"`
	Authors      map[string]int `json:"authors"`
}

// Analyze runs bus factor analysis on the given git repository directory.
// threshold controls what fraction of commits constitutes "ownership" (default 0.5).
func Analyze(dir string, threshold float64) ([]FileResult, error) {
	files, err := trackedFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("listing tracked files: %w", err)
	}

	var results []FileResult
	for _, f := range files {
		authors, err := fileAuthors(dir, f)
		if err != nil {
			continue // skip files with no history (e.g. binary, submodule)
		}
		if len(authors) == 0 {
			continue
		}

		bf := computeBusFactor(authors, threshold)
		topAuthor, topCount := topContributor(authors)
		total := totalCommits(authors)

		results = append(results, FileResult{
			File:         f,
			BusFactor:    bf,
			TopAuthor:    topAuthor,
			TopAuthorPct: float64(topCount) / float64(total) * 100,
			TotalAuthors: len(authors),
			Authors:      authors,
		})
	}

	// Sort by bus factor ascending (riskiest first), then by top author pct descending.
	sort.Slice(results, func(i, j int) bool {
		if results[i].BusFactor != results[j].BusFactor {
			return results[i].BusFactor < results[j].BusFactor
		}
		return results[i].TopAuthorPct > results[j].TopAuthorPct
	})

	return results, nil
}

func trackedFiles(dir string) ([]string, error) {
	cmd := exec.Command("git", "ls-files")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}
	return strings.Split(raw, "\n"), nil
}

func fileAuthors(dir, file string) (map[string]int, error) {
	cmd := exec.Command("git", "log", "--follow", "--format=%aN", "--", file)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseAuthorCounts(strings.TrimSpace(string(out))), nil
}

// parseAuthorCounts takes newline-separated author names and returns counts.
func parseAuthorCounts(output string) map[string]int {
	if output == "" {
		return nil
	}
	counts := make(map[string]int)
	for _, line := range strings.Split(output, "\n") {
		name := strings.TrimSpace(line)
		if name != "" {
			counts[name]++
		}
	}
	return counts
}

// computeBusFactor returns the minimum number of authors whose combined
// commits exceed threshold fraction of total commits.
func computeBusFactor(authors map[string]int, threshold float64) int {
	total := totalCommits(authors)
	if total == 0 {
		return 0
	}

	// Sort authors by commit count descending.
	type kv struct {
		author string
		count  int
	}
	sorted := make([]kv, 0, len(authors))
	for a, c := range authors {
		sorted = append(sorted, kv{a, c})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	cumulative := 0
	target := int(float64(total)*threshold) + 1 // strictly more than threshold
	for i, entry := range sorted {
		cumulative += entry.count
		if cumulative >= target {
			return i + 1
		}
	}
	return len(authors)
}

func topContributor(authors map[string]int) (string, int) {
	var best string
	var bestCount int
	for a, c := range authors {
		if c > bestCount {
			best = a
			bestCount = c
		}
	}
	return best, bestCount
}

func totalCommits(authors map[string]int) int {
	total := 0
	for _, c := range authors {
		total += c
	}
	return total
}
