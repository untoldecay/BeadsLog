# [feature] Solo/local-only mode at bd init (BeadsLog-paj)

**Date:** 2026-08-06

## Problem
Users working in team repos had no first-class way to keep beads data strictly
local: no init-time choice, `bd doctor` nagged about unpushed sync branches,
and the pre-commit hook silently stages `.beads/` into normal commits — leaking
beads data through regular feature-branch pushes.

## Work Done
- **New declarative config key `sync-mode`** (`"auto"` | `"local-only"`), with
  `sync.mode` alias following the `sync.branch` → `sync-branch` precedent
  (internal/config/config.go, yaml_config.go keyAliases).
- **`bd init --solo`** (cmd/bd/init_solo.go): writes `sync-mode: local-only`,
  `no-push: true`, `daemon.auto-sync: false`, then offers two git postures:
  1. **Invisible** (default): `.beads/` into `.git/info/exclude` via existing
     `setupForkExclude`; hooks + merge driver skipped (hooks would fail
     `git add`-ing excluded files).
  2. **Local branch**: `beads-local` sync branch via existing `syncbranch.Set`
     + `createSyncBranch`, never pushed.
  Mutually exclusive with `--team`; composes with `--stealth` (skips question).
- **Doctor respects the mode**: `CheckSyncBranchHealth` returns OK
  ("remote checks skipped") when `sync-mode: local-only` (cmd/bd/doctor/git.go).
- **Config template**: solo section documented in generated config.yaml
  (init_helpers.go).
- **Bug fix found via e2e**: `createSyncBranch` (init_team.go) had an inverted
  guard (`!=` instead of `==`) so it NEVER switched back to the original
  branch — team wizard users were silently left on the sync branch. Fixed.

## Validation
- Unit: TestApplySoloConfig (cmd/bd/init_solo_test.go) — three keys written.
- E2E in scratch repos: both wizard paths verified — config keys written,
  `.beads/` invisible to `git status` (option 1), `beads-local` created with
  user back on `main` (option 2), doctor shows "Local-only mode — skipped".
- `go vet` clean; internal/config + doctor suites green. cmd/bd suite has 6-7
  failures (TestInitCommand, TestIsForkOfBeads_*, TestVersionChangesCoverage,
  TestOutputContextFunction, etc.) — **verified pre-existing on clean baseline**
  by stashing all changes and re-running; not introduced by this work.

## Final Session Summary
**Final Status:** Implemented and e2e-validated on branch `dev/solo-mode-paj`;
issue BeadsLog-paj stays in_progress until merge. Also created BeadsLog-d2c
(copy unification, proposals in _rules/_analysis/).
**Key Learnings:**
- The pre-commit hook's `git add .beads/` is the leak vector solo mode must
  neutralize; git-exclude and sync-branch are the two working postures.
- Follow the `sync-branch` yaml-alias precedent for new dotted config keys —
  viper can't read flat dotted yaml keys as nested paths.
