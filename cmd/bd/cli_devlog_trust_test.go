//go:build integration
// +build integration

package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/untoldecay/BeadsLog/internal/storage/sqlite"
)

// setupDevlogTrustFixture creates a repo with two synced sessions, then
// deletes one file and re-syncs so it becomes a ghost (is_ghost=1).
// Returns tmpDir and the ghost session's title.
func setupDevlogTrustFixture(t *testing.T) (string, string) {
	t.Helper()
	tmpDir := setupCLITestDB(t)
	devlogDir := filepath.Join(tmpDir, "_rules", "_devlog")
	if err := os.MkdirAll(devlogDir, 0755); err != nil {
		t.Fatalf("Failed to create devlog dir: %v", err)
	}

	alphaContent := `# Alpha Session

Working on AlphaService integration.

### Architectural Relationships
- AlphaComponent -> AlphaService (calls)
`
	ghostContent := "# Phantom Session\n\nThis file will be deleted to create a ghost.\n"

	if err := os.WriteFile(filepath.Join(devlogDir, "2026-01-01_alpha.md"), []byte(alphaContent), 0644); err != nil {
		t.Fatalf("Failed to write alpha session: %v", err)
	}
	ghostPath := filepath.Join(devlogDir, "2026-01-02_phantom.md")
	if err := os.WriteFile(ghostPath, []byte(ghostContent), 0644); err != nil {
		t.Fatalf("Failed to write phantom session: %v", err)
	}

	indexContent := `## Work Index

| Subject | Problems | Date | Devlog |
|---------|----------|------|--------|
| [feature] Alpha Session | Alpha integration | 2026-01-01 | [2026-01-01_alpha.md](2026-01-01_alpha.md) |
| [test] Phantom Session | Ghost fixture | 2026-01-02 | [2026-01-02_phantom.md](2026-01-02_phantom.md) |
`
	if err := os.WriteFile(filepath.Join(devlogDir, "_index.md"), []byte(indexContent), 0644); err != nil {
		t.Fatalf("Failed to write index: %v", err)
	}

	runBDInProcess(t, tmpDir, "config", "set", "devlog_dir", "_rules/_devlog")
	runBDInProcess(t, tmpDir, "devlog", "sync")

	// Ghost the phantom session: delete its file and re-sync
	if err := os.Remove(ghostPath); err != nil {
		t.Fatalf("Failed to remove phantom file: %v", err)
	}
	runBDInProcess(t, tmpDir, "devlog", "sync")

	return tmpDir, "[test] Phantom Session"
}

// ghostSessionID returns the session ID of the ghost in the fixture DB.
func ghostSessionID(t *testing.T, tmpDir string) string {
	t.Helper()
	store, err := sqlite.New(context.Background(), filepath.Join(tmpDir, ".beads", "beads.db"))
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	defer store.Close()
	var id string
	if err := store.UnderlyingDB().QueryRow("SELECT id FROM sessions WHERE is_ghost = 1").Scan(&id); err != nil {
		t.Fatalf("Failed to find ghost session: %v", err)
	}
	return id
}

func TestCLI_DevlogTrust_FreshInstallHasNoGhosts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow CLI test in short mode")
	}
	// Regression (BeadsLog-a2g): the index template seeded a sample row whose
	// file never existed, so every fresh install started with 1 ghost.
	tmpDir := setupCLITestDB(t)
	runBDInProcess(t, tmpDir, "devlog", "initialize")
	runBDInProcess(t, tmpDir, "config", "set", "devlog_dir", "_rules/_devlog")

	out := runBDInProcess(t, tmpDir, "devlog", "sync")
	if strings.Contains(out, "ghost") || strings.Contains(out, "incomplete") {
		t.Errorf("fresh install should have no ghosts or incompletes:\n%s", out)
	}
	if !strings.Contains(out, "No sessions recorded yet") {
		t.Errorf("empty devlog space should get a friendly message, not an error:\n%s", out)
	}

	store, err := sqlite.New(context.Background(), filepath.Join(tmpDir, ".beads", "beads.db"))
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	defer store.Close()
	var ghosts int
	_ = store.UnderlyingDB().QueryRow("SELECT COUNT(*) FROM sessions WHERE is_ghost = 1").Scan(&ghosts)
	if ghosts != 0 {
		t.Errorf("fresh install has %d ghost(s)", ghosts)
	}
}

func TestCLI_DevlogTrust_GhostExclusion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow CLI test in short mode")
	}
	tmpDir, ghostTitle := setupDevlogTrustFixture(t)

	out := runBDInProcess(t, tmpDir, "devlog", "resume", "--last", "5")
	if strings.Contains(out, ghostTitle) {
		t.Errorf("resume returned ghost session:\n%s", out)
	}
	if !strings.Contains(out, "Alpha Session") {
		t.Errorf("resume missing live session:\n%s", out)
	}

	out = runBDInProcess(t, tmpDir, "devlog", "list")
	if strings.Contains(out, ghostTitle) {
		t.Errorf("list returned ghost session:\n%s", out)
	}
	if !strings.Contains(out, "Alpha Session") {
		t.Errorf("list missing live session:\n%s", out)
	}

	out = runBDInProcess(t, tmpDir, "devlog", "search", "Phantom")
	if strings.Contains(out, "2026-01-02_phantom") || strings.Contains(out, ghostTitle) {
		t.Errorf("search returned ghost session:\n%s", out)
	}

	out = runBDInProcess(t, tmpDir, "devlog", "status")
	if !strings.Contains(out, "Ghosts:") {
		t.Errorf("status should still report ghost count:\n%s", out)
	}
}

func TestCLI_DevlogTrust_ShowGhost(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow CLI test in short mode")
	}
	// Must exec the binary: the ghost path in show calls os.Exit(1)
	tmpDir, _ := setupDevlogTrustFixture(t)
	ghostID := ghostSessionID(t, tmpDir)

	cmd := exec.Command(testBD, "--no-daemon", "devlog", "show", ghostID)
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "BEADS_NO_DAEMON=1")
	out, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("Expected non-zero exit for ghost show")
	}
	sOut := string(out)
	if !strings.Contains(sOut, "ghost") || !strings.Contains(sOut, "bd devlog prune") {
		t.Errorf("Expected ghost explanation with prune hint, got:\n%s", sOut)
	}
	if strings.Contains(sOut, "no such file or directory") {
		t.Errorf("Raw OS error leaked to user:\n%s", sOut)
	}
}

func TestCLI_DevlogTrust_SyncSummary(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow CLI test in short mode")
	}
	tmpDir, _ := setupDevlogTrustFixture(t)

	out := runBDInProcess(t, tmpDir, "devlog", "sync")
	if !strings.Contains(out, "ghost session(s) in index") || !strings.Contains(out, "bd devlog prune") {
		t.Errorf("sync should report ghosts with prune hint:\n%s", out)
	}
	if !strings.Contains(out, "incomplete session(s)") || !strings.Contains(out, "bd devlog verify --fix") {
		t.Errorf("sync should report incomplete sessions with verify hint:\n%s", out)
	}
}

func TestCLI_DevlogTrust_SyncReconcilesDeindexedSessions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow CLI test in short mode")
	}
	// A session whose index row AND file are both deleted is never visited by
	// the sync row loop — reconciliation must still mark it as a ghost.
	tmpDir, ghostTitle := setupDevlogTrustFixture(t)
	devlogDir := filepath.Join(tmpDir, "_rules", "_devlog")

	// Un-ghost it in the DB to simulate the stale pre-reconciliation state,
	// then remove its index row so the row loop can't re-mark it.
	store, err := sqlite.New(context.Background(), filepath.Join(tmpDir, ".beads", "beads.db"))
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	if _, err := store.UnderlyingDB().Exec("UPDATE sessions SET is_ghost = 0, is_missing = 0 WHERE is_ghost = 1"); err != nil {
		t.Fatalf("Failed to reset ghost flag: %v", err)
	}
	store.Close()

	indexPath := filepath.Join(devlogDir, "_index.md")
	indexBytes, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("Failed to read index: %v", err)
	}
	var kept []string
	for _, line := range strings.Split(string(indexBytes), "\n") {
		if !strings.Contains(line, "phantom") {
			kept = append(kept, line)
		}
	}
	if err := os.WriteFile(indexPath, []byte(strings.Join(kept, "\n")), 0644); err != nil {
		t.Fatalf("Failed to rewrite index: %v", err)
	}

	runBDInProcess(t, tmpDir, "devlog", "sync")

	out := runBDInProcess(t, tmpDir, "devlog", "resume", "--last", "5")
	if strings.Contains(out, ghostTitle) {
		t.Errorf("resume returned de-indexed session with missing file:\n%s", out)
	}
	if ghostID := ghostSessionID(t, tmpDir); ghostID == "" {
		t.Error("expected de-indexed session to be marked as ghost")
	}
}

func TestCLI_DevlogTrust_PruneRemovesIndexRows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow CLI test in short mode")
	}
	// prune must also drop the ghost's _index.md row, otherwise the next
	// sync re-ingests it and the ghost resurrects.
	tmpDir, ghostTitle := setupDevlogTrustFixture(t)
	devlogDir := filepath.Join(tmpDir, "_rules", "_devlog")

	out := runBDInProcess(t, tmpDir, "devlog", "prune")
	if !strings.Contains(out, "Pruned 1 ghost") || !strings.Contains(out, "Removed 1 dead row(s)") {
		t.Errorf("prune should report DB and index cleanup:\n%s", out)
	}

	indexBytes, err := os.ReadFile(filepath.Join(devlogDir, "_index.md"))
	if err != nil {
		t.Fatalf("Failed to read index: %v", err)
	}
	if strings.Contains(string(indexBytes), "phantom") {
		t.Errorf("index still contains the pruned ghost's row:\n%s", indexBytes)
	}

	// The critical regression: a re-sync must NOT resurrect the ghost
	runBDInProcess(t, tmpDir, "devlog", "sync")
	store, err := sqlite.New(context.Background(), filepath.Join(tmpDir, ".beads", "beads.db"))
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	defer store.Close()
	var ghosts int
	_ = store.UnderlyingDB().QueryRow("SELECT COUNT(*) FROM sessions WHERE is_ghost = 1").Scan(&ghosts)
	if ghosts != 0 {
		t.Errorf("ghost resurrected after prune + sync: %d ghost(s)", ghosts)
	}

	out = runBDInProcess(t, tmpDir, "devlog", "resume", "--last", "5")
	if strings.Contains(out, ghostTitle) {
		t.Errorf("resume returned resurrected ghost:\n%s", out)
	}
}

func TestCLI_DevlogTrust_StatusHonesty(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow CLI test in short mode")
	}
	tmpDir, _ := setupDevlogTrustFixture(t)

	// Simulate a legacy, never-enriched session
	store, err := sqlite.New(context.Background(), filepath.Join(tmpDir, ".beads", "beads.db"))
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	if _, err := store.UnderlyingDB().Exec("UPDATE sessions SET enrichment_status = 0 WHERE is_ghost = 0"); err != nil {
		t.Fatalf("Failed to downgrade enrichment status: %v", err)
	}
	store.Close()

	out := runBDInProcess(t, tmpDir, "devlog", "status")
	if strings.Contains(out, "All memory optimized") {
		t.Errorf("status claims optimized despite unenriched/incomplete/ghost sessions:\n%s", out)
	}
	if !strings.Contains(out, "Unenriched:") {
		t.Errorf("status should show unenriched count:\n%s", out)
	}
	if !strings.Contains(out, "Memory health:") {
		t.Errorf("status should show health warning when repairs are needed:\n%s", out)
	}
}

func TestCLI_DevlogTrust_JSONOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow CLI test in short mode")
	}
	tmpDir, ghostTitle := setupDevlogTrustFixture(t)

	// list --json: valid array, ghost excluded
	out := runBDInProcess(t, tmpDir, "devlog", "list", "--json")
	var listRows []map[string]interface{}
	if err := json.Unmarshal([]byte(out), &listRows); err != nil {
		t.Fatalf("list --json is not valid JSON: %v\n%s", err, out)
	}
	if len(listRows) != 1 {
		t.Errorf("list --json expected 1 live session, got %d:\n%s", len(listRows), out)
	}
	if strings.Contains(out, ghostTitle) {
		t.Errorf("list --json contains ghost:\n%s", out)
	}

	// resume --json: valid array, ghost excluded
	out = runBDInProcess(t, tmpDir, "devlog", "resume", "--last", "5", "--json")
	var resumeRows []map[string]interface{}
	if err := json.Unmarshal([]byte(out), &resumeRows); err != nil {
		t.Fatalf("resume --json is not valid JSON: %v\n%s", err, out)
	}
	if strings.Contains(out, ghostTitle) {
		t.Errorf("resume --json contains ghost:\n%s", out)
	}

	// show --json: object with content
	if len(listRows) == 1 {
		id, _ := listRows[0]["id"].(string)
		out = runBDInProcess(t, tmpDir, "devlog", "show", id, "--json")
		var shown map[string]interface{}
		if err := json.Unmarshal([]byte(out), &shown); err != nil {
			t.Fatalf("show --json is not valid JSON: %v\n%s", err, out)
		}
		if content, _ := shown["content"].(string); !strings.Contains(content, "AlphaService") {
			t.Errorf("show --json missing file content:\n%s", out)
		}
	}

	// status --json: struct with honest counts
	out = runBDInProcess(t, tmpDir, "devlog", "status", "--json")
	var status map[string]interface{}
	if err := json.Unmarshal([]byte(out), &status); err != nil {
		t.Fatalf("status --json is not valid JSON: %v\n%s", err, out)
	}
	if ghosts, _ := status["ghosts"].(float64); ghosts != 1 {
		t.Errorf("status --json expected 1 ghost, got %v:\n%s", status["ghosts"], out)
	}

	// entities --json: valid array
	out = runBDInProcess(t, tmpDir, "devlog", "entities", "--json")
	var entities []map[string]interface{}
	if err := json.Unmarshal([]byte(out), &entities); err != nil {
		t.Fatalf("entities --json is not valid JSON: %v\n%s", err, out)
	}

	// impact --json: valid array, no lipgloss borders
	out = runBDInProcess(t, tmpDir, "devlog", "impact", "AlphaService", "--json")
	var impact []map[string]interface{}
	if err := json.Unmarshal([]byte(out), &impact); err != nil {
		t.Fatalf("impact --json is not valid JSON: %v\n%s", err, out)
	}
	if strings.Contains(out, "╭") {
		t.Errorf("impact --json contains table borders:\n%s", out)
	}
}
