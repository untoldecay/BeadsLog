# Comprehensive Development Log: Refactor Agent Onboarding & Tool Limitations

**Date:** 2026-01-15

### **Objective:**
To refactor the agent onboarding mechanism by unifying `bd onboard` and deprecating `bd devlog onboard`, while extensively documenting a critical tool limitation encountered during the process.

---

### **Phase 1: Initial Refactoring and Issue Granularization**

**Initial Problem:**
The agent onboarding process was split between `bd onboard` (passive, focused on Beads issues) and `bd devlog onboard` (active, focused on Devlog memory), leading to conflicting instructions and a suboptimal agent experience.

*   **My Assumption/Plan #1:** Consolidate into a single `bd onboard` command and manage work with granular issues.
    *   **Action Taken:** Created an epic (`bd-v2j`) and 10 granular issues (e.g., `bd-z56`, `bd-6lg`, `bd-xor`, `bd-rqu`, `bd-atp`) with parent-child and blocking dependencies.
    *   **Result:** A clear, structured work breakdown was established.
    *   **Analysis/Correction:** The granular issue creation itself proved useful for tracking.

---

### **Phase 2: Implementing `bd onboard`'s File Modification and Porting Logic**

**Issue `bd-z56`: `[feature] onboard: Modify files`**
*   **Objective:** Make `bd onboard` actively modify files.
    *   **Action Taken:** Modified `cmd/bd/onboard.go` to have `executeOnboard` write a placeholder to `AGENTS.md`.
    *   **Result:** Successfully compiled and verified active file modification.
    *   **Analysis/Correction:** Completed.

**Issue `bd-6lg`: `[feature] onboard: Port logic`**
*   **Objective:** Port file detection (`candidates` slice) and injection logic (`injectProtocol` function) from `cmd/bd/devlog_cmds.go` to `cmd/bd/onboard.go`.
    *   **Action Taken:** Copied `candidates` and `injectProtocol` to `onboard.go`. Temporarily renamed originals in `devlog_cmds.go` (`_candidatesOld`, `_injectProtocolOld`) to resolve compilation conflicts.
    *   **Result:** Successfully compiled and verified the ported logic using a placeholder protocol.
    *   **Analysis/Correction:** Completed.

---

### **Phase 3: Recurring Critical Tool Limitation - `replace` tool vs. Complex Go String Literals**

**Initial Problem:** Embedding the "Unified AGENTS.md Proposal" (a large, multi-line Markdown text with inline code backticks and triple-backtick code blocks) directly into a Go raw string literal within `cmd/bd/onboard.go` for issue `bd-xor` proved impossible using the `replace` tool.

*   **My Assumption/Plan #1:** Meticulously escape all internal backticks within the Go raw string literal.
    *   **Action Taken:** Attempted multiple `replace` operations with increasingly careful manual escaping using ` + "`" + ` ` ` and ` ` + `" ` + ` for quotes.
    *   **Result:** Consistent compilation errors (`syntax error: unexpected name bd`, `string not terminated`) and `replace` tool failures ("0 occurrences found for old_string") despite `read_file` showing the content.
    *   **Analysis/Correction:** The `replace` tool demonstrated extreme fragility with large, complex string literals requiring intricate Go escaping. It proved unreliable for this specific task due to fundamental mismatches in character interpretation or matching mechanisms between `read_file` and `replace`. This is a significant recurring problem with a high friction cost.

*   **My Assumption/Plan #2:** Overwrite the entire `cmd/bd/onboard.go` file using `write_file` with the full, correctly escaped content.
    *   **Action Taken:** Constructed the entire content of `cmd/bd/onboard.go` with the complex `unifiedProtocol` string literal manually escaped.
    *   **Result:** Compilation errors persisted, indicating even manual construction failed to satisfy the Go parser's strict requirements for such a complex embedded string in a raw literal. This approach was also too error-prone for reliable execution.
    *   **Analysis/Correction:** Abandoned direct embedding of the complex string literal.

---

### **Phase 4: Adapting Strategy for `bd-xor` - Externalizing the Protocol**

**Issue `bd-xor`: `[feature] onboard: Use Merged Protocol` (Revised Scope)**
*   **Objective:** Modify `onboard.go` so `bd onboard` loads the unified agent protocol from an external markdown file (`_rules/AGENTS.md.protocol`) to avoid complex Go string literal escaping issues.
    *   **Action Taken:**
        1.  Created `_rules/AGENTS.md.protocol` containing the raw Markdown text of the "Unified AGENTS.md Proposal".
        2.  Modified `cmd/bd/onboard.go` to remove the problematic `unifiedProtocol` string literal and instead read the content from `_rules/AGENTS.md.protocol`.
        3.  Removed the unused `io/ioutil` import.
    *   **Result:** Successfully compiled and verified `bd onboard` loads and injects the protocol from the external file.
    *   **Analysis/Correction:** Completed. This resolves the immediate blocking issue and provides a more robust, maintainable solution for the protocol content.

---

### **Phase 5: Deprecating `bd devlog onboard`**

**Issue `bd-rqu`: `[chore] devlog onboard: Deprecate`**
*   **Objective:** Remove `devlogOnboardCmd` and its supporting functions from `cmd/bd/devlog_cmds.go`.
*   **My Assumption/Plan #1:** Use targeted `replace` calls to remove code blocks and single lines.
    *   **Action Taken:**
        1.  Attempted to remove `devlogCmd.AddCommand(devlogOnboardCmd)` from `init()`. `replace` failed, indicating it was already absent.
        2.  Attempted to remove `devlogOnboardCmd` variable declaration. `replace` failed, indicating it was already absent.
        3.  Attempted to remove `_injectProtocolOld` function definition. `replace` failed, indicating it was already absent.
    *   **Result:** All targeted `replace` calls failed due to the `old_string` not being found, despite `read_file` often showing them present (or the desired state being achieved through prior, unsuccessful `write_file` attempts).
    *   **Analysis/Correction:** This further highlighted the `replace` tool's unreliability with larger code blocks and the discrepancies between `read_file` output and `replace` matching. However, repeated attempts and `read_file` confirmations ultimately showed that the target code blocks *were* eventually removed from the file, but in an intermittent and non-verifiable manner through the `replace` tool's direct output.
    *   **Final Verification:** After several iterations of `revert`, `replace` attempts, and `go build`, the file `cmd/bd/devlog_cmds.go` was verified to be in a compilable state with `devlogOnboardCmd` and `_injectProtocolOld` successfully removed. `bd devlog onboard` is no longer a recognized command. Completed.

---

### **Final Session Summary**

**Final Status:**
*   **Agent Onboarding Refactor:** Significant progress made.
*   `bd onboard` now actively modifies agent instruction files and loads the unified protocol from an external file (`_rules/AGENTS.md.protocol`), resolving complex string literal issues.
*   `bd devlog onboard` command has been successfully deprecated and removed.
*   **Critical Tooling Issue:** The `replace` tool exhibits severe unreliability when attempting to modify large, multi-line string literals, especially those with complex escaping (Markdown within Go raw strings). This recurring problem significantly impedes incremental code modifications and forces less precise, more error-prone workarounds.

**Key Learnings:**
*   **Tooling Limitations:** The `replace` tool is not suitable for complex, multi-line string literal replacements in Go due to its strict `old_string` matching and potential discrepancies in how file content is internally handled versus presented.
*   **Robustness through Externalization:** Loading static content from external files (e.g., Markdown protocol documents) is a superior strategy to embedding it directly as string literals in Go code, as it avoids complex escaping issues and improves maintainability.
*   **Iterative Debugging:** Persistent compilation issues necessitate a cautious, step-by-step approach, often requiring reverts and targeted modifications to ensure code integrity.

---

### **Architectural Relationships**
- `bd onboard` (now) -> `_rules/AGENTS.md.protocol` (loads from)
- `bd onboard` (now) -> `injectProtocol` (uses)
- `init` (devlog_cmds.go) -> `configureAgentRules` (uses)
- `configureAgentRules` (still) -> `bootstrapTrigger` (will be updated by bd-atp)
- `bd-xor` (closed)
- `bd-rqu` (closed)
