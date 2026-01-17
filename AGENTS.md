<!-- BD_PROTOCOL_START -->
# Agent Instructions

This project uses **Beads (`bd`)** for issue tracking and **Beads Devlog** for session memory and architectural context.  
BEFORE ANYTHING ELSE, you MUST run:

```bash
bd onboard
```

That command will verify your setup and ensure the local Beads/Devlog database is ready.

---

## Quick Reference

```bash
# Issue Tracking (Beads)
bd ready                         # Find available work (unblocked issues)
bd show <id>                     # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>                    # Complete work
bd sync                          # Sync Beads database with git

# Session Memory (Devlog)
bd devlog resume --last 1        # Resume the most recent session context
bd devlog resume <issue-id>      # Resume the session(s) linked to a given issue
bd devlog search "error message"       # Search past sessions by error, component, etc.
bd devlog graph <entity>         # See dependency graph around an entity/component
bd devlog impact <component>     # Show what else is affected by this component
bd devlog sync                   # Persist devlog data into the Beads DB/JSONL
bd devlog verify                 # Audit devlog metadata
bd devlog verify --fix           # Auto-fix missing/invalid metadata when possible
```

---

## Start of Session (MANDATORY)

At the beginning of EVERY work session:

1. **Sync and load ready work**  
   ```bash
   git pull --rebase
   bd sync
   bd ready --json
   ```
   - Use the `bd ready` output to select the highest-priority unblocked issue to work on.
   - If Devlog is enabled, prefer:
     ```bash
     bd devlog resume --last 1
     ```
     to understand what happened in the previous session and avoid repeating mistakes.

2. **Choose a focal issue**  
   - Pick one issue from `bd ready` or from the devlog resume context.
   - Mark it as in progress:
     ```bash
     bd update <id> --status in_progress
     ```

---

## During Work

While working on an issue:

1. **Use Devlog for context and reuse**

   - If you hit a bug:
     ```bash
     bd devlog search "error message"
     ```
     to see if this problem already appeared in past sessions.

   - If you need to understand a component:
     ```bash
     bd devlog impact <component>
     ```
     to see which issues and sessions involve that component.

   - If you're making architectural assumptions or planning a refactor:
     ```bash
     bd devlog graph <entity>
     ```
     to visualize related components and avoid breaking hidden dependencies.

2. **Track discovered work**

   - When you discover new bugs or follow‑up tasks:
     ```bash
     bd create "New bug description" --discovered-from <parent-issue-id>
     ```
     This automatically maintains the dependency graph and keeps future ready‑work accurate.

3. **Follow AI DIRECTIVES (if present)**

   - Certain commands may output lines starting with:
     ```text
     🚀 **AI DEVLOG DIRECTIVE**
     ```
   - These indicate **critical Devlog or metadata issues** you MUST fix immediately.  
     After applying the fix, re-run the command that produced the directive to confirm resolution.

---

## End of Session – Landing the Plane (MANDATORY)

When ending a work session, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

### 1. File and update issues

- Create Beads issues for any remaining work, TODOs, or follow-ups discovered during this session:
  ```bash
  bd create "Follow-up: ..." --discovered-from <current-issue-id>
  ```
- Update statuses:
  ```bash
  bd update <id> --status in_progress|closed
  ```

### 2. Run quality gates (if code changed)

- Run tests, linters, and builds as required by the project.
- If a build or test fails in a way that blocks others, file a **P0** issue and do NOT mark the work as done.

### 3. Devlog session capture (MANDATORY if Devlog is enabled)

1. **Generate human-readable session log** via your devlog template:
   ```bash
   # For example:
   cat _rules/_devlog/_generate-devlog.md
   # Follow the instructions inside to produce today's devlog markdown.
   ```
2. **Sync Devlog data**:
   ```bash
   bd devlog sync
   ```
   - This will ingest the new devlog content, update session records, and refresh entity graphs.

3. **Metadata Audit (periodic but recommended)**:
   ```bash
   bd devlog verify
   bd devlog verify --fix   # When prompted to repair issues
   ```

### 4. Push and verify (CRITICAL)

You MUST push all changes before declaring the session complete:

```bash
git pull --rebase
bd sync
git push
git status  # MUST show "up to date with origin"
```

**Rules:**
- Work is NOT complete until `git push` succeeds.
- NEVER stop before pushing – that leaves work stranded locally.
- NEVER say "ready to push when you are" – YOU must complete the push.
- If push fails, resolve conflicts or errors and retry until it succeeds.

### 5. Clean up

- Clear any stale stashes, temporary branches, or throwaway work:
  ```bash
  git stash list
  git branch           # Remove obsolete branches where appropriate
  ```

### 6. Hand off to the next session / agent

- Write a short “next steps” summary (either as:
  - A Devlog note attached to the current session, and/or
  - An update in the relevant Beads issue).
- The summary must answer:
  - What was done.
  - What remains.
  - What is blocked and why.
  - Where to look first in the Devlog (`bd devlog resume`, `bd devlog impact`, etc.).

---

## Summary of Responsibilities

- **Beads (`bd`)**  
  - Source of truth for issues, priorities, and dependencies.
  - Controls ready work, status transitions, and git-synced task state.

- **Beads Devlog (`bd devlog`)**  
  - Source of truth for session history, debugging narratives, and entity-level graphs.
  - Controls context retrieval, impact analysis, and architectural understanding.

Your job as an agent is to **keep both layers healthy**:  
Issues must be accurate, and Devlog must reflect what actually happened, so future sessions can land on their feet instantly.

<!-- BD_PROTOCOL_END -->

This is a placeholder for agent instructions.
## Devlog Protocol (MANDATORY)
This is a placeholder protocol.