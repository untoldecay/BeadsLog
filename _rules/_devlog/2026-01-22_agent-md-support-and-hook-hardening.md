# Comprehensive Development Log: Support AGENT.md and Local Hook Fallback

**Date:** 2026-01-22

### **Objective:**
To resolve issues reported with the `opencode` AI coding tool, which uses `AGENT.md` (singular) and often operates in environments without a global `bd` binary. The goal was to ensure `AGENT.md` is correctly onboarded and that git hooks are robust enough to use a local `bd` fallback.

---

### **Phase 1: Expanding Candidate Detection**

**Initial Problem:** `bd onboard` and `bd init` were only looking for `AGENTS.md` (plural), missing tools that use the singular `AGENT.md` or `CODEBASE.md`.

*   **My Assumption/Plan #1:** Add `AGENT.md` and `CODEBASE.md` to the central `Candidates` list.
    *   **Action Taken:** Modified `cmd/bd/onboard.go` to include the new filenames.
    *   **Result:** `bd onboard` now correctly detects these files.

---

### **Phase 2: Hardening Onboarding Logic**

**Initial Problem:** `bd onboard` and `finalizeOnboarding` were too strict, only updating files if they already contained specific tags or the bootstrap trigger.

*   **My Assumption/Plan #1:** Update the logic to ensure that if a candidate file exists but is empty or missing the trigger, it gets initialized correctly.
    *   **Action Taken:** Modified `onboard.go` to prepend the bootstrap trigger during onboarding and the full protocol during finalization if the file was previously unmanaged.
    *   **Result:** `AGENT.md` is now reliably populated even if it starts as an empty file.

---

### **Phase 3: Robust Git Hooks with Local Fallback**

**Initial Problem:** Git hooks were failing in environments where `bd` was not in the global `PATH`, even if a local `./bd` binary was present.

*   **My Assumption/Plan #1:** Update the hook templates to prioritize `./bd` over the global `bd` command.
    *   **Action Taken:** 
        1. Modified `cmd/bd/templates/hooks/pre-commit` to use a `BD_CMD` variable with local fallback.
        2. Updated `cmd/bd/init_git_hooks.go` to use the same logic in the `preCommitHookBody`.
    *   **Result:** Verified in sandbox that hooks now correctly use the local binary, ensuring enforcement works during development.

---

### **Final Session Summary**

**Final Status:** `AGENT.md` and `CODEBASE.md` are now first-class citizens in the onboarding flow. Git hooks are hardened against missing global binaries.
**Key Learnings:**
*   Always provide a local binary fallback in automation scripts (hooks) to support developer environments where global installation might be skipped.
*   Logical "singular vs plural" naming variations are common in AI tools; the `Candidates` list should be as inclusive as possible.

---

### **Architectural Relationships**
<!-- Format: [From Entity] -> [To Entity] (relationship type) -->
- bd onboard -> AGENT.md (singular support)
- bd init -> configureAgentRules (uses updated Candidates)
- pre-commit hook -> ./bd (local fallback)
