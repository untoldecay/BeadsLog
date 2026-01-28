# BeadsLog

**Git-backed devlog sessions that capture why you built things, for AI agents and teams.**

BeadsLog extends [Beads](https://github.com/steveyegge/beads) with automated session mapping and background AI enrichment. It turns your natural devlog narratives into a permanent, version-controlled knowledge graph.

```bash
# Install
go install github.com/untoldecay/BeadsLog/cmd/bd@latest

# Initialize in repo
bd init

# Connect AI agent
bd onboard
```

## 🔄 The BeadsLog Workflow

BeadsLog is engineered for the era of AI coding agents where context is the most valuable resource. The workflow follows a high-efficiency iterative loop:

1. **Probing:** Efficient building begins with architectural awareness. Rather than forcing agents to "brute-force" the entire codebase using slow, token-intensive searches, BeadsLog provides an instant structural probe. By querying the local knowledge graph via `graph`, `impact`, and `entities`, you can immediately pinpoint the exact components and dependencies relevant to your task.
2. **Mapping:** As work progresses, the system automatically records the evolving architecture, intent, and decision-making process. This transforms the session into a structured map of technical relationships and historical context.
3. **Iterating:** This creates a continuous, high-velocity development loop: *acquire context -> implement -> iterate -> record context*. Every "Bead" added to the chain ensures the project's collective memory grows stronger and more precise.

By offloading slow AI tasks to a background process, your CLI interactions remain instant while your architectural foresight grows automatically.

## 🌟 Real Use Cases

**🔍 New agent debugging a complex flow:**
```bash
bd devlog search "auth session timeout"
# Returns: "Switched from Redis to Memcached (memory spike fix),
#          impacts AuthService and SessionManager"
```

**🤝 Architectural impact before refactoring:**
```bash
bd devlog impact "PostgresConnector"
# Returns: "Used by UserManager, BillingService, and AnalyticsJob."
```

**⚡ Resuming context after a 3-day break:**
```bash
bd devlog resume --last 1
# Returns: "Last session: Fixed redirect loop, left off at validation tests."
```

## 📚 Docs

- [Use Cases](docs/USE_CASES.md) — Real-world scenarios for agents and teams.
- [Devlog](docs/DEVLOG.md) — Understanding the "Bead" concept and the narrative format.
- [Devlog Architecture](docs/DEVLOG_ARCHITECTURE.md) — How the background AI and crystallization engine works.
- [Issue Architecture](docs/ARCHITECTURE.md) — Standard Beads data model and sync mechanism.
- [Hooks](docs/HOOKS.md) — Automating your workflow with Git integration.
- [Visualization](docs/VISUALIZATION.md) — Exploring Search, Impact, and the Graph.
- [Commands](docs/COMMANDS.md) — Full categorized CLI reference.
- [Installation](docs/INSTALLING.md) — Setup guide and platform requirements.

***