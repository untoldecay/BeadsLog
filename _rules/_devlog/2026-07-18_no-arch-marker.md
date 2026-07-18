# [feature] "No architectural changes" marker for verify

**Date:** 2026-07-18

## Problem
`bd devlog verify` flagged administrative devlogs (READMEs, file moves, review
reports) as "incomplete" forever: the health/verify query treats any session
lacking entities OR typed edges as incomplete, with no escape hatch. This drove
agents to run `--fix-ai`, whose default 1B local model (`llama3.2:1b`, 5s budget)
hallucinated spurious edges just to clear the warning — polluting the graph that
retrieval-led reasoning depends on. Root cause: no way for a legitimately
non-architectural session to declare it has nothing to link.

## Work Done
- Added an opt-in sentinel: a devlog writes `_No architectural changes_` in its
  relationships block to mark itself complete.
- Both copies of the incomplete-check query (`devlogHealthCounts` and the
  `verify` command in `cmd/bd/devlog_cmds.go`) now exclude sessions whose
  narrative contains the marker via `AND narrative NOT LIKE '%No architectural
  changes%'`. Additive change — unmarked sessions behave exactly as before.
- Marked sessions never reach `--fix-ai`, so no hallucinated backfill edges.
- Authoring guidance updated: the `⚠️ MANDATORY: Architectural Relationships`
  block, the stub-template comment, and `docs/DEVLOG.md` now tell agents to write
  the marker for no-edge sessions.
- Regression test `TestCLI_DevlogVerify_NoArchMarkerNotFlagged`: a marked admin
  devlog with no arrows (0 edges) is reported complete; fails if the exclusion
  is removed.

Deliberately skipped (from the caveat report): making the extractor configurable
(already is — `entity_extraction.primary_extractor` / `ollama.model` are live
config keys), mandating edges at write time (template already did), and
confidence-gating AI extraction (separate, larger change).

## Final Session Summary
**Final Status:** Shipped the no-arch marker; verify no longer flags
administrative devlogs and no longer invites 1B-model edge hallucination to clear
them. 3 files, ~5 lines + 1 regression test, backward-compatible.
**Key Learnings:** A false positive with no escape hatch is an active hazard — it
pushed agents toward a worse tool (hallucinated edges). The cheap fix was an
opt-in marker at the query layer, not smarter extraction.

### Architectural Relationships
- bd-verify -> no-arch-marker (skips)
- devlogHealthCounts -> no-arch-marker (skips)
- no-arch-marker -> entity_deps (bypasses)
