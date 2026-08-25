# [research] GitHub/Notion sync brick: feasibility & v1 plan

**Date:** 2026-08-06

_No architectural changes_

## Problem
Wanted a new integration brick mirroring bd issues to GitHub Issues (and maybe
Notion) while BeadsLog stays the source of truth. Needed a feasibility
assessment and a v1 plan before committing to build.

## Work Done
- Codebase assessment: `Issue.ExternalRef` + `Issue.SourceSystem`
  (internal/types/types.go:49-50) already exist and are indexed — no migration
  needed. The Linear integration (cmd/bd/linear.go + internal/linear/) is a
  complete template for the brick (pull/push/dry-run/prefer-local, incremental
  timestamp sync, config via bd config). Partial Jira scaffolding also exists.
- External research: upstream untoldecay/BeadsLog has no GitHub/Notion integration
  and has diverged to Dolt storage — greenfield for us. GitHub side is
  low-risk (go-github v89, `gh auth token` reuse, `since=` polling, throttle
  ~1 create/sec for bulk export). Notion side is a YAGNI candidate for v1: no
  up-to-date Go SDK (API 2025-09-03 data-sources split), and Notion natively
  mirrors GitHub Issues via synced databases.
- Chose scope Level 2 on the bidirectionality ladder: one-way push mirror +
  status-only pull-back (close/reopen). Dodges echo loops, field-mapping
  guesswork, and delete semantics — where all prior art (git-bug, Unito) says
  the complexity lives.
- Wrote the plan: `_rules/_plans/github-notion-issue-sync.md`.
- Created reactivation issue **BeadsLog-3c2** (paused until picked up).
- Fixed a `bd sync` failure on the way: JSONL held 30 legacy `bd-` prefixed
  issues vs DB prefix `BeadsLog-`; resolved with the documented
  `bd sync --rename-on-import` (references updated automatically).

## Final Session Summary
**Final Status:** Plan written and persisted; issue BeadsLog-3c2 open as the
reactivation handle. No code changes.
**Key Learnings:**
- The Linear brick is the proven mold for any external tracker sync; copy
  internal/linear/ structure rather than designing from scratch.
- Legacy `bd-` prefix issues in JSONL break `bd sync` imports; the supported
  fix is `--rename-on-import`, which rewrites IDs and updates references.
