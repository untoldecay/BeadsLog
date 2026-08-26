# bd bench — Run 8 (RECALL, settled protocol — the matched pair to Run 6)

**Date:** 2026-08-25
**Why:** Runs 1–5 tested recall but with an inconsistent, noisy, partly-primed judge (±25% noise,
one primed judge, reused bd answers). This run re-tests recall ONCE under the **settled Run 6/7
protocol** (5 access tiers, fresh arms, unprimed blind judge, pre-registered coverage checklist +
fact-verification) so the recall finding has the same rigor as the synthesis finding. It is the
**recall half of the matched pair**: Run 6 = synthesis, Run 8 = recall, identical method.

Runs 1–5 are hereby demoted to "how the method was developed" (a dev trail), not measurements.

## Question (recall — "why did shipped code end up this way")
*"Why is the drawer editor's live typing sync split into two pipes, why is each timed the way it is
(throttle vs debounce), and what breaks if you get it wrong?"*

Chosen so the answer naturally hits the **100ms-vs-50ms contract drift** — letting an UNPRIMED judge
re-catch it via fact-check, testing whether the drift misleads on recall.

## Ground truth (verified in code BEFORE pre-registering — the B6 lesson)
| Fact | Code | Contract 02 |
|---|---|---|
| Pipe-1 throttle | **50ms** (`NoteEditorView.vue:1064`) | ~100ms (`02:15`) → **DRIFT, code is truth** |
| Pipe-2 debounce | **800ms** (`:1069`) | ~800ms (`02:16`) → match |

## Result (blind judge, coverage checklist C1–C10, all ms values fact-checked)
| Rank | Arm | Cov | Note |
|---|---|---|---|
| 1 | **grep-TERSE** | 10/10 | Only arm to hit every item (incl. C8 title + C10 no-metadata-on-stream) AND flag the drift. Tightest contract line-mapping. |
| 2 | grep-FULL | 9/10 | Drift flagged, both cascade traces (24s+38s), timestamp guard; missed C10. |
| 3 | grep-NODEVLOG | 9/10 | Drift flagged, best mid-burst→caret-jump causal chain; missed C8. |
| 4 | grep-CODE-ONLY | 8/10 | Correct values but **could not flag drift (no contract to compare)**; missed C10. |
| 5 | **bd** | 7/10 | **Last on the checklist** — but the ONLY arm to surface the tuning HISTORY (outside the rubric). |

**All five stated 50ms and 800ms correctly. The drift misled no one.**

## The three findings

**1. Recall coverage is a wash — confirmed under the tight protocol.** All arms landed 7–10/10 and every
arm got the timing values right. This is the same null result runs 1–5 produced, now without the judge
noise/priming defects — so it's the *robust* version. The "why" is redundant across code comments,
contract 02, and the devlogs; no retrieval method has a decisive coverage edge.

**2. The drift is a PLANNING hazard, not a RECALL hazard.** In Run 7 (planning) the stale "no toast" /
"~100ms" contracts misled arms that cited them as design authority. Here (recall) **every arm opened the
code file that holds the real value**, so nobody parroted the stale 100ms — three arms explicitly flagged
the drift, and even code-only (which *can't* detect drift, having no contract) still reported the correct
50ms because it read the code. **Lesson: curated memory drifts, but it only misleads when it's trusted as
authority instead of verified against code — which is exactly what planning does and recall doesn't.**

**3. bd ranked LAST on the checklist but owns the one thing nothing else had.** bd alone reconstructed the
tuning trajectory (persistence 500→200, dual-pipe split 100/800, latency pass stream 100→50 + editor
300→150, measured cross-window ~410→210ms) and honestly noted 800ms was never re-tuned — from two
devlogs. This is **outside the 10-item checklist** (the rubric asks *what the design is + what breaks*,
never *how the numbers were arrived at*), which is why it cost bd coverage points. It is also bd's
**consistent, established niche across the whole benchmark** (Run 4 Q1, Run 6 C6): the *why-with-no-code-
home* — parked experiments, superseded values, measured outcomes.

## The matched pair (the article's backbone)
| | Run 6 — SYNTHESIS ("should we build X?") | Run 8 — RECALL ("why is shipped X this way?") |
|---|---|---|
| #1 | grep-FULL 8.5 | grep-TERSE 10 |
| bd | #2 tie-adjacent (8.5) — unique: abandoned-work risk | **#5 on rubric (7)** — unique: tuning history |
| worst | grep-CODE-ONLY 6.5 — **mis-framed as greenfield** | grep-CODE-ONLY 8 — fine on facts, can't detect drift |
| takeaway | **memory changed the decision** | **memory didn't change the recall; code carries it** |

**Combined, uniform-method conclusion:** durable memory earns its keep when you're **deciding what to
build next** (surfaces prior designs + abandoned risks — a decision-quality win), and washes out when
you're **re-deriving why shipped code works** (the code + comments + contracts already carry it). bd's
specific, repeatable edge in BOTH runs is the same narrow thing: **recall of the "why" that has no code
home** — history, abandoned paths, measured outcomes. Its retrieval tooling did NOT out-cover thorough
grep on either task; a grep tier won #1 both times.

## Complete bd scoreboard (every run, by question type)
Scales differ: runs 1–5 = prose head-to-head bd-vs-one-grep-arm (/20, ±25% judge noise);
runs 6–8 = coverage across 5 access tiers (/10). Read placements, not raw cross-scale numbers.

| Run | Question type | bd score | bd placement | Winner |
|---|---|---|---|---|
| 1 | recall (1 Q) | 20/20 | — | **bd** 20–16 |
| 2 | recall (4 Q) | 18/20 | — | grep 20–18 |
| 3 | recall (4 Q, strict no-devlog) | 19/20 | — | grep 20–19 |
| 4 | recall (4 Q, terse commits) | 20/20 | — | **bd** 20–17 |
| 5 | recall (4 Q, code-only) | 15/20 | — | grep(code) 20–15 |
| 6 | **synthesis** (feasibility) | 8.5/10 | **#2 of 5** | grep-FULL 8.5 |
| 7a | **planning** (pill, touchy) | 8.5/10 | #3 of 5 | grep-FULL 9 |
| 7b | **planning** (palette, new) | 8/10 | **#2 of 5** | grep-FULL 10 |
| 8 | recall (matched protocol) | 7/10 | **#5 of 5** | grep-TERSE 10 |

**Pattern:** on **planning/synthesis** bd sits **upper-middle (#2–#3)** and always uniquely surfaces the
abandoned/historical "why." On **recall** bd is a coin-flip within noise (runs 1–5) and **last on a strict
coverage rubric** (run 8) — because the rubric credits *what breaks*, not the *history* bd alone carries.
A grep tier took #1 on every 5-tier run. bd never took #1 on coverage.

## Important caveat — bd is an INDEX for grep, measured here as its ADVERSARY
This benchmark pitted bd *against* grep (separate arms) to isolate each. **That is not how bd is meant to
be used.** In real use bd is the **index**: `bd devlog search/graph` tells you *where* the answer lives,
then you grep/read *that spot* — instead of brute-forcing a wall of code for the needle. bd + grep are
**partners, not rivals.**

Three consequences the adversarial framing understates:
1. **The Run-8 bd arm already WAS bd+grep** — it used bd to locate, then read the code bd pointed to. It
   got 50ms right AND was the only arm with the tuning history. It ranked #5 **only because the coverage
   rubric didn't score history** — as a *combined* method it was arguably the most complete answer, just
   not the way this rubric counts.
2. **"grep tier beat bd" is a SMALL-REPO result.** The winning grep arms already knew where to look on a
   familiar ~repo. On a large/unfamiliar codebase, blind grep drowns in false hits; **bd-pointed grep is
   the whole point** — aim before you dig. This bench never exercised that scale (a standing caveat since
   Run 3), so it is a FLOOR for bd, not its ceiling.
3. **bd's measured, non-noisy edge = efficiency** (fewer round-trips, faster wall-time in every run) — which
   is exactly the "aim, don't brute-force" property. The coverage checklist doesn't reward it, but it is
   the real day-to-day value: less context burned finding the right file.

**Corrected framing for the article:** don't claim "bd out-answers grep." Claim **"bd aims grep"** —
it finds the right place fast (efficiency) and remembers the why-with-no-code-home (history/abandoned
paths). Coverage of shipped code is a wash because the code carries it; the win is *getting there cheaply*
and *knowing what isn't in the code at all.*

## Method notes
- Fresh arms (no reused bd answers, unlike runs 3–5). Single unprimed judge; judge re-grepped to verify
  every ms value before scoring — no arm singled out.
- Pre-registration (`2026-08-25_run8_recall_checklist_prereg.md`) committed with ground truth verified in
  code first — deliberately avoiding the Run-7 B6 stale-claim error.
- One recall question, one judge — directional, not statistical (same caveat as all runs).

## Daemon hygiene
Baseline `72371` only; `--no-daemon` on the bd arm; zero strays across 5 arms + judge. Restored.
