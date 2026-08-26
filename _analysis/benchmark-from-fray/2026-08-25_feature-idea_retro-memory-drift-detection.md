# Feature idea for BeadsLog — orphan-code + existence-claim drift detection

**Date:** 2026-08-25
**Status:** IDEA / not started. Captured from a `bd bench` finding for a future agent to pick up.
**For the next agent:** this is a proposed BeadsLog capability, born from a real incident in this repo.
Read the incident first (it's the justification), then the proposal. The retro devlog it references is
`_rules/_devlog/2026-08-25_retro-dormant-toast-store.md`.

## The incident that motivates this
During Run 7 of the memory benchmark (`2026-08-25_bd-bench_run7_two-planning-tasks.md`), 2 of 5 agents —
**and the human's own pre-registered checklist** — wrongly concluded "no toast system exists."
- Reality: `src/stores/notificationStore.js` + `src/components/NotificationToast.vue` shipped
  2025-06-07 (commit `400a233`), wired into `Bottombar.vue`/`TopBar.vue`.
- It was landed inside an **unrelated `fix:` commit** with **no devlog and no lifecycle marker**.
- A contract still asserted **"no toast system yet."**
- So: retrieval-led agents couldn't see the code (nothing recorded it), and agents that trusted the
  contract were **actively misled**. The third such contract-drift found in this benchmark
  (cf. the 100ms→50ms throttle drift in Run 5).

**Root diagnosis:** BeadsLog is **write-triggered with no read-back from code.** You must *remember* to
record; nothing scans the code to verify memory is still true. Write-only memory drifts by construction.

## The proposal — close the code↔memory loop (two checks)

### 1. Orphan-code detection ("what code has no memory?")
Invert the graph. Instead of only "what did we record," answer *"what code exists that no devlog / issue
/ contract references?"*
- Scan tracked source files (e.g. `src/**`); for each, check whether its path (or a declared symbol)
  appears in any devlog/contract/issue.
- Files with **zero** references = **unmemoried** → surface them so a human/agent can attach a lifecycle
  record (or confirm "intentionally trivial, ignore").
- Especially valuable for **stores/components/composables** — the artifacts that carry behavior.
- Cheap v0: a grep script over `src/**` × `_rules/_devlog|_contracts` + git. No new infra.

### 2. Existence-claim verification ("does memory contradict the code?")
Memory that asserts *"no X exists"* / *"X is not implemented"* is a **falsifiable claim** — verify it.
- Detect existence-negations in contracts/devlogs ("no toast system", "not yet built", "does not exist").
- Grep the code for the named entity; if it's present, flag **DRIFT** with both locations.
- This is the check that would have caught the toast contract directly.
- Ties into the existing `bd devlog verify --fix` surface — add a "claim vs code" pass.

## Why this is the strategic bet (direction B)
The human is weighing two directions: (A) position BeadsLog on efficiency + memory use-cases, or
(B) make it more powerful/instant (LSP-style memory server + drift-verification + semantic search).
**This incident is the concrete, honest proof-point for B.** The whole benchmark's recurring weakness
was drift (curated memory going stale vs code); orphan + existence-claim detection is the direct
countermeasure, and it's the capability that turns "memory you must remember to write" into "memory that
notices when it's wrong."

## Companion discipline fixes (not the feature, but do them too)
- **Contract wording rule:** never write "no X exists" — annotate the greppable fact instead
  ("dormant X exists at <path>; no *blessed* X yet, see devlog"). A contract must not contradict grep.
- **Pre-commit guard (~20 lines):** new file under `src/stores/*` or `src/components/*` whose path is in
  no devlog ⇒ fail the check. Forces a lifecycle record at land-time — where the toast store slipped through.

## Concrete next steps if picked up
1. Prototype orphan-scan as a standalone script over this repo; count how many `src/**` files are unmemoried (baseline signal).
2. Prototype existence-claim scan over `_rules/_contracts` + `_rules/_devlog`; see how many negations grep contradicts.
3. If signal is real, propose the `bd` subcommand surface (e.g. `bd devlog orphans`, `bd devlog verify --claims`).
