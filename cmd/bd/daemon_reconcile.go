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
		// Find any session on this branch that has a commit_sha
		var lastSHA string
		err := store.UnderlyingDB().QueryRowContext(ctx, "SELECT commit_sha FROM sessions WHERE branch LIKE ? AND commit_sha IS NOT NULL AND commit_sha != '' ORDER BY timestamp DESC LIMIT 1", branch+"%").Scan(&lastSHA)
		if err == nil && lastSHA != "" {
			found, _ := findPatchInHead(lastSHA)
			if found {
				isMerged = true
				log.log("Branch %s was squashed/rebased into HEAD (found via patch-id)", branch)
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

// findPatchInHead checks if a commit logic exists in HEAD even if branch was squashed/deleted
func findPatchInHead(commitSHA string) (bool, error) {
	if commitSHA == "" {
		return false, nil
	}

	// git patch-id of the target commit
	patchIDCmd := exec.Command("sh", "-c", fmt.Sprintf("git show %s | git patch-id", commitSHA))
	output, err := patchIDCmd.Output()
	if err != nil {
		return false, err
	}
	
	parts := strings.Fields(string(output))
	if len(parts) < 1 {
		return false, fmt.Errorf("invalid patch-id output")
	}
	targetPatchID := parts[0]

	// Check if this patch-id exists in HEAD (last 100 commits for performance)
	// This is still slow. We might want to optimize this.
	checkCmd := exec.Command("sh", "-c", fmt.Sprintf("git log -n 100 --format=%%H | while read sha; do git show $sha | git patch-id; done | grep ^%s", targetPatchID))
	if err := checkCmd.Run(); err == nil {
		return true, nil
	}

	return false, nil
}
