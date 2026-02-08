# Comprehensive Development Log: Migrate Protocol Tags to XML

**Date:** 2026-02-08

### **Objective:**
To improve agent compliance with the BeadsLog protocol by switching from HTML-style comment tags (`<!-- BD_PROTOCOL_START -->`) to semantic XML tags (`<beads_protocol>`). The hypothesis is that models may treat HTML comments as low-priority metadata, whereas XML tags are often interpreted as structured instructions. A migration strategy is implemented to support legacy tags in existing repositories.

---

### **Phase 1: Analysis & Brainstorming**

**Initial Problem:** Agents occasionally ignored the "Always-On Protocol" or required explicit prompting to use `bd` commands, suggesting the protocol's "weight" in the context window was insufficient.

*   **My Assumption/Plan #1:** The use of `<!-- ... -->` delimiters might be signaling "comment/ignore" to the LLM.
    *   **Action Taken:** Analyzed `AGENT.md` content. Confirmed instructions are *between* tags, not *inside* them, but the framing might still be an issue.
    *   **Result:** Proposed switching to `<beads_protocol>` to signal "mandatory structure".

---

### **Phase 2: Implementation - Tag Migration**

**Initial Problem:** Changing the tag constants in `cmd/bd/protocol.go` would break `bd onboard` for existing repositories, as it wouldn't recognize the old tags to update them.

*   **My Assumption/Plan #1:** Implement a "Legacy Tag" detection and upgrade path.
    *   **Action Taken:**
        1.  Defined `LegacyProtocolStartTag` (`<!-- BD_PROTOCOL_START -->`) in `cmd/bd/protocol.go`.
        2.  Updated `ProtocolStartTag` to `<beads_protocol>`.
        3.  Updated `finalizeOnboarding` in `cmd/bd/onboard.go` to detect legacy tags and replace the entire block with the new format.
        4.  Updated `injectBootstrapTrigger` in `cmd/bd/devlog_cmds.go` to respect legacy tags as a valid "initialized" state (preventing downgrade).

---

### **Phase 3: Testing & Verification**

**Initial Problem:** Need to verify that existing "Legacy Ready" repos are upgraded, and new repos work correctly.

*   **My Assumption/Plan #1:** Run a sandbox test simulating a legacy repo.
    *   **Action Taken:** Created `_sandbox/tag-migration-test/test.sh`.
    *   **Steps:**
        1.  Initialize repo with legacy HTML tags in `AGENT.md`.
        2.  Run `bd ready`.
        3.  Verify tags changed to `<beads_protocol>`.
        4.  Run `bd onboard` to verify idempotency (no downgrade).
    *   **Result:** Test passed. Migration and idempotency confirmed.

---

### **Final Session Summary**

**Final Status:** **Implemented & Verified.** The protocol now uses semantic `<beads_protocol>` tags. Existing repositories will automatically migrate to the new format upon running `bd onboard` or `bd ready`.
**Key Learnings:**
*   **Migration Patterns:** When changing delimiters in managed files, always keep the old delimiters as constants to allow for detection and seamless upgrade.
*   **Semantic Framing:** LLMs are sensitive to wrapping. XML tags are a strong signal for "structured data/instructions" compared to HTML comments.

---

### **Architectural Relationships**
<!-- Format: [From Entity] -> [To Entity] (relationship type) -->
- finalizeOnboarding -> LegacyProtocolStartTag (detects)
- finalizeOnboarding -> ProtocolStartTag (writes)
- AGENT.md -> beads_protocol (contains)
