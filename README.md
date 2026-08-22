# BeadsLog

**Git-backed project memory that scales from solo developers to distributed AI agent teams.**

BeadsLog extends [Beads](https://github.com/steveyegge/beads) with automated session mapping, background enrichment, and lifecycle tracking. It turns natural language devlogs into a stateful, verifiable knowledge graph so agents never build on dead code or repeat historical mistakes.

## 🚀 Quick Start

Requires [Go](https://go.dev/doc/install) (1.24+).

```bash
# 1. Install
go install github.com/untoldecay/BeadsLog/cmd/bd@latest

# 2. Initialize in repo
bd init

# 3. Connect AI agent
bd onboard
```

> [!NOTE]
> **Folder naming:** entity extraction filters common words and trunk branch names (`master`, `main`, `test`, `core`, ...) to keep the knowledge graph clean. If your project folder is named one of these (e.g. a worktree checked out as `master/`), BeadsLog automatically tracks it as `<name>-repository`. Manually written relationship arrows targeting a filtered word are dropped **with a warning** — rename the entity in your arrow (e.g. `master-repository`) to keep the edge.

## 🛠 Useful commands

Your agents handle the heavy lifting. As a human, you only need to know one command: [Check all commands here](docs/COMMANDS.md).

- **Stay aligned:** `bd catchup` (See what your team/agents completed since last checked)

## 🔄 The Agent Workflow

BeadsLog provides agents with dynamic context and architectural awareness, eliminating the need to guess or brute-force the repository.

### Semantic Search & Exploration
Agents use specialized tools to map the project's evolution and dependencies:
- **`search`**: FTS5-powered retrieval of historical decisions and narratives.
- **`impact`**: Identifies all components semantically dependent on a target entity.
- **`entities`**: Discovers the most significant architectural components by relationship density.
- **`path`**: Traces the chain of events and sessions linking two disparate entities.

### Stateful Iteration
Acquire context → Implement → Iterate → Record.

## 🎯 Use Cases

| Goal | Command |
|---|---|
| **Resume Context** | `bd devlog resume --last 1` |
| **Audit Architecture** | `bd devlog impact "AuthService"` |
| **Team Sync** | `bd catchup` |
| **Kill Experiments** | `bd devlog abandon --scope branch:feature-x --message "..."` |
| **Trace Merges** | `bd devlog search "modularization" --status active` |

## 📚 Docs

- [Running Modes](docs/RUNNING_MODES.md) — Team, solo, and switching: which mode to run and why.
- [Use Cases](docs/USE_CASES.md) — Real-world scenarios for agents and teams.
- [Project Catchup](docs/CATCHUP.md) — Staying aligned with team progress.
- [Lifecycle States](docs/LIFECYCLE.md) — Managing active, paused, and abandoned work.
- [Devlog](docs/DEVLOG.md) — Understanding the "Bead" concept and the narrative format.
- [Devlog Architecture](docs/DEVLOG_ARCHITECTURE.md) — How the background AI and crystallization engine works.
- [Ollama Enrichment](docs/OLLAMA_ENRICHMENT.md) — Configuring background semantic analysis.
- [Issue Architecture](docs/ARCHITECTURE.md) — Standard Beads data model and sync mechanism.
- [Hooks](docs/HOOKS.md) — Automating your workflow with Git integration.
- [Installation](docs/INSTALLING.md) — Setup guide and platform requirements.

***
