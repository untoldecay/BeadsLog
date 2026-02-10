# Development Log: Finalize OpenCode A/B/C Eval Harness & Worktree Stability

**Date:** 2026-02-10
**Session:** sess-abc-final

### **Objective:**
To implement a production-grade A/B/C evaluation harness for the agent protocol, ensuring reliable execution in isolated sandboxes and detailed performance reporting.

---

### **Phase 1: The A/B/C Harness Implementation**
**Problem:** A simple A/B test (Implicit vs Explicit) didn't provide enough baseline data to quantify the "cost of blindness."
**Action:** Expanded the harness in `cmd/bd/eval_task_opencode.go` to support a triple-run methodology:
1.  **Implicit**: Standard protocol via AGENT.md.
2.  **Explicit**: Prompted with "CRITICAL PROTOCOL" instructions.
3.  **Blind**: Restricted with "Do NOT use any bd commands" to establish a pure brute-force baseline.

---

### **Phase 2: Logic-Based Reporting & Archiving**
**Problem:** Evaluation reports were ephemeral and lacked qualitative depth. "First Tools" wasn't enough to describe the strategy.
**Action:** 
- Refactored `scoreTrace` to detect specific **Logic States**: "Map FIRST + Verify LATER", "Verify BEFORE Map", "Map ONLY (Risk)", and "Pure Brute-Force".
- Implemented `archiveTrace` to persist reconstructed tool calls (with arguments) to `eval/history.jsonl`.
- Updated `bd eval report` in `cmd/bd/eval_cmds.go` to read from history and perform aggressive task clustering via Levenshtein distance, grouping A/B/C runs even if the prompts vary slightly due to instructions.

---

### **Phase 3: Solving the Worktree Index Trap**
**Problem:** The agent's file modifications were being silently reverted by Git inside the worktree sandboxes.
**Diagnosis:** Git worktree creation can leave a stale or shared index that triggers autorevert on "dirty" states.
**Action:** Implemented the "Detached HEAD + Forced Reset" strategy in `createEvalWorktree`:
1.  Used `git worktree add -d` to avoid branch pointer conflicts.
2.  Forced index reset via `git rm --cached -r .`, `git read-tree HEAD`, and `git checkout-index -a --force`.
3.  Disabled `core.fileMode` to prevent permission-based noise.
**Verification:** Validated the fix with a manual reproduction test; modifications now persist correctly under `git status`.

---

### **Phase 4: UX & Cleanup**
**Action:** 
- Added a 10-line live stream viewport during eval runs for real-time visibility.
- Integrated `charmbracelet/huh` for interactive post-run management.
- Hardened `bd eval clean` to prune detached orphan worktrees from both local and `/tmp` directories.

---

### **Key Learnings:**
- **Automated Worktrees:** True isolation requires decoupling the index from the parent repository completely via `read-tree`.
- **Prompt Sensitivity:** Agents follow the protocol significantly better when "Map FIRST" is phrased as a CRITICAL requirement rather than a suggestion.

---

### **Architectural Relationships**
- eval_task_opencode.go -> createEvalWorktree (isolation)
- eval_cmds.go -> eval/history.jsonl (historical truth)
- eval_task_opencode.go -> archiveTrace (persistence)
- scoreTrace -> LogicDesc (qualitative mapping)
