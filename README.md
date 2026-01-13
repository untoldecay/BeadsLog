# bd - BeadsLog

**Distributed, git-backed graph issue tracker & session memory for AI agents.**

[![License](https://img.shields.io/github/license/untoldecay/BeadsLog)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/untoldecay/BeadsLog)](https://goreportcard.com/report/github.com/untoldecay/BeadsLog)

BeadsLog forks Beads to provide a persistent, structured memory for coding agents. It merges a dependency-aware task graph with a semantic **Devlog** system, allowing agents to "remember" past sessions, understand architecture, and solve recurring problems instantly.

---

## ☯️ Philosophy: The Loop

BeadsLog is designed around a cycle of **Planning** (Forward) and **Reflection** (Backward).

| System | Direction | Purpose | Question Answered |
| :--- | :--- | :--- | :--- |
| **Tasks** | ➡️ **Forward** | Planning & Execution | *What do we need to do next?* |
| **Devlog** | ⬅️ **Backward** | Context & Memory | *How and why did we do that?* |

**The Workflow Loop:**
1.  **Resume:** Load history (`bd devlog resume --last 1`).
2.  **Context:** Check dependencies/impact (`bd devlog impact`).
3.  **Execute:** Code the task.
4.  **Log:** Document the session (`_rules/_devlog/_generate-devlog.md`).
5.  **Sync:** Ingest changes (`git commit` triggers auto-sync).

---

## ⚡ Quick Start

No external databases required. It's a single binary.

```bash
# 1. Install (requires Go)
go install github.com/untoldecay/BeadsLog/cmd/bd@latest

# 2. Initialize (Setup Tasks & Devlog assets)
bd init

# 3. AI Onboarding (Run once per repo by the first agent)
bd devlog onboard
```

---

## 🧠 Power Scenarios

*   **"Resume Context"**: Start a new chat session with full context of what you did last time.  
    `bd devlog resume --last 1`
*   **"I've seen this error before..."**: Find the exact session where you fixed a specific bug.  
    `bd devlog search "nginx 400 error"`
*   **"Refactoring Anxiety"**: See every component that historically depends on an entity before you change it.  
    `bd devlog impact "auth-hook"`
*   **"Architectural Graph"**: Visualize the parent/child/related tree of a component.  
    `bd devlog graph "AuthService"`

---

## 📚 Command Reference

| Command | Usage | Description |
| :--- | :--- | :--- |
| **Help** | `bd devlog --help` | Display all devlog commands and usage details. |
| **Onboard** | `bd devlog onboard` | Enrolls AI agent into the **MANDATORY Devlog Protocol**. |
| **Search** | `bd devlog search "query"` | Hybrid search across sessions, narratives, and entities. |
| **Resume** | `bd devlog resume [--last N]` | Finds relevant context or shows the last N sessions. |
| **Impact** | `bd devlog impact "entity"` | Shows components that depend on or relate to a specific entity. |
| **Graph** | `bd devlog graph "entity"` | Visualizes the architectural dependency tree. |
| **Verify** | `bd devlog verify [--fix]` | Audits for missing metadata and generates recovery directives. |
| **List** | `bd devlog list [--type]` | Lists chronological sessions (fix, feature, chore, etc.). |
| **Show** | `bd devlog show <date>` | Displays the full content of a specific session log. |
| **Sync** | `bd devlog sync` | Manually triggers ingestion of new or updated devlogs. |
| **Hooks** | `bd devlog install-hooks` | Installs git hooks for automatic background synchronization. |
| **Reset** | `bd devlog reset` | **Truncates** all devlog tables (sessions, entities) for a fresh start. |
| **Status** | `bd devlog status` | Checks configuration, stats, and git hook health. |

---

## 🛠 Core Features (Standard Beads)

*   **Git as Database:** Issues stored as JSONL in `.beads/`. Versioned with your code.
*   **Zero Conflict:** Hash-based IDs allow concurrent agent work without merge conflicts.
*   **Invisible Infrastructure:** SQLite local cache for fast queries; auto-sync via git hooks.
*   **Self-Healing Index:** Strict linting and AI instructions protect `_index.md` from corruption.

## 📝 Documentation

*   [Installing](docs/INSTALLING.md)
*   [Agent Workflow](AGENT_INSTRUCTIONS.md)
*   [Sync Branch Mode](docs/PROTECTED_BRANCHES.md)
*   [Troubleshooting](docs/TROUBLESHOOTING.md)