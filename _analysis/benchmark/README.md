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
| 04 | This repo (full codebase) | feasibility/impact/dependency assessment | `run-04_feasibility-integration-picker.md` |

## Headline finding — bd's value scales with task complexity

| Task type | Tokens | Tool-calls | Answer quality (blind judge) |
|---|---|---|---|
| **Simple recall** ("what fixed bug X?") | tie | bd ~half | **tie** |
| **Architectural reasoning** ("assess feasibility/impact/deps of Y") | **bd −17%** | **bd −33%** | **bd 17/20 vs 11/20** |

- On flat recall a capable agent can grep, **bd ties or its output-verbosity loses** — brute-force grep is a strong baseline when keywords are guessable and the corpus is greppable.
- The one durable recall-level edge is **tool-calls / round-trips (~half)**.
- bd **wins decisively once the task requires understanding the system** — it recalls the *right* prior decisions and maps dependencies instead of brute-forcing exploration, and that grounding yields a *more correct* assessment (see Run 04's storage-design call).

## Honest caveats
- Recall result (Run 03) is n=4 with per-round judges — solid.
- Feasibility result (Run 04) is n=1 — directional; needs 2–3 more runs to harden before publishing a number.
- The bd arm = "agent WITH bd added to its normal tools" vs grep-only. That is the intended comparison (bd augments the agent).
- Judges are blind + shuffled to avoid pro-bd bias; in Run 03 the judge picked grep in 2 of 4 rounds.
