# [task] Consolidate bd upgrade into a single command (BeadsLog-q9o.2)

**Date:** 2026-08-08

_No architectural changes_

## Problem
`bd upgrade` was a parent with four subcommands (`status`, `review`, `ack`,
`check`) — misleading and redundant. `ack` also became obsolete once seen-state
went per-machine (q9o.3). Users want one command: am I current? what's new?
update?

## Work Done
- `cmd/bd/upgrade.go` rewritten: `bd upgrade` (no subcommands) shows current
  version, fetches the remote version from GitHub, and — if newer — surfaces
  that version's changelog Features and prompts to install. Flags `--install`
  (skip prompt) and `--json` ({current, latest, upgrade_available}).
- Deleted `status`, `review`, `ack`, and the `check` subcommand (its behavior
  IS the root now).
- `cmd/bd/upgrade_check.go`: added `fetchRemoteChangelog` (one fetch → version +
  best-effort latest Features via `extractLatestFeatures`); `fetchRemoteVersion`
  is now a thin wrapper for the background check. Background-check banner points
  to `bd upgrade`.
- The 🔄 post-`list`/`ready` upgrade banner (`version_tracking.go`) no longer
  references the deleted `bd upgrade review`; it now points to `bd refresh`
  (lands in q9o.1).
- Kept `pluralize` (test-covered) in upgrade.go.

## Validation
- Unit: TestExtractLatestFeatures (newest-version-only extraction + malformed
  input returns nil, no panic). TestPluralize still green.
- E2E Test 26: removed subcommands absent from help; `--json` carries
  current_version (network-tolerant). Test 25 updated to exercise `bd onboard`
  for the per-machine write instead of the now-removed `ack`. Suite 26/26 green.
- Live: `bd upgrade`/`--json` hit GitHub, report up-to-date correctly.

## Final Session Summary
**Final Status:** q9o.2 done on dev/refresh-postupdate. Remaining: q9o.1 (bd refresh).
**Key Learnings:**
- Deleting `ack` broke the q9o.3 test that used it — sibling-issue coupling;
  the per-machine write is now proved via `onboard` instead.
