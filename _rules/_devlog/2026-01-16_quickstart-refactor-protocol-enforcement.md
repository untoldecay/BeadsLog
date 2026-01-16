# Comprehensive Development Log: Quickstart Refactor & Devlog Protocol Enforcement

**Date:** 2026-01-16

### **Objective:**
To refactor the `bd quickstart` command into a unified entry point supporting both Task (forward) and Devlog (backward) workflows, fix outdated references in `bd init`, and enforce the placement of the Devlog Protocol at the top of agent instruction files for better visibility. Later in the session, the focus shifted to hardening the `bd onboard` command by embedding the protocol into the binary and implementing tag-based replacement for safer updates. Finally, `bd init` was updated to support multi-agent files, automatic versioning was implemented, and index integrity checks were added.

---

### **Phase 1: Quickstart Refactor & Init Fixes**

**Initial Problem:**
1.  `bd init` pointed users to a non-existent command: `bd devlog quickstart`.
2.  `bd quickstart` only covered core task management, ignoring the "BeadsLog" memory features.
3.  The CLI lacked a cohesive "getting started" experience for the dual-mode nature of the tool.

*   **My Assumption/Plan #1:** I should just delete the broken link in `bd init`.
    *   **Analysis:** This would leave a documentation gap. A better approach is to implement the missing functionality.
    *   **Correction:** Decided to refactor `bd quickstart` to handle both use cases via flags.

*   **My Assumption/Plan #2:** Create a new `bd devlog quickstart` subcommand.
    *   **Analysis:** This fragments the "help" experience.
    *   **Correction:** Implemented `bd quickstart --tasks` and `bd quickstart --devlog` within the existing `cmd/bd/quickstart.go`. The default `bd quickstart` now serves as a high-level menu/router.

*   **Action Taken:**
    *   Refactored `cmd/bd/quickstart.go` to use `qsTasks` and `qsDevlog` boolean flags.
    *   Implemented three print functions: `printOverview`, `printTasksQuickstart`, and `printDevlogQuickstart`.
    *   Updated `cmd/bd/init.go` to point to the new flagged commands.

*   **Result:**
    *   `bd quickstart` now shows a "Choose your path" menu.
    *   `bd init` outputs correct, actionable next steps.

---

### **Phase 2: Agent Onboarding Protocol Enforcement**

**Initial Problem:**
The `bd onboard` command (formerly `bd devlog onboard`) was appending the mandatory Devlog Protocol to the *bottom* of agent instruction files (e.g., `GEMINI.md`). This risked agents ignoring the protocol due to context window truncation or low prioritization.

*   **My Assumption/Plan #1:** Appending is safer to avoid overwriting user content.
    *   **Analysis:** While safe, it's ineffective for "MANDATORY" protocols that must be read first.
    *   **Correction:** Changed logic to **prepend** the protocol, preserving existing content below it.

*   **Action Taken:**
    *   Modified `injectProtocol` in `cmd/bd/onboard.go`.
    *   Changed string concatenation from `content + protocol` to `protocol + "\n\n" + content`.
    *   Manually removed the old protocol section from `GEMINI.md` to test re-injection.
    *   Ran `bd onboard`.

*   **Result:**
    *   `GEMINI.md` now has the "Devlog Protocol" at the very top.
    *   Existing context remains intact below.

---

### **Phase 3: Protocol Embedding and Tag-Based Replacement**

**Initial Problem:**
The `bd onboard` command failed when run in other repositories because it tried to read `_rules/AGENTS.md.protocol`, a file that only exists in the BeadsLog source repo. Additionally, blindly prepending the protocol on every run would cause duplication if the user edited the file.

*   **My Assumption/Plan #1:** Go `embed` is the standard way to handle this.
    *   **Action Taken:** Created `cmd/bd/protocol.go` with the protocol content as a const string (simulating embed to avoid directory structure complexities with `go:embed` across packages).

*   **My Assumption/Plan #2:** Use HTML comments as tags to mark the protocol section for future updates.
    *   **Action Taken:**
        *   Defined `<!-- BD_PROTOCOL_START -->` and `<!-- BD_PROTOCOL_END -->`.
        *   Rewrote `injectProtocol` in `cmd/bd/onboard.go` to search for these tags.
        *   **Logic:**
            *   **Tags Found:** Replace content *between* tags with the new protocol (Update).
            *   **Tags Missing/Broken:** Prepend the full block (Start Tag + Protocol + End Tag) to the file (Install).
            *   **Idempotency:** If content is identical, do nothing.

*   **Result:**
    *   The binary is now self-contained (no external file dependency).
    *   Updates are safe and targeted.

---

### **Phase 4: Comprehensive Sandbox Testing**

**Initial Problem:**
I needed to verify the complex logic of "Fresh Install", "Update Existing", "Fix Broken Tags", and "Fresh No-Tag File" without risking my own `GEMINI.md`.

*   **My Assumption/Plan #1:** Use the existing `_sandbox/_utils/setup_init_tests.py` framework.
    *   **Action Taken:**
        *   Added scenarios `Test-11` to `Test-15` covering all edge cases.
        *   Generated the sandbox environments.
        *   Manually ran `bd onboard` in each environment and inspected the output.
    *   **Result:**
        *   **Fresh:** Created file with tags.
        *   **Existing:** Prepended block with tags.
        *   **Outdated:** Replaced content between tags correctly.
        *   **Garbage:** Detected broken tags and prepended a fresh block (safety fallback).

---

### **Phase 5: Init Hardening & Auto-Versioning**

**Initial Problem:**
`bd init` only looked for the first available agent file and appended instructions, ignoring other agents (multi-agent setup). Also, `bd --version` was showing a hardcoded version without git hash context. Finally, `bd init` was silent about pre-existing index corruption.

*   **My Assumption/Plan #1:** Update `bd init` to find all candidate files.
    *   **Action Taken:** Refactored `configureAgentRules` in `devlog_cmds.go` to iterate over all `Candidates` (exported from `onboard.go`) and offer to update them all. Changed injection logic to **prepend** the bootstrap trigger.

*   **My Assumption/Plan #2:** Use `ldflags` to inject version info.
    *   **Action Taken:**
        *   Updated `Makefile` to inject `main.Commit`, `main.Branch`, and `main.Build`.
        *   Fixed `cmd/bd/main.go` to use the same rich printing logic as the `bd version` subcommand.
        *   Added `bd version bump [major|minor|patch]` command.

*   **My Assumption/Plan #3:** Add integrity check to `init`.
    *   **Action Taken:** Modified `initializeDevlog` to call `parseIndexMD` on existing `_index.md` files. If it fails, `bd init` now warns the user and suggests running `bd devlog sync` for the fix.

*   **Result:**
    *   `bd init` is now multi-agent aware, enforces top-posting, and checks data integrity.
    *   `bd --version` shows full context: `bd version 0.47.1 (e592d692: dev/beads-devlog-synergy@e592d6926f46)`.

---

### **Phase 6: Automatic Build Counters**

**Initial Problem:**
The version string was static (`dev` or short hash), making it hard to order builds chronologically without checking git history.

*   **My Assumption/Plan #1:** Add a monotonic counter to the version string.
    *   **Action Taken:** Updated `Makefile` to include `git rev-list --count HEAD` in the `BUILD` variable injected via `ldflags`.
    *   **Result:** Version output is now `0.47.1 (dev.<count>.<hash>)`, providing immediate visual ordering of builds.

---

### **Final Session Summary**

**Final Status:**
*   **Quickstart:** Unified `--tasks` and `--devlog` modes.
*   **Onboarding:**
    *   **Embedded:** No external dependencies.
    *   **Top-Posted:** Protocol is always at the top of the file.
    *   **Tag-Managed:** Safe, idempotent updates using `<!-- BD_PROTOCOL_... -->` tags.
*   **Init:** Multi-agent aware, prepends triggers, checks index integrity.
*   **Versioning:** Fully automated build injection, monotonic counters, and `bump` command.
*   **Testing:** Full coverage via sandbox scenarios.

**Key Learnings:**
*   **CLI UX:** Unified entry points reduce confusion.
*   **Agent Prompts:** Mandatory instructions must be at the top.
*   **Tool Portability:** Never rely on local source files (like `_rules/`) for a distributed binary. Always embed assets.
*   **Idempotency:** Tag-based content replacement is far superior to simple append/prepend for managing injected code/text in user files.
*   **Go LDFlags:** When building from root, `-X main.Var` works if the variables are in `package main`.
*   **Parsing Logic:** `parseIndexMD` only returns errors on *malformed* tables, not *missing* tables. Integrity checks need to align with parser strictness.

---

### **Architectural Relationships**
- bd quickstart -> printDevlogQuickstart (calls)
- bd onboard -> injectProtocol (modifies files)
- injectProtocol -> GEMINI.md (prepends/updates content)
- injectProtocol -> protocol.go (reads const)
- bd init -> configureAgentRules (calls)
- bd init -> parseIndexMD (checks integrity)
- configureAgentRules -> injectBootstrapTrigger (modifies files)
- Makefile -> main.Commit (injects build var)
