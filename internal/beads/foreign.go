package beads

import (
	"path/filepath"

	"github.com/untoldecay/BeadsLog/internal/git"
)

// ForeignBeadsDir returns the .beads directory in use when it belongs to a
// different checkout than the one bd runs in: a linked worktree sharing the
// main checkout's database, a redirect, or BEADS_DIR pointing elsewhere.
// It returns "" when the database lives in this checkout or outside git.
//
// Commands that touch files under .beads/ with git (checkout, status, commit)
// must run against the owner of this directory, not the current checkout,
// or they read and write a copy the database never sees (BeadsLog-7hz).
func ForeignBeadsDir() string {
	beadsDir := FindBeadsDir()
	if beadsDir == "" {
		return ""
	}
	root := git.GetRepoRoot()
	if root == "" {
		return ""
	}
	if sameDir(filepath.Dir(beadsDir), root) {
		return ""
	}
	return beadsDir
}

func sameDir(a, b string) bool {
	if ra, err := filepath.EvalSymlinks(a); err == nil {
		a = ra
	}
	if rb, err := filepath.EvalSymlinks(b); err == nil {
		b = rb
	}
	return filepath.Clean(a) == filepath.Clean(b)
}
