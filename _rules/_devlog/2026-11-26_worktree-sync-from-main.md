# Comprehensive Development Log: bd sync in a linked worktree rolled back the branch's issues.jsonl (BeadsLog-7hz)

**Date:** 2026-11-26

## Problem
Running `bd sync` in Fray's linked worktree (`worktree-1`, branch `feature/protected-note`, no upstream) printed
`git status failed: exit status 128`, fell back to `--from-main`, and replaced the branch's `.beads/issues.jsonl`
(141 issues) with origin/main's June copy (79), staged. The "147 unchanged" import line came from the main
checkout's JSONL, so the output looked healthy. No data was lost from the database. Reproduced identically with
0.65.0 and 0.66.0-dev.1: pre-existing, not caused by the worktree-safety PRs.

### **Phase 1: Root cause**

*   **Action Taken:** Traced three mismatches between "the checkout bd runs in" and "the checkout that owns the database":
    1. `config.Initialize` walked up from cwd and read the worktree's own committed `config.yaml`, which predates `sync-branch: beads-metadata` (set on develop). The database is the main checkout's.
    2. `gitHasUncommittedBeadsChanges` ran `git status <main checkout>/.beads/issues.jsonl` from the worktree: git refuses paths outside the repository (exit 128). Only redirects were handled.
    3. With no sync branch and no upstream, `doSyncFromMain` ran `git checkout origin/main -- .beads/` in the worktree, then imported the JSONL next to the database: two different files.
*   **Result:** `restoreBeadsDirFromBranch` had the same flaw on the sync-branch path, and would have failed (or discarded the main checkout's uncommitted `.beads`) once the config was fixed.

### **Phase 2: Fix**

*   **Action Taken:** `beads.ForeignBeadsDir()` returns the `.beads` in use when its parent is not this checkout's root (linked worktree, redirect, BEADS_DIR). Empty outside git.
*   **Action Taken:** `config.UseProjectBeadsDir(dir)` makes `Initialize` read that directory's `config.yaml` first, falling back to the cwd walk when it has none. main.go sets it before `config.Initialize`.
*   **Action Taken:** sync_git.go: `beadsDirIsElsewhere()` (redirect or foreign) runs `git status` from the owning checkout and skips the post-sync `.beads` restore, as redirects already did.
*   **Action Taken:** `doSyncFromMain` refuses when the database belongs to another checkout (auto fallback and explicit `--from-main`), naming the checkout to run it in and the config.yaml to set `sync-branch` in. The misleading "using --from-main mode" line is no longer printed first.
*   **Result:** Fray-like repro: with sync-branch in the main checkout, the worktree syncs through `beads-metadata` and its files stay clean; without it, bd refuses and changes nothing.

### **Phase 3: Tests**

*   **Action Taken:** `TestForeignBeadsDir` (main checkout, inside .beads, worktree with and without its own .beads, outside git) and `TestUseProjectBeadsDir`. E2E Test 37: sync-branch followed from a worktree, refusal (auto and explicit) leaves the branch copy intact, and the main checkout still syncs from main on a branch without upstream.
*   **Result:** E2E 37/37. Go short suite: the same 9 pre-existing failures; `TestDetectPrefix` fails on the base commit too when run under this repo's config.yaml (it reads `issue-prefix: BeadsLog`).

## Final Session Summary

**Final Status:** Done on fix/worktree-sync (stacked on feature/dev-channel). Fray worktree-1 file restored (141 lines, clean).
**Key Learnings:**
*   **Database owner, not cwd:** every git command and config read about `.beads` must use the checkout that owns the database.
*   **An import that reports "unchanged" can hide a wrong file:** the import read a different JSONL from the one the checkout replaced.
*   **Known limit:** a worktree branch that intentionally edits config.yaml no longer sees its own copy while it shares the main database.

### Architectural Relationships
- ForeignBeadsDir -> FindBeadsDir (uses)
- ForeignBeadsDir -> git.GetRepoRoot (uses)
- config.Initialize -> UseProjectBeadsDir (reads)
- doSyncFromMain -> ForeignBeadsDir (guards with)
- gitHasUncommittedBeadsChanges -> beadsDirIsElsewhere (uses)
- restoreBeadsDirFromBranch -> beadsDirIsElsewhere (uses)
- syncbranch.Get -> config.yaml (reads)
