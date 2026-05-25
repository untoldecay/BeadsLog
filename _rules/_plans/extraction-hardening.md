# BeadsLog Feature Plan: Extraction Hardening & Graph Correctness
**Date:** 2026-05-21
**Status:** Planned

## 1. Goal
Address fragmentation and noise in the architectural graph by prioritizing structured human intent over heuristic inference.

## 2. Core Principle
*Structured human intent beats heuristic inference, and heuristic inference beats nothing.*

## 3. Implementation Phases

### Phase 1: Graph Correctness (Noise Reduction)
- **Conservative Regex Extraction:**
    - Modify `RegexExtractor` to be less aggressive.
    - Exclude common noise (e.g., common English words, CSS properties).
    - Increase confidence of explicit matches (keywords like `service`, `modal`) vs. generic CamelCase.
- **Explicit Edge Promotion:**
    - Ensure the `### Architectural Relationships` markdown block is the "Gold Standard".
    - Modify the pipeline to ensure explicit edges from this block outrank any AI-inferred entities or relationships.

### Phase 2: Manual Curation (Entity Aliasing)
- **Phase 2.1 (The Tool):** Implement `bd devlog alias <TargetEntity> <Alias1,Alias2,...>` command.
    - Allows manually collapsing fragmented nodes (e.g., `pill-mode`, `pill-layer`) into a single canonical entity (`DrawerPill`).
    - Merges relationships in the `entities` and `entity_deps` tables.
- **Phase 2.2 (The Discovery / Hinting):** Opportunistic Hinting.
    - When running `bd devlog search` or `bd devlog graph`, perform a fast SQL overlap check (e.g., Jaccard similarity) to find entities sharing >80% of sessions.
    - Add a proactive hint to the CLI output: `💡 OPPORTUNITY: 'drawerpill' and 'pill-layer' appear to be the same. Run 'bd devlog alias drawerpill pill-layer' after verification.`
    - *Agent behavior:* Agents must verify the context before executing the alias command.

### Phase 3: History Recovery (Git Re-anchoring)
- **Squash-Merge Awareness:**
    - Extend the `bd-daemon`'s `ReconcileGitFacts` worker.
    - When a session SHA is missing from Git but its logic is found in `HEAD` (via `patch-id`), automatically update the `commit_sha` in the `sessions` table to the new squash/rebase commit SHA.
    - This ensures `bd devlog resume` and SHA-based lookups continue to work after repo cleanup.

## 4. Technical Details

### 4.1 Regex Tuning
Current: Matches all `[A-Z][a-z]+(?:[A-Z][a-z]+)+`.
Proposed: Require a minimum of 5 characters and check against a blacklist of generic terms.

### 4.2 Explicit Edges
The `ExtractRelationships` function in `regex.go` is already robust. The improvement will be in the merging logic in `pipeline.go` to give these relationships higher weight/priority.

### 4.3 Re-anchoring
Update `cmd/bd/daemon_reconcile.go`:
```go
if found {
    isMerged = true
    // NEW: Update the session table with the new SHA found in HEAD
    _, _ = db.ExecContext(ctx, "UPDATE sessions SET commit_sha = ? WHERE id = ?", foundSHA, sessionID)
}
```

## 5. Verification Plan
- **Unit Tests:** Verify new `RegexExtractor` filtering and SQL Jaccard similarity logic.
- **E2E Sandbox:** 
    1. Create fragmented entities.
    2. Check if search/graph outputs the alias hint.
    3. Run `alias` command.
    4. Verify graph is collapsed.
    5. Perform a squash merge.
    6. Wait for daemon to re-anchor sessions.
    7. Verify `bd devlog resume` works with new SHA.
