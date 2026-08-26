# Benchmarks: where durable memory actually helps

**The question isn't "does BeadsLog beat `grep`."** That framing is a trap: in real use `bd` *aims* grep — it points to where an answer lives, then you read that spot — so they're partners, not rivals. The real question is **when a durable memory layer changes the outcome, and when the code already carries it.**

We ran two benchmarks — a first pass on this repo, then a harder, more self-critical one on a second, unusually well-documented codebase ([Fray traces](../_analysis/benchmark-from-fray/)). The honest synthesis:

## What we found

- **Re-deriving "why shipped code works" is a wash.** On a well-documented repo the reasoning is redundant across code comments, commits, and design notes — strip any three and the fourth still answers. No retrieval method had a decisive coverage edge, and a blind judge scored *identical* text with a ±25% swing, larger than any BeadsLog-vs-grep gap. Don't read per-round "winners" as signal.
- **Deciding what to build next is where memory pays.** On a feasibility task, the arm with only live code mis-framed a feature as greenfield and missed an abandoned-corruption history the team had already fought — advice that would re-open a solved problem. Durable memory prevented it. This is the clearest, non-noisy win — but it's the *memory* (devlogs + design notes), not `bd`'s tooling specifically; thorough grep over the same notes tied it.
- **BeadsLog's own repeatable edges are narrow and real:** the **"why with no code home"** — abandoned experiments, numbers you tried and reverted, designs you decided against, which live nowhere in the code — and **semantic recall** when your words don't match the code's identifiers (grep confidently pinned the wrong mechanism; the concept index found the right one).
- **The efficiency case is unproven, not disproven.** On a small, familiar repo, plain grep was as fast or faster — `bd`'s index step is overhead when grep already finds things easily. "Aim grep so you brute-force less" needs a large or unfamiliar corpus we haven't tested. No efficiency claim until then.

> **Learned the hard way:** curated memory can go *stale vs the code* (a design note said ~100ms; the code said 50ms). Memory helps when checked against the code and misleads when trusted as authority — which is why it's most valuable for *deciding*, and least necessary for *re-reading* shipped behavior.

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

This is where durable memory pays — it surfaces prior designs and abandoned experiments the raw code no longer shows. (The win is *having* the memory; thorough grep over the same notes matches it. The retrieval tooling isn't what wins here.)

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

These early per-run "winners" looked encouraging, but a follow-up showed the judge swings ±25% on identical text — so the individual results are within noise, not signal. The rigorous re-test lives in [`_analysis/benchmark-from-fray/`](../_analysis/benchmark-from-fray/): pre-registered coverage checklists, fact-verified ground truth, unprimed judges, on a second codebase. Its verdict is the honest one and it's what "What we found" above reports — recall is a wash, memory changes *decisions*, and on a small familiar repo plain grep was as efficient or better.

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

- **`grep` is a strong baseline — often the winner.** On a well-documented, familiar repo it reconstructs "why" from code + comments + commits, and finds files as fast as `bd` or faster. A grep tier took #1 on every multi-tier run.
- **Judge noise is large.** Identical text scored across a ±25% range. Treat single-run results as directional at best; the signal only survives across many runs and pre-registered checklists.
- **Curated memory can go stale vs the code.** A design note said ~100ms where the code said 50ms. "Authoritative" is only as good as its upkeep — verify against code before trusting it.
- **The efficiency / "aim grep" win is untested at scale.** On a small familiar repo `bd`'s index step is overhead. Its plausible payoff — a large or unfamiliar codebase where blind grep drowns in false hits — is the single most valuable experiment we haven't run. Don't claim it yet.
- **`bd` is meant as an index for grep, not its rival.** These runs pit them apart to isolate each; real use is `bd` locates → you read that spot. Partners, not opponents.
- **What's plausible but unproven** (would favor `bd`, honestly): cross-session recall over months, impact/graph analysis, shared team memory, lifecycle badges. This bench probed small, greppable "why" questions — a fair *floor* for `bd`, not its ceiling.

---

## Product findings the benchmark surfaced

The benchmark doubled as a bug finder:

1. **Cite committed filenames, not `sess-xxxxx`.** A judge flagged BeadsLog's *correct* answer as a "grounding risk" because it cited opaque session IDs; grep cited `issues.jsonl:40`. Search now returns the committed filename.
2. **Deterministic graph.** A one-line `ollama.go` fix (`temperature=0, seed=42`) closes ~90% of cross-machine drift.
3. **`bd devlog export` and `bd devlog watch`** — both real ~80-line features, surfaced as feasibility questions and shipped.
