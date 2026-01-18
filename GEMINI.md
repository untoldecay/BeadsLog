BEFORE ANYTHING ELSE: run 'bd onboard'

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


# Beads (`bd`) - Gemini Context

## Project Overview
**Beads** is a distributed, git-backed graph issue tracker designed specifically for AI agents. It provides a structured, persistent memory for agents, replacing messy markdown plans with a dependency-aware graph.

*   **Core Philosophy:** Issues are stored as JSONL in git (`.beads/issues.jsonl`), enabling versioning, branching, and merging like code.
*   **Performance:** Uses a local SQLite cache for millisecond-latency queries.
*   **Conflict-Free:** Uses hash-based IDs (e.g., `bd-a1b2`) to prevent collisions in distributed/multi-agent environments.
*   **Architecture:** Three-layer model: CLI -> SQLite (Local Cache) -> JSONL (Git-Synced Source of Truth).

## Architecture & Key Concepts
*   **Daemon:** A background process (`internal/daemon/`) that handles auto-syncing between SQLite and JSONL, and serves RPC requests.
*   **Dual Mode:** The CLI attempts to connect to the daemon (RPC) but falls back to direct SQLite access if the daemon is unavailable.
*   **Molecules & Wisps:**
    *   **Molecules:** Template work items defining structured workflows.
    *   **Wisps:** Ephemeral, local-only child tasks. They are *never* synced to JSONL until "squashed" into a permanent digest.
*   **Data Model:**
    *   **Issues:** The core unit of work.
    *   **Dependencies:** Typed links (blocks, parent-child, related) forming a graph.

## Development & Conventions

### Build & Run
*   **Build:** `make build` or `go build -o bd ./cmd/bd`
*   **Install:** `make install`
*   **Run:** `./bd <command>` (e.g., `./bd init`, `./bd create "Task"`)

### Testing
*   **Framework:** Standard `go test`.
*   **Fast Tests (Unit):** `go test -short ./...` (Default for dev).
*   **Full Tests (Integration):** `go test ./...` (Run before commit).
*   **Dual-Mode Testing:** **CRITICAL.** Commands must be tested in both Direct and Daemon modes. Use the `RunDualModeTest` helper in `cmd/bd/dual_mode_test.go`.
    ```go
    RunDualModeTest(t, "test_name", func(t *testing.T, env *DualModeTestEnv) {
        // Test logic using env.CreateIssue(), env.GetIssue(), etc.
    })
    ```
*   **Benchmarks:** `make bench` (generates CPU profiles).

### Code Quality
*   **Linting:** `golangci-lint run ./...`. Note: The project has a baseline of ~100 known warnings; focus on avoiding *new* issues.
*   **Formatting:** Standard `gofmt`.

### Devlog System (Memory & Context)
The Devlog system allows you to search past sessions, analyze architectural impact, and resume context. Use these commands to be a smarter agent.

| Goal | Command | Description |
| :--- | :--- | :--- |
| **Resume Context** | `bd devlog resume --last 1` | Get the full content of the last session to catch up. |
| **Find Solution** | `bd devlog search "error message"` | Search past logs for similar errors or topics. |
| **Check Impact** | `bd devlog impact "component-name"` | See what depends on a component before modifying it. |
| **Visualize** | `bd devlog graph "entity"` | See the dependency tree of an entity. |
| **List History** | `bd devlog list --type feature` | See a timeline of specific work types. |
| **Sync** | `bd devlog sync` | Manually ingest new logs (usually handled by git hooks). |

### Critical Rules
1.  **NEVER modify `.beads/issues.jsonl` manually.** This file is the database source of truth. Always use CLI commands or internal helpers to modify issues. CI will fail if this file is manually edited in a PR.
2.  **Respect the Architecture:** Changes to core logic usually involve `internal/storage` (SQLite) and `internal/types`. CLI commands in `cmd/bd` should remain thin wrappers around internal logic or RPC calls.

## Directory Structure
*   `cmd/bd/`: CLI entry point and command implementations (Cobra).
*   `internal/`: Core application logic.
    *   `storage/`: Database interfaces and SQLite implementation.
    *   `types/`: Core data structures (Issue, Dependency, etc.).
    *   `daemon/`: Background process logic.
    *   `importer/` & `export/`: Logic for syncing between SQLite and JSONL.
    *   `molecules/`: Logic for Wisps and workflows.
*   `.beads/`: Local storage directory (contains `beads.db`, `issues.jsonl`, config).

## Project Goals (Mission & PRD)

### Mission
**Build a shared devlog session history system.**

The Devlog Beads project forks the Beads issue tracker into a graph-powered development session memory system that imports existing markdown devlogs (index.md + dated session files), automatically extracts software entities from problem descriptions, builds architectural dependency graphs between components/modal/hooks/endpoints, and enables hybrid text+graph retrieval to resume debugging with complete session history and related-context discovery across multi-session fixes/features/enhancements.

### Complete Devlog Implementation PRD

**Summary**
The Devlog Beads project forks the Beads issue tracker into a graph-powered development session memory system that imports existing markdown devlogs (index.md + dated session files), automatically extracts software entities from problem descriptions, builds architectural dependency graphs between components/modal/hooks/endpoints, and enables hybrid text+graph retrieval to resume debugging with complete session history and related-context discovery across multi-session fixes/features/enhancements.

#### Phase 1: Fork & Schema (30min)

1.  **Fork Beads:** Clone `steveyegge/beads` to `devlog`.
2.  **Complete Schema (`internal/db/schema.go`):**
    *   Add `sessions` table (id, title, timestamp, status, type, filename, narrative).
    *   Add `entities` table (id, name, type, first_seen, mention_count).
    *   Add `session_entities` table (session_id, entity_id, relevance).
    *   Add `entity_deps` table (from_entity, to_entity, relationship, discovered_in).

#### Phase 2: Import Script (30min)
*   Create `cmd/devlog/import-md.go` to import `index.md` into `sessions` and extract entities.

#### Phase 3: Graph Queries (20min)
*   Implement `GetEntityGraph` in `internal/queries/graph.go` using recursive CTEs.
*   Implement `devlog graph` command.

#### Phase 4: CLI Commands (20min)
*   `devlog list --type fix`
*   `devlog show <date>`
*   `devlog search <query>` (Hybrid text+graph)
*   `devlog entities` (Top entities)
*   `devlog impact <entity>`
*   `devlog resume <query>`

#### Phase 5: Test & Deploy (10min)
*   Init project, import data, test graph, and commit.

#### Success Criteria
*   Index.md imported to sessions DB.
*   Graph commands show relationships.
*   Full narratives preserved.
*   Git-tracked JSONL.