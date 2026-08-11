# Solo/team mode transitions: devlog isolation, continuity, reversible round-trip

## Problem
Switching a shared repo between team and solo mode leaked state and clobbered
history. Three concrete faults: (1) solo mode only excluded `.beads/`, leaving
the team `_rules/_devlog/` graph shared; (2) solo's `sync-mode: local-only`,
`no-push`, `daemon.auto-sync: false` were written to the *tracked*
`config.yaml`, which `.git/info/exclude` cannot hide, so solo settings leaked
to the team on the next commit; (3) there was no clean team→solo→team round
trip — no way to carry team history into solo, publish solo work back, or
reverse the excludes/config.

## Solution
Implemented the epic (BeadsLog-9vd) on `dev/solo-team-transitions`.

- **Separate excluded devlog dir** (`cmd/bd/solo_transition.go`): solo repoints
  `devlog_dir` to `_rules/_devlog-solo/` and adds `.beads/` + the solo dir to
  `.git/info/exclude`. The committed team `_devlog/` is never touched. Fixed a
  clobber where `initializeDevlog` reset `devlog_dir` back to the default after
  the wizard set it — it now only writes `devlog_dir` when unset
  (`devlog_cmds.go`).
- **Fresh vs Continuity** (`init_solo.go`): team→solo offers a clean solo graph
  or copying the committed team devlogs into the solo dir (devlog sync scans a
  single dir, so a copy is the pragmatic carry).
- **Per-machine config via `config.local.yaml`** (`internal/config`): a new
  git-excluded override merged over `config.yaml` by viper
  (`MergeInConfig`). Invisible/stealth modes write solo settings there
  (`SetLocalYamlConfig`), so nothing leaks; local-branch/non-git modes still
  use `config.yaml`. Verified `config get sync.mode` → `local-only`,
  `no-push` → `true`, with `git status` clean.
- **Reversible publish (merge-in default)** (`init_team.go` Step 0): when the
  team wizard detects `devlog_dir == _devlog-solo`, it publishes new solo
  devlogs into the team dir, restores `devlog_dir`, removes the excludes,
  deletes `config.local.yaml`, and removes the now-redundant solo dir. Warns
  that un-excluding `.beads/` merges local issue state via per-issue LWW.

Round trip (team → solo invisible+continuity → solo devlog → solo→team publish)
verified clean end to end. Unit tests added for exclude add/remove (preserving
the user's own excludes), copy/publish, and the config-leak guard.

## Entities
- solo_transition.go
- config.local.yaml
- devlog_dir
- transitionToSolo
- transitionToTeam
- SetLocalYamlConfig
