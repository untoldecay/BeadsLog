# Run 9 — pre-registration: cold-locate + semantic-mismatch, efficiency-scored

**Committed BEFORE the arms run.** Tests the ONE condition every prior run skipped: finding a needle
you don't know the location of, from a question whose words DON'T match the code's identifiers. This is
where bd is *supposed* to win (aim grep). If it doesn't, bd is marginal even at its strength.

## Design
- **6 questions** in plain user/functional language, deliberately avoiding code identifiers.
- **2 arms, both may read code:**
  - **Arm A — bd→grep:** for EACH question, FIRST `bd devlog search/graph` (`--no-daemon`) to locate the
    area, THEN open/grep the pointed code to pin exact file:line.
  - **Arm B — grep-only:** code + `git grep`/read only. NO bd.
- **Primary metric = EFFICIENCY to the correct answer:** tool-calls + tokens + wall (from completion
  metadata). **Secondary = HIT/MISS** vs the ground truth below (right file = hit; +exact line/function = clean hit).
- Neither arm sees this file or the ground truth.

## Questions + GROUND TRUTH (verified in code 2026-08-25)
| # | Question (as given to arms) | Ground-truth answer |
|---|---|---|
| Q1 | When a note has nothing in it, the editor shows faint prompt text — where is the decision that it "looks empty enough" to show that prompt? | `src/components/TsStackWysiwygEditor/index.vue` — `isEmpty`/`domEmpty` (`:57` render, `:880-881` comment). |
| Q2 | Dragging a note to file it into a space, when several spaces could receive it — where is the choice of which single space (or ask the user) decided? | `src/utils/spaceDropCandidates.js` — `resolveSpaceDrop`/`resolveTrayCandidate` (`:10`,`:36`). |
| Q3 | The floating always-on-top tab can change size — where does the code deliberately avoid a resize technique to stop visual corruption on macOS? | `main.js:5619-5620` (also `:2356`) — "Do NOT resize the BrowserWindow… avoids macOS compositor transparency corruption from repeated setBounds." |
| Q4 | Two windows show the same note — where does the code stop an older/slower update from clobbering newer typed text? | `src/stores/notes.js:653-654` — monotonic `updatedAt`/`timestamp` guard (`if (!timestamp \|\| timestamp >= localTimestamp)`). |
| Q5 | The blurred/frosted background can be told to stop working temporarily to save resources — where is that switch? | `src/glass/strategies/MicroCaptureStrategy.js:337` `forcePause()` (state `:36`, honored `:127`). |
| Q6 | When a note's content becomes the actual file on disk — where does that conversion/write happen? | `src/services/libraryService.js:82` `updateNote` → DB + FS (`:371` `writeFileSync`; markdown folder = SSOT). |

## Scoring
- HIT = correct file. CLEAN HIT = correct file + right function/line. MISS = wrong file or none.
- Efficiency read from each arm's completion metadata (tool_uses, subagent_tokens, duration_ms).
- I score HIT/MISS myself against this fixed table (objective: right file or not), transparently.

## Prediction (stated before results)
Arm A (bd→grep) reaches the right file in FEWER tool-calls on the questions whose concept is in the
devlog/graph index; on questions with no memory trace (e.g. Q5 glass forcePause, Q1 placeholder) bd
search returns little and A falls back to grep — parity there. Arm B (grep-only) burns more calls on the
semantic-mismatch questions (query words ≠ code terms) and risks a MISS on the ones buried far from any
matching string. **If A is NOT more efficient / higher-hit, bd is marginal even here → pick direction A
(use-cases), not B (build power).**
