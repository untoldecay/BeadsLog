# BeadsLog Feature Plan: Devlog Lifecycle Branch State
**Date:** 2026-05-17
**Status:** Planned

## 1. Context & Goal
Agents (and users) currently suffer from a "Merge-State Blindspot". When reading history from `bd devlog search`, an agent may see a completed session and assume its code is present in the current baseline (the working tree). If the branch was archived, abandoned, or paused, the agent will hallucinate dependencies on code that was never merged.

**Goal:** Provide context-aware relevance for historical sessions based on Human Intent and Git Reachability.
- Prevent agents from treating dead branches as baselines.
- Expose the *Reason* a path was abandoned so the agent doesn't repeat historical mistakes.
- Ensure the state updates automatically when branches are deleted or merged, minimizing manual tagging overhead.

## 2. The Core State Model

We expose 3 visible states and 1 automatic trust badge:

*   `active`: The current working line.
*   `paused`: A side branch that is sleeping (not the current line, not merged).
*   `abandoned`: A path explicitly marked dead by a human with a reason.
*   `validated` (Badge): A Git fact proving the code landed in the mainline.

### Machine Rules (State Derivation)
*   **Active + Validated:** The session is an ancestor of `HEAD` (`git merge-base --is-ancestor`) or survived via squash/cherry-pick (`git patch-id`).
*   **Active (Unvalidated):** The session belongs to the currently checked-out branch.
*   **Paused:** The session is on a side-branch (or deleted branch) that is *not* validated, and not the current branch.
*   **Abandoned:** Overrides everything. Set manually via CLI.

## 3. Data Architecture

We must decouple the "Branch Intent" from individual sessions to prevent tagging 20 files for a single abandoned branch.

*   **SQLite Table:** `branch_states`
    *   `state` (TEXT)
    *   `scope_type` (TEXT) - e.g., branch, entity, file
    *   `scope_ref` (TEXT) - e.g., `feature/xyz`, `entity:UnifiedContentView`
    *   `short_reason` (TEXT)
    *   `full_reason_ref` (TEXT) - Link to the special devlog entry
    *   `actor`, `timestamp`, `commit_sha`, `branch_ref`
*   **Special Devlogs:** Every manual state change (`pause`, `abandon`) generates a `[state-change]` devlog session so the decision itself is indexed and searchable.

## 4. Components & CLI Flow

### 4.1 Explicit Commands (Manual Intent)
*   `bd devlog pause --scope <target> --message "<short_reason>"`
*   `bd devlog abandon --scope <target> --message "<short_reason>"`
*   *(Both write to `branch_states` SQL and create a special devlog file).*

### 4.2 Daemon (Git Fact Reconciliation)
The `bd-daemon` acts as a "Badge Reconciler" (not a state engine).
*   **Cheap Checks (On CLI use):** Checks `git merge-base` and current branch to determine immediate display state for search results.
*   **Expensive Checks (Background):** Runs `git patch-id` logic for unresolved/archived records to find squash-merge survivors.

### 4.3 UI & Search Rendering
When rendering devlog results in `search`, `list`, or `resume`:
*   Display Badge + Short Reason.
*   Examples: `[ACTIVE · VALIDATED]`, `[PAUSED · Waiting on API]`, `[ABANDONED · Memory leak]`.

### 4.4 Proximity Warnings (The Guardrail)
When an agent or user runs `bd ready` or `bd devlog resume` on a specific file or entity, `bd` checks for scope overlap in `branch_states`.
*   If working on `UnifiedContentView.vue` and an `ABANDONED` scope matches it, print:
    `⚠️ CAUTION: UnifiedContentView.vue is within the scope of ABANDONED sess-123 ("Memory leak").`

## 5. Implementation Roadmap (Epics/Issues)

1.  **Phase 1: Database & Scope Engine:** Create `branch_states` migration and SQLite queries.
2.  **Phase 2: Intent Commands & Logging:** Implement `bd devlog pause` and `bd devlog abandon` with special devlog generation.
3.  **Phase 3: Daemon & Reachability:** Implement Git ancestor checks, patch-id fuzzy matching, and the Daemon caching loop.
4.  **Phase 4: UI & Guardrails:** Update `search`/`list` outputs with badges and implement the Scope Overlap warnings on `resume`/`ready`.