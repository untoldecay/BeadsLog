# bd bench — Run 7 (TWO planning tasks: modify-touchy + new-feature, across the access spectrum)

**Date:** 2026-08-25
**Hypothesis under test (user's):** *planning* questions on (1) a NEW feature and (2) modifying a
TOUCHY subsystem "will show even more the value" of durable memory than the recall runs did — sharp
enough to formalize in an article.

**Result: the hypothesis did not hold.** The touchy-subsystem task was NOT the sharpest memory-value
demo — it was the sharpest demonstration that **code is the freshest truth and curated memory drifts.**
grep-FULL (widest code+contract read) won BOTH tasks; bd placed #2–#3; the highest-leverage findings on
both tasks came from thorough CODE reading, not from bd's retrieval; and a THIRD contract-drift surfaced
that only the code-reading arms caught.

## Design
Two planning tasks, each answered by the same 5-arm access spectrum, blind-judged on **pre-registered
10-item coverage checklists** (committed BEFORE any answer existed — see `2026-08-25_run7_checklists_prereg.md`).

- **Task B — MODIFY touchy:** pill → "expression surface" (drop images on pill w/o opening drawer →
  enlarge to mini-UI to add + pre-categorize images into spaces; pill morphs to show toast notifications
  on web-extension capture).
- **Task A — NEW feature:** global command palette / quick-switcher (fuzzy-jump to space/note/action).
- **Arms:** bd (`devlog search/graph/impact` + any file) · grep-FULL (code+contracts+devlogs) ·
  grep-NODEVLOG · grep-TERSE (+contracts, subject-only commits) · grep-CODE-ONLY.

## Coverage results (blind judges, verified contested facts against the repo)

### Task B — pill (touchy)
| Rank | Arm | Cov | Note |
|---|---|---|---|
| 1 | **grep-FULL** | 9/10 | Correct on toast (names both files); **sole finder of the focus-blur paradox** (focusing pill for mini-UI inputs blurs drawer → close sequence fires → drawer closes) — the least-obvious real blocker. |
| 2 | grep-CODE | 8.5 | Correct on toast; verified `pill.js` no-Pinia internals; **only arm to flag the existing drag-to-space feature OVERLAPS the task → reuse (commits da2d777/92f716a/9c5e3fb)** — biggest scope-cut. |
| 3 | **bd** | 8.5 | Correct on toast ("wrong window, no IPC reaches pill"); cited pill-v2 + pill-zorder devlogs for compositor history. Clean, but no unique load-bearing find the code arms lacked. |
| 4 | grep-TERSE | 6.5 | Best compositor/data-model detail (invisible-paint warmup 34:57, "no membership field, untaggable trays") but **WRONG on toast — trusted contract 51's stale "no toast system yet."** |
| 5 | grep-NODEVLOG | 5.5 | Thinnest; **WRONG on toast (net-new).** |

### Task A — palette (new)
| Rank | Arm | Cov | Note |
|---|---|---|---|
| 1 | **grep-FULL** | 10/10 | **Only arm to find the existing reusable fuzzy ranker** `matchCommands` (slashCommands.js:76, contract 43) — everyone else proposed rebuilding it. |
| 2 | **bd** | 8 | Uniquely cited the PAUSED `drawer-search-esc-chain` devlog ("#1 recurring bug class") + space-switch-debounce devlog; best ESC-regression grounding. **Missed matchCommands** (proposed substring). |
| 3 | grep-NODEVLOG | 7 | Clean architecture (Cmd+J free binding, focusStack mirror). Missed matchCommands. |
| 4 | grep-CODE | 6 | Found UniversalPicker.vue reuse + loop-regression commit SHAs; asserted "no fuzzy lib." |
| 5 | grep-TERSE | 4 | Wrong shortcut line (:23), proposed hand-writing the ~15-line scorer **that already exists and is exported.** |

## The three findings

**1. Thoroughness > tooling — again, and decisively.** grep-FULL won both tasks by reading widest.
It alone found the two highest-leverage reuse assets: the existing **toast store** (pill task) and the
existing **fuzzy `matchCommands` ranker** (palette task). **bd MISSED the fuzzy matcher** and proposed a
substring search — the single most consequential coverage miss in the run. Finding matchCommands was a
`grep matchCommands src/` away; it lives in the editor's slash-command layer, so arms that scoped
"fuzzy search" to the search UI (SearchResults/UniversalPicker) never saw it. bd's semantic retrieval
did not surface it either.

**2. Third contract-drift, caught only by code.** Contract 51 says "no toast system yet." The code has
`notificationStore.js` + `NotificationToast.vue` (drawer-scoped Pinia). The arms that **read code**
(FULL, bd, CODE) got it right ("reuse via IPC to the pill"); the arms that **trusted the contract**
(TERSE, NODEVLOG) got it flatly wrong (net-new / doesn't exist). **My own pre-registered checklist B6
made the same drift error** — I trusted the "no toast" framing and had to correct it post-run. This is
the same failure mode as Run 5's 100ms→50ms drift: curated memory is only as fresh as its last edit;
code is ground truth.

**3. The touchy task did NOT separate memory from code.** grep-CODE placed **#2 on the pill** — it found
the store, the no-Pinia gotcha, and the feature-overlap reuse purely from code + commit hashes. The
"modify a touchy subsystem" framing, predicted to *widen* memory's edge, instead let the code arm tie
the top and exposed a memory liability (the stale contract). bd's unique contributions stayed marginal:
pill-v2/pill-zorder history (pill) and the paused ESC-chain devlog (palette) — real color, not decisive
coverage.

## What this means for the article (honest)
- **Do NOT position memory on "modify a touchy subsystem."** On a well-commented codebase the code IS
  the memory for how shipped subsystems work, and stale contracts/devlogs can actively mislead. This run
  is evidence *against* that pitch.
- **The defensible memory pitch stays what Run 6 found:** forward **synthesis/feasibility** decisions,
  specifically surfacing **things with no code home** — abandoned/parked experiments, measured outcomes,
  cross-session narrative, and "we already litigated this risk." That's a decision-quality win.
- **bd's robust, non-noisy edges** remain **efficiency** (fewer round-trips on recall) and **recall of
  the homeless "why."** Its retrieval did NOT beat wide grep on planning-coverage here, and once (the
  fuzzy matcher) it lost a high-leverage reuse asset that a one-line grep found.
- **The sharpest, most trustworthy claim to publish** is the inverse of the original hypothesis: *durable
  memory pays off for deciding what to build next; for understanding what already shipped, read the code —
  and keep your contracts in sync, because a stale invariant is worse than no invariant.*

## Method caveats
- Two tasks, single coverage judge each, one repo — directional not statistical.
- Judges verified the two contested facts (toast store, fuzzy matcher) directly against source before scoring.
- Pre-registered checklist item B6 was itself wrong (trusted a stale contract); corrected in-place with
  strikethrough rather than silently — the error is itself a data point (drift fools the checklist author too).
- Arm outputs were condensed faithfully into the judge prompts; judges re-grepped to confirm load-bearing claims.

## Daemon hygiene
Baseline `72371` only; zero strays across all 10 report arms + 2 judges. Environment restored.
