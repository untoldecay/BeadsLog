# A/B Test — Protocol Footprint: Modified Lean (Round 2)

**Date:** 2026-08-09
**Change from Round 1:** Added one line to the lean protocol —
`Context: _rules/_orchestration/PROJECT_CONTEXT.md (PRD, tech stack, conventions)
— consult before planning` — to recover the PROJECT_CONTEXT loading that Round 1
lean agents skipped. PROJECT_CONTEXT.md also offered as an on-demand reference.

## Method
Same as Round 1: 4 fresh subagents, identical login-timeout task, planning
compliance scored on the 6-pt beads-loop rubric.

## Results

| Trial | Loop score | Loaded WORKING_PROTOCOL | Loaded PROJECT_CONTEXT |
|---|---|---|---|
| T1 | 6/6 | ✅ | ❌ |
| T2 | 6/6 | ✅ | ❌ |
| T3 | 6/6 | ✅ | ✅ |
| T4 | 6/6 | ✅ | ✅ |
| **Mean** | **6.0 / 6** | 4/4 | **2/4** |

## Findings
- **Loop compliance unchanged: 6.0/6** — modified lean still matches full on
  resume→map→claim→devlog→close→sync. The extra line cost nothing here.
- **PROJECT_CONTEXT recovery is partial: 0/4 (R1) → 2/4 (R2).** A single bullet
  pointer helps but does NOT reliably induce context loading — on a concrete
  bug-fix task, half the agents still deprioritized it. Full protocol got 4/4
  because it leads with "Context:" prominently AND repeats retrieval-first
  discipline throughout.
- Interpretation: either (a) make the pointer more imperative (a numbered
  MANDATORY step, not a bullet), or (b) accept that PROJECT_CONTEXT loading
  isn't load-bearing for well-scoped tasks and let the agent pull it only when
  the task is architectural/ambiguous.

## Token cost
Modified lean ≈ 155 + 15 = **~170 tokens** vs full **~738** → still ~4.3× smaller.

## Verdict so far (Rounds 1–2)
- Lean+deferred preserves the core workflow loop at ~1/4 the token cost. Strong.
- The one soft loss (proactive PROJECT_CONTEXT loading) is partially fixable and
  arguably not critical.
- Confidence limits unchanged: n=4, single task, single model, PLANNING not
  execution. Round 3 (long-context pressure) is the real stress test.

## Next — Round 3 (proposed)
Seed subagents with a cleaned long-run session export (~30–60k tokens of prior
work) BEFORE the protocol + task, to test whether the lean protocol survives
context pressure — the real production condition where a short protocol competes
with a large working context. See devlog for the sanitization plan.
