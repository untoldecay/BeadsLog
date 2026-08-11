# Solo-invisible hides the full beads footprint from a non-beads team

## Problem
A solo dev on a shared repo whose team does NOT use beads: `bd init` injects the
`<beads_protocol>` block (or the pre-onboard bootstrap trap) into agent files
(AGENTS.md, CLAUDE.md, .cursorrules, GEMINI.md, …), scaffolds
`_rules/_orchestration/` + `_rules/_devlog/`, and adds a `merge=beads` line to
`.gitattributes`. All of it is tracked/committable, so following the AGENTS.md
close protocol (`git add -A` → commit → push) pollutes teammates who never opted
into beads. Previously solo only hid `.beads/` (BeadsLog-9vd, kf7). (BeadsLog-0mw)

## Solution
Key insight: agents read these files from **disk**, not git. So hiding the
footprint from git — while leaving it on disk — keeps solo fully functional and
the team never sees beads.

In `cmd/bd/solo_transition.go`, `transitionToSolo` now hides the whole footprint
in one exclude write plus skip-worktree:
- `soloFootprintDirs` (`_rules/_orchestration/`, `_rules/_devlog/`) → excluded
  (only hides untracked instances; a team that committed them uses beads and is
  left alone).
- `footprintFiles` = agent Candidates + `.gitattributes`; `agentFileHasProtocol`
  detects a bd marker (`<beads_protocol>`/legacy tag/`BootstrapTrap` for agent
  files, `merge=beads` for .gitattributes). Untracked → excluded; tracked →
  skip-worktree (their beads content stays on disk for the solo agent while
  git ignores it and the team's committed copy is untouched).
- `.beads/` handled as before (exclude + skip-worktree, kf7).

`transitionToTeam` reverses all of it: clears skip-worktree on the tracked
footprint files and removes every possible exclude pattern.

`protocolCommittedFiles` + a wizard warning cover the wrinkle: skip-worktree
can't hide a block already in git history, so if the block/trap is committed the
user is told to strip it.

Decisions (user): skip-worktree the shared agent file (keep the block local)
rather than relocating it; make full hiding the default for solo-invisible.

## Verification
e2e Test 32 (new): team baseline with its own beads-free AGENTS.md → solo dev
adds beads + goes invisible → `git add -A` after a private issue stages nothing
beads (incl. .gitattributes); AGENTS.md keeps beads content on disk, is
skip-worktree'd, and its committed copy still holds the team's rules; rejoin
clears skip-worktree. Full 32-test suite + solo unit tests green.

## Entities
- solo_transition.go
- beadsFootprintExcludes
- footprintFiles
- agentFileHasProtocol
- BootstrapTrap
- transitionToSolo
