# Protocol Footprint A/B — Summary & Recommendations

**Dates:** 2026-08-09 · Rounds 1–3 · subagent A/B (fresh agents, no CLAUDE.md/hooks)

## Question
Can the always-on beads protocol shrink from ~738 tokens to a lean ~155–170
tokens (deferring detail to WORKING_PROTOCOL.md) without losing workflow
compliance — to cut the resident/injection footprint?

## Result across 3 rounds (6-pt beads-loop rubric)

| Round | Condition | Full | Lean |
|---|---|---|---|
| 1 | clean slate | 6.0/6 | 6.0/6 |
| 2 | lean + Context pointer | — | 6.0/6 |
| 3 | ~30k-token prior-context pressure | 6.0/6 | 6.0/6 |

**Lean matched full on every round, every criterion, including under context
pressure.** Progressive disclosure worked (agents loaded WORKING_PROTOCOL.md on
demand; some complied even without loading it).

## The one soft spot
Proactive PROJECT_CONTEXT.md loading: full 4/4 → lean 0/4 (R1), recovered to
2/4 with a one-line `Context:` bullet (R2). A bullet isn't enough for reliable
context-loading; if that matters, make it a numbered MANDATORY step.

## Key reframe
Protocol injection is **per-session + per-compaction, NOT per-message** (the
per-message hook is `mail check`, not prime). The original "burns tokens every
message" premise was wrong — but there's still real waste (below).

## Concrete recommendations (in priority order)

1. **Dedupe the SessionStart double-injection — free win, do first.**
   Today SessionStart both (a) loads the CLAUDE.md `<beads_protocol>` block (the
   harness always loads CLAUDE.md) AND (b) runs `prime --hook` re-emitting the
   same ~738 tokens. Change the SessionStart hook to inject the **lean** form
   (or nothing, since CLAUDE.md already carries the protocol), and keep **full**
   `prime` only on **PreCompact** (where it restores the protocol after
   truncation). Saves ~700 tokens/session, zero behavior change.

2. **Ship a lean `.beads/PRIME.md` override — reversible, no code.**
   Lean core + a stronger (numbered, not bulleted) PROJECT_CONTEXT pointer.
   Run it on real sessions and watch for drift. This is the zero-risk way to
   A/B in production.

3. **Only if live sessions stay compliant:** make lean the default prime output
   and shrink the CLAUDE.md `<beads_protocol>` block to match — landing the
   full ~4–5× footprint reduction everywhere.

## Confidence & limits
Strong directional signal (3 rounds, consistent 6.0/6), but: planning-compliance
not live execution; single model; small n; benign (non-adversarial) context.
Recommendation #2 (live PRIME.md A/B) is the bridge from "directionally proven"
to "production-confident."
