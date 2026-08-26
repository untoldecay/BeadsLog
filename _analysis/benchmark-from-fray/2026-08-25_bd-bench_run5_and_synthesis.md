# bd bench — Run 5 (code-only) + full synthesis (runs 1–5)

> ⚠️ **CAPTION (added 2026-08-25):** the bd arm (Arm A) in runs 1–5 was **bd-mostly-ALONE, not bd+grep** —
> retrieval-led, and it did not reliably drop to code to verify (this run is the PROOF: bd reported the
> throttle as "~100ms" from contract 02, never opening `NoteEditorView.vue:1061` to see the real 50ms).
> That misrepresents real bd usage (index → then grep the pointed spot). Do NOT cite runs 1–5 as
> bd-vs-grep evidence. The conclusion-bearing runs are **6 (synthesis) and 8 (recall)**, whose bd arms DID
> index-then-read-code. See `2026-08-25_bd-bench_run8_recall-matched-protocol.md` "bd is an INDEX for grep".

**Date:** 2026-08-25

## Run 5 — "typical other user": code only (no contracts, no devlogs, no rich commits)
Requested because most BeadsLog users won't have the contract discipline. Arm B: code (+ comments)
+ a terse conventional log + `git show -- <path>` diffs. Arm A (bd) = Run 2's answers reused.

| Arm | Tokens | Tool-uses | Wall | Judge |
|---|---|---|---|---|
| A — BeadsLog | 32,452 | 10 | 45s | 15/20 |
| B — **code only** | 37,251 | 14 | 78s | **20/20** |

**Per-question:** Q1 = tie, Q2 = code, Q3 = tie, Q4 = **code**. **Code-only won 20 vs 15.**

Two things drove it:
1. **The code comments are the memory.** They carry the rationale AND trace numbers (e.g. "~20-24×/s", the uncapped-pool comment at `DrawerPanelView.vue:1095`), so code-only reconstructed Q1–Q3 fully.
2. **bd inherited a STALE number.** bd's Q4 said Pipe-1 throttle "~100ms" (from **contract 02**); the code actually uses **50ms** (`NoteEditorView.vue:1061`). The contract drifted from the code, and bd trusted the contract. Code-only read ground truth and was *more correct*.

## Two honesty caveats on Run 5
- **I primed this judge** — I told it to verify the 50ms-vs-100ms discrepancy. Earlier judges weren't cued and didn't penalize bd's 100ms. So part of bd's low score here is my prompt, not pure blind judgment. Disclosed, not hidden.
- Even so, the underlying fact (contract stale vs code) is real and independently verifiable.

## The bigger finding: JUDGE NOISE dominates
bd's answer was **byte-identical** across runs 2–5 (reused). Four different blind judges scored that same text **18, 19, 20, 15** — a **±5/20 (25%) swing**. Grep's scores ranged 16–20. **The two ranges overlap almost entirely**, so the per-run "winners" are within noise. See `2026-08-25_bd-bench_scorecurve.svg`.

## Full scoreboard
| Run | Q | grep sees | bd | grep | winner |
|---|---|---|---|---|---|
| 1 | 1 | devlogs+commits+contracts | 20 | 16 | bd |
| 2 | 4 | devlogs+commits+contracts | 18 | 20 | grep |
| 3 | 4 | commits+contracts (no devlogs) | 19 | 20 | grep |
| 4 | 4 | contracts (no devlogs/commits) | 20 | 17 | bd |
| 5 | 4 | code only | 15 | 20 | grep(code) |

## Honest conclusions (what the 5 runs actually support)
1. **Answer quality is a statistical wash on this corpus.** Judge variance (±25% on identical text) is larger than any bd-vs-grep gap. Do not read the individual winners as signal.
2. **The rationale is redundant across FOUR layers** — devlogs, commit bodies, code comments, contracts. Strip any three and the fourth still answers most "why" questions. That redundancy is a discipline achievement, and it *compresses bd's marginal recall value toward zero on shipped work*.
3. **Code (+ rich comments) is the freshest, strongest single source.** Curated memory (contracts, devlogs) can go **stale vs code** — bd lost Q4 to exactly that (100ms vs 50ms). "Authoritative" is only as good as its upkeep.
4. **bd's one genuinely-unique win** (Run 4 Q1) was the granular detail of a **parked/reverted** experiment — no code trace, so the devlog was the only home. That is bd's real niche here: **abandoned paths, measured outcomes, cross-cutting narrative.**
5. **bd's robust, non-noisy edge = efficiency.** Fewer tool-calls and ~2–4× faster wall-time than the grep arms that had to reconstruct from scattered sources — consistent across every run.
6. **Not exercised (would favor bd, honestly):** large corpora, cross-session recall over months, semantic-mismatch queries at scale, impact/graph, lifecycle badges, shared team memory. This bench probed small, greppable "why" questions on an unusually well-documented repo — a fair FLOOR for bd, not its ceiling.

## Method caveats (for posterity)
- Single judge per run; N=1 or 4 questions; one repo. Directional, not statistical.
- Arm A reused across runs 3–5 (its method/corpus was the invariant under test).
- One judge (Run 5) was primed on a specific discrepancy — noted above.

## Daemon hygiene
Baseline `72371` only; zero strays across all five runs. Environment restored each time.
