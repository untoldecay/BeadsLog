package main

import (
	"os"
	"strings"
)

// syncBranchExcludeHeader labels the .git/info/exclude block that keeps beads
// data off the user's work branch when a dedicated sync branch is configured.
const syncBranchExcludeHeader = "# Beads sync branch — keep beads data off work branches (bd init)"

// excludeBeadsFromWorkBranch adds .beads/ to .git/info/exclude so a manual
// 'git add -A' (or an agent's blanket stage) never stages beads data onto the
// user's work branch. In sync-branch mode the beads data lives on the dedicated
// branch (committed via an internal worktree that force-adds), so hiding it from
// the work tree is safe and desired. Idempotent; a no-op outside a git repo.
func excludeBeadsFromWorkBranch() {
	path, err := gitExcludePath()
	if err != nil {
		return // not a git repo
	}
	content := ""
	if b, err := os.ReadFile(path); err == nil { // #nosec G304 - path from git-dir
		content = string(b)
	}
	if containsExactPattern(content, ".beads/") {
		return // already excluded (e.g. by solo mode)
	}
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += "\n" + syncBranchExcludeHeader + "\n.beads/\n"
	// #nosec G306 - git exclude is a plain config file
	_ = os.WriteFile(path, []byte(content), 0644)
}
