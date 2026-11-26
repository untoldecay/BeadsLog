# Comprehensive Development Log: Devlog index links outside the devlog directory became ghosts and were pruned (BeadsLog-y8l)

**Date:** 2026-11-26

## Problem
Fray's devlog index links six old documents outside the devlog directory, e.g.
`[09_Security-Implementation.md](../_documentation/09_Security-Implementation.md)`. All six exist on develop, in the
main checkout and in worktree-1, yet their sessions were ghosts. A simulated `bd devlog prune` from a main checkout
(throwaway clone, copied database) deleted them: 933 -> 927 sessions. The worktree-safety fix (BeadsLog-svg.1)
did not cover this: it is a path bug, not a branch bug.

### **Phase 1: Root cause**

*   **Action Taken:** Followed the filename through sync, presence, show and prune.
*   **Result:** Every step reduced the link to its base name inside the devlog directory:
    `SyncSession` read `_rules/_devlog/09_Security-Implementation.md` (the comment said "use base filename to prevent
    double-path"), `Presence.Lookup/Read` only knew devlog-directory listings keyed by base name, `show` did the same,
    and prune matched index rows by `](basename`. So sync marked the sessions as ghosts, `RestoreFound` never found
    the files, and prune deleted them. Reproduced with dev.3: "Missing log session, _rules/_devlog/09_Security.md".

### **Phase 2: Fix**

*   **Action Taken:** `devlog.ResolvePath(dir, filename)`: plain names stay in the devlog directory; links with directories are followed; links that repeat the directory (`_rules/_devlog/x.md`) still resolve by base name, as before.
*   **Action Taken:** `Presence` handles "nested" filenames (relative with directories, or absolute outside the devlog dir) by repository path: Here on disk, Elsewhere at the same path in another worktree or on any branch tip (one `cat-file --batch-check` per path, cached), Unknown outside the repository. Plain names keep the existing fast base-name scan.
*   **Action Taken:** Missing directories are canonicalized through their nearest existing ancestor (`canonicalAbove`), so `/var` vs `/private/var` on macOS does not push a path "outside the repository" on a branch where the directory does not exist.
*   **Action Taken:** `SyncSession` and `show` use `ResolvePath`; sync passes an absolute path to presence. Prune removes index rows by the link as written for nested files.

### **Phase 3: Tests**

*   **Action Taken:** `TestPresenceLinkOutsideDevlogDir` (here, absolute, same base name is a different file, other branch with Read, other worktree, deleted everywhere, outside the repo is Unknown) and `TestResolvePath`. E2E Test 38: linked doc synced with its content; a false ghost from an older bd is restored by sync and kept by prune; on a branch without the doc it is not a ghost; deleted everywhere it is pruned with its index row. Test 38 fails on dev.3 at the first step.
*   **Result:** E2E 38/38; Go short suite: the same 9 pre-existing failures.
*   **Action Taken:** Fray simulation on a clone with all 23 Fray refs and the copied database: 933 sessions / 25 missing / 6 ghosts -> sync: 0 ghosts, 19 missing, 42 reported as living on other branches -> prune: 0 deleted, 933 kept.
*   **Result:** A first clone without Fray's other branches showed 42 ghosts: simulations must carry every ref, or presence correctly concludes the files are gone.

## Final Session Summary

**Final Status:** Done on fix/worktree-sync. Fray's real database untouched (backup in ~/.local/share/bd-backup/fray-20261126-0227).
**Key Learnings:**
*   **A path is not a name:** reducing links to base names was a shortcut for one directory layout that silently broke every other one.
*   **"Absent" needs the whole repository:** a clone missing branches proves nothing about deletion.

### Architectural Relationships
- SyncSession -> ResolvePath (uses)
- Presence.Lookup -> lookupPath (delegates to)
- lookupPath -> findPath (uses)
- findPath -> git cat-file (uses)
- PruneGhosts -> RestoreFound (uses)
- devlogShowCmd -> ResolvePath (uses)
