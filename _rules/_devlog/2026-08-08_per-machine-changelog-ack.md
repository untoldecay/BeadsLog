# [fix] Per-machine changelog-seen state — no cross-clone ack leak (BeadsLog-q9o.3)

**Date:** 2026-08-08

## Problem
"Changelog seen / ack" state lived in COMMITTED files: `last-seen-changelog-version`
in `config.yaml` (written by `maybeShowChangelog`) and `LastBdVersion` in
`metadata.json` (written by `bd upgrade ack`). Acking on one clone committed the
marker; a teammate pulling it never saw the "what's new" for that version. The
upgrade DETECTION layer was already correct (per-machine gitignored
`.local_version`) — only the seen/ack state leaked.

## Work Done
- Added per-machine helpers `readLocalChangelogSeen` / `writeLocalChangelogSeen`
  backed by a new gitignored `.beads/.local_changelog_seen`
  (cmd/bd/version_tracking.go).
- `maybeShowChangelog` (onboard.go) now reads/writes the per-machine file
  instead of `config.yaml`; swapped the `config` import for `beads`.
- `bd upgrade ack` (upgrade.go) now writes the per-machine file and no longer
  touches the deprecated `metadata.json` LastBdVersion; dropped the now-unused
  `configfile` import.
- Gitignored `.local_changelog_seen`; removed the stale committed
  `last-seen-changelog-version: 0.50.0` line from this repo's config.yaml
  (now inert everywhere — first run on each machine re-derives).

## Validation
- E2E Test 25: ack writes the per-machine file, NOT config.yaml, and it's
  git-ignored; a fresh clone (per-machine file absent) still shows the changelog
  on `onboard`, then records its own seen-state. Full suite 25/25 green.
- Manual: acked on "machine A" → "machine B" (rm the file) still shows changelog,
  then suppresses on re-run.

## Final Session Summary
**Final Status:** BeadsLog-q9o.3 done on branch dev/refresh-postupdate. Unblocks
q9o.2 (bd upgrade consolidation can now drop `ack` entirely).
**Key Learnings:**
- Detection was already per-machine via `.local_version`; the seen/ack state was
  the only committed leak. Same "shared vs per-machine" split that bit us on the
  prefix work — committed state that should be per-clone.
