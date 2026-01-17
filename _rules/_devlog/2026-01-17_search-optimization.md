# Comprehensive Development Log: Unified Search & Graph Optimization

**Date:** 2026-01-17

### **Objective:**
To replace the naive substring search with a robust Hybrid Search Engine (BM25 + Entity Graph Expansion) and upgrade `bd devlog` commands (`graph`, `impact`) to be "human-friendly" with fuzzy matching and suggestions.

---

### **Phase 1: Foundation (FTS5)**

**Initial Problem:**
SQLite's `LIKE` operator is insufficient for relevance ranking and linguistic matching (stemming). We needed full-text search capabilities.

*   **Assumption:** We can add `fts5` virtual tables without breaking existing schema.
    *   **Action Taken:**
        *   Updated `internal/storage/sqlite/schema.go` to add `sessions_fts` and `entities_fts` using `content=` option (external content tables) to save space.
        *   Added triggers (`ai`, `ad`, `au`) to keep indexes in sync with base tables.
    *   **Result:** Schema definition valid.

*   **Assumption:** New indexes need population for existing data.
    *   **Action Taken:**
        *   Created migration `044_populate_fts.go`.
        *   Logic: Check if base tables have data; if so, run `INSERT INTO ... VALUES('rebuild')`.
    *   **Result:** Ensures seamless upgrade for existing users.

*   **Verification:**
    *   Created `internal/storage/sqlite/fts_check_test.go` to confirm `ncruces/go-sqlite3` supports FTS5.
    *   Result: Passed.

---

### **Phase 2: Core Logic (Hybrid Search)**

**Initial Problem:**
We needed a way to combine text relevance (BM25) with architectural context (Entity Graph).

*   **Implementation:**
    *   Created `internal/queries/search.go`.
    *   Implemented `HybridSearch`:
        1.  **BM25:** Query `sessions_fts` for ranked text matches.
        2.  **Entity ID:** Query `entities_fts` for matching entity names.
        3.  **Expansion:** Find sessions linked to those entities via `session_entities`.
        4.  **Merge:** Combine results, boosting score if a session matches both text and entity.
    *   **Correction during coding:** SQLite FTS5 `bm25()` returns lower values for better matches (negative/small), so I implemented score subtraction logic to boost relevance.

*   **Fuzzy Suggestions:**
    *   Created `internal/queries/fuzzy.go`.
    *   Implemented `SuggestEntities`: Tries FTS prefix match (`term*`) first, falls back to `LIKE %term%`.

---

### **Phase 3: Search Command Refactor**

**Initial Problem:**
`bd devlog search` was a simple SQL `LIKE`.

*   **Action Taken:**
    *   Refactored `cmd/bd/devlog_cmds.go` (`devlogSearchCmd`).
    *   Wired up `HybridSearch`.
    *   Added flags: `--strict` (disable expansion), `--text-only`, `--limit`, `--json`.
    *   Added "Did you mean?" output when zero results found.

---

### **Phase 4: Graph & Impact Refactor**

**Initial Problem:**
`graph` and `impact` commands required exact names, frustrating users who didn't know the exact capitalization or spelling.

*   **Action Taken:**
    *   Refactored `devlogImpactCmd`:
        *   Now searches for **all** matching entities (fuzzy).
        *   Groups output by Entity Name.
    *   Refactored `devlogGraphCmd`:
        *   Now handles multiple targets (Multi-root visualization).
        *   Uses `GetEntityGraphExact` for each resolved target to avoid double-fuzzy issues.
    *   Added `GetEntityGraphExact` to `internal/queries/graph.go` and removed legacy `GetEntityGraph`.

---

### **Phase 5: Cleanup & Polish**

**Action Taken:**
*   Updated `docs/CLI_REFERENCE.md` with new Devlog Management section.
*   Fixed a syntax error in `search.go` (un-escaped quote) found during `make build`.
*   Verified compilation success.

---

### **Final Session Summary**

**Final Status:**
The `bd devlog` suite is now powered by a robust search engine. Searching for "modal" ranks relevant sessions first and finds related architectural components. Graph and Impact commands are forgiving and helpful.

**Key Learnings:**
*   **SQLite FTS5 Content Tables:** Using `content='table'` is excellent for keeping DB size down, but requires manual `rebuild` on migration.
*   **BM25 Scoring:** It's counter-intuitive (lower is better), which requires careful handling when merging with custom scores.
*   **UX Pattern:** The "Did you mean?" pattern is essential for CLI agents to avoid dead ends.

---

### **Architectural Relationships**
- HybridSearch -> sessions_fts (uses)
- HybridSearch -> entities_fts (uses)
- HybridSearch -> session_entities (joins)
- SuggestEntities -> entities_fts (uses)
- devlogSearchCmd -> HybridSearch (calls)
- devlogImpactCmd -> SuggestEntities (calls)
