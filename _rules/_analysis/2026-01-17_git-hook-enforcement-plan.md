# Analysis: Git Hook Strategy for Enforcing Devlog

**Date:** 2026-01-17
**Status:** Plan Definition
**Objective:** Define the architecture for a `pre-commit` hook that strictly enforces the Devlog Protocol (data generation) without relying on unreliable automation or loops.

---

## 1. Problem Statement

We have a "Silent Failure" risk:
1.  **User/Agent** does work.
2.  **User/Agent** commits.
3.  **Result:** The code is committed, but the context (Devlog) is lost. The `pre-commit` hook currently only flushes database state; it doesn't ensure *new* narrative exists.

We need a gatekeeper that says: **"No Devlog? No Commit."**

---

## 2. The Logic: Enforcement, Not Generation

**Crucial Design Decision:**
The hook must **NOT** attempt to generate the log. It lacks the context window and reasoning of the LLM. Its sole job is **Compliance Verification**.

### 2.1 The "Loophole" Defense
A naive check ("is `_index.md` staged?") is insufficient. An agent could just stage an empty or unchanged file to bypass the check.

**Robust Verification Logic:**
1.  **Scope Check:** Does this commit involve "work"?
    *   *If* only `.gitignore` or `README.md` changed -> Maybe skip?
    *   *Better:* Assume ALL commits need a log unless strictly whitelisted (e.g., automated version bumps).
2.  **Configuration Check:** Is enforcement enabled?
    *   `devlog.enforce_on_commit` in `.beads/config.yaml`.
3.  **Content Verification:**
    *   Is `_rules/_devlog/_index.md` staged?
    *   Does it have **actual diffs** compared to HEAD? (`git diff --cached --exit-code`)
    *   Are there **new** or **modified** session logs (`YYYY-MM-DD_*.md`) in the devlog directory?

---

## 3. Architecture

### 3.1 New Configuration (`.beads/config.yaml`)
```yaml
devlog:
  enabled: true
  enforce_on_commit: true   # <--- The Switch
  dir: "_rules/_devlog"
```

### 3.2 New CLI Commands

**`bd config`**
Manages the configuration without re-running init.
*   `bd config set <key> <value>`
*   `bd config get <key>`

**`bd check` (Internal / CI Tool)**
The brain of the hook.
*   `bd check --hook pre-commit`
*   **Returns 0:** Success (Log present or enforcement disabled).
*   **Returns 1:** Failure (Block commit).
*   **Output:** A structured, helpful error message guiding the agent/user.

### 3.3 The Hook Script (`.beads-hooks/pre-commit`)
```bash
# ... existing flush logic ...

# Compliance Check
if ! bd check --hook pre-commit; then
    echo "❌ Commit rejected by Beads Devlog Policy."
    exit 1
fi
```

---

## 4. User Experience (The "Guide")

When the hook blocks a commit, it must provide the **exact steps** to resolve it.

**Error Message:**
```text
❌ BLOCKER: Devlog Update Missing

This project requires a devlog update for every commit.

👉 AGENT ACTION REQUIRED:
1. Run the devlog generation prompt:
   cat _rules/_devlog/_generate-devlog.md
2. Generate the session log (YYYY-MM-DD_*.md).
3. Update the index (_rules/_devlog/_index.md).
4. Stage these files: git add _rules/_devlog/
5. Retry the commit.
```

---

## 5. Implementation Plan

1.  **Config Layer:** Implement `bd config` and update `config.yaml` schema.
2.  **Logic Layer:** Implement `bd check` command with git-diff logic.
3.  **Hook Layer:** Update `pre-commit` template and existing hook.
4.  **Init Layer:** Update `bd init` to ask the "Enforce?" question.

---
