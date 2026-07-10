//go:build integration
// +build integration

package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/untoldecay/BeadsLog/internal/storage/sqlite"
)

// setupBatch2Fixture: two synced sessions whose manual arrows produce
// separator-variant duplicate entities (OllamaExtractor / ollama-extractor).
func setupBatch2Fixture(t *testing.T) string {
	t.Helper()
	tmpDir := setupCLITestDB(t)
	devlogDir := filepath.Join(tmpDir, "_rules", "_devlog")
	if err := os.MkdirAll(devlogDir, 0755); err != nil {
		t.Fatalf("Failed to create devlog dir: %v", err)
	}

	one := `# First Session

Working on OllamaExtractor integration.

### Architectural Relationships
- OllamaExtractor -> ExtractionPipeline (feeds)
`
	two := `# Second Session

Refining the ollama-extractor confidence scoring.

### Architectural Relationships
- ollama-extractor -> ConfidenceScoring (tunes)
`
	os.WriteFile(filepath.Join(devlogDir, "2026-01-01_first.md"), []byte(one), 0644)
	os.WriteFile(filepath.Join(devlogDir, "2026-01-02_second.md"), []byte(two), 0644)

	indexContent := `## Work Index

| Subject | Problems | Date | Devlog |
|---------|----------|------|--------|
| [feature] First Session | Extractor work | 2026-01-01 | [2026-01-01_first.md](2026-01-01_first.md) |
| [feature] Second Session | Scoring work | 2026-01-02 | [2026-01-02_second.md](2026-01-02_second.md) |
`
	os.WriteFile(filepath.Join(devlogDir, "_index.md"), []byte(indexContent), 0644)

	runBDInProcess(t, tmpDir, "config", "set", "devlog_dir", "_rules/_devlog")
	return tmpDir
}

func TestCLI_Batch2_SyncAutoAlias(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow CLI test in short mode")
	}
	tmpDir := setupBatch2Fixture(t)

	out := runBDInProcess(t, tmpDir, "devlog", "sync")
	if !strings.Contains(out, "Auto-merged") {
		t.Errorf("sync should report auto-merged duplicates:\n%s", out)
	}

	// Only one ollamaextractor variant survives in the entity list
	out = runBDInProcess(t, tmpDir, "devlog", "entities", "--json")
	variants := strings.Count(strings.ToLower(out), "ollama")
	if variants > 1 {
		t.Errorf("expected a single merged ollama entity, output:\n%s", out)
	}

	// The variant name resolves via the registry
	store, err := sqlite.New(context.Background(), filepath.Join(tmpDir, ".beads", "beads.db"))
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	defer store.Close()
	var n int
	_ = store.UnderlyingDB().QueryRow("SELECT COUNT(*) FROM entity_aliases").Scan(&n)
	if n != 1 {
		t.Errorf("expected 1 registry alias, got %d", n)
	}
}

func TestCLI_Batch2_EntityBonus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow CLI test in short mode")
	}
	tmpDir := setupBatch2Fixture(t)
	runBDInProcess(t, tmpDir, "devlog", "sync")

	out := runBDInProcess(t, tmpDir, "devlog", "search", "ExtractionPipeline", "--json")
	var resp struct {
		Results []struct {
			ID          string  `json:"id"`
			EntityBonus float64 `json:"entity_bonus"`
		}
	}
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("search --json is not valid JSON: %v\n%s", err, out)
	}
	if len(resp.Results) == 0 {
		t.Fatalf("expected results for ExtractionPipeline:\n%s", out)
	}
	found := false
	for _, r := range resp.Results {
		if r.EntityBonus == 0.75 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected an entity_bonus of 0.75 for the graph-linked session:\n%s", out)
	}

	// Merged variant spellings must earn the bonus too: hyphenated queries
	// are token-split, and variant names resolve via the alias registry.
	for _, q := range []string{"ollama-extractor", "OllamaExtractor"} {
		out := runBDInProcess(t, tmpDir, "devlog", "search", q, "--json")
		if err := json.Unmarshal([]byte(out), &resp); err != nil {
			t.Fatalf("search %q --json invalid: %v\n%s", q, err, out)
		}
		bonus := false
		for _, r := range resp.Results {
			if r.EntityBonus == 0.75 {
				bonus = true
			}
		}
		if !bonus {
			t.Errorf("query %q: expected entity_bonus 0.75 via alias/joined-token match:\n%s", q, out)
		}
	}
}

func TestCLI_Batch2_PruneNoise(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow CLI test in short mode")
	}
	tmpDir := setupBatch2Fixture(t)
	runBDInProcess(t, tmpDir, "devlog", "sync")

	// Seed legacy junk directly (predates extraction-time filtering)
	store, err := sqlite.New(context.Background(), filepath.Join(tmpDir, ".beads", "beads.db"))
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	if _, err := store.UnderlyingDB().Exec("INSERT INTO entities (id, name, mention_count) VALUES ('junk1', 'map it step', 3)"); err != nil {
		t.Fatalf("Failed to seed junk entity: %v", err)
	}
	store.Close()

	out := runBDInProcess(t, tmpDir, "devlog", "prune", "--noise")
	if !strings.Contains(out, "Pruned 1 noise") {
		t.Errorf("prune --noise should report the junk entity:\n%s", out)
	}

	store, _ = sqlite.New(context.Background(), filepath.Join(tmpDir, ".beads", "beads.db"))
	defer store.Close()
	var n int
	_ = store.UnderlyingDB().QueryRow("SELECT COUNT(*) FROM entities WHERE name = 'map it step'").Scan(&n)
	if n != 0 {
		t.Error("junk entity survived prune --noise")
	}
	_ = store.UnderlyingDB().QueryRow("SELECT COUNT(*) FROM entities WHERE name LIKE '%extractionpipeline%'").Scan(&n)
	if n == 0 {
		t.Error("legit entity was wrongly pruned")
	}
}

func TestCLI_Batch2_GraphHTML(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow CLI test in short mode")
	}
	tmpDir := setupBatch2Fixture(t)
	runBDInProcess(t, tmpDir, "devlog", "sync")

	htmlPath := filepath.Join(tmpDir, "out", "graph.html")
	out := runBDInProcess(t, tmpDir, "devlog", "graph", "ExtractionPipeline", "--html", htmlPath)
	if !strings.Contains(out, "exported") {
		t.Errorf("graph --html should confirm export:\n%s", out)
	}

	content, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("HTML file not written: %v", err)
	}
	html := string(content)
	if !strings.Contains(html, "force-graph") {
		t.Error("HTML missing force-graph library reference")
	}
	if !strings.Contains(strings.ToLower(html), "extractionpipeline") {
		t.Error("HTML missing root entity in graph data")
	}
}
