# bd bench — Run 4 (FAIR: terse commits, no devlogs, contracts kept)

> ⚠️ **CAPTION (added 2026-08-25):** the bd arm (Arm A) in runs 1–5 was **bd-mostly-ALONE, not bd+grep** —
> retrieval-led, and it did not reliably drop to code to verify (proof: in Run 5 it reported the throttle
> as "~100ms" from contract 02, never opening `NoteEditorView.vue:1061` to see the real 50ms). That
> misrepresents real bd usage (index → then grep the pointed spot). Do NOT cite runs 1–5 as bd-vs-grep
> evidence. The conclusion-bearing runs are **6 (synthesis) and 8 (recall)**, whose bd arms DID
> index-then-read-code. See `2026-08-25_bd-bench_run8_recall-matched-protocol.md` "bd is an INDEX for grep".

**Date:** 2026-08-25
**Design rationale:** Run 3 stripped devlogs but left the rich commit BODIES (which feed the
user's Mattermost channel — they carry the "why"). Run 5 (deny contracts) was rejected as unfair:
contracts are the user's own spec-first discipline, NOT part of BeadsLog, and are present regardless.
So Run 4 keeps the user's real always-on discipline (code + code comments + contracts + conventional
commits) and strips only the two BeadsLog-adjacent crutches: **the devlogs and the rich commit bodies**.
Arm B got a **genericized terse `fix:/feat:/chore:` log (subjects only)**, code + contracts, and
`git show -- <path>` for diffs. Arm A (bd) = Run 2's answers reused (constant). Same 4 questions.

## Results
| Arm | Tokens | Tool-uses | Wall time | Judge score (blind) |
|---|---|---|---|---|
| A — BeadsLog | 32,452 | 10 | 45s | **20/20** |
| B — terse grep (no devlogs, no bodies) | 30,799 | 14 | 94.6s | 17/20 |

**Per-question:** Q1 = **bd**, Q2 = tie, Q3 = tie, Q4 = **bd**. bd won 20 vs 17.

- **Q1 (why Phase 2 parked) — bd wins:** the granular polish-tail (entry blink → busy per-hop curtain → tray reflow → 1-frame width snap) and the hard metrics (22.4s→5.1s, worst frame 1833→582ms) live ONLY in the devlog. Phase 2 was *reverted*, so there's no code trace; terse grep got a generic "diminishing returns" with a self-admitted inference caveat.
- **Q2 / Q3 — tie:** their rationale is ALSO in code comments (the ~38s-Vue-patch trace, the uncapped-pool comment at `DrawerPanelView.vue:1113`) and contracts (10, 02) — terse grep mined those fully.
- **Q4 — bd edges it:** terse grep had a minor §24/§25 mis-cite of contract 02.

## The finding (this time the prediction held)
On the FAIR test — strip devlogs + rich commit bodies but keep the user's other real discipline —
**bd separates, and exactly where predicted:** on the "why" that has **no other home**. Abandoned/
reverted experiments leave no code trace, and quantitative outcomes aren't in code or contracts, so
the devlog is the *only* place that rationale survives. On questions whose rationale the user ALSO
puts in code comments + contracts (Q2, Q3), it stays a wash.

## Honest synthesis (runs 1–4)
| Run | grep sees devlogs? | grep sees rich commit bodies? | contracts? | Quality winner |
|---|---|---|---|---|
| 1 (1 Q) | yes | yes | yes | **bd** 20–16 |
| 2 (4 Q) | yes | yes | yes | grep 20–18 |
| 3 (4 Q) | no | yes | yes | grep 20–19 |
| 4 (4 Q) | **no** | **no** | yes | **bd** 20–17 |

**Conclusion:** grep ties or wins as long as it can read EITHER the devlogs OR the rich commit bodies —
both carry the "why." Remove both (Run 4) and **bd wins**, because the fully-granular rationale of
**abandoned paths + measured outcomes** exists only in the devlog. The questions that stayed tied are
the ones whose rationale the user redundantly captured in **code comments + contracts** — that
discipline, not BeadsLog, is what makes grep competitive.

**So bd's marginal value on this repo = capturing the "why" that structurally can't live anywhere else:**
parked/reverted experiments, quantitative results, and cross-cutting narrative — delivered as one
authored account instead of scavenging four sources. Its other claimed decisive edges (large corpora,
cross-session recall over months, semantic-mismatch queries at scale, impact/graph, lifecycle badges)
were not exercised here and would only widen the gap.

**Efficiency footnote:** bd was also faster/leaner in every run (fewer round-trips, 45s vs 94–173s
wall-time for the grep arms that had to reconstruct from scattered sources).

## Daemon hygiene
Baseline `72371` only; no strays across any run. Environment restored.
