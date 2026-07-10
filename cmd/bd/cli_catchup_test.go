//go:build integration
// +build integration

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// setupCatchupFixture: two sessions on a feature branch, one on main, with a
// shared entity, then an explicit abandon on the feature branch.
func setupCatchupFixture(t *testing.T) string {
	t.Helper()
	tmpDir := setupCLITestDB(t)
	devlogDir := filepath.Join(tmpDir, "_rules", "_devlog")
	if err := os.MkdirAll(devlogDir, 0755); err != nil {
		t.Fatalf("Failed to create devlog dir: %v", err)
	}

	one := `# [feature] Payment retry queue

## Problem
Retries were dropped silently under load.

### Architectural Relationships
- RetryQueue -> PaymentGateway (feeds)
`
	two := `# [fix] Payment retry timeouts

## Problem
Slow gateways hit the retry timeout.

**Final Status:** Retry queue hardened and verified.

### Architectural Relationships
- RetryQueue -> TimeoutPolicy (uses)
`
	three := `# [docs] Update contributor guide

## Problem
The contributor guide was stale.
`
	os.WriteFile(filepath.Join(devlogDir, "2026-07-09_retry-queue.md"), []byte(one), 0644)
	os.WriteFile(filepath.Join(devlogDir, "2026-07-09_retry-timeouts.md"), []byte(two), 0644)
	os.WriteFile(filepath.Join(devlogDir, "2026-07-09_docs.md"), []byte(three), 0644)

	now := time.Now()
	d1 := now.Add(-2 * time.Hour).Format("2006-01-02 15:04")
	d2 := now.Add(-1 * time.Hour).Format("2006-01-02 15:04")
	indexContent := fmt.Sprintf(`## Work Index

| Subject | Problems | Author | Agent | Date | Branch | Devlog |
|---------|----------|--------|-------|------|--------|---------|
| [feature] Payment retry queue | Retries were dropped silently under load. | alice | TestAgent | %s | feat/pay | [2026-07-09_retry-queue.md](2026-07-09_retry-queue.md) |
| [fix] Payment retry timeouts | Slow gateways hit the retry timeout. | bob | TestAgent | %s | feat/pay | [2026-07-09_retry-timeouts.md](2026-07-09_retry-timeouts.md) |
| [docs] Update contributor guide | The contributor guide was stale. | carol | TestAgent | %s | main | [2026-07-09_docs.md](2026-07-09_docs.md) |
`, d1, d2, d2)
	os.WriteFile(filepath.Join(devlogDir, "_index.md"), []byte(indexContent), 0644)

	runBDInProcess(t, tmpDir, "config", "set", "devlog_dir", "_rules/_devlog")
	runBDInProcess(t, tmpDir, "devlog", "sync")
	runBDInProcess(t, tmpDir, "devlog", "pause", "--scope", "branch:feat/pay", "--message", "waiting on review")
	return tmpDir
}

func TestCLI_Catchup_Digest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow CLI test in short mode")
	}
	tmpDir := setupCatchupFixture(t)

	out := runBDInProcess(t, tmpDir, "catchup", "--digest")

	if !strings.Contains(out, "📦 feat/pay") {
		t.Errorf("digest missing branch arc:\n%s", out)
	}
	if !strings.Contains(out, "Started with: Retries were dropped silently under load.") {
		t.Errorf("digest missing arc narrative start:\n%s", out)
	}
	if !strings.Contains(out, "Currently: Retry queue hardened and verified.") {
		t.Errorf("digest missing arc final status:\n%s", out)
	}
	if !strings.Contains(out, "PAUSED since you last looked") || !strings.Contains(out, "waiting on review") {
		t.Errorf("digest missing attached lifecycle delta:\n%s", out)
	}
	// alice + bob deduped on the branch arc
	if !strings.Contains(out, "alice, bob") {
		t.Errorf("digest missing deduped authors:\n%s", out)
	}
}

func TestCLI_Catchup_DigestJSON_And_Ack(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow CLI test in short mode")
	}
	tmpDir := setupCatchupFixture(t)

	out := runBDInProcess(t, tmpDir, "catchup", "--digest", "--json")
	var resp struct {
		Groups []struct {
			Key       string `json:"key"`
			Kind      string `json:"kind"`
			Narrative string `json:"narrative"`
		} `json:"groups"`
	}
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("digest --json invalid: %v\n%s", err, out)
	}
	if len(resp.Groups) < 2 {
		t.Errorf("expected at least 2 groups, got %d:\n%s", len(resp.Groups), out)
	}
	if resp.Groups[0].Key != "feat/pay" || resp.Groups[0].Narrative == "" {
		t.Errorf("branch group malformed in JSON: %+v", resp.Groups[0])
	}

	// --ack sets the checkpoint; a following digest is empty.
	// Explicit =false values defeat cobra flag stickiness across in-process runs.
	runBDInProcess(t, tmpDir, "catchup", "--digest", "--ack", "--json=false")
	out = runBDInProcess(t, tmpDir, "catchup", "--digest", "--ack=false", "--json=false")
	if !strings.Contains(out, "Nothing new") {
		t.Errorf("post-ack digest should be empty:\n%s", out)
	}
}

func TestCLI_Lifecycle_OngoingState(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow CLI test in short mode")
	}
	tmpDir := setupCatchupFixture(t)

	// Off-current-branch unmerged sessions derive ONGOING, but the explicit
	// pause on feat/pay wins for its scope; revive it with 'devlog ongoing'.
	out := runBDInProcess(t, tmpDir, "devlog", "ongoing", "--scope", "branch:feat/pay", "--message", "resuming after context switch")
	if strings.Contains(strings.ToLower(out), "error") {
		t.Fatalf("devlog ongoing failed:\n%s", out)
	}

	out = runBDInProcess(t, tmpDir, "devlog", "search", "retry")
	if strings.Contains(out, "PAUSED") {
		t.Errorf("scope revived as ongoing must not render PAUSED:\n%s", out)
	}
	if !strings.Contains(out, "ONGOING") {
		t.Errorf("expected ONGOING badge after revive:\n%s", out)
	}

	// ongoing must not trigger proximity warnings on resume
	out = runBDInProcess(t, tmpDir, "devlog", "resume", "retry")
	if strings.Contains(out, "This work is") {
		t.Errorf("ongoing state should not warn on resume:\n%s", out)
	}
}
