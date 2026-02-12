# Development Log: Enhanced Eval Trace & Bulletproof Isolation

**Date:** 2026-02-11
**Session:** sess-enhanced-eval

### **Objective:**
To enhance the evaluation harness with reasoning capture, tool output logging, improved efficiency reporting, and a bulletproof isolation lifecycle.

---

### **Phase 1: Trace Enhancement (Icons, Reasoning, & Results)**

**Initial Problem:** Evaluation logs lacked the agent's chain-of-thought, tool execution results, and visual categorization, hindering protocol analysis.

*   **Action Taken:** Updated `OpenCodeTrace` struct to capture `reasoning` and `reasoning_tokens`.
*   **Action Taken:** Implemented `getToolCategoryIcon` to visually distinguish between Hydration (◐), Mapping (●), and Verification (⚪).
*   **Action Taken:** Updated `aggregateEvent` to capture `tool_result` outputs and fixed the logging logic in `runOpenCodeTest` to write them to `.readable.log`.
*   **Action Taken:** Enhanced the comparative report matrix with strategy-weighted emojis and robust efficiency calculations (handling missing Base tokens).
*   **Action Taken:** Softened Base sandbox instructions to encourage natural brute-force behavior instead of early-exit.
*   **Result:** Live logs and reports now provide high-fidelity insight into agent logic, performance, and actual tool results.

---

### **Phase 2: Bulletproof Isolation & Non-Destructive Backups**

**Initial Problem:** Persistent "ghost" worktrees and file reversions caused by shared git index corruption and destructive `git stash` operations.

*   **Diagnosis:** `git worktree add` shares the main repo's `.git` directory. Agents running git commands in sandboxes were accidentally reverting files in the main repo.
*   **Action Taken:** Rewrote `createEvalWorktree` to use **Total Isolation** via `git init` (Fresh repo clones instead of worktrees).
*   **Action Taken:** Replaced `git stash` with **Invisible Snapshots** (using `git commit-tree` plumbing) to create safety backup branches without clearing the workspace or moving off the current branch.
*   **Action Taken:** Added `OPENCODE_DISABLE_GIT=1` and `OPENCODE_NO_PARENT_CONFIG=1` to sandbox environments to prevent OpenCode from interacting with host git configuration.
*   **Result:** The "reversion loop" is broken. Sandboxes are physically separated from the project tree, and work is safeguarded in visible backup branches.

---

### **Final Session Summary**

**Final Status:** **Complete.** The evaluation harness is now professional-grade, safe, and robust.
**Key Learnings:**
*   **Physical Isolation:** For tools that automate Git, shared worktrees are too risky for uncommitted work.
*   **Plumbing Power:** Git plumbing (`write-tree`, `commit-tree`) allows for high-fidelity backups without the side effects of Porcelain commands like `stash`.

---

### **Architectural Relationships**
- runOpenCodeEval -> createBackupBranch (Invisible Snapshot)
- runOpenCodeEval -> createEvalWorktree (Physical Isolation)
- runOpenCodeEval -> getSandboxEnv (Config Hardening)
- aggregateEvent -> Tool Output Capture
