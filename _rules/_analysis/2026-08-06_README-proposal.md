# PROPOSAL — New README.md (not applied)

> Canonical definition all other surfaces must derive from:
> **"BeadsLog is git-backed project memory for AI coding agents. It combines an issue tracker (`bd`, extending Beads) with a devlog knowledge graph, so agents never repeat mistakes, rebuild dead code, or lose decisions between sessions."**
>
> Naming rule: **BeadsLog** = the system · **bd** = the CLI · **Beads** = the upstream issue-tracker engine we extend · **devlog** = the memory layer.

---

# BeadsLog

**Git-backed project memory for AI coding agents.**

AI agents forget everything between sessions. They re-make decisions, re-try abandoned approaches, and break components they can't see. BeadsLog fixes this: every session's decisions, fixes, and dead ends are recorded, indexed, and queryable — by the next agent, on any machine, via `git pull`.

## What it is

Two halves, one system:

- **`bd` — a dependency-aware issue tracker** (extends [Beads](https://github.com/untoldecay/BeadsLog)): tasks, bugs, and their dependency chains, all in the terminal, JSON-output for agents.
- **The devlog — a session memory graph**: markdown session logs turned into a searchable, verifiable knowledge graph. `bd devlog resume` loads yesterday's context; `bd devlog search` finds the fix from three weeks ago; `bd devlog graph`/`impact` shows what depends on what *before* you touch it. Sessions carry lifecycle states (`VALIDATED` / `PAUSED` / `ABANDONED` + reason) so abandoned experiments are never mistaken for baseline truth.

## How it stores things (and why Git doesn't fight it)

| Layer | Where | In Git? |
|---|---|---|
| SQLite DB — queries, search, graph | each machine | ❌ ignored, never conflicts |
| `issues.jsonl` — **one issue per line** (JSONL, not JSON) | repo | ✅ auto-merged per-issue by a custom merge driver |
| Devlogs — one markdown file per session | repo | ✅ append-only, near-zero overlap |

Different issues live on different lines → Git merges them like different functions. Same-issue edits are resolved per-field by the `bd` merge driver (last-write-wins on timestamp). You never see conflict markers. Memory travels with `git pull` — same branch, same PR, same review as the code.

## Quick start

```bash
bd init          # set up in your repo (solo, team, or contributor mode)
bd onboard       # activate an agent: sync DB, ingest devlogs, verify graph
bd ready         # see what's ready to work on
```

## Learn more (progressive disclosure)

| I want to… | Read |
|---|---|
| See real workflows | [USE_CASES.md] |
| Understand session lifecycle | [LIFECYCLE.md] |
| Write & query devlogs | [DEVLOG.md] |
| Catch up after time away | [CATCHUP.md] |
| See the internals | [ARCHITECTURE.md] / [DEVLOG_ARCHITECTURE.md] |
| Local enrichment with Ollama | [OLLAMA_ENRICHMENT.md] |
| Git hooks reference | [HOOKS.md] |
| Install | [INSTALLING.md] |

---

## Why this draft (rationale, not part of the README)

1. **Leads with the memory identity** (the differentiator) and demotes "issue tracker" to a component — resolving the two-identities drift.
2. **Imports the two strongest sections of the 2026-08-06 explainer**: the 3-failure-modes pitch and the 3-layer/JSONL no-conflict table — the exact objections outsiders raise (cf. Slack debate).
3. **States the naming rule once** so every other surface can reference it.
4. Keeps progressive disclosure: 30-second pitch → mechanics table → quick start → deep links.
