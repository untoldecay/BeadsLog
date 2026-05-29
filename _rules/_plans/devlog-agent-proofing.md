# BeadsLog Feature Plan: Devlog Agent-Proofing & UX Optimizations
**Date:** 2026-05-26
**Status:** Planned

## 1. Goal
Reduce friction and data integrity risks for both human and AI agents by automating maintenance tasks, improving visibility of "orphaned" data, and ensuring repo-wide synchronization is the default behavior.

## 2. Optimization Features

### 2.1 Auto-Adoption Warnings (`bd devlog sync`)
- **Problem:** Agents create devlogs but forget to `record` them. `sync` currently ignores them silently.
- **Action:** During `sync`, scan the devlog directory for `.md` files not present in the index.
- **UX:** Print a prominent warning: `⚠️  Found 3 orphaned files. Run 'bd devlog verify --fix' to adopt them.`

### 2.2 Auto-Flush Metadata
- **Problem:** Housekeeping changes (`alias`, `unalias`, `pause`, `abandon`) are only saved to local SQLite. If the agent forgets to `sync`, the team never sees the cleanup.
- **Action:** Automatically trigger a `flushMetadata` operation at the end of these commands to update `.beads/aliases.jsonl` and `.beads/issues.jsonl`.

### 2.3 Atomic `record` (Stub Generation)
- **Problem:** Two-step process (create file -> run record) is error-prone.
- **Action:** If the `--file` provided to `bd devlog record` does not exist, create it automatically with a default template.

### 2.4 Non-Interactive `prune`
- **Problem:** `bd doctor --fix` is too heavy and interactive for ghost session cleanup.
- **Action:** Add `bd devlog prune` (non-interactive) to strictly purge sessions that no longer have a corresponding file on disk.

### 2.5 Preferred Casing for Entities
- **Problem:** Database stores lowercase, but human/AI readability is better with original casing.
- **Action:** Display entities in the UI using their "Preferred Casing" (the version first encountered or specified in an alias).

### 2.6 Protocol Reinforcement
- **Action:** Update `PROTOCOL.md` to mandate `bd devlog verify --fix` as part of the session close checklist.

## 3. Implementation Roadmap
1.  **Registry/Metadata Helpers:** Implement `flushMetadata` and `pruneGhosts`.
2.  **Command Updates:** Update `sync`, `record`, `alias`, `pause`.
3.  **UI/Display Polish:** Adjust entity rendering.
4.  **Protocol Documentation:** Update orchestration rules.
