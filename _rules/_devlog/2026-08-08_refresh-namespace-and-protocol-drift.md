# [fix] bd refresh: correct Namespace verdict + protocol-drift check (BeadsLog-d7j)

**Date:** 2026-08-08

## Problem
Two issues found when running `bd refresh` live on this repo:
1. The Namespace line surfaced `bd sync`'s internal noise —
   "Ignoring prefix mismatches (all are tombstones): [bd- 106, explicit- 1]" —
   as if it were a namespace verdict. Those are benign old-prefix tombstones;
   no namespace change occurred.
2. After a binary update, an agent shouldn't need the full heavyweight
   `bd onboard` (which reruns session-init scaffolding and rewrites protocol).
   The only genuine need is to acknowledge a NEW protocol when it changed.

## Work Done (cmd/bd/refresh.go)
- **Namespace fix**: `refreshNamespace` now reports only a real adoption (the
  "✓ Adopted upstream prefix …" line) and otherwise says
  "in sync (no namespace change)". The tombstone-cleanup noise is never shown.
- **Protocol-drift check**: new `refreshProtocol(reinject)` compares the block
  between the protocol tags in each agent file (Candidates) against this
  binary's embedded `FullBootloader`. If a file's block differs, it re-injects
  ONLY that block — never PROJECT_CONTEXT or user prose. Files without a
  protocol block are skipped (that stays onboard's job). A "Protocol:" line is
  added to the refresh summary. This gives "acknowledge the new protocol" on
  update without the full onboard.

## Notable side effect (intended)
Running refresh on this repo re-injected the protocol block in 9 agent files
(CLAUDE.md, AGENTS.md, etc.): the committed protocol was genuinely behind the
0.59.0 binary's `FullBootloader` (expanded step 4: write devlog first, use
`bd now`). Diff confirmed only the tagged block changed; surrounding prose
untouched. Committed as part of this change.

## Validation
- Unit: TestRefreshProtocol — outdated block reinjected, current block left
  alone, no-block file skipped, prose outside tags preserved, idempotent.
- E2E Test 27 extended: refresh prints a "Protocol:" line and reports it
  CURRENT on a freshly-inited repo (same binary → no drift). Suite 27/27 green.
- Live: Namespace now "in sync (no namespace change)"; Protocol correctly
  detected + fixed real drift.

## Final Session Summary
**Final Status:** BeadsLog-d7j done on branch dev/refresh-protocol-drift.
**Key Learnings:**
- Extracting a "result" line from a subcommand's stdout must match the specific
  success signal, not a loose keyword — "prefix" also matched benign noise.
- Full onboard is overkill for updates; a surgical protocol-block drift check is
  the right granularity. Orchestration scaffolding is already write-if-missing,
  so it never clobbers PROJECT_CONTEXT — the earlier fear was mostly unfounded.
