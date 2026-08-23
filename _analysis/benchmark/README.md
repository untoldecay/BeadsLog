# BeadsLog retrieval benchmark — trace log

Traces of live A/B benchmarks comparing an agent answering questions about this
repo **with BeadsLog retrieval** (`bd devlog search/graph/impact` + normal tools)
vs **brute-force grep only**. Each run spawns two arms (same task, same model,
isolated contexts) and, from Run 03 on, a **blind judge** (given both answer sets
anonymized + shuffled) scoring correctness / significance / impact-awareness /
actionability. Metrics: tokens + tool-calls (from sub-agent usage) + judge score.

Method is codified in the `bd eval bench` command and the `audience-test`-style loop.

## Runs

| # | Corpus | Question style | File |
|---|---|---|---|
| 01 | Tasklet fixture (8 devlogs, since dropped) | keyword recall | `run-01_fixture-keyword.md` |
| 02 | This repo (104 devlogs) | keyword recall | `run-02_real-keyword.md` |
| 03 | This repo (104 devlogs) | fuzzy/human recall, 4 rounds + judges | `run-03_real-fuzzy-4rounds.md` |
| 04 | This repo (full codebase) | feasibility: init integration-picker | `run-04_feasibility-integration-picker.md` |
| 05 | This repo (full codebase) | feasibility: reproducible graph / drift | `run-05_feasibility-drift.md` |
| 06 | This repo (full codebase) | feasibility: `bd devlog export` | `run-06_feasibility-export.md` |
| 07 | This repo (full codebase) | feasibility: `bd devlog watch` | `run-07_feasibility-watch.md` |

## Headline finding — bd's value scales with task complexity

| Task type | Sample | Tokens | Tool-calls | Answer quality (blind judge) |
|---|---|---|---|---|
| **Recall** ("what fixed bug X?") | n=4 rounds | tie | **bd ~half** (7 vs 13.5) | **tie** (18 vs 18) |
| **Engineering reasoning** ("assess feasibility/impact/deps of Y") | n=4 tasks | **bd usually cheaper** (median 76k vs 89k) | **bd usually fewer** (median 53 vs 64.5) | **bd +2** (median 16.5 vs 14.5) |

- On flat recall a capable agent can grep, **bd ties on quality and tokens**; its one durable edge is **fewer tool-calls / round-trips (~half)**.
- bd **wins as the task requires understanding the system** — it recalls prior decisions and maps dependencies instead of brute-forcing exploration. It won quality in **3 of 4** feasibility tasks (Run 04 +6, Run 06 +4, Run 07 +1) and was cheaper in 3 of 4.
- **Not absolute:** grep can reconstruct committed history (issues.jsonl, devlogs) with more effort — it won Run 05 by 1, and crushed Run 06 on tokens. bd sometimes over-explores.

## Product findings surfaced by the benchmark
1. **Cite committed filenames, not `sess-xxxxx`.** In Run 05 the judge flagged bd's *correct* answer as a "grounding risk" because it cited opaque devlog session IDs it couldn't verify; grep cited `issues.jsonl:40` line numbers. `bd devlog search/graph` should return the committed filename for verifiability.
2. **Deterministic graph** (Run 05) — a one-line `ollama.go` gap (`temperature=0, seed=42`) closes ~90% of cross-machine drift; `graph.jsonl` (BeadsLog-5xf) is the 100% design.
3. **`bd devlog export`** (Run 06) and **`bd devlog watch`** (Run 07) are both real ~80-line features reusing existing machinery.

## Honest caveats
- Recall (Run 03) n=4 with per-round judges; feasibility now n=4 — both directional but consistent.
- The bd arm = "agent WITH bd added to its normal tools" vs grep-only — the intended comparison (bd augments the agent).
- Judges blind + shuffled to avoid pro-bd bias; grep still won 3 of 8 judged rounds overall (Run03 R2/R4, Run05).
