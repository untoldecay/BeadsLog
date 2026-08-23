# Run 03 — This repo, fuzzy/human recall, 4 rounds + blind judges

- **Date:** 2026-08-23
- **Corpus:** BeadsLog `_rules/_devlog/` (104 files).
- **Model:** sonnet (arms + judges). **Method upgrades over Run 02:** Arm A uses
  compact `bd ... --json`; questions are fuzzy ("how a teammate asks", no exact
  keywords); a **blind judge** scores both answer sets anonymized + shuffled.
- **Fixed question set across all 4 rounds** (cleanest variance measurement).

## Questions (fuzzy)
1. I'm worried beads clutters my team's branches when I sync — did we do anything about that?
2. If I use beads just for myself on a shared repo, could my private stuff leak to my team?
3. The relationship suggestions felt like junk a while back — what was wrong, and did we fix it?
4. Did we start anything and decide not to rely on it, or leave it half-finished?
5. How do I go from working solo back to sharing with the team without breaking things?

## Per-round data

| Round | bd: tok · calls · judge | grep: tok · calls · judge | Judge winner |
|---|---|---|---|
| R1 | 29,236 · 6 · 18/20 | 27,982 · 17 · 17/20 | bd +1 |
| R2 | 30,240 · 7 · 15/20 | 31,030 · 12 · 19/20 | grep +4 |
| R3 | 29,262 · 7 · 18/20 | 25,126 · 10 · 15/20 | bd +3 |
| R4 | 38,240 · 10 · 19/20 | 29,090 · 15 · 20/20 | grep +1 |
| **Median** | **~29.8k · 7 · 18** | **~28.5k · 13.5 · 18** | **2–2 tie** |

## Blind judge verdicts (Set 1/Set 2 were shuffled per round; de-anonymized here)
- **R1 (bd 18, grep 17):** bd surfaced the sync-branch fix (`2lt`) as a *separate* Q1 concern grep missed, and split Q3 into two bugs (`rob`+`819`). grep edged correctness on a `bd team` vs `bd init` naming nit.
- **R2 (grep 19, bd 15):** the **bd arm hallucinated a non-existent `bd team` command** on Q5 and missed a second abandoned item (lean-protocol); grep got both right.
- **R3 (bd 18, grep 15):** bd named the sync-branch-pollution root cause for Q1 and split Q3's two bugs; grep collapsed them.
- **R4 (grep 20, bd 19):** grep's Q4 named two abandoned threads (GitHub/Notion + Gemini eval-harness pivot) with distinct fates; bd named one-and-a-half.

## Conclusion
- **Answer quality: dead tie** (bd median 18, grep median 18; each won 2 of 4 rounds).
  bd twice hallucinated a `bd team` command — a real downside.
- **Tokens: tie** (~29.8k vs ~28.5k).
- **Tool-calls: bd wins every round** (median 7 vs 13.5 — ~half the round-trips).

On greppable recall, bd's only durable edge is fewer round-trips. Quality and tokens
are ties. This motivated Run 04: test a task that requires *reasoning about the
system*, not just recalling a fact.
