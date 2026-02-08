# Validation Report: Worktree-Based Eval System

**Date:** 2026-02-08
**Feature:** `bd eval task` (Worktree Sandboxing)
**Status:** ✅ **FEASIBLE (with External Runner)**

## 1. Feasibility Analysis

### ✅ Worktree Isolation (Verified)
The core concept of using `git worktree` for isolation is sound.
- **Challenge:** `bd`'s discovery logic (`FindBeadsDir`) explicitly prioritizes the **Main Repository Root** when inside a worktree. This breaks isolation by default.
- **Solution:** The harness **MUST** explicitly inject `BEADS_DIR=<worktree_path>/.beads` into the environment of any child process it spawns. This overrides the default discovery and forces usage of the sandbox database.

### ✅ Execution Capability (Solved)
Previous gap: `bd` could not run the agent itself.
**Solution:** Use **`gemini-cli`** as the explicit runner.
- **Integration:** The `eval_task.go` implementation will spawn `exec.Command("gemini", ...)` inside the worktree.
- **Authentication:** Inherits `GEMINI_API_KEY` from the user's environment (standard behavior).
- **Isolation:** Uses `--session-id` unique to the trace to ensure no context leakage.
- **Context:** Injects `../BEADSLOG_AGENTS.MD` (parent path relative to worktree) as system context.

## 2. Validated Workflow

1.  **Stash:** `bd eval task` auto-stashes current work.
2.  **Worktrees:** Creates `eval-test_implicit` and `eval-test_explicit`.
3.  **Init:** Runs `bd init` in each worktree with `BEADS_DIR` override.
4.  **Run (Implicit):** Spawns `gemini --session-id X ...` with task description.
5.  **Run (Explicit):** Spawns `gemini --session-id Y ...` with task + "USE BD COMMANDS".
6.  **Trace:** Captures stdout/stderr and audit logs.
7.  **Report:** `bd eval report` aggregates data.
8.  **Cleanup:** Removes worktrees and restores stash.

## 3. Implementation Plan

1.  **`cmd/eval_task.go`:** Implement the harness logic.
    *   Add `runEvalTest` wrapper around `gemini` binary.
    *   Add environment injection `BEADS_DIR` for child processes.
2.  **Dependencies:** Ensure user has `gemini` installed (check via `exec.LookPath`).
3.  **Testing:** Manual E2E test with `gemini` installed in sandbox.

This plan is now fully validated and ready for implementation.