# Analysis: Refactoring Agent Onboarding (Beads + Devlog)

**Date:** 2026-01-15
**Status:** Plan Validation
**Objective:** Validate the strategy to unify `bd onboard` and `bd devlog onboard` into a single, authoritative command as per the "Beads Devlog Synergie PRD".

---

## 1. Current State Assessment

We currently have a split-brain situation regarding agent instructions:

| Feature | `bd onboard` | `bd devlog onboard` |
| :--- | :--- | :--- |
| **Source File** | `cmd/bd/onboard.go` | `cmd/bd/devlog_cmds.go` |
| **Action** | **Passive** (Prints text to stdout for copy-paste) | **Active** (Modifies `AGENTS.md` directly) |
| **Content** | Points to `bd prime` for workflow | Injects "Devlog Protocol" (Resume/Graph/Sync) |
| **Trigger** | None (User manual execution) | Bootstrapped by `bd init` via `configureAgentRules` |
| **Result** | Agents know about Issues but not Devlog | Agents know about Devlog but might miss Issues context |

**The Conflict:**
`bd init` injects a trigger to run `bd devlog onboard`. If an agent runs this, they get the Devlog protocol. If a user manually runs `bd onboard` (as suggested by older docs), they get the `bd prime` pointer. The two systems are competing for the "instruction space" in `AGENTS.md`.

---

## 2. Target Architecture (per PRD)

The goal is a **Single Source of Truth** where `bd` is the owner.

*   **Command:** `bd onboard` (Unified)
*   **Behavior:** Active (Modifies files, strictly idempotent).
*   **Content:** A merged "Beads + Devlog" protocol (PRD Section 4).
    *   Replaces the dependency on `bd prime` in `AGENTS.md` with explicit "Start/During/End" steps.
    *   Integrates `bd ready` (Work) and `bd devlog resume` (Context) into a single flow.

---

## 3. Refactoring Plan

### Step 3.1: Promote `bd onboard`
*   **Refactor `cmd/bd/onboard.go`:**
    *   Change from `Run: printInstructions` to `Run: injectProtocol`.
    *   Port the file detection and injection logic from `devlog_cmds.go`.
    *   **Crucial Change:** Update the injection string to the full "Merged Protocol" defined in the PRD.

### Step 3.2: Retire `bd devlog onboard`
*   **Update `cmd/bd/devlog_cmds.go`:**
    *   Remove `devlogOnboardCmd` (or make it a hidden alias to `bd onboard` for backward compatibility).
    *   Update `configureAgentRules` (used by `init`) to inject the **new bootstrap trigger**:
        *   *Old:* `BEFORE ANYTHING ELSE: run 'bd devlog onboard'`
        *   *New:* `BEFORE ANYTHING ELSE: run 'bd onboard'`

### Step 3.3: Migration Logic
The new `bd onboard` must be smart enough to handle existing files:
1.  **Clean Slate:** If `AGENTS.md` is empty -> Write full protocol.
2.  **Bootstrap:** If "BEFORE ANYTHING ELSE..." is present -> Replace it with full protocol.
3.  **Old Devlog Protocol:** If "## Devlog Protocol" is present -> Detect and **Upgrade** to the merged protocol (since the old one is missing the "Issue Tracking" parts).
4.  **Old `bd prime` Pointer:** If "Run `bd prime`" is present -> Replace with full protocol.

---

## 4. Gap Analysis & Risks

### 4.1 `bd prime` Role Shift
*   **Current:** `AGENTS.md` relies on `bd prime` to dynamically generate instructions.
*   **Proposed:** `AGENTS.md` contains static instructions.
*   **Implication:** `bd prime` remains useful for context hooks (e.g., `bash-agent` startup), but `AGENTS.md` becomes self-contained. This reduces token usage (no need to run `bd prime` to know *how* to work) but means `AGENTS.md` won't automatically reflect config changes (like "stealth mode") unless re-onboarded.
*   **Mitigation:** The static protocol is robust enough to cover 90% of cases. Special configs can still be handled by agents reading `bd config` or `bd prime` if they really need to.

### 4.2 File Targeting
*   `devlog_cmds.go` currently targets 8 different files (`.windsufrules`, `.cursorrules`, etc.).
*   `onboard.go` currently only mentions `AGENTS.md` and `.github/copilot-instructions.md`.
*   **Action:** The new `bd onboard` will adopt the comprehensive list from `devlog_cmds.go`.

---

## 5. Conclusion

The plan is solid. It simplifies the user experience (`bd onboard` = "Fix my agent setup") and ensures agents have a holistic view of the system (Work + Memory) from the start.

**Next Steps:**
1.  User confirmation of this analysis.
2.  Execution of the code changes (Refactor `onboard.go`, cleanup `devlog_cmds.go`).
