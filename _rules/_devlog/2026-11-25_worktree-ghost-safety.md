# Comprehensive Development Log: Worktree ghost safety (BeadsLog-svg.1)

**Date:** 2026-11-25

## Problem
Agent_Collab handoff 133ebe11 (from Fray): all git worktrees share one beads.db,
but `_rules/_devlog/*.md` is versioned per branch. `bd devlog sync` in a worktree
on an older branch marked 162 live sessions as ghosts, hiding them from
resume/search/list, and `bd devlog prune` would then delete them. The reverse also
happened: a main-checkout sync ghosted a devlog that only existed on an unmerged
branch.

### **Phase 1: Reproduce before designing**

*   **Action Taken:** Reproduced every claim in an isolated sandbox (own HOME, no daemon).
*   **Result:** Confirmed, with three corrections to the ask. `git log --all -- <path>` also matches files deleted long ago, so ghosts could never be real again. The `sha=` in index links is HEAD before the devlog commit, so it never contains the file. The handoff test only covered a never-committed file.
*   **Result:** Found two paths the handoff missed. The shared post-commit/post-merge hook runs sync in every worktree, and `bd doctor --fix` had its own unguarded copy of prune.

### **Phase 2: Presence**

*   **Action Taken:** New package `internal/devlog` with `Presence`: a file is Here, Elsewhere, Absent or Unknown. Elsewhere = on a local/remote branch TIP, on another worktree's HEAD, or on disk in another worktree. Git is queried only when a file is missing here. Tips are resolved with one `cat-file --batch-check` and each distinct tree is listed once.
*   **Action Taken:** Unknown (git failed inside a repo) never marks or deletes. GIT_DIR & co. are stripped from the child environment because hooks export them.
*   **Action Taken:** `Restriction()` says when a checkout may not decide at all: a linked worktree (unless the repo is bare, which has no main checkout), or a checkout redirected to another repository's database.

### **Phase 3: Wiring**

*   **Action Taken:** `SyncSession` and the sync reconciliation use Presence. Elsewhere keeps the narrative; a narrative already replaced by the problem text is fetched back from git. `RestoreFound` un-marks false ghosts.
*   **Action Taken:** `PruneGhosts` is now the only deletion path, used by `devlog prune` and `doctor --fix`, in one transaction. The doctor copy matched index rows by title and could drop a live session's row.
*   **Action Taken:** `devlog show` falls back to the file's branch or worktree.

### **Phase 4: Tests**

*   **Result:** 12 unit tests in internal/devlog; E2E Test 34 (handoff scenarios plus hook, doctor, committed-then-deleted, un-marking, narrative). E2E 34/34; no new Go failures vs main (9 short and 4 integration failures pre-exist).

## Final Session Summary

**Final Status:** Done on fix/devlog-worktree-ghosts.
**Key Learnings:**
*   **"Not here" is not "gone":** a shared database must not trust one checkout's disk.
*   **Tips, not history:** history answers "did it ever exist", which is the wrong question for ghosts.
*   **Known limits:** a devlog deleted on main stays un-ghosted while a stale branch still has it. A separate clone's unpushed branches are invisible to the clone that owns the database; that session can still be ghosted and pruned there, and comes back when the branch is merged and synced. Content overwrite from older branches is tracked in BeadsLog-5zn.

### Architectural Relationships
- devlog-sync -> DevlogPresence (uses)
- SyncSession -> DevlogPresence (uses)
- devlog-prune -> PruneGhosts (uses)
- DevlogGhosts -> PruneGhosts (uses)
- PruneGhosts -> DevlogPresence (depends on)
- devlog-show -> DevlogPresence (uses)
- reconcileGhosts -> RestoreFound (uses)
