# Run 8 — pre-registered coverage checklist (recall question, settled protocol)

**Committed BEFORE any arm answer exists.** This is the RECALL counterpart to Run 6 (synthesis),
under the identical Run 6/7 protocol: 5 access tiers, fresh arms, unprimed blind judge, coverage
checklist + fact-verification. Purpose: give the recall finding the same rigor as synthesis so the
matched pair (Run 6 recall-vs-synthesis... — actually Run 6 = synthesis, Run 8 = recall) is method-uniform.

## Question (a genuine "why did shipped code end up this way")
*"Why is the drawer editor's live typing sync split into two pipes, and why is each timed the way it is
(throttle vs debounce)? What breaks if you get it wrong?"*

Chosen because its answer naturally hits the **100ms-vs-50ms contract drift**, letting an UNPRIMED
judge re-catch it via fact-check (not because I told it to). The "why" is redundant across code
comments (the Dual-Pipe block in `NoteEditorView.vue`), contract 02, and the typing-perf devlogs.

## Ground truth (VERIFIED in code before pre-registering — the B6 lesson)
- Code `src/views/NoteEditorView.vue:1061`: Pipe 1 = `throttle(..., 50)` — **50ms**.
- Code `NoteEditorView.vue`: Pipe 2 = `debounce(..., 800)` — **800ms**.
- Contract `02_live_sync_durability.md:15`: Pipe 1 "~100ms" — **STALE vs code (50ms).**
- Contract `:16`: Pipe 2 "~800ms" — matches code.

## Checklist (10 items; each PRESENT / PARTIAL / MISSED)
- **C1** TWO pipes: Pipe 1 = throttled *stream* (transient, sibling broadcast, NO DB/FS); Pipe 2 = debounced *persistence* (formal DB/FS save).
- **C2** Pipe 1 value = **50ms in code** (contract says ~100ms → DRIFT; code is truth). *[fact-verify the number]*
- **C3** Pipe 2 value = **800ms** (code + contract agree).
- **C4** WHY split: live cross-window typing must be fast + cheap (no disk/keystroke); durable save can be lazy/debounced — a latency-vs-durability trade.
- **C5** Transient updates MUST bypass DB/FS in the main process (no disk write per keystroke).
- **C6** Self-transient hazard: a transient MUST NOT mutate the ORIGINATING window's reactive note; self-broadcast = no-op, only CROSS-window re-mints identity (`sourceWindowId !== currentWindowId`).
- **C7** WHY C6 matters: per-keystroke re-render of the covered grid / every mounted keep-alive space = the typing-stutter bug this design prevents.
- **C8** Title: optimistic title flows THROUGH the two pipes (extracted in the content-update handler); a separate per-keystroke title write to the store is FORBIDDEN (re-renders every note card).
- **C9** Correctness invariant: every edit MUST eventually persist (transient ≠ sole source of truth); siblings still show live typing; acceptance = chars in siblings within 150ms.
- **C10** Heavy metadata extraction (tags/tasks) MUST NOT run during the throttled stream pipe (keep Pipe 1 cheap).

## Arms (5 access tiers, same question)
- **bd** — retrieval-led: `bd devlog search/graph/impact` first (`--no-daemon`), may open files it points to.
- **grep-FULL** — code + contracts + devlogs, grep/read anything.
- **grep-NODEVLOG** — code + contracts; MUST NOT read `_rules/_devlog/`.
- **grep-TERSE** — code + contracts + subject-only commit log; no devlogs, no rich commit bodies.
- **grep-CODE-ONLY** — code + comments + terse commits; NO contracts, NO devlogs.

## Judge
One blind judge, reports shuffled/anonymized, scores all 5 vs the checklist (present/partial/missed +
coverage n/10 + rank). **Unprimed instruction:** "verify EVERY numeric claim (throttle/debounce ms)
against the code" — not singling out any arm. Report which arms drifted on C2.

## Prediction (stated before results)
Recall stays a wash or slight code-edge; fact-verification dings whichever arms trusted contract 02's
~100ms (expected: bd + the contract-having grep tiers) while code-only/full-that-read-code get 50ms right.
Confirms — does not overturn — the standing conclusion.
