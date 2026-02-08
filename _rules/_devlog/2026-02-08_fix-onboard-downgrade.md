# Comprehensive Development Log: Fix Onboard Instruction Downgrade

**Date:** 2026-02-08

### **Objective:**
To resolve a bug in `bd onboard` where it would incorrectly overwrite the "Full Protocol" in agent instruction files (like `AGENT.md` or `GEMINI.md`) with the initial "bootstrap trap" string. This occurred because the idempotency check in `injectBootstrapTrigger` only looked for the trap string itself, ignoring the presence of the more advanced protocol tags.

---

### **Phase 1: Investigation & Analysis**

**Initial Problem:** Users reported that `bd onboard` was resetting their `AGENT.md` to a "SETUP IN PROGRESS" state even after they had already unlocked the full project context.

*   **My Assumption/Plan #1:** The `bd onboard` command is designed to be idempotent and safe to run multiple times. It likely lacks a check for the "Full Protocol" state.
    *   **Action Taken:** Analyzed `cmd/bd/onboard.go` and `cmd/bd/devlog_cmds.go`.
    *   **Result:** Confirmed that `injectBootstrapTrigger` (called by `bd onboard` via `configureAgentRules`) checked for the existence of the `bootstrapTrigger` string. If missing, it proceeded to migrate "legacy content" and overwrite the file with the trigger. Since a "Ready" file has the trigger replaced by the `FullBootloader`, the check failed, and the downgrade occurred.

---

### **Phase 2: Implementation - Protocol Awareness**

**Initial Problem:** `injectBootstrapTrigger` needs to recognize the `Full Protocol` as a terminal/valid state.

*   **My Assumption/Plan #1:** Add a check for `ProtocolStartTag` in `injectBootstrapTrigger`.
    *   **Action Taken:** Modified `cmd/bd/devlog_cmds.go`.
    *   **Result:** Added a check: `if strings.Contains(sContent, ProtocolStartTag) { return false }`. This ensures that if the file already contains `<!-- BD_PROTOCOL_START -->`, the function returns early without modification.

---

### **Phase 3: Testing & Verification**

**Initial Problem:** Verify that the fix prevents the downgrade while still allowing the initial trap to be planted in new/legacy files.

*   **My Assumption/Plan #1:** Create a reproduction script in a sandbox environment.
    *   **Action Taken:** Created `_sandbox/beadslog-976-test/test.sh`. The script:
        1.  Initializes a beads repo.
        2.  Simulates a "Ready" state by manually writing the Full Protocol to `AGENT.md`.
        3.  Runs `bd onboard`.
        4.  Verifies that `AGENT.md` still contains the Full Protocol and NOT the bootstrap trap.
    *   **Result:** The test passed successfully.

---

### **Final Session Summary**

**Final Status:** **Fixed.** `bd onboard` is now safe to run in initialized environments without risk of downgrading agent instruction files.
**Key Learnings:**
*   **State Machine Logic:** When implementing idempotent setup commands, always check for the "most advanced" state first to avoid accidental regressions/downgrades.
*   **Sandbox Testing:** Using isolated repository simulations is the most reliable way to verify CLI behavior that modifies the filesystem.

---

### **Architectural Relationships**
<!-- Format: [From Entity] -> [To Entity] (relationship type) -->
- injectBootstrapTrigger -> ProtocolStartTag (recognizes)
- bd onboard -> injectBootstrapTrigger (calls)
- AGENT.md -> ProtocolStartTag (contains)
