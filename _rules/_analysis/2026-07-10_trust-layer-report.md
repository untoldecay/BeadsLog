# BeadsLog — Project Report (Trust Layer Focus)

**Date:** 2026-07-10
**Author:** Claude (Fable 5), from devlog history (`bd devlog search/entities/status`), project docs, and a full codebase sweep.

---

## 1. What the project is

BeadsLog is a fork of Steve Yegge's **Beads** issue tracker, evolved into a **git-backed project memory system for AI agents**. The core bet: agents fail not because they can't code, but because they lack durable context. The tool gives them three interlocking memories:

1. **An issue graph** (inherited from Beads): dependency-aware work items stored as JSONL in git, cached in SQLite, hash-based IDs for conflict-free multi-agent use.
2. **A devlog memory** (the fork's contribution): markdown session narratives (`_rules/_devlog/*.md`) ingested into SQLite, FTS5/BM25 indexed, with entities auto-extracted (regex + optional Ollama) and linked into an architectural graph queryable via `search`, `graph`, `impact`, `path`, `resume`.
3. **A behavioral protocol** (`AGENTS.md`/`CLAUDE.md`): always-on instructions forcing retrieval-led reasoning — query the graph before code, record sessions after, sync through git.

Architecture: CLI (Cobra) → SQLite (cache, FTS5) → JSONL/markdown (git-synced source of truth), plus an optional Unix-socket daemon for background export, auto-commit, and AI enrichment. ~130k lines of Go, 791 files, 0.73 test-to-source ratio, very low TODO debt.

The v0.51→0.54 arc is **agent-proofing**: stub detection, "Success Trap" prevention, write-first workflows, catchup automation. The project dogfoods itself and its pain points are visible in its own data.

## 2. Key features

| Feature | What it does | Maturity |
|---|---|---|
| **Hybrid devlog search** | BM25 over FTS5, query decomposition (phrase → NEAR → AND → OR), Go-side phrase/proximity/recency bonuses, related-entity expansion | Solid |
| **Entity graph** | Auto-extracted entities, manual `A -> B (type)` arrows, recursive CTEs, depth cap 5 | Works; noisy |
| **Lifecycle tracking** | `VALIDATED / PAUSED / ABANDONED` states so agents don't build on dead code | The standout differentiator |
| **`resume` / `catchup`** | One-command context reload (agent) / team digest (human) | Core loop, works |
| **Background AI enrichment** | Daemon: regex pass → optional Ollama crystallization, graceful fallback | Solid |
| **Issue tracking + deps** | `ready`/`blocked`/`epic`, typed dep graph, JSONL merge driver | Mature (inherited) |
| **Agent-proofing suite** | `verify --fix`, stub detection, collision checks, auto-metadata | Active battleground |
| **Eval harness** | Worktree-isolated A/B/C agent runs | Paused; traces without analysis tooling |

## 3. Better for the agent

**a. Finish single-step `record` (Phase 2 plan — the data proves the need).** The DB shows **80 of 148 sessions incomplete** and **7 ghosts**; the four most recent index entries point to files that don't exist. That's the "2-Step Trap" from `_rules/_plans/devlog-agent-proofing-phase2.md` happening in production. Alternative A (write-first: `record --file` parses an existing complete file) makes an incomplete stub structurally impossible rather than caught-later.

**b. Token-efficient, machine-first output.** Devlog commands render lipgloss box-drawing tables — expensive and lossy for the primary consumer (snippets truncated mid-word inside `╭──╮` borders). Every devlog read command needs first-class `--json` or a compact non-TTY mode. Issue commands already have `--json` everywhere; the devlog side lags.

**c. Entity graph signal-to-noise.** Top-30 entities include `ulti-layered` (truncation artifact), `map it step`, `huh.select`, and three spellings of the same thing (`ollama`, `ollamaextractor`, `ollama-extractor`). Jaccard co-occurrence similarity is already computed — auto-merge high-confidence duplicates during enrichment (with `unalias` as undo). Also: the entity bonus in search scoring is wired but set to 0.

**d. Collapse protocol startup cost.** The loop asks for `resume` + `graph`/`search` + `ready` — three or four round-trips before work. A single `bd context` bundling last-session summary + ready work + memory-health warnings into one compact block would raise protocol compliance (agents skip steps under context pressure; one command is harder to skip).

**e. Ghost/error UX.** `bd devlog show <ghost>` fails with a raw `open ... no such file or directory`. The DB knows it's a ghost — say so and point at `bd devlog prune`. Every error an agent hits is a fork toward the brute-force grep spiral the protocol tries to prevent; error messages are protocol enforcement.

**f. Keep test data out of production memory.** The recent index is polluted with `Gemini-Test-Agent` stubs that became ghosts. Devlog writes during eval/test runs should be sandboxed (worktree isolation already exists) or tagged and excluded from `resume`/`search` by default.

## 4. Better for the human

- **`catchup` is the right flagship.** A `--digest` grouping by feature arc with lifecycle deltas ("X was abandoned since you last looked") would make it a standup replacement.
- **Memory health at a glance.** `bd devlog status` reported "✨ All memory optimized" while 54% of sessions are incomplete. Reconcile into one honest health score; surface it in `catchup`.
- **Graph visualization.** `bd devlog graph --html` export — terminal CTE trees don't communicate structure to people.
- **Trust indicators.** Extend lifecycle badges with provenance: which commit validated it, which test.

## 5. Optimizations

**Code structure**
- `cmd/bd/devlog_cmds.go` is 2,970 lines — split by verb. Same for `internal/rpc/server_issues_epics.go` (2,367).
- Hand-built SQL across 93 files in `storage/sqlite/` — where drift bugs will breed.

**Performance**
- Entity alias suggestion is O(n²) Jaccard — fine at 779 entities, not at 10k. Pre-bucket by token or shortlist via FTS.
- Daemon polls every 5s; fsnotify would drop latency and idle CPU.
- Search bonuses computed in Go after fetching rows — move into SQL before LIMIT as session counts grow.

**Data integrity — the biggest one**
- Three sources of truth for devlogs: markdown files, `_index.md`, SQLite — currently disagreeing (ghosts, incompletes). Stable IDs (`file.md?id=sess-xxx`) were the right move; next step is a fully reconciling `sync`: files on disk win, index rebuilt from files, DB from index, every time. `verify --fix` and `prune` are repair tools that must be remembered — same trap as the two-step record. Fold them into `sync` by default.
- Immediate hygiene: `bd devlog prune` (7 ghosts), triage the 80 incompletes (bulk `verify --fix` or a `--legacy-ok` waterline).

## 6. Overall take

**The differentiator isn't search — it's lifecycle.** Plenty of tools do RAG over notes. Almost nothing tells an agent *"this path was ABANDONED, here's why, don't build on it."* That plus git as substrate (versioned, branchable, serverless, merges like code) is a defensible position. Retrieval-led reasoning is the right instinct: agents' failure mode is confident fabrication, and forcing a graph query first is a cheap countermeasure.

**The honest tension: the tool's weakest point is its own adoption loop.** It works when agents follow the protocol, and its own database shows agents half-following it. The team has diagnosed this correctly — every release moves enforcement from *instructions* to *mechanisms*. Write-first record is the next and biggest move (BeadsLog-j6c already says so).

**Utility verdict:** for a solo developer running agents across long-lived projects, close to a killer tool — it attacks context loss, the biggest tax on agentic development. Team primitives exist (attribution, branch tracking, shared aliases) but are less battle-tested. Engineering quality is high; remaining work is about *trustworthiness of the memory itself* — entity noise, sync drift, output ergonomics. Fix the trust layer and the pitch ("agents never build on dead code or repeat historical mistakes") stops being aspirational and becomes simply true.
