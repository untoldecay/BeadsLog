# A/B Test — Protocol Footprint under Long-Context Pressure (Round 3)

**Date:** 2026-08-09
**Question:** Does the lean protocol survive when buried under a large prior
context — the real production condition where a short protocol competes with a
big working context and might get "forgotten"?

## Method
- Seeded each subagent with a ~30k-token **prior session context** (sanitized,
  assembled from this repo's committed devlog history + analysis — absolute
  paths stripped; the one secret-pattern hit was a false positive: "di**sk-**heavy").
- Agent reads the context file first, THEN receives the protocol variant + a
  NEW task (login-endpoint timeout — a deliberate context switch from the bd
  internals in the prior context).
- 2 arms × 3 trials = 6 subagents. Same 6-pt beads-loop rubric.

## Results

| Arm | Trials | Mean score |
|---|---|---|
| Full (~738 tok) | 3 | **6.0 / 6** |
| Lean+deferred (~170 tok) | 3 | **6.0 / 6** |

Every trial, both arms: resume → map (graph/search/impact) → claim (ready +
update) → devlog record → close with bd sync + bd devlog sync → push. Code
reading/editing only after retrieval. **Zero degradation from context pressure.**

## Qualitative observations under pressure
- All 6 agents correctly recognized the **context switch** (prior work = eval
  harness / bd internals; new task = login) and re-applied the protocol fresh —
  no context bleed, no drift into the old task.
- Several agents reasoned BETTER given the rich prior context: unprompted
  ponytail-style notes ("fix the shared function all callers route through, not
  the reported path"), conditional steps, branch isolation. The committed devlog
  history carried repo conventions into their plans — a positive side effect of
  BeadsLog's own memory being in context.
- File-loading metric is noisy here (the deferred refs were inlined in the lean
  prompt, so "FILES_LOADED: none" agents still used the info). Not a clean
  measure in this setup; the loop-compliance rubric is the reliable signal.

## Verdict
The lean protocol's predicted weak point — surviving context pressure — showed
**no weakness**. Lean = full at 6.0/6 across all three rounds (planning
compliance), including under 30k tokens of competing context. Strong green light
to adopt the lean footprint.

## Caveats (unchanged)
- Planning-compliance, not live execution over a multi-turn real session.
- Single model, single task family, n small (3–4 per arm per round).
- A truly adversarial context (one actively pulling toward a different workflow)
  wasn't tested; the prior context here was benign bd history.
