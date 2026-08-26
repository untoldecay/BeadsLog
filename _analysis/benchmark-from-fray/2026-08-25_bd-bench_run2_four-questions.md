# bd bench — Run 2 (four questions)

> ⚠️ **CAPTION (added 2026-08-25):** the bd arm (Arm A) in runs 1–5 was **bd-mostly-ALONE, not bd+grep** —
> retrieval-led, and it did not reliably drop to code to verify (proof: in Run 5 it reported the throttle
> as "~100ms" from contract 02, never opening `NoteEditorView.vue:1061` to see the real 50ms). That
> misrepresents real bd usage (index → then grep the pointed spot). Do NOT cite runs 1–5 as bd-vs-grep
> evidence. The conclusion-bearing runs are **6 (synthesis) and 8 (recall)**, whose bd arms DID
> index-then-read-code. See `2026-08-25_bd-bench_run8_recall-matched-protocol.md` "bd is an INDEX for grep".

**Date:** 2026-08-25
**bd version:** 0.65.0
**Mode:** parallel sub-agents (Arm A / Arm B / blind judge), isolated contexts. Each arm answered all 4 in one context.

## Questions (cross-session, decision-heavy)
1. Why was space-switch "Phase 2" (content-follows-cursor) PARKED, and what single fix shipped instead?
2. What caused the drawer note-editor TYPING STUTTER, and what were the fixes?
3. Why is the keep-alive space pool effectively UNCAPPED now, and what earlier constraint did that depend on?
4. What is the live-sync "two-pipe" model, and why must transient updates avoid heavy metadata extraction?

## Results
| Arm | Tokens | Tool-uses | Judge score (blind) |
|---|---|---|---|
| A — BeadsLog (`bd devlog`) | 32,452 | **10** | 18/20 |
| B — brute grep | 34,766 | 15 | **20/20** |

- Token ratio A/B: **0.93** (BeadsLog ~7% fewer)
- Tool-call ratio A/B: **0.67** (10 vs 15 — bigger gap than run 1's 3v4)
- Blind judge (didn't know which was which): both factually clean, **no errors** in either set. Picked **grep (Set 2)** for precision — it named specific code sites (`stores/views.js`, `pendingViewId`, the `traySlotWidths` singleton), cited contract §sections, and pulled in a second corroborating devlog (`2026-04-19_cross-window-typing-latency.md`, ~300ms / 10×-per-sec figures) that BeadsLog only summarized abstractly.

## Honest read
**A split, and the confound matters:** grep produced the higher-quality answer — but it spent **15 tool-calls (50% more)** to read more files and surface that extra detail. BeadsLog reached a comparable, fully-correct answer with **fewer round-trips and fewer tokens**. So the measured story is: on this small, freshly-indexed corpus, **bd's edge is efficiency (fewer tool-calls/tokens), not answer quality** — grep, given more calls, matches or exceeds it. This is exactly the protocol's stated expectation (don't fabricate a bd win; ties/losses on quality are valid on a greppable corpus; bd's repeatable edge is round-trips, growing with scale + semantic-mismatch wording + tasks grep can't do).

## Combined verdict (run 1 + run 2)
- **Efficiency:** BeadsLog won both runs on tokens (0.94, 0.93) and tool-calls (0.75, 0.67); the tool-call gap widened with more questions.
- **Quality (blind judge):** 1–1 — bd won run 1 (semantic recall of a decision whose wording differed from source), grep won run 2 (more file reads → more precise code-site detail).
- **Takeaway:** on THIS corpus (small, well-keyworded, fresh commit-mapped devlogs) the two tie on answer quality; bd's durable advantage is fewer round-trips. bd's decisive wins (cross-session persistence, lifecycle badges, impact/graph, large corpora) weren't exercised by these greppable questions.

## Methodology caveat (added after the fact)
Arm B grepped **`_rules/_devlog/`** — which IS BeadsLog's raw memory. So runs 1–2
measure the **search-tooling delta** (bd search vs grep) with the curated devlog
corpus held constant, NOT the value of keeping devlogs at all. Grep's win here is
riding on the BeadsLog artifact (someone wrote the "why" into a devlog). Run 3
removes that confound: strict Arm B may read code + contracts + `git log` only, NOT
`_rules/_devlog/`.

## Daemon hygiene
Baseline `72371` only; no strays across either run (`--no-daemon` on Arm A; Arm B/judge used no bd). Environment restored both times.
