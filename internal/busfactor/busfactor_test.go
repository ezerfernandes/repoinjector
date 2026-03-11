package busfactor

import (
	"testing"
)

func TestParseAuthorCounts(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect map[string]int
	}{
		{
			name:   "empty",
			input:  "",
			expect: nil,
		},
		{
			name:  "single author multiple commits",
			input: "Alice\nAlice\nAlice",
			expect: map[string]int{
				"Alice": 3,
			},
		},
		{
			name:  "multiple authors",
			input: "Alice\nBob\nAlice\nCharlie\nBob\nAlice",
			expect: map[string]int{
				"Alice":   3,
				"Bob":     2,
				"Charlie": 1,
			},
		},
		{
			name:  "whitespace handling",
			input: "  Alice  \nBob\n\nAlice\n",
			expect: map[string]int{
				"Alice": 2,
				"Bob":   1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAuthorCounts(tt.input)
			if tt.expect == nil && got != nil {
				t.Fatalf("expected nil, got %v", got)
			}
			if tt.expect != nil {
				if len(got) != len(tt.expect) {
					t.Fatalf("expected %d authors, got %d: %v", len(tt.expect), len(got), got)
				}
				for k, v := range tt.expect {
					if got[k] != v {
						t.Errorf("author %q: expected %d, got %d", k, v, got[k])
					}
				}
			}
		})
	}
}

func TestComputeBusFactor(t *testing.T) {
	tests := []struct {
		name      string
		authors   map[string]int
		threshold float64
		expect    int
	}{
		{
			name:      "single author",
			authors:   map[string]int{"Alice": 10},
			threshold: 0.5,
			expect:    1,
		},
		{
			name:      "two equal authors",
			authors:   map[string]int{"Alice": 5, "Bob": 5},
			threshold: 0.5,
			expect:    2,
		},
		{
			name:      "three equal authors at 50%",
			authors:   map[string]int{"Alice": 10, "Bob": 10, "Charlie": 10},
			threshold: 0.5,
			expect:    2,
		},
		{
			name:      "dominant author",
			authors:   map[string]int{"Alice": 90, "Bob": 5, "Charlie": 5},
			threshold: 0.5,
			expect:    1,
		},
		{
			name:      "empty map",
			authors:   map[string]int{},
			threshold: 0.5,
			expect:    0,
		},
		{
			name:      "many small contributors",
			authors:   map[string]int{"A": 3, "B": 3, "C": 3, "D": 3, "E": 3},
			threshold: 0.5,
			expect:    3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeBusFactor(tt.authors, tt.threshold)
			if got != tt.expect {
				t.Errorf("expected bus factor %d, got %d", tt.expect, got)
			}
		})
	}
}

func TestTopContributor(t *testing.T) {
	authors := map[string]int{"Alice": 10, "Bob": 5, "Charlie": 3}
	name, count := topContributor(authors)
	if name != "Alice" || count != 10 {
		t.Errorf("expected Alice/10, got %s/%d", name, count)
	}
}

func TestTotalCommits(t *testing.T) {
	authors := map[string]int{"Alice": 10, "Bob": 5, "Charlie": 3}
	total := totalCommits(authors)
	if total != 18 {
		t.Errorf("expected 18, got %d", total)
	}
}
