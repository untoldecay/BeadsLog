# Comprehensive Development Log: Optimize Agent Instructions for Codebase Inquiries

**Date:** 2026-01-20

### **Objective:**
To enforce a "Beads First" workflow where agents must use `bd` and `bd devlog` tools to gather context *before* answering user questions about the codebase or architecture. This ensures answers are grounded in historical context and existing tasks.

---

### **Phase 1: Protocol Optimization**

**Initial Problem:** Agents were defaulting to reading files or `grep` when asked about code, often missing the "why" stored in Devlog or duplicating existing tasks.

*   **My Assumption/Plan #1:** I need to modify the "During Work" section of the agent protocol to explicitly trigger on user questions.
    *   **Action Taken:** Updated `_rules/AGENTS.md.protocol` to add a new "Codebase Inquiry & Exploration (MANDATORY FIRST STEP)" section.
    *   **Result:** The new section explicitly lists `bd devlog search`, `bd devlog impact`, and `bd search` as required first steps for general questions, dependency checks, and status checks.
    *   **Analysis/Correction:** The original plan to just "optimize" was too vague. I made it concrete by defining specific triggers ("IF the user asks...") and mandatory actions.

*   **My Assumption/Plan #2:** Running `bd onboard` will propagate changes to all agent files.
    *   **Action Taken:** Ran `bd onboard`.
    *   **Result:** `GEMINI.md` was updated. `AGENTS.md` was skipped as "up to date" (likely because it was manually modified or just matching protocol, but `GEMINI.md` update confirms propagation works).

---

### **Phase 2: Fix Stale Protocol in Binary**

**Initial Problem:**
The user reported that even after installing the new binary, `bd onboard` was injecting the *old* protocol in other projects.

*   **My Assumption/Plan #1:** I assumed `go:embed` was reading `_rules/AGENTS.md.protocol` directly.
    *   **Investigation:** Search history revealed `[feature] Protocol Embedding` session where `go:embed` was avoided due to directory restrictions.
    *   **Finding:** The protocol is hardcoded as a `const string` in `cmd/bd/protocol.go` to simulate embedding. This file must be manually updated when the protocol changes.
    *   **Action Taken:** Manually updated `cmd/bd/protocol.go` with the new content from `_rules/AGENTS.md.protocol`.

*   **Result:**
    *   The `bd` binary (when rebuilt) will now carry the correct, updated protocol string.
    *   `bd onboard` will now correctly update agents in other repositories.

---

### **Phase 3: Automate Protocol Embedding**

**Initial Problem:**
Manual synchronization of `cmd/bd/protocol.go` with `_rules/AGENTS.md.protocol` is technical debt and error-prone. The user removed a redundant string from the protocol source, but the binary would still have the old version unless manually updated.

*   **My Assumption/Plan #1:** Use `go generate` to automate the synchronization.
    *   **Action Taken:**
        *   Created `cmd/bd/gen_protocol.go`: A generator program that reads the protocol markdown and outputs it as a Go string literal to `protocol_content.go`.
        *   Updated `cmd/bd/protocol.go`: Added `//go:generate` directive and removed the hardcoded string.
        *   Ran `go generate ./cmd/bd`.
    *   **Result:** `cmd/bd/protocol_content.go` is now automatically generated from `_rules/AGENTS.md.protocol`. Future updates only require running `go generate`.

---

### **Final Session Summary**

**Final Status:** The agent protocol is optimized, the redundant "onboard" instruction is removed, and the binary embedding process is fully automated via `go generate`.
**Key Learnings:**
*   Explicit "IF/THEN" rules in agent instructions are more effective than general guidelines.
*   `bd onboard` effectively syncs the central protocol to agent-specific files.
*   `go generate` is a powerful tool for embedding content that defies standard `go:embed` path restrictions, effectively eliminating manual synchronization technical debt.

---

### **Architectural Relationships**
<!-- Format: [From Entity] -> [To Entity] (relationship type) -->
- _rules/AGENTS.md.protocol -> GEMINI.md (propagates to)
- bd onboard -> _rules/AGENTS.md.protocol (reads)
- cmd/bd/protocol.go -> bd onboard (embeds content)
- cmd/bd/gen_protocol.go -> _rules/AGENTS.md.protocol (reads)
- cmd/bd/gen_protocol.go -> cmd/bd/protocol_content.go (generates)
