# Comprehensive Development Log: Validate Worktree-Based Eval System

**Date:** 2026-02-08

### **Objective:**
To analyze and validate the implementation plan for `bd eval task`, a revamped evaluation system using `git worktree` for isolated A/B testing between "Implicit" (Always-On Protocol) and "Explicit" (Tool-Specific instructions) agent sessions.

---

### **Phase 1: Feasibility Analysis & Confrontation**

**Initial Problem:** `bd` needs a way to run isolated evals without disk-heavy repo copies and without manual setup.

*   **Assumption:** `git worktree` provides perfect isolation.
    *   **Verification:** Created `_sandbox/wt-test/` to simulate worktree usage. 
    *   **Discovery:** Confirmed that `bd` discovery logic (`FindBeadsDir`) prioritizes the **Main Repository Root** even when inside a worktree. This would cause evals to share the same database, breaking isolation.
    *   **Correction:** The implementation must explicitly set `BEADS_DIR` environment variable for all child processes to force local sandbox database usage.

---

### **Phase 2: The Runner Gap**

**Initial Problem:** `bd` is a tracker, not an agent runner. It cannot execute prompts autonomously.

*   **My Assumption/Plan #1:** Integrate with an external runner.
    *   **Action Taken:** Analyzed `_rules/_analysis/2026-01-08_GeminiCli-wrapper.md`.
    *   **Result:** Validated that wrapping `gemini-cli` is feasible. We can use `--session-id` for session isolation and pipe the task via `stdin`.
    *   **Architecture:** `bd eval task` will act as a **Harness**, managing the lifecycle (Stash -> Worktree -> Gemini -> Report -> Cleanup).

---

### **Final Session Summary**

**Final Status:** **Validated & Ready.** The architectural gaps (Isolation and Execution) have been solved via `BEADS_DIR` injection and `gemini-cli` wrapping.
**Key Learnings:**
*   **Environment Primacy:** In a complex tool like `bd` that has "smart" path discovery, environment variables are the most reliable way to enforce sandbox boundaries.
*   **Harness Pattern:** CLI tools should focus on environment orchestration rather than trying to build internal agent loops when external robust runners exist.

---

### **Architectural Relationships**
<!-- Format: [From Entity] -> [To Entity] (relationship type) -->
- bd eval task -> git worktree (orchestrates)
- bd eval task -> gemini-cli (wraps)
- bd eval task -> BEADS_DIR (enforces isolation)
- bd eval report -> interactions.jsonl (aggregates)
