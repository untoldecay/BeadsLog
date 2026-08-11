# Solo mode: stop already-tracked .beads/ files leaking via skip-worktree

## Problem
Solo invisible mode hid beads with `.git/info/exclude`, which only affects
UNTRACKED files. In an established team-beads repo, `.beads/issues.jsonl` and
`config.yaml` are already committed, so exclusion could not hide the solo
user's local mutations. Following the AGENTS.md SESSION CLOSE PROTOCOL
(`git add -A` → commit → push), the agent's private issues leaked to the team
(and risked LWW conflicts). Verified: fresh repos were safe; established team
repos leaked. `bd sync` itself does not auto-commit, so the trigger was the
protocol's explicit git steps. (BeadsLog-kf7)

## Solution
Git's native skip-worktree bit is the mechanism for locally diverging a tracked
file without staging/committing/pushing it. In `cmd/bd/solo_transition.go`:
- `gitTrackedBeadsFiles()` — `git ls-files -z .beads/` lists the tracked files
  (issues.jsonl, config.yaml, aliases.jsonl, …).
- `setSkipWorktree(files, skip)` — flips `--skip-worktree` / `--no-skip-worktree`.
- `transitionToSolo` sets skip-worktree on those files (paired with the existing
  `.beads/` exclude that covers untracked files).
- `transitionToTeam` clears it first, deliberately re-exposing local issue state
  to merge back (per-issue LWW) — consistent with the publish/merge-in default.

The daemon caveat from memory (`git-rebase-vs-bd-daemon`: skip-worktree +
auto-commit fighting a rebase) is disarmed because solo runs with
`daemon.auto-sync: false` — the daemon never auto-commits, so it can't race a
rebase. Documented the residual ceiling (a manual rebase/pull touching the file)
and the upgrade path (separate untracked DB/JSONL) in a ponytail comment.

e2e Test 31 extended: after going solo it asserts issues.jsonl carries the
skip-worktree bit, that a private solo issue + `git add -A` stages nothing under
`.beads/`, and that rejoining the team clears the bit. Full suite green.

## Entities
- solo_transition.go
- setSkipWorktree
- gitTrackedBeadsFiles
- transitionToSolo
- transitionToTeam
