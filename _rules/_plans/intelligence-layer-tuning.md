# Plan: Intelligence Layer Tuning (Signal-to-Noise)

**Status:** Planned
**Objective:** Refine the Intelligence Layer features based on agent feedback to improve discovery precision, resolve entity resolution friction, and fix data integrity edge cases.

---

## **1. Fuzzy Pathfinding (P1)**
The `path` command is currently too strict, requiring exact entity names.
- **Action:** Integrate `queries.SuggestEntities` and fuzzy matching into `devlogPathCmd`.
- **UX:** If entities aren't found exactly, show "Did you mean?" suggestions (similar to `graph`).

## **2. Architectural Entity Ranking (P1)**
The `entities` command is currently a word frequency counter, surfacing noise like "border-radius" or "used".
- **Action:** Change the ranking metric from "Mentions" (session_entities) to **"Degree Centrality"** (number of unique relationships in `entity_deps`).
- **Filter:** Implement a simple blacklist/heuristic to ignore common CSS properties and generic verbs.
- **Command:** Add `--sort=relationships` (default) and `--sort=mentions`.

## **3. Search UI Enhancements (P1)**
- **Date visibility:** Add the session date to the inline search results (currently only in `show`).
- **Snippet Polish:** Ensure the snippets in both `search` and `list` are as clean as possible.

## **4. Data Integrity Investigation (P1)**
Investigate report of mismatched session content (ID vs Narrative).
- **Hypothesis:** `bd devlog sync` might be miscalculating IDs when multiple sessions share a filename, or the `SyncSession` lookup-by-filename logic is too aggressive.
- **Action:** Verify ID stability and collision handling in `devlog_core.go`.

## **5. Graph Filtering (P2)**
- **Action:** Add `--type [uses|contains|refactored_to]` flag to `bd devlog graph` to allow filtering dense architectural nodes.

---

## **Implementation Steps**

### **Phase 1: Entity & Path Precision (P1)**
- [ ] Implement fuzzy resolution in `devlogPathCmd`.
- [ ] Refactor `entities` query to count relationships.
- [ ] Add basic stop-word filter for common noise.

### **Phase 2: UI & UX Polish (P1)**
- [ ] Update `ui.SearchResultItem` and `RenderResultsWithContext` to show Dates.
- [ ] Ensure `--limit` is prominently documented in help.

### **Phase 3: Integrity & Filtering (P2)**
- [ ] Add unit test for multi-session-per-file sync scenarios.
- [ ] Implement `--type` filter for `bd devlog graph`.
