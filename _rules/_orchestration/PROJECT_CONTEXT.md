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


## Legacy Content from AGENTS.md (at init)
This is a placeholder for agent instructions.
## Devlog Protocol (MANDATORY)
This is a placeholder protocol.

## Legacy Content from GEMINI.md (at init)
BEFORE ANYTHING ELSE: run 'bd onboard'

## Legacy Content from .github/copilot-instructions.md (at init)
# GitHub Copilot Instructions for Beads

## Project Overview

**beads** (command: `bd`) is a Git-backed issue tracker designed for AI-supervised coding workflows. We dogfood our own tool for all task tracking.

**Key Features:**
- Dependency-aware issue tracking
- Auto-sync with Git via JSONL
- AI-optimized CLI with JSON output
- Built-in daemon for background operations
- MCP server integration for Claude and other AI assistants

## Tech Stack

- **Language**: Go 1.21+
- **Storage**: SQLite (internal/storage/sqlite/)
- **CLI Framework**: Cobra
- **Testing**: Go standard testing + table-driven tests
- **CI/CD**: GitHub Actions
- **MCP Server**: Python (integrations/beads-mcp/)

## Coding Guidelines

### Testing
- Always write tests for new features
- Use `BEADS_DB=/tmp/test.db` to avoid polluting production database
- Run `go test -short ./...` before committing
- Never create test issues in production DB (use temporary DB)

### Code Style
- Run `golangci-lint run ./...` before committing
- Follow existing patterns in `cmd/bd/` for new commands
- Add `--json` flag to all commands for programmatic use
- Update docs when changing behavior

### Git Workflow
- Always commit `.beads/issues.jsonl` with code changes
- Run `bd sync` at end of work sessions
- Install git hooks: `bd hooks install` (ensures DB ↔ JSONL consistency)

## Issue Tracking with bd

**CRITICAL**: This project uses **bd** for ALL task tracking. Do NOT create markdown TODO lists.

### Essential Commands

```bash
# Find work
bd ready --json                    # Unblocked issues
bd stale --days 30 --json          # Forgotten issues

# Create and manage (ALWAYS include --description)
bd create "Title" --description="Detailed context" -t bug|feature|task -p 0-4 --json
bd update <id> --status in_progress --json
bd close <id> --reason "Done" --json

# Search
bd list --status open --priority 1 --json
bd show <id> --json

# Sync (CRITICAL at end of session!)
bd sync  # Force immediate export/commit/push
```

### Workflow

1. **Check ready work**: `bd ready --json`
2. **Claim task**: `bd update <id> --status in_progress`
3. **Work on it**: Implement, test, document
4. **Discover new work?** `bd create "Found bug" --description="What was found and why" -p 1 --deps discovered-from:<parent-id> --json`
5. **Complete**: `bd close <id> --reason "Done" --json`
6. **Sync**: `bd sync` (flushes changes to git immediately)

**IMPORTANT**: Always include `--description` when creating issues. Issues without descriptions lack context for future work.

### Priorities

- `0` - Critical (security, data loss, broken builds)
- `1` - High (major features, important bugs)
- `2` - Medium (default, nice-to-have)
- `3` - Low (polish, optimization)
- `4` - Backlog (future ideas)

## Project Structure

```
beads/
├── cmd/bd/              # CLI commands (add new commands here)
├── internal/
│   ├── types/           # Core data types
│   └── storage/         # Storage layer
│       └── sqlite/      # SQLite implementation
├── integrations/
│   └── beads-mcp/       # MCP server (Python)
├── examples/            # Integration examples
├── docs/                # Documentation
└── .beads/
    ├── beads.db         # SQLite database (DO NOT COMMIT)
    └── issues.jsonl     # Git-synced issue storage
```

## Available Resources

### MCP Server (Recommended)
Use the beads MCP server for native function calls instead of shell commands:
- Install: `pip install beads-mcp`
- Functions: `mcp__beads__ready()`, `mcp__beads__create()`, etc.
- See `integrations/beads-mcp/README.md`

### Scripts
- `./scripts/bump-version.sh <version> --commit` - Update all version files atomically
- `./scripts/release.sh <version>` - Complete release workflow
- `./scripts/update-homebrew.sh <version>` - Update Homebrew formula

### Key Documentation
- **AGENTS.md** - Comprehensive AI agent guide (detailed workflows, advanced features)
- **AGENT_INSTRUCTIONS.md** - Development procedures, testing, releases
- **README.md** - User-facing documentation
- **docs/CLI_REFERENCE.md** - Complete command reference

## Important Rules

- ✅ Use bd for ALL task tracking
- ✅ Always use `--json` flag for programmatic use
- ✅ Run `bd sync` at end of sessions
- ✅ Test with `BEADS_DB=/tmp/test.db`
- ❌ Do NOT create markdown TODO lists
- ❌ Do NOT create test issues in production DB
- ❌ Do NOT commit `.beads/beads.db` (JSONL only)

---

**For detailed workflows and advanced features, see [AGENTS.md](../AGENTS.md)**


## Devlog Protocol (MANDATORY)

This repo uses **Beads Devlog** to persist context and prevent regressions. You must follow this protocol:

### 1. 🟢 Start of Session
- **Resume Context:** Run `bd devlog resume --last 1` to load the immediate history.
- **Goal:** Understand what happened last time to avoid repeating mistakes.

### 2. 🟡 During Work
- **Bug Encounter:** If you hit a bug, run `bd devlog search "error message"` to see if it's a known issue.
- **Information Request:** If asking about a component, run `bd devlog impact <component>` to understand dependencies.
- **Planning:** If making a plan, check `bd devlog graph <entity>` to verify architectural assumptions.

### 3. 🔴 End of Session
- **Log & Sync:**
  1. Generate the session log using `_rules/_prompts/generate-devlog.md`.
  2. Run `bd devlog sync` to persist it to the database.
