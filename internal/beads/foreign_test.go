package beads

import (
	"path/filepath"
	"testing"
)

// BeadsLog-7hz: a linked worktree that uses the main checkout's database must
// know the database is not its own, so sync never runs git on the wrong copy.
func TestForeignBeadsDir(t *testing.T) {
	root := t.TempDir()
	main := filepath.Join(root, "main")
	b := mkBeads(t, main, "beads.db", "config.yaml")
	gitRun(t, main, "init", "-q")
	gitRun(t, main, "commit", "-q", "--allow-empty", "-m", "x")
	wt := filepath.Join(root, "wt")
	gitRun(t, main, "worktree", "add", "-q", wt)

	inDir(t, main)
	if got := ForeignBeadsDir(); got != "" {
		t.Errorf("main checkout: ForeignBeadsDir = %q, want empty", got)
	}
	inDir(t, filepath.Join(main, ".beads"))
	if got := ForeignBeadsDir(); got != "" {
		t.Errorf("inside main .beads: ForeignBeadsDir = %q, want empty", got)
	}

	// Worktree without its own .beads, and with a committed copy of it
	for _, own := range []bool{false, true} {
		if own {
			mkBeads(t, wt, "config.yaml", "issues.jsonl")
		}
		inDir(t, wt)
		if got := ForeignBeadsDir(); !samePath(got, b) {
			t.Errorf("worktree (own .beads=%v): ForeignBeadsDir = %q, want %q", own, got, b)
		}
	}

	// Outside git there is no other checkout to confuse
	plain := filepath.Join(root, "plain")
	mkBeads(t, plain, "beads.db")
	inDir(t, plain)
	if got := ForeignBeadsDir(); got != "" {
		t.Errorf("outside git: ForeignBeadsDir = %q, want empty", got)
	}
}
