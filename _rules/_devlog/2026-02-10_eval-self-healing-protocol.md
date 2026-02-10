# Development Log: Eval Self-Healing Protocol (BeadsLog-3l0)

**Date:** 2026-02-10
**Session:** sess-self-healing

### **Objective:**
To implement a "sanitize-on-entry" protocol for the evaluation harness, preventing artifact accumulation and ensuring cross-project safety.

---

### **Implementation Details**

#### **1. Project Fingerprinting**
Implemented `getProjectHash` to generate a unique short identifier based on the absolute path of the repository. This is used to namespace all temporary evaluation artifacts.

#### **2. Pre-flight Janitor**
Added a cleanup phase at the very start of `runOpenCodeEval`. It scans the global `git worktree list` and force-removes any worktrees matching the current project's fingerprint. This ensures that even after a hard crash (`SIGKILL` or timeout), the next run starts with a clean registry.

#### **3. Stash Recovery UI**
Developed an interactive detection system for leftover `bd eval auto-stash` entries. 
-   **Detects**: Orphaned stashes from previous interrupted runs.
-   **Actions**: Offers `Restore (Pop)`, `Delete (Clear)`, or `Ignore`.
-   **Safety**: Uses targeted deletion (`git stash drop`) to avoid touching the user's non-evaluation stashes.

#### **4. Parallel Safety**
Temporary sandbox directories are now named `beads-eval-<proj_hash>-<ts>`. This prevents a `bd eval task` run in Project A from accidentally "healing" (deleting) an active eval sandbox in Project B.

---

### **Verification**
Successfully validated the protocol by simulating a timeout/crash (via `Ctrl+C` and 5-minute environment timeout). The subsequent run correctly detected the orphaned stash and pruned the registered worktrees from `/tmp`.

---

### **Architectural Relationships**
- runOpenCodeEval -> getProjectHash (isolation)
- runOpenCodeEval -> pruneProjectOrphans (sanitization)
- runOpenCodeEval -> handleLeftoverStashes (recovery)

### **Refinement: Logic-Based Metrics & Base Reference**
Updated the A/B/C harness to use a static **Base** reference (formerly "Blind") for efficiency calculations.
-   **Base Run**: Hardcoded to `1.0x` efficiency (Reference).
-   **Logic Phases**: Defined strict sequence: `Hydration` -> `Mapping` -> `Verification`.
-   **Efficiency Formula**: `(Tokens(Base) / Tokens(Run)) * StrategyBonus`.
    -   Optimal Strategy Bonus: 1.5x
    -   Shallow Strategy Bonus: 0.8x
    -   Disordered Strategy Bonus: 0.5x
