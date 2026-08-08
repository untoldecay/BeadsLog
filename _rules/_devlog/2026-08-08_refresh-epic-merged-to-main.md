# [release] Frictionless post-update epic landed on main (v0.59.0)

**Date:** 2026-08-08

_No architectural changes_

## Problem
The frictionless post-update epic (BeadsLog-q9o) was complete on
`dev/refresh-postupdate` but local `main` was stale (still at the
GitHub/Notion plan commit). Land the epic on main and cut the release.

## Work Done
- Ran the full `_sandbox/run_e2e_tests.sh` suite: 27/27 green.
- Fast-forwarded local `main` to `dev/refresh-postupdate` (74c0c8f3, v0.59.0) —
  clean ff, no merge commit needed. Branch kept (also already on origin from the
  earlier release cut); no further pushes.
- Epic contents now on main:
  - q9o.1 — `bd refresh`, one-command post-update.
  - q9o.2 — `bd upgrade` consolidated to a single command.
  - q9o.3 — per-machine changelog-seen state (`.local_changelog_seen`), no
    cross-clone ack leak.

## Final Session Summary
**Final Status:** Epic q9o merged to local main at v0.59.0; release cut via
scripts/cut-release.sh. Branch retained locally.
**Key Learnings:**
- main had fallen far behind because the whole solo-mode → alias → refresh arc
  lived on feature branches with releases cut there; a periodic ff-merge to main
  keeps tags anchored to main history.
