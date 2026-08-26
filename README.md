# BeadsLog

**Agent-first codebase memory.**

BeadsLog remembers what your project actually became — the decisions, the dead
ends, which components depend on which — recorded as you and your agent work and
stored as a searchable graph in git.

That memory pays off twice. Agents reload full context in seconds — the
architecture, the past decisions, why the last change was made — so you stop
re-explaining your project every session. And it catches drift: BeadsLog records
what you and your agent actually decided, so when an agent writes thousands of
lines a session, the places the code wandered from the plan stay visible instead
of compounding.

Built on [Beads](https://github.com/steveyegge/beads), a git-native issue
tracker for AI agents.

**For you if:**
- You vibecode on a team at high pace. Everyone's agents share one memory and build on each other's work instead of colliding or redoing it.
- You're a solo vibecoder juggling many projects. Pick one back up after weeks; your agent reloads the context in seconds.
- You don't have the dev reflexes yet — git hygiene, writing down scope and architecture, coordinating work across people. BeadsLog handles that.
- You re-explain your project to your agent every session.

**Not for you if:** you're a seasoned dev who codes without AI agents.

**Built for vibecoders, agent-agnostic** — works with Claude Code, Cursor, Codex, ….

## Quick Start

Requires [Go](https://go.dev/doc/install) 1.24+.

```bash
go install github.com/untoldecay/BeadsLog/cmd/bd@latest
bd init        # set up in this project — private by default, nothing is shared unless you choose to
bd onboard     # ask your agent to run this; it reads the setup and connects itself
```

> [!TIP]
> BeadsLog keeps its notes on a separate git branch, not mixed into the code you
> and your team work on. Working solo, or on a team that doesn't use BeadsLog?
> [Running Modes](docs/RUNNING_MODES.md) shows the fully private setup.

## What agents do with it

As your agent works, it writes a short devlog — plain markdown, a few sentences.
You don't write these; the agent does:

> **Switched auth to webhook-first validation** — the sync response timed out under load. Touches `AuthService`, `StripeClient`.

Weeks later, any agent reads that back before touching the code — no
re-explaining, no `grep`-and-guess. Your agent runs these, not you (local and
instant — no network, no waiting):

| Your agent's job | Command it runs |
|---|---|
| Load where I left off | `bd devlog resume --last 1` |
| Find a past decision | `bd devlog search "stripe timeout"` |
| See what breaks if I change X | `bd devlog impact "AuthService"` |
| See the key components | `bd devlog entities` |
| Trace how A connects to B | `bd devlog path "A" "B"` |
| See what the team shipped | `bd catchup` |

Trust badges keep the history honest: `[🟢 VALIDATED]` code landed in main;
`[⏸ PAUSED]` and `[🚫 ABANDONED]` warn agents off dead branches. See
[Lifecycle](docs/LIFECYCLE.md).

The loop — acquire context, implement, record — runs as your agent works. The
one command you run yourself: `bd catchup`.

<details>
<summary>Folder-naming note</summary>

Entity extraction filters common words and branch names (`master`, `main`,
`test`, `core`, …). If your project folder is named one of these, BeadsLog tracks
it as `<name>-repository`. A hand-written relationship arrow targeting a filtered
word is dropped with a warning — rename the entity (e.g. `master-repository`) to
keep the edge.
</details>

## Does it hold up?

We benchmarked it honestly — losing runs included. On a small, well-documented
repo, re-deriving *why shipped code works* is a wash: the code, comments, and
commits already carry the reasoning, and a capable `grep` keeps pace. BeadsLog
earns its keep in two narrower, sturdier places — **deciding what to build next**
(it surfaces prior designs and abandoned experiments, so you don't re-open a
problem the team already closed) and **remembering what the code can't**: the path
you abandoned, the number you tried and reverted, the design you decided against.
See the [benchmarks](docs/BENCHMARK.md) — the method, every run, and the losses
left in.

## Docs

**Start here**
- [Running Modes](docs/RUNNING_MODES.md) — team, solo, switching: which to run and why.
- [Commands](docs/COMMANDS.md) — full command reference.
- [Use Cases](docs/USE_CASES.md) — real scenarios for agents and teams.

**Concepts**
- [Devlog](docs/DEVLOG.md) — the "Bead" and the narrative format.
- [Lifecycle States](docs/LIFECYCLE.md) — active, paused, abandoned, validated.
- [Devlog Architecture](docs/DEVLOG_ARCHITECTURE.md) — how the graph is built.
- [Benchmarks](docs/BENCHMARK.md) — BeadsLog vs `grep`, blind-judged; run your own with `bd bench`.

**Operate**
- [Catchup](docs/CATCHUP.md) · [Hooks](docs/HOOKS.md) · [Ollama Enrichment](docs/OLLAMA_ENRICHMENT.md)
- [Issue Architecture](docs/ARCHITECTURE.md) · [Installation](docs/INSTALLING.md)
