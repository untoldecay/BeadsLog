# bd bench — Run 9 (cold-locate + semantic-mismatch, efficiency-scored)

**Date:** 2026-08-25
**Why:** the one condition every prior run skipped — finding a needle you don't know the location of,
from a question whose words DON'T match the code's identifiers. This is where bd is *supposed* to win
("aim grep"). Two arms, both may read code: **A = bd→grep** (bd locates, then read the pointed code) vs
**B = grep-only** (no bd, no `_rules/`). 6 questions in plain user language. Primary metric = EFFICIENCY
to the correct file; secondary = hit/miss vs pre-registered ground truth (`2026-08-25_run9_coldlocate_prereg.md`).

## Efficiency (from completion metadata)
| Arm | Tool-calls | Tokens | Wall |
|---|---|---|---|
| A — bd→grep | 24 | 39,462 | 97s |
| **B — grep-only** | **18** | **38,755** | **75s** |

**grep-only was MORE efficient** — fewer calls, fewer tokens, faster. bd's index step ADDED a round-trip
per question that grep didn't need, because on a small familiar repo grep finds things fast anyway.

## Hit / miss (vs ground truth; 2 corrections noted)
| Q (concept) | Arm A bd→grep | Arm B grep-only |
|---|---|---|
| Q1 placeholder "empty enough" | ✓ `index.vue:887 isEmpty` | ✓ `index.vue:887 isEmpty` |
| Q2 drop-to-space resolution | ✓ `ViewBar.vue:271` + `spaceDropCandidates.js` | ✓ `spaceDropCandidates.js:36` |
| Q3 pill resize compositor | ✓ `main.js:5620` | ✓ `main.js:5619` |
| Q4 stale/older update guard | ~ found *focus* re-render guard (`NoteEditorView.vue:954`), NOT the age guard | ✓ **exact** age guard `notes.js:654` |
| Q5 blur "stop to save resources" | ✓ **`MicroCaptureStrategy.js:320 pause`** (glass engine) | ✗ **MISS** — `main.js:2384` vibrancy (wrong mechanism) |
| Q6 content→disk write | ✓ `libraryFS.js:17 writeNote` | ✓ `libraryFS.js:17 writeNote` |

**Score: 5/6 each, different one wrong.** A missed the *specific* age guard on Q4; B missed the glass
pause switch on Q5.

Ground-truth corrections (both arms beat my pre-reg): **Q6** real write is `libraryFS.js:writeNote`
(`content.md`), not the `libraryService` wrapper I pre-registered. **Q5** the switch is the strategy's
`pause`/`paused` (enforced at `:127`); `forcePause` is a sibling of the same mechanism — either counts.

## The finding — the two claims, tested at last
1. **Efficiency ("bd aims grep so you brute-force less"): NOT demonstrated here — grep-only won.** On a
   small familiar codebase grep already locates fast, so bd's extra index round-trip is pure overhead
   (24 vs 18 calls). **This is the predicted FLOOR:** bd's efficiency edge needs SCALE / unfamiliarity —
   a corpus where blind grep drowns in false hits — and this repo never provides it. So the efficiency
   pitch remains **unproven, not disproven**; it simply can't be shown at this size.
2. **Semantic recall (query words ≠ code words): REAL but narrow — bd won exactly 1 of 6 (Q5).** When the
   user's vocabulary ("blurred/frosted background… stop to save resources") didn't match any code
   identifier, grep's literal search ("vibrancy/blur/disable") confidently pinned the WRONG mechanism,
   while bd's concept index ("disable backdrop blur performance") mapped to the right glass strategy. That
   is bd's genuine, repeatable edge — and it showed up on precisely the question built to trigger it.
3. **Everywhere the query words DID hint the code (Q1–Q3, Q6): dead tie.** grep matched the terms directly;
   bd's index step added nothing but latency.

## Strategic verdict (answers "verify our readings")
- **Confirms the floor:** on this repo bd is NOT faster than grep — it's slightly slower. Any "bd is more
  efficient / more powerful than grep" claim CANNOT be made from this codebase; it requires a large-repo
  test we have not run. Do not publish an efficiency-beats-grep claim.
- **bd's demonstrable edges stay the two narrow ones:** (a) **semantic recall** on vocabulary-mismatch
  queries (Q5 here; ~1 in 6), and (b) the **why-with-no-code-home** (abandoned paths / tuning history,
  from Runs 4/6/8). Both real, both modest.
- **So: direction A (position on use-cases + memory, NOT raw speed-vs-grep) is the honest pitch.** The
  "make it more powerful/instant" case (direction B) is only worth building if you first prove the
  efficiency edge on a LARGE codebase — the single most valuable un-run experiment left.

## Method notes
- Efficiency read directly from subagent completion metadata (tool_uses/tokens/duration) — no judge needed.
- Hit/miss scored by author against the fixed pre-reg table; 2 ground-truth items were corrected in the
  arms' favor (disclosed above) — the arms were more precise than the pre-registration, not less.
- Q4/Q5 each have mild question-design ambiguity (two valid anti-clobber guards; two blur mechanisms);
  noted rather than hidden. One question-pair, one repo — directional.

## Daemon hygiene
Baseline `72371` only; `--no-daemon` on Arm A; zero strays. Restored.
