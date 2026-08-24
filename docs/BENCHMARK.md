# Benchmarks: BeadsLog vs `grep`

**Question:** does querying a BeadsLog graph actually beat an agent brute-forcing `grep` over the same history?

**Finding:** it depends on the *kind* of question — and we measured both, honestly. Grep ties on flat lookups; BeadsLog pulls ahead as the question needs understanding the system.

## The headline

| Question type | Sample | Answer quality (blind judge) | Retrieval cost |
|---|---|---|---|
| **Recall** — "what fixed X?" | n=4 rounds | tie (18 vs 18) | BeadsLog **~half the tool-calls** (7 vs 13.5) |
| **Engineering reasoning** — "assess impact/deps of Y" | n=4 tasks | **BeadsLog +2** (median 16.5 vs 14.5) | cheaper in **3 of 4** (median 76k vs 89k tokens) |

The value scales with task complexity: on a small, greppable corpus a capable agent with good keywords ties; BeadsLog's durable edge is **fewer round-trips**, and its decisive wins come when the task requires recalling prior decisions and mapping dependencies instead of brute-forcing exploration.

---

## Test harness and corpus

Every run used the same setup, so the only variable is the retrieval method:

- **Harness:** Claude Code, spawning isolated sub-agents through its Task/Agent tool — one per arm plus the judge, each with its own clean context.
- **Model:** Claude Sonnet for both arms and the judge, identical every run. Same model on both sides means a win reflects the *retrieval method*, not raw model strength.
- **Corpus:** this repository — 104 committed devlogs and the entity graph derived from them. Runs 04–07 also gave both arms the full Go codebase (800+ files) to explore.
- **The BeadsLog arm is the agent *with* `bd` added to its normal tools**, not a bd-only agent — the question is whether retrieval helps an already-capable agent, so grep-and-read stays available to it too.
- **Metrics source:** tokens and tool-calls come from each sub-agent's own usage report (no estimation); answer quality comes from the blind judge's 1–5 scores.
- **Dates:** runs 01–03 earlier; runs 04–07 on 2026-08-23.

---

## 1. How it works

Every run spawns isolated sub-agents so contexts never leak between arms:

- **Arm A — BeadsLog:** may use `bd devlog search / graph / impact / resume` (with `--no-daemon`) plus reading files.
- **Arm B — brute-force:** may use only `grep` / `ls` / `read` over the devlog directory. No `bd` at all.
- **Blind judge (3rd agent):** gets both answer sets **anonymized and shuffled** as "Set 1" / "Set 2" — never told which used BeadsLog. Scores each 1–5 on correctness, significance (conveys *why*), impact/dependency awareness, and actionability.

Both arms get identical questions and the same instruction: answer in 2–3 sentences, include *why it mattered* and what it affects, and cite the source.

**Measured per arm:** tokens · tool-calls · correctness · blind-judge score.

---

## 2. The two question types

The result splits cleanly by question type, so we test both.

**Recall** — a fact lives in one devlog; the question is whether the agent can find it.

> "Did we ever give up on an approach for cross-machine determinism?"

Phrased the way a teammate actually asks — fuzzy, no exact keywords from the source. Keyword-exact questions ("what does function Y do?") make grep trivial and don't test retrieval.

**Engineering reasoning** — the answer requires stitching together decisions, dependencies, and abandoned branches across sessions.

> "Assess the feasibility, impact, and dependencies of adding an interactive integration-picker to `bd init`. Comprehensive report."

This is where a graph of prior decisions beats re-deriving context from raw files.

---

## 3. The runs

Full traces live in [`_analysis/benchmark/`](../_analysis/benchmark/).

| # | Corpus | Question style |
|---|---|---|
| 01 | Tasklet fixture (8 devlogs) | keyword recall |
| 02 | This repo (104 devlogs) | keyword recall |
| 03 | This repo (104 devlogs) | fuzzy/human recall, 4 rounds + judges |
| 04 | This repo (full codebase) | feasibility: init integration-picker |
| 05 | This repo (full codebase) | feasibility: reproducible graph / drift |
| 06 | This repo (full codebase) | feasibility: `bd devlog export` |
| 07 | This repo (full codebase) | feasibility: `bd devlog watch` |

BeadsLog won answer quality in 3 of 4 feasibility tasks (Run 04 +6, Run 06 +4, Run 07 +1) and was cheaper in 3 of 4. Grep won Run 05 by 1 and crushed Run 06 on tokens — it can reconstruct committed history with more effort.

---

## 4. Reproduce it on your own repo

```bash
bd bench "why did we switch storage?"   # benchmark one question
bd bench                                  # agent generates its own question set
```

`bd bench` prints a harness-agnostic protocol the current agent runs: it spawns the two arms plus a blind judge, measures tokens and tool-calls, and reports honestly — a tie or a loss on a greppable corpus is a valid result.

> **Daemon hygiene:** the protocol spawns multiple sub-agents that each run `bd`, and every `bd` call can auto-start a background daemon. Left unchecked they form a swarm that re-exports `issues.jsonl` and can push corrupted data. `bd bench` includes a hygiene section — snapshot daemons first, pass `--no-daemon` everywhere, and stop strays with `bd daemon stop <workspace-path|pid>` (never `pkill -9` or `bd daemon killall`, which hit unrelated projects).

---

## 5. Honest caveats

- `grep` is a strong baseline. A capable agent with good keywords reconstructs committed history (issues.jsonl, devlogs) with more effort — it won 3 of 8 judged rounds overall.
- The BeadsLog arm is "agent **with** bd added to its normal tools" vs grep-only — the intended comparison (bd *augments* the agent).
- Judges were blind and answer sets shuffled, to keep pro-BeadsLog bias out.
- Samples are directional (n=4 per regime), but consistent across runs.
- BeadsLog's edge grows with corpus size, with questions worded differently from the source (semantic recall), and on things grep can't do at all: cross-session memory, lifecycle badges, shared team graph, impact analysis.

---

## Product findings the benchmark surfaced

The benchmark doubled as a bug finder:

1. **Cite committed filenames, not `sess-xxxxx`.** A judge flagged BeadsLog's *correct* answer as a "grounding risk" because it cited opaque session IDs; grep cited `issues.jsonl:40`. Search now returns the committed filename.
2. **Deterministic graph.** A one-line `ollama.go` fix (`temperature=0, seed=42`) closes ~90% of cross-machine drift.
3. **`bd devlog export` and `bd devlog watch`** — both real ~80-line features, surfaced as feasibility questions and shipped.
