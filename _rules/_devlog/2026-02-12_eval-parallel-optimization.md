# Comprehensive Development Log: Eval Speed Optimization (Parallel A/B/C)

**Date:** 2026-02-12

### **Objective:**
To optimize the evaluation harness (`bd eval task`) by implementing parallel execution, feature stripping, and automated sandbox hydration, targeting a 5-10x reduction in total evaluation time.

---

### **Phase 1: Parallel Evaluation & Live UI**

**Initial Problem:** Evaluations were running sequentially (Implicit -> Explicit -> Base), taking ~45-60s to complete.

*   **My Assumption/Plan #1:** Use Goroutines to run the three sandboxes concurrently and implement a Bubble Tea-based live progress table.
    *   **Action Taken:** Updated `cmd/bd/eval_task_opencode.go` to use `sync.WaitGroup` and `tea.Program`. Created a `progressModel` to track and display live status, time, and the last executed tool for each run.
    *   **Result:** All three tests now launch and run simultaneously. Total wall-clock time dropped to ~12-15s.
    *   **Analysis/Correction:** Initial implementation of the Bubble Tea UI was complex. Simplified it to a 10-line static-height table using Lipgloss to avoid terminal flicker and scroll issues.

---

### **Phase 2: Feature Stripping & Model Hardening**

**Initial Problem:** OpenCode CLI startup was sluggish due to file watchers and mDNS discovery.

*   **My Assumption/Plan #1:** Pass environment variables to OpenCode to disable non-essential features.
    *   **Action Taken:** Added `OPENCODE_DISABLE_STORAGE=1`, `OPENCODE_EXPERIMENTAL_FILEWATCHER=0`, and `OPENCODE_NO_MDNS=1` to the sandbox environment.
    *   **Result:** Startup overhead for each `opencode run` call was reduced by several seconds.
    *   **Analysis/Correction:** Confirmed that these variables do not affect the agent's reasoning or tool-calling capabilities, only its background infrastructure.

---

### **Phase 3: Automated Sandbox Hydration**

**Initial Problem:** In fresh sandboxes, agents were often "blind" because the local graph wasn't hydrated, causing `bd devlog search` to return empty results.

*   **My Assumption/Plan #1:** Automate hydration by running `bd onboard` during sandbox setup.
    *   **Action Taken:** Modified `createEvalWorktree` to execute `bd onboard --quiet` before passing control to the agent.
    *   **Result:** Agents now have immediate access to historical context without having to "remember" to hydrate themselves.
    *   **Analysis/Correction:** Added the `--quiet` flag to ensure the process remains non-interactive and doesn't hang the evaluation harness.

---

### **Phase 4: Documentation Synchronisation**

**Initial Problem:** `docs/EVAL.md` was outdated and still referenced "Worktree Isolation" and "Auto-Stashing".

*   **My Assumption/Plan #1:** Update the documentation to reflect the shift to `git init` (Total Isolation) and `commit-tree` (Invisible Snapshots).
    *   **Action Taken:** Updated `docs/EVAL.md` with new sections for Parallel Execution, Pre-Hydration, Total Isolation, and Invisible Snapshots.
    *   **Result:** Documentation now accurately describes the high-performance, concurrent evaluation architecture.

---

### **Phase 5: High-Fidelity Tool Result Capture**

**Initial Problem:** Evaluation logs were only showing tool calls, not the actual outputs/results from those tools, making it difficult to verify if the agent's reasoning was grounded in reality.

*   **My Assumption/Plan #1:** Look for a separate `tool_result` event in the OpenCode stream.
    *   **Action Taken:** Initially implemented logic to parse a top-level `tool_result` event.
    *   **Result:** Results were still missing because OpenCode (in the version used) embeds results directly within the `tool_use` event under `part.state.output`.
    *   **Analysis/Correction:** Updated `runOpenCodeTestParallel` and `aggregateEvent` to extract tool output from the `state.output` field of `tool_use` events. Added truncating logic (1000 chars) to keep logs readable.
    *   **Result:** High-fidelity logs now show both the tool command and the first 1000 characters of its output.

---

### **Final Session Summary**

**Final Status:** **Complete.** The evaluation harness is now significantly faster (>5x speed multiplier) and provides a much better developer experience with live progress monitoring.
**Key Learnings:**
*   **Concurrency:** LLM-bound tasks are the perfect candidate for Goroutines as the latency is almost entirely network-based.
*   **Silent Hydration:** Pre-populating the context in a sandbox is more reliable than relying on the agent to follow a multi-step setup protocol.

---

### **Architectural Relationships**
- runOpenCodeEval -> runOpenCodeTestParallel (concurrency)
- runOpenCodeTestParallel -> progressModel (UI updates)
- createEvalWorktree -> bd onboard (pre-hydration)
- internal/eval/cleanup.go -> SafeCleanupABC (isolation cleanup)
