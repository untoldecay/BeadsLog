# Comprehensive Development Log: Finalize OpenCode Eval Harness

**Date:** 2026-02-10

### **Objective:**
To finalize the `bd eval task` command by implementing a robust, isolated execution harness using Git Worktrees, real-time output visibility, automated stashing, and accurate token counting.

---

### **Phase 1: Worktree Isolation & Stability**

**Initial Problem:** The evaluation harness was silently reverting file changes made by the agent inside the worktree. This was caused by the worktree sharing the index or state with the main repo, or Git autorevert mechanisms triggering on dirty states.

*   **My Assumption/Plan #1:** Use `git worktree add` with standard branch checkout.
    *   **Action Taken:** Implemented basic worktree creation.
    *   **Result:** Files modified by the agent (OpenCode) were reverted immediately.
    *   **Analysis/Correction:** The worktree index was likely stale or conflicted. Switched to a **Detached HEAD** approach (`git worktree add -d`) and forced a fresh index reset (`git rm --cached -r .`, `git read-tree HEAD`, `git checkout-index -a --force`). Also disabled `core.fileMode` to prevent permission-based diffs.

*   **My Assumption/Plan #2:** The OpenCode agent might be triggering git actions that conflict with the harness.
    *   **Action Taken:** Updated `opencode.json` configuration to explicitly disable `autoAdd`, `autoCommit`, and set `indexUpdate` to `manual`.
    *   **Result:** Stability achieved. Agent modifications now persist correctly within the sandbox.

---

### **Phase 2: Token Counting & Visibility**

**Initial Problem:** The evaluation report was showing 0 tokens and missing detailed tool usage.

*   **My Assumption/Plan #1:** Tokens are available in the `usage` event.
    *   **Action Taken:** Implemented `aggregateEvent` to parse `usage` events.
    *   **Result:** Still 0 tokens.
    *   **Analysis/Correction:** OpenCode streams token usage in `step_finish` events within a `part.tokens` object. Updated `aggregateEvent` to parse this nested structure. Added a unit test `eval_token_test.go` to verify the logic.

*   **My Assumption/Plan #2:** Users need to see what's happening during the run.
    *   **Action Taken:** Implemented a 10-line scrolling viewport in the CLI output that shows real-time tool calls (e.g., `🛠️  [TOOL] bash(ls -la)`).
    *   **Result:** excellent visibility into the agent's actions without flooding the console.

---

### **Phase 3: UX & Safety**

**Initial Problem:** Running evals in a dirty workspace risks data loss or contamination.

*   **My Assumption/Plan #1:** Abort if dirty.
    *   **Action Taken:** Added a `checkDirty` function.
    *   **Correction:** Too restrictive for a developer tool. Implemented **Automatic Stashing**.
    *   **Action Taken:** Added `createStash` and `restoreStash` to `eval_task_opencode.go`. The harness now automatically stashes changes before creating worktrees and pops them after cleanup.

*   **My Assumption/Plan #2:** Post-run workflow needs to be flexible.
    *   **Action Taken:** Integrated `charmbracelet/huh` for an interactive post-run menu: `Quit (Clean up)`, `Restart`, or `New Task`.

---

### **Final Session Summary**

**Final Status:** **Complete.** The `bd eval task` command is now a production-grade evaluation harness. It supports A/B testing (Implicit vs Explicit protocols), accurate token metering, safe execution via worktrees, and rich reporting.
**Key Learnings:**
*   **Git Worktrees:** For true isolation, always use detached HEAD and force a complete index reset. Shared branches/indices are a trap for automated tools.
*   **OpenCode Streams:** Token usage data is often nested in lifecycle events (`step_finish`) rather than top-level `usage` events.

---

### **Architectural Relationships**
<!-- Format: [From Entity] -> [To Entity] (relationship type) -->
- eval_task_opencode.go -> createEvalWorktree (uses)
- createEvalWorktree -> git (executes)
- runOpenCodeEval -> createStash (calls)
- runOpenCodeEval -> restoreStash (calls)
- aggregateEvent -> TokenUsage (updates)
