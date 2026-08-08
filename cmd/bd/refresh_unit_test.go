package main

import (
	"os"
	"strings"
	"testing"
)

// TestRefreshProtocol covers the drift check: an outdated protocol block is
// detected and re-injected; a current block is left alone; a file with no
// protocol block is skipped (that's onboard's job, not refresh's); and
// content outside the tags is never touched.
func TestRefreshProtocol(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	desired := ProtocolStartTag + "\n" + restoreCodeBlocks(FullBootloader) + "\n" + ProtocolEndTag

	// AGENTS.md: outdated block wrapped in user prose → should drift + reinject.
	stalePre := "# My project notes\n\n"
	stalePost := "\n\n## More user prose kept verbatim\n"
	stale := stalePre + ProtocolStartTag + "\nOLD PROTOCOL\n" + ProtocolEndTag + stalePost
	if err := os.WriteFile("AGENTS.md", []byte(stale), 0644); err != nil {
		t.Fatal(err)
	}
	// CLAUDE.md: already current → no drift.
	if err := os.WriteFile("CLAUDE.md", []byte("intro\n"+desired+"\noutro\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// GEMINI.md: no protocol block at all → skipped.
	if err := os.WriteFile("GEMINI.md", []byte("no protocol here\n"), 0644); err != nil {
		t.Fatal(err)
	}

	drifted := refreshProtocol(true)

	if len(drifted) != 1 || drifted[0] != "AGENTS.md" {
		t.Fatalf("expected only AGENTS.md to drift, got %v", drifted)
	}

	got, _ := os.ReadFile("AGENTS.md")
	gs := string(got)
	if !strings.Contains(gs, restoreCodeBlocks(FullBootloader)) {
		t.Error("AGENTS.md was not re-injected with the current protocol")
	}
	if strings.Contains(gs, "OLD PROTOCOL") {
		t.Error("stale protocol still present after re-inject")
	}
	if !strings.HasPrefix(gs, stalePre) || !strings.HasSuffix(gs, stalePost) {
		t.Error("user prose outside the protocol tags was modified")
	}

	// GEMINI.md untouched.
	if g, _ := os.ReadFile("GEMINI.md"); string(g) != "no protocol here\n" {
		t.Error("file without a protocol block should be skipped")
	}

	// Idempotent: second run finds no drift.
	if d := refreshProtocol(true); len(d) != 0 {
		t.Fatalf("expected no drift on second run, got %v", d)
	}
}
