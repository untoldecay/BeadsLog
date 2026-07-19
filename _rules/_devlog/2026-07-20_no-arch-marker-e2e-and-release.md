# [release] No-Arch Marker: e2e validation & v0.55.1

**Date:** 2026-07-20

## Problem
The `_No architectural changes_` marker (shipped in commit 450e35c1) was only
covered by a Go unit test that exercised the `devlog sync` path. Before
releasing, it needed validation against the real agent flow (`bd devlog record`)
in a real repo, and then a proper version bump/release.

## Work Done

### Phase 1: Real e2e validation in sandbox/
- Wrote `sandbox/test_no_arch_marker.sh` — drives the actual `bd` binary
  (`--no-daemon`) through `bd devlog record` (the real agent path, not
  hand-written orphan files) and asserts on `verify` output.
- Proves three paths: unmarked admin session IS flagged; marker-at-creation is
  NOT flagged; and a flagged session CLEARS after the marker is added + re-synced
  (the historical-devlog path). All 4 assertions green, self-cleaning via `prune`.
- **Discovery the unit test missed:** `bd devlog record` stores
  `problem + fileBody` as the session narrative, so the marker lands in the
  column `verify` checks — but only the `record`/first-import path was untested.
  Also `record` is idempotent on file hash, so re-recording an unchanged file is
  a no-op; adding the marker changes the file, so the hash changes and the
  narrative refreshes. Early red runs were stale test state (reused filenames),
  not a feature bug — fixed with PID-scoped names.

### Phase 2: Release v0.55.1 via scripts/cut-release.sh
- Pre-drafted release notes in the three release-note files (CHANGELOG.md,
  cmd/bd/info.go, internal/changelog/changelog.go) with the version already at
  0.55.1, so bump-version's sed edits were no-ops.
- **Gotcha handled:** bump-version's commit `git add` list omits
  internal/changelog/changelog.go, so its CurrentVersion bump + new entry would
  be dropped from the release commit. Pre-staging changelog.go landed it in the
  commit (verified: commit 51132787 contains all four files).
- Ran cut-release.sh non-interactively (patch → 0.55.1, skip in-script notes
  since pre-drafted, publish to main, tag). Pushed origin/main + tag v0.55.1;
  the earlier feature commit 450e35c1 rode up in the same push.

## Final Session Summary
**Final Status:** v0.55.1 released and pushed (commit 51132787, tag v0.55.1);
no-arch marker validated end-to-end via the real `bd devlog record` flow.
**Key Learnings:**
- Unit tests that exercise one import path can miss another: `record` and `sync`
  store narrative differently; test the path agents actually use.
- Release tooling can silently drop a file from the commit — verify the release
  commit's file list, don't trust the script's `git add` to be complete.

### Architectural Relationships
- cut-release -> bump-version (uses)
- bump-version -> changelog.go (omits-from-commit)
- no-arch-marker -> bd-verify (skips)
- test-no-arch-marker -> devlog-record (validates)
