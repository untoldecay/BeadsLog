# [research] Protocol footprint A/B — Round 3 (long-context) + verdict

**Date:** 2026-08-09

_No architectural changes_

## Problem
Rounds 1–2 showed lean protocol (~170 tok) matches full (~738 tok) on the
beads-loop rubric on a clean slate. Open question: does lean survive under
realistic context pressure, where a short protocol competes with a large prior
context and might be "forgotten"?

## Work Done
- Assembled a ~30k-token sanitized prior-session context from this repo's
  committed devlog history + analysis (absolute paths stripped; lone
  secret-pattern hit was a false positive in "disk-heavy").
- 2 arms × 3 trials: each subagent reads the context, then gets the protocol
  variant + a context-switch task (login timeout).
- Both arms scored **6.0/6** — no degradation under pressure.
- Wrote Round 3 results + a consolidated SUMMARY.md with recommendations under
  `_rules/_analysis/ABTest/`.

## Findings
- Lean = full at 6.0/6 across all 3 rounds, including 30k-token pressure. The
  predicted weak point didn't materialize.
- Bonus: rich prior context (bd's own devlogs) improved plan quality — agents
  spontaneously reasoned root-cause-at-shared-function, conditional steps.
- Top recommendation is a FREE win independent of the lean question: dedupe the
  SessionStart double-injection (CLAUDE.md block + prime --hook both emit the
  same ~738 tok) → keep full prime only on PreCompact. ~700 tok/session saved.
- Next after that: ship a lean `.beads/PRIME.md` override and A/B it live.

## Final Session Summary
**Final Status:** A/B experiment complete (rounds 1–3). Lean footprint validated
for planning compliance. No code changes yet — recommendations recorded for a
decision.
**Key Learnings:**
- The lean protocol's resilience under context pressure was the crux, and it
  held — the optimization is safe to pursue.
- Biggest immediate saving isn't the lean rewrite at all; it's removing the
  redundant SessionStart re-injection.
