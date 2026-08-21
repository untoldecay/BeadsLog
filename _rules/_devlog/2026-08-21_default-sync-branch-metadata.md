# Default beads commits to a dedicated sync branch, never the work branch

## Problem
Plain `bd init` never runs the team wizard, so the "Is your main branch
protected?" prompt (the only place `beads-metadata` was set up) was never seen.
With `sync.branch` unset, `bd sync` committed and pushed beads **to the current
work branch** (e.g. `develop`), polluting shared work. A user hit exactly this:
`bd sync` pushing devlogs/issues onto `develop`. (BeadsLog-2lt)

## Solution
Make a dedicated sync branch the default so beads never lands on a work branch.

- **`cmd/bd/init.go`**: when a git remote exists (the only case where beads can
  leak via push) and none of `--branch/--inline/--solo/--stealth` apply, default
  `branch = defaultSyncBranch` ("beads-metadata"). A dedicated branch sidesteps
  the GH#807 worktree collision (it's never checked out in the work tree). Added
  a `--inline` flag to restore commit-to-current-branch.
- **`cmd/bd/init_team.go`**: dropped the "protected main?" prompt entirely — a
  dedicated branch is safe whether or not main is protected (that's all the
  question decided) — replaced with an info message. Always sets `beads-metadata`
  (honoring an already-configured name). Simplified the summary accordingly.
- **`cmd/bd/sync.go`**: made the CLI sync-branch path no-remote-safe — skip the
  worktree fetch and skip the push when there's no `origin` (mirrors the daemon's
  `syncBranchPull` guard), so `beads-metadata` also works on purely local repos.

Push behavior is unchanged and still governed by mode: solo = `no-push`, team =
auto-push. `beads-metadata` being pushed in team mode is fine — it's a separate
branch, not shared work. Solo is untouched (already isolates via exclude /
`beads-local`).

## Verification
e2e Test 33: repo with a bare remote → plain `bd init` on `develop` defaults to
`beads-metadata`; after create+sync, `issues.jsonl` is on local + `origin`
`beads-metadata`, `develop` and `origin/develop` carry **zero** beads;
`--inline` sets no sync branch. Local-only repos keep the inline path (whole
existing e2e suite unchanged, all green).

## Entities
- init.go
- init_team.go
- sync.go
- defaultSyncBranch
- beads-metadata
