# bd bench — Run 1 (single question)

> ⚠️ **CAPTION (added 2026-08-25):** the bd arm (Arm A) in runs 1–5 was **bd-mostly-ALONE, not bd+grep** —
> retrieval-led, and it did not reliably drop to code to verify (proof: in Run 5 it reported the throttle
> as "~100ms" from contract 02, never opening `NoteEditorView.vue:1061` to see the real 50ms). That
> misrepresents real bd usage (index → then grep the pointed spot). Do NOT cite runs 1–5 as bd-vs-grep
> evidence. The conclusion-bearing runs are **6 (synthesis) and 8 (recall)**, whose bd arms DID
> index-then-read-code. See `2026-08-25_bd-bench_run8_recall-matched-protocol.md` "bd is an INDEX for grep".

**Date:** 2026-08-25
**bd version:** 0.65.0
**Command:** `bd bench "how did the media edge-handoff avoid the handoff bleep?"` (protocol), run with the user's framing question below.
**Mode:** parallel sub-agents (Arm A / Arm B / blind judge), isolated contexts.

## Question
"What is Fray's media-card drag handoff strategy — HTML5 internal drag vs native OS (Finder) export? How did the team arrive at the current approach, and what earlier approach was tried or used before it? Explain why the HTML5-vs-native distinction mattered and what the handoff depends on."

Chosen because it needs cross-time decision history (2026-04-19 native-drag → 2026-05-21 native-only after `fix/drag-in-out-2-zones` → 2026-08-25 renderer-side handoff), and the wording ("native vs html strategy") does not match the source text (semantic-recall test).

## Results
| Arm | Tokens | Tool-calls | Judge score | Time |
|---|---|---|---|---|
| A — BeadsLog (`bd devlog`) | 21,504 | 3 | 20/20 | 29.3s |
| B — brute grep | 22,885 | 4 | 16/20 | 17.6s |

- Token ratio A/B: **0.94** (BeadsLog ~6% fewer)
- Tool-call ratio A/B: **0.75** (3 vs 4)
- Blind judge (didn't know which was which): **picked BeadsLog** (Set 2). Both verified correct against the devlog; BeadsLog preferred for nailing the two-path architecture (`application/fray-item` internal vs `startDrag` native) and correctly attributing the earlier approach to the *planned main-process 2-zone `GlobalMouseEventAddon` poll* revised to the renderer handoff. Grep framed the predecessor as the ⌥-modifier hack (actually contract 51, superseded) and missed the MIME split.

## Arm A answer (BeadsLog)
Two paths: text/task/link ride HTML5 DnD via `application/fray-item` for internal filing; media export via Electron `startDrag`. Planned a main-process two-zone `GlobalMouseEventAddon` poll but found the handoff can fire from the renderer while the pointer is ~55px inside; HTML5 can't hand off to `startDrag` mid-flight and dies at the window boundary, so media had to be a mouse-driven drag firing `native-drag-start` on `pointermove` before the edge. Seamlessness depends on the ghost matching the OS drag image (size/aspect/cursor-centered) + dropping it on `drag-image-started` + 1 rAF, plus suppressing the inner `<img>`/`<video>` native `dragstart` via preventDefault on mousedown. Source: 2026-08-25 devlog (contracts 51 & 52; commits 0c07670, 3209c8a, 0efea16).

## Arm B answer (grep)
Mouse-driven pointermove drag with a custom ghost handing off to native `startDrag`/`native-drag-start` from the renderer while ~55px inside, swapping the ghost for the Finder image at the edge. Replaced pure HTML5 drag (2026-04-19_native-drag.md) + an ⌥-modifier filing hack (da2d777); HTML5 can't call `startDrag` mid-flight and dies at the window boundary, so it could never file AND export in one gesture. Depends on the mouse-driven engine, cursor vs `.drawer-view` edge rect, suppressing inner `<img>`/`<video>` dragstart. Source: 2026-08-25 devlog (contract 52, cf474b5/3209c8a) + 2026-04-19_native-drag.md.

## Honest read
On this small, freshly commit-mapped corpus, grep is a strong baseline and TIED on raw tokens. BeadsLog's edge appeared where predicted: fewer tool-calls (3 vs 4) and semantic recall of a decision whose wording differed from the source — which is why the blind judge rated its answer higher (20 vs 16). Single question / one judge pass / tiny corpus → directional, not statistical.

## Methodology caveat (added after the fact)
Arm B was allowed to grep **`_rules/_devlog/`** — but the devlogs ARE BeadsLog's
raw memory (`bd devlog record` indexes those exact files). So this run does NOT
measure "BeadsLog vs no-BeadsLog"; it measures the narrower **search-tooling delta**
(`bd devlog search` vs `grep`) while HOLDING the curated devlog corpus constant.
The big BeadsLog value — that the decision history exists in writing at all — was
handed to grep for free, which is why grep stayed competitive. See Run 3 for the
stricter test (grep forbidden from `_rules/_devlog/`).

## Daemon hygiene
Baseline `72371` only; no strays spawned (`--no-daemon` on Arm A; Arm B/judge used no bd). Environment restored.
