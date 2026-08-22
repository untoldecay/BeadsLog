# BeadsLog

**Git-backed project memory for AI coding agents.**

BeadsLog records what your agents do as they work — decisions, dead ends, and which components depend on which — and turns it into a searchable graph stored in git. Agents query it before they code, so they don't rebuild deleted code or repeat past mistakes.

Built on [Beads](https://github.com/steveyegge/beads).

## Quick Start

Requires [Go](https://go.dev/doc/install) 1.24+.

```bash
go install github.com/untoldecay/BeadsLog/cmd/bd@latest
bd init        # solo? private? see Running Modes
bd onboard     # connect your AI agent
```

> [!TIP]
> `bd init` keeps beads off your work branch by default (a dedicated `beads-metadata` branch when a remote exists). Working privately, or on a team that doesn't use beads? See [Running Modes](docs/RUNNING_MODES.md).

## What agents do with it

Agents query the graph instead of guessing from `grep` and `ls`. One command per job:

| Job | Command |
|---|---|
| Load where I left off | `bd devlog resume --last 1` |
| Find a past decision | `bd devlog search "stripe timeout"` |
| See what breaks if I change X | `bd devlog impact "AuthService"` |
| See the key components | `bd devlog entities` |
| Trace how A connects to B | `bd devlog path "A" "B"` |
| See what the team shipped | `bd catchup` |

Trust badges keep history honest: `[🟢 VALIDATED]` code landed in main; `[⏸ PAUSED]`/`[🚫 ABANDONED]` warn agents off dead branches. See [Lifecycle](docs/LIFECYCLE.md).

The loop: **acquire context → implement → record.** You only ever run one command yourself: `bd catchup`.

<details>
<summary>Folder-naming note</summary>

Entity extraction filters common words and branch names (`master`, `main`, `test`, `core`, …). If your project folder is named one of these, BeadsLog tracks it as `<name>-repository`. A hand-written relationship arrow targeting a filtered word is dropped with a warning — rename the entity (e.g. `master-repository`) to keep the edge.
</details>

## Docs

**Start here**
- [Running Modes](docs/RUNNING_MODES.md) — team, solo, switching: which to run and why.
- [Commands](docs/COMMANDS.md) — full command reference.
- [Use Cases](docs/USE_CASES.md) — real scenarios for agents and teams.

**Concepts**
- [Devlog](docs/DEVLOG.md) — the "Bead" and the narrative format.
- [Lifecycle States](docs/LIFECYCLE.md) — active, paused, abandoned, validated.
- [Devlog Architecture](docs/DEVLOG_ARCHITECTURE.md) — how the graph is built.

**Operate**
- [Catchup](docs/CATCHUP.md) · [Hooks](docs/HOOKS.md) · [Ollama Enrichment](docs/OLLAMA_ENRICHMENT.md)
- [Issue Architecture](docs/ARCHITECTURE.md) · [Installation](docs/INSTALLING.md)
