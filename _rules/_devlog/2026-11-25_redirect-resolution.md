# Comprehensive Development Log: One redirect resolver, one lookup order (BeadsLog-svg.2)

**Date:** 2026-11-25

## Problem
Agent_Collab handoff 133ebe11, asks 4-7. `bd worktree create` wrote redirects relative
to `.beads` while `FollowRedirect` resolves from the project root, so deep or unrelated
paths landed one directory off. `FindBeadsDir` and `FindDatabasePath` could pick
different places in a worktree, `bd worktree info` ignored the worktree's redirect, and
a broken redirect only warned before falling back to another database.

### **Phase 1: Diagnosis**

*   **Result:** Five separate readers interpreted the file: FollowRedirect (project root), worktree_cmd, routing and activity (relative to `.beads`), and two doctor checks that joined even absolute paths onto the cwd, so doctor rejected valid redirects.
*   **Result:** `worktree info` was not a /tmp symlink problem: `findLocalBeadsDir` always chose the main repository in a worktree.
*   **Result:** The directory/database disagreement appears when the owning clone has config or JSONL but no database. The `~/.beads/default.db` fallback in the handoff was not reproduced; the path only exists in quickstart help text.

### **Phase 2: Fix**

*   **Action Taken:** `ResolveRedirect` is the only interpreter: absolute, then project-root relative, then `.beads` relative (legacy), preferring a directory with project files. All readers use it.
*   **Action Taken:** `worktreeRedirectTarget` is checked first by both `FindBeadsDir` and `findDatabaseInTree`. `worktree create` writes absolute paths.
*   **Action Taken:** `CheckRedirect` gates every database command in main.go. `where`, `worktree` and `doctor` still run; `--db`/`BEADS_DB` bypass it.

### **Phase 3: The commit blocker the tests caught**

*   **Result:** Installed pre-commit hooks run `bd check --hook` and `bd sync --flush-only` and abort the commit on failure, so the new error would have blocked code commits. Inside a git hook those two now explain and exit 0 without reading or writing anything.

### **Phase 4: Tests**

*   **Result:** Unit tests in internal/beads/redirect_test.go; E2E Test 35 (handoff tests 5-7, legacy relative, broken, hook). E2E 35/35; no new Go failures vs main.

## Final Session Summary

**Final Status:** Done on fix/beads-redirects (stacked on fix/devlog-worktree-ghosts).
**Key Learnings:**
*   **One file format, one parser:** every copy of the resolver had drifted.
*   **A new hard error must be tested through git hooks:** bd runs inside other people's commits.

### Architectural Relationships
- FollowRedirect -> ResolveRedirect (uses)
- CheckRedirect -> ResolveRedirect (uses)
- FindBeadsDir -> worktreeRedirectTarget (uses)
- findDatabaseInTree -> worktreeRedirectTarget (uses)
- worktree_cmd.go -> ResolveRedirect (uses)
- CheckRedirectTargetValid -> ResolveRedirect (uses)
- routes.go -> FollowRedirect (uses)
