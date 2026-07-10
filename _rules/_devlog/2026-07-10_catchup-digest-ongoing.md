# [feature] Catchup Feature Digest & Ongoing Lifecycle State

**Date:** 2026-07-10

## Problem
The flat catchup feed gave humans a wall of items with no shape, and off-branch work displayed as PAUSED — making merely context-switched branches look deliberately frozen to teammates and agents.

### **Phase 1: By-Feature Digest**

*   **Action Taken:** `GroupCatchupDelta()` (internal/queries/catchup.go) organizes the delta into feature arcs: branch is the primary key; trunk sessions group by dominant shared entity (greedy); leftovers form General. State changes attach to their arc by scope match.
*   **Action Taken:** `bd catchup --digest` renders each arc with sessions, deduped authors, top entities, and attached lifecycle deltas ("PAUSED since you last looked — reason"). `--json` added for both flat and digest modes.

### **Phase 2: Deterministic Arc Narratives**

*   **Action Taken:** `arcNarrative()` builds a 2-3 line prose summary per arc from real session content: the first session's Problem paragraph ("Started with:") plus the latest session's `**Final Status:**` line ("Currently:"), with fallbacks. No AI at render time; nothing invented.

### **Phase 3: Ongoing Lifecycle State**

*   **Action Taken:** Added `StateOngoing` to types. The automatic derivation for off-current-branch unmerged work (internal/queries/search.go) now yields ONGOING instead of PAUSED — PAUSED is reserved for explicit `bd devlog pause`.
*   **Action Taken:** New `bd devlog ongoing --scope --message` command (reusing handleStateChange) as the explicit un-pause/revive. New `[🔄 ONGOING]` badge. `CheckProximityWarnings` now filters to paused/abandoned only, so ongoing work never warns.

### **Phase 4: Precision Fix Found by Tests**

*   **Initial Problem:** A state change could survive `--ack` forever: `branch_states` timestamps carry sub-second precision (RFC3339Nano) while the ack checkpoint stored second-precision RFC3339, and comparisons are lexicographic.
*   **Action Taken:** Ack checkpoint and the delta filter both use RFC3339Nano now.

### **Phase 5: Tests & Docs**

*   **Action Taken:** `TestGroupCatchupDelta`/`TestArcNarrativeFinalStatus` (queries), plus 3 CLI tests (digest render, digest JSON + ack cutoff, ongoing state/badge/no-warning). Cobra flag stickiness across in-process runs defeated with explicit `--flag=false`.
*   **Action Taken:** docs/CATCHUP.md (digest section) and docs/LIFECYCLE.md (state table + revive command) updated.

## Final Session Summary

**Final Status:** Catchup digest and ongoing state shipped (issue BeadsLog-1w1); real-repo digest groups the day's work into a single `feature/enhance-fable` arc with an accurate two-line narrative and 16 closed issues.
**Key Learnings:**
*   **Derived states must not impersonate explicit ones:** auto-deriving PAUSED for off-branch work destroyed the signal that pause is a deliberate act with a reason.
*   **Lexicographic timestamp comparison demands one format:** mixing RFC3339 and RFC3339Nano in the same column comparison creates same-second immortal rows.

### Architectural Relationships
- bd catchup -> GroupCatchupDelta (uses)
- GroupCatchupDelta -> arcNarrative (uses)
- HybridSearch -> StateOngoing (derives)
- devlog ongoing -> handleStateChange (uses)
- CheckProximityWarnings -> paused filter (refines)
