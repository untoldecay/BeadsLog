# [research] A/B test: lean vs full protocol footprint (BeadsLog-q9o follow-up)

**Date:** 2026-08-09

_No architectural changes_

## Problem
The always-on beads protocol is ~738 tokens and is injected at SessionStart +
each PreCompact (not per-message — the per-message hook is `mail check`, not
prime). It's also double-loaded (CLAUDE.md block + `prime --hook`). Question:
can a lean ~155-token protocol that defers detail to WORKING_PROTOCOL.md hold
workflow compliance, to cut the resident footprint?

## Work Done (subagent A/B test)
- Built a controlled A/B: fresh subagents (no CLAUDE.md/hooks), identical
  login-timeout task, output = ordered command plan, scored on a 6-pt beads-loop
  rubric (resume/map/claim/devlog/close-with-sync/no-code-first).
- **Round 1** (2 arms × 4): full (~738 tok) vs lean+deferred (~155 tok). Both
  scored 6.0/6. Lean's only gap: 0/4 loaded PROJECT_CONTEXT.md (full: 4/4).
- **Round 2** (4 trials): lean + a `Context: PROJECT_CONTEXT.md` pointer line.
  Compliance still 6.0/6; PROJECT_CONTEXT loading recovered to 2/4 (partial).
- Recorded both rounds under `_rules/_analysis/ABTest/`.

## Findings
- Lean+deferred matches full on the core workflow loop at ~1/4 the tokens.
- Progressive disclosure works: 3/4 (R1) lean agents loaded WORKING_PROTOCOL.md
  on demand; one complied without loading anything.
- The single soft regression (proactive PROJECT_CONTEXT loading) is only
  partially fixed by a one-line pointer — needs a stronger cue or is non-critical.
- Free win independent of all this: dedupe the SessionStart double-injection
  (~700 tok/session).

## Caveats
- n=4, single task, single model, planning-compliance not execution. Directional.
- Round 3 planned: seed subagents with a sanitized long-run session export to
  test the lean protocol under realistic context pressure (the real failure
  mode). Sanitization: strip tokens/secrets/absolute paths from a transcript, or
  reuse the already-safe committed devlog history as the long-context proxy.

## Final Session Summary
**Final Status:** Rounds 1–2 done and recorded; lean approach validated for
planning compliance. Round 3 (long-context) proposed, not yet run.
**Key Learnings:**
- Protocol injection is per-session + per-compaction, NOT per-message — reframes
  the whole optimization (the original "burns tokens every message" premise was
  wrong).
- A short imperative core + on-demand references preserves compliance; prominence
  matters for optional context loading (a buried bullet gets skipped half the time).

## Round 3 addendum (long-context pressure)
Seeded subagents with a ~30k-token sanitized prior-session context, then the
protocol variant + a context-switch task. 2 arms × 3 trials. BOTH arms 6.0/6 —
no degradation. The lean protocol's predicted weak point (surviving context
pressure) showed no weakness. Consolidated verdict + recommendations in
_rules/_analysis/ABTest/SUMMARY.md. Top rec: dedupe the SessionStart
double-injection (~700 tok/session, free), then live-A/B a lean .beads/PRIME.md.
