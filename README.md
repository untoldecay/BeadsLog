# bd - BeadsLog

**Distributed, git-backed graph issue tracker & session memory for AI agents.**

[![License](https://img.shields.io/github/license/untoldecay/BeadsLog)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/untoldecay/BeadsLog)](https://goreportcard.com/report/github.com/untoldecay/BeadsLog)

BeadsLog forks Beads to provide a persistent, structured memory for coding agents. It merges a dependency-aware task graph with a semantic **Devlog** system, allowing agents to "remember" past sessions, understand architecture, and solve recurring problems instantly.

## ⚡ Quick Start

No external databases or complex dependencies required. It's a single binary.

```bash
# One-liner install (requires Go)
go install github.com/untoldecay/BeadsLog/cmd/bd@latest

# Initialize in your repo (automatically sets up Tasks & Devlogs)
bd init
```

## 🧠 The Devlog System
...
### 🔄 Usage Workflow

1.  **Resume:** Agent starts by running `bd devlog resume --last 1`.
2.  **Work:** Code the task as usual.
3.  **Log:** Generate a log entry (using the prompt in `_rules/_devlog/_generate-devlog.md`).
4.  **Sync:** `git commit` or manual `bd devlog sync`.

---

## 📚 Command Reference

| Command | Usage | Description |
| :--- | :--- | :--- |
| **Onboard** | `bd devlog onboard` | Enrolls AI agent into the **MANDATORY Devlog Protocol**. |
| **Search** | `bd devlog search "query"` | Hybrid search across session titles, narratives, and entities. |
| **Resume** | `bd devlog resume [--last N]` | Finds relevant context or shows the last N sessions. |
| **Impact** | `bd devlog impact "entity"` | Shows what other components depend on or relate to a specific entity. |
| **Graph** | `bd devlog graph "entity"` | Visualizes the dependency tree (parent/child/related) of an entity. |
| **Verify** | `bd devlog verify [--fix]` | Audits sessions for missing metadata and generates recovery directives. |
| **List** | `bd devlog list [--type]` | Lists chronological sessions. Filter by `fix`, `feature`, `chore`, etc. |
| **Show** | `bd devlog show <date>` | Displays the full content of a specific session log. |
| **Sync** | `bd devlog sync` | Manually triggers ingestion of new or updated devlogs. |
| **Reset** | `bd devlog reset` | **Truncates** all devlog tables (sessions, entities) for a fresh start. |
| **Status** | `bd devlog status` | Checks configuration, stats, and git hook health. |

---

## ☯️ Philosophy: Tasks vs. Devlogs

BeadsLog is designed around a cycle of **Planning** and **Reflection**.

| Feature | Direction | Purpose | Question Answered |
| :--- | :--- | :--- | :--- |
| **Beads (Core)** | ➡️ **Forward** | Planning & Execution | *What do we need to do next?* |
| **Devlog** | ⬅️ **Backward** | Context & Memory | *How and why did we do that?* |

**The Workflow Loop:**
1.  **Plan:** Create a Task (`bd create`).
2.  **Context:** Search the Devlog for similar past work (`bd devlog resume`).
3.  **Execute:** Do the work.
4.  **Reflect:** Write a Devlog entry.
5.  **Close:** Complete the Task (`bd close`).

## 🛠 Core Features (Standard Beads)

*   **Git as Database:** Issues stored as JSONL in `.beads/`. Versioned with your code.
*   **Zero Conflict:** Hash-based IDs (`bd-a1b2`) allow concurrent agent work without merge conflicts.
*   **Invisible Infrastructure:** SQLite local cache for millisecond queries; background daemon for auto-sync.
*   **Compaction:** "Memory decay" summarizes old closed tasks to save token context.

## 📝 Documentation

*   [Installing](docs/INSTALLING.md)
*   [Agent Workflow](AGENT_INSTRUCTIONS.md)
*   [Sync Branch Mode](docs/PROTECTED_BRANCHES.md)
*   [Troubleshooting](docs/TROUBLESHOOTING.md)
