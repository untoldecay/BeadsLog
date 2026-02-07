# Comprehensive Development Log: Implement Vercel-style Always-On Agent Protocol

**Date:** 2026-02-07

### **Objective:**
To improve agent reliability and compliance by adopting the "Always-On Protocol" pattern (inspired by Vercel's research). Instead of relying on agents to conditionally load `PROTOCOL.md`, we inject the core retrieval and mapping instructions directly into the agent's system prompt (via `AGENT.md`), removing decision friction.

---

### **Phase 1: Analysis & Template Update**

**Initial Problem:** Agents often skip the "Map the Landscape" step because it requires an extra file read (`PROTOCOL.md`).

*   **My Assumption/Plan #1:** Replace the `FullBootloader` template (injected by `bd ready`) with the compressed, always-on protocol.
    *   **Action Taken:** Modified `cmd/bd/init_templates.go`. Replaced `FullBootloader` content with the `BEADSLOG AGENTS.MD` format.
    *   **Action Taken:** Updated `ProtocolMdTemplate` and `WorkingProtocolMdTemplate` to serve as "Detailed References" rather than primary instruction files, linking back to the always-on context.
    *   **Result:** The templates now reflect the new strategy.

---

### **Phase 2: Verification Strategy**

**Initial Problem:** Need to ensure the "Trap -> Onboard -> Unlock" flow still works and correctly upgrades the file.

*   **My Assumption/Plan #1:** Run a manual simulation or existing tests.
    *   **Action Taken:** Created a dedicated test script `_sandbox/powersearch-e2e/test_protocol.sh`.
    *   **Action Taken:** The script initializes a repo, verifies the trap, runs `bd onboard` and `bd ready`, and then inspects `AGENT.md` for the new headers and lack of legacy instructions.
    *   **Challenge:** Initial test failed because I forgot that `bd onboard` only *shows* instructions; `bd ready` is what actually triggers the file update (`finalizeOnboarding`).
    *   **Correction:** Updated the test script to run `$BD ready` after onboarding.
    *   **Result:** Test passed. `AGENT.md` was correctly transformed from the trap to the Always-On Protocol.

---

### **Final Session Summary**

**Final Status:** **Implemented & Verified.** The system now injects a robust, retrieval-first protocol directly into agent instructions upon initialization.
**Key Learnings:**
*   **Workflow Nuance:** `bd onboard` is a guide; `bd ready` is the activator. This separation allows users/agents to read the guide before committing to the protocol, but tests must account for the two-step process.
*   **Template Management:** Centralizing templates in Go constants (`init_templates.go`) makes updates easy but requires a binary rebuild to test.

---

### **Architectural Relationships**
<!-- Format: [From Entity] -> [To Entity] (relationship type) -->
- FullBootloader -> AGENT.md (injects_into)
- finalizeOnboarding -> AGENT.md (writes)
- bd ready -> finalizeOnboarding (calls)
- ProtocolMdTemplate -> PROTOCOL.md (generates)
