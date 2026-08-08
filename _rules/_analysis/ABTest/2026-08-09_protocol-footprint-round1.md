# A/B Test — Protocol Footprint: Full vs Lean+Deferred (Round 1)

**Date:** 2026-08-09
**Question:** Can we shrink the always-on protocol (~738 tokens) to a lean form
(~155 tokens) that defers detail to `WORKING_PROTOCOL.md`, without losing
beads-workflow compliance?

## Method
- 2 arms × 4 trials = 8 fresh subagents (blank slate — no CLAUDE.md, no hooks).
- Identical task: *"users report the login endpoint times out intermittently
  under load — please fix it."* Output = ordered command plan, first action →
  session close. Planning-compliance only (no execution).
- **Arm A (Full):** full `bd prime --full` protocol injected (~738 tok).
- **Arm B (Lean):** `bd prime --mcp` lean protocol (~155 tok) +
  `WORKING_PROTOCOL.md` provided as an ON-DEMAND reference (progressive
  disclosure).
- Rubric (6 pts): RESUME · MAP-before-code · CLAIM · DEVLOG-record ·
  CLOSE-with-sync · no-code-read-before-retrieval.

## Results

| Criterion | Full (~738 tok) | Lean+deferred (~155 tok) |
|---|---|---|
| RESUME | 4/4 | 4/4 |
| MAP before code | 4/4 | 4/4 |
| CLAIM (ready + update) | 4/4 | 4/4 |
| DEVLOG record | 4/4 | 4/4 |
| CLOSE with sync | 4/4 | 4/4 |
| No code before retrieval | 4/4 | 4/4 |
| **Mean** | **6.0 / 6** | **6.0 / 6** |

**Lean matched full on every criterion.**

## Findings
1. **Progressive disclosure works and is robust.** 3/4 lean agents loaded
   `WORKING_PROTOCOL.md` on demand. The 4th loaded NOTHING and still produced a
   fully compliant loop — the lean core alone was sufficient.
2. **One regression — PROJECT_CONTEXT.md.** 4/4 full agents loaded it (the full
   protocol names it: "Context: …PROJECT_CONTEXT.md"); 0/4 lean agents did (the
   lean protocol omits the pointer). Fix: add one `Context:` line to lean
   (~15 tok, still ~4× smaller). → tested in Round 2.

## Token math
- Resident footprint: 738 → ~170 tok (lean + context pointer) ≈ **4–5× smaller**
  at every injection point (session start + each compaction).
- Separately: SessionStart currently DOUBLE-loads — the CLAUDE.md protocol block
  (harness-loaded) plus `prime --hook` re-emitting the same ~738 tok. Deduping
  that saves ~700 tok/session regardless of the lean change.

## Caveats
- n=4, single task, single model, PLANNING-compliance (not execution over a long
  real session). Strong directional signal, not production statistics.
- Clean bug-fix task maps neatly onto the loop; ambiguous/multi-step-epic tasks
  may separate the arms more.
- Some agents hallucinated file paths (e.g. "loaded CLAUDE.md" with no CLAUDE.md
  present) — noise, not signal.

## Raw per-trial (loaded files / score)
- Full F1–F4: 6,6,6,6. All loaded PROJECT_CONTEXT.md (+ some CLAUDE.md).
- Lean L1–L4: 6,6,6,6. Loaded WORKING_PROTOCOL.md: L1,L2,L3 yes; L4 none.
  None loaded PROJECT_CONTEXT.md.

## Next
- Round 2: lean + `Context:` pointer — does PROJECT_CONTEXT loading recover?
- Round 3 (proposed): seed subagents with a cleaned long-run session export to
  test compliance under realistic context pressure.
