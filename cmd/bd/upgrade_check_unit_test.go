package main

import (
	"strings"
	"testing"
)

// TestExtractLatestFeatures verifies the best-effort parser pulls the newest
// version's Features bullets out of a changelog.go source snippet.
func TestExtractLatestFeatures(t *testing.T) {
	src := `
const CurrentVersion = "0.59.0"

var entries = []Entry{
	{
		Version: "0.59.0",
		Date:    "2026-08-09",
		Features: []string{
			"Added bd refresh, the one-command post-update.",
			"Consolidated bd upgrade into a single command.",
		},
		Protocol: []string{"n/a"},
	},
	{
		Version: "0.58.0",
		Features: []string{
			"Older feature that must NOT be returned.",
		},
	},
}`

	got := extractLatestFeatures(src)
	if len(got) != 2 {
		t.Fatalf("expected 2 features from the latest version, got %d: %v", len(got), got)
	}
	if !strings.Contains(got[0], "bd refresh") || !strings.Contains(got[1], "single command") {
		t.Errorf("unexpected features: %v", got)
	}
	for _, f := range got {
		if strings.Contains(f, "Older feature") {
			t.Errorf("leaked an older version's feature: %q", f)
		}
	}

	// Malformed input must not panic and returns nil.
	if f := extractLatestFeatures("no features here"); f != nil {
		t.Errorf("expected nil for malformed input, got %v", f)
	}
}
