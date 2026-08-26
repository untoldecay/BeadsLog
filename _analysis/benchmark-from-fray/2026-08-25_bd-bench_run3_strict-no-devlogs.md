# bd bench — Run 3 (STRICT: grep forbidden from the devlogs)

> ⚠️ **CAPTION (added 2026-08-25):** the bd arm (Arm A) in runs 1–5 was **bd-mostly-ALONE, not bd+grep** —
> retrieval-led, and it did not reliably drop to code to verify (proof: in Run 5 it reported the throttle
> as "~100ms" from contract 02, never opening `NoteEditorView.vue:1061` to see the real 50ms). That
> misrepresents real bd usage (index → then grep the pointed spot). Do NOT cite runs 1–5 as bd-vs-grep
> evidence. The conclusion-bearing runs are **6 (synthesis) and 8 (recall)**, whose bd arms DID
> index-then-read-code. See `2026-08-25_bd-bench_run8_recall-matched-protocol.md` "bd is an INDEX for grep".

**Date:** 2026-08-25
**bd version:** 0.65.0
**Purpose:** remove the Run 1–2 confound. Arm B may NOT read `_rules/_devlog/` (BeadsLog's
raw memory) — only **code + contracts + `git log`/`git show`**. Same 4 "why" questions as Run 2.
Arm A (bd) is Run 2's answers reused (its method/corpus is unchanged), judged blind vs the new strict arm.

## Questions
Same four as Run 2 (Phase-2 park, typing stutter, uncapped pool, two-pipe sync).

## Results
| Arm | Tokens | Tool-uses | Wall time | Judge score (blind) |
|---|---|---|---|---|
| A — BeadsLog (`bd devlog`) | 32,452 | 10 | 45s | 19/20 |
| B — **strict grep** (no devlogs) | 31,022 | 8 | **173s** | **20/20** |

- Blind judge: **strict grep won 20 vs 19.** Both factually clean, no errors. Strict grep edged it on **verifiability/actionability** — it anchored every claim to a specific **commit hash** (adc13bf, a1abe6a, da70093, 9fc1e32, c9f137b — all verified) and an exact code line (`DrawerPanelView.vue:1113-1119`), plus an extra correct fact (image-caption char-drop fix). bd's answer read cleaner but leaned on looser citations (approx. addendum number, contract 49) and packed fewer load-bearing anchors — costing it one actionability point.

## The finding (this overturns my prediction)
I predicted bd would win the strict test *decisively* because "the why isn't in the code." **Wrong.** The strict arm reconstructed all four whys — and slightly beat bd — because on THIS repo the decision history is captured **redundantly** across:
- **rich commit messages** (I write detailed ones: the park reason, the numbers, the mechanism),
- **code comments** (the uncapped-pool rationale is a comment at `DrawerPanelView.vue:1113`),
- **contracts** (the two-pipe model is contract 02; glass forcePause is contract 10).

So the dedicated devlog layer added **little marginal quality** for these questions — the same reasoning already lives in git + code + contracts.

**Cost side:** strict grep paid ~**4× wall-time** (173s vs 45s) — mining `git log`/`git show` + code is heavy per call — for comparable tokens/tool-calls. bd was faster and no less correct.

## Honest combined verdict (runs 1–3)
| Run | grep devlog access | Quality winner | bd efficiency |
|---|---|---|---|
| 1 (1 Q) | yes | **bd** 20–16 | leaner (0.94 tok, 3 vs 4 calls) |
| 2 (4 Q) | yes | grep 20–18 | leaner (0.93 tok, 10 vs 15 calls) |
| 3 (4 Q) | **no** | strict grep 20–19 | faster (45s vs 173s), ~= tokens/calls |

- **Answer quality: a wash** (bd 1, grep 2) — even DENIED the devlogs, grep matched bd by mining git + code + contracts.
- **bd's consistent edge here = efficiency/speed** (fewer round-trips in 1–2; ~4× faster wall-time in 3), NOT unique recall.
- **Why:** this repo has unusually disciplined commits/comments/contracts, so the "why" is redundant across artifacts — no retrieval method has a decisive quality edge. A repo WITHOUT that discipline would favor the devlog memory much more.
- **Not exercised (would favor bd):** large corpora, cross-session recall over months, semantic-mismatch queries at scale, lifecycle badges, impact/graph, shared team memory. This bench only probed small, greppable "why" questions on a redundantly-documented repo — a fair floor for bd, not its ceiling.

## Caveat on Run 3's design
Arm A was Run 2's bd answers reused (method/corpus identical), not a fresh run — acceptable since the variable under test was Arm B's devlog access. Single judge pass, 4 questions — directional.

## Daemon hygiene
Baseline `72371` only; no strays (strict grep + judge used no bd). Environment restored.
