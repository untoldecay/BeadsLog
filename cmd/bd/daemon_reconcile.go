package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/untoldecay/BeadsLog/internal/storage/sqlite"
	"github.com/untoldecay/BeadsLog/internal/types"
)

func ReconcileGitFacts(ctx context.Context, store *sqlite.SQLiteStorage, log *daemonLogger) error {
	log.log("Starting Git fact reconciliation cycle")

	db := store.UnderlyingDB()

	// 1. Get all branches from sessions
	rows, err := db.QueryContext(ctx, "SELECT DISTINCT branch FROM sessions WHERE branch IS NOT NULL AND branch != 'N/A' AND branch != ''")
	if err != nil {
		log.log("Error querying branches for reconciliation: %v", err)
		return err
	}
	defer rows.Close()

	var branches []string
	for rows.Next() {
		var branch string
		if err := rows.Scan(&branch); err == nil {
			// Clean "(local)" suffix if present
			branch = strings.TrimSuffix(branch, " (local)")
			branch = strings.TrimSpace(branch)
			if branch != "" {
				branches = append(branches, branch)
			}
		}
	}

	log.log("Found %d branches to reconcile", len(branches))

	for _, branch := range branches {
		log.log("Reconciling branch: %s", branch)
		reconcileBranch(ctx, store, branch, log)
	}

	log.log("Git fact reconciliation cycle complete")
	return nil
}

func reconcileBranch(ctx context.Context, store *sqlite.SQLiteStorage, branch string, log *daemonLogger) {
	// Check if branch exists locally or remotely
	exists := true
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	if err := cmd.Run(); err != nil {
		// Try remote
		cmd = exec.Command("git", "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+branch)
		if err := cmd.Run(); err != nil {
			exists = false
		}
	}

	// Check if merged into current HEAD
	isMerged := false
	cmd = exec.Command("git", "merge-base", "--is-ancestor", branch, "HEAD")
	if err := cmd.Run(); err == nil {
		// Validated if it's on a mainline branch OR current branch is mainline
		// and this branch was merged into it.
		// For simplicity, we also check against 'main' explicitly.
		mainCheck := exec.Command("git", "merge-base", "--is-ancestor", branch, "main")
		if err := mainCheck.Run(); err == nil {
			isMerged = true
		} else {
			// Check if current branch IS mainline
			branchOut, _ := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
			curr := strings.TrimSpace(string(branchOut))
			if curr == "main" || curr == "master" || curr == "develop" {
				isMerged = true
			}
		}
	}

	// 3. Deep check if branch is deleted but code landed (Squash-Merge survival)
	if !isMerged && !exists {
		// Find all sessions on this branch that have a commit_sha
		rows, err := store.UnderlyingDB().QueryContext(ctx, "SELECT id, commit_sha FROM sessions WHERE branch LIKE ? AND commit_sha IS NOT NULL AND commit_sha != ''", branch+"%")
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var sessionID, oldSHA string
				if err := rows.Scan(&sessionID, &oldSHA); err == nil {
					found, newSHA, _ := findPatchInHead(oldSHA)
					if found && newSHA != "" {
						isMerged = true
						// RE-ANCHOR: Update the session with the new SHA found in HEAD
						_, err = store.UnderlyingDB().ExecContext(ctx, "UPDATE sessions SET commit_sha = ? WHERE id = ?", newSHA, sessionID)
						if err == nil {
							log.log("Re-anchored session %s: %s -> %s (squash-merge survivor)", sessionID, oldSHA[:8], newSHA[:8])
						}
					}
				}
			}
		}
	}

	// Update cache
	cache := types.BranchCache{
		BranchName:    branch,
		IsMerged:      isMerged,
		IsDeleted:     !exists,
		LastCheckedAt: time.Now(),
	}

	if err := store.UpdateBranchCache(ctx, cache); err != nil {
		log.log("Error updating branch cache for %s: %v", branch, err)
	}
}

// findPatchInHead checks if a commit logic exists in HEAD even if branch was squashed/deleted.
// Returns (found, newSHA, error)
func findPatchInHead(commitSHA string) (bool, string, error) {
	if commitSHA == "" {
		return false, "", nil
	}

	// git patch-id of the target commit
	patchIDCmd := exec.Command("sh", "-c", fmt.Sprintf("git show %s | git patch-id", commitSHA))
	output, err := patchIDCmd.Output()
	if err != nil {
		return false, "", err
	}
	
	parts := strings.Fields(string(output))
	if len(parts) < 1 {
		return false, "", fmt.Errorf("invalid patch-id output")
	}
	targetPatchID := parts[0]

	// Check if this patch-id exists in HEAD (last 100 commits for performance)
	// We use 'git log' to iterate and 'git show | git patch-id' to compare
	// Format: <patch-id> <commit-hash>
	checkCmd := exec.Command("sh", "-c", fmt.Sprintf("git log -n 100 --format=%%H | while read sha; do patch=$(git show $sha | git patch-id); echo \"$patch $sha\"; done | grep ^%s", targetPatchID))
	out, err := checkCmd.Output()
	if err == nil {
		line := strings.TrimSpace(string(out))
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			// fields[0] is patch-id, fields[1] is the inner patch-id, fields[2] is the commit hash
			return true, fields[2], nil
		}
	}

	return false, "", nil
}
