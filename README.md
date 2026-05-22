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

## 🛠 Commands for Humans

Your agents handle the heavy lifting. As a human, you only need a few commands to steer the project:

- **Stay aligned:** `bd catchup` (See what your team/agents completed since yesterday)
- **Check tool updates:** `bd upgrade check`
- **Verify health:** `bd devlog status`

## 🔄 The Agent Workflow

BeadsLog is engineered for agents. Context is retrieved dynamically, not guessed:

1. **Probing:** Query the knowledge graph (`graph`, `impact`, `entities`) to pinpoint dependencies instantly.
2. **Mapping:** The system tracks lifecycle states (`active`, `paused`, `abandoned`) and Git reachability so history matches the actual codebase.
3. **Iterating:** Acquire context -> implement -> iterate -> record.

## 🎯 Use Cases

- **Resume Context:** Pick up exactly where you left off.  
  `bd devlog resume --last 1`
- **Audit Architecture:** See the impact of a component before changing it.  
  `bd devlog impact "AuthService"`
- **Team Sync:** See what was modularized or abandoned since yesterday.  
  `bd catchup`
- **Kill Experiments:** Explicitly mark a flawed path so no one retries it.  
  `bd devlog abandon --scope branch:feature-x --message "Memory leak"`
- **Trace Merges:** Prove that code actually landed in the baseline.  
  `bd devlog search "modularization" --status active`

## 📚 Docs

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
