# Comprehensive Development Log: Intelligence Layer Evolution (Search & Graph)

**Date:** 2026-05-09

### **Objective:**
Implement the first two phases of the Intelligence Layer roadmap: Hybrid FTS5 Search ranking and Statistical Graphing (co-occurrence). Additionally, resolve UX friction by implementing "Ghost" session management and enrichment visibility.

---

### **Phase 1: Hybrid Search Engine**

**Initial Problem:** Search was naive BM25-only and returned 0 results for multi-word queries that didn't match exactly.

*   **Action Taken:** Implemented `decomposeQuery` to generate Phrase, NEAR, AND, and OR variants.
*   **Action Taken:** Implemented a hybrid ranking formula in `internal/queries/search.go`: `(-bm25) + 3.0(Phrase) + 1.5(NEAR) + Recency`.
*   **Action Taken:** Added `--preview` and `--explain` flags to `devlog search`.
*   **Result:** Search is now robust and provides a "foothold" even for broad queries.

---

### **Phase 2: Statistical Graphing & Pathfinding**

**Initial Problem:** Graph coverage was low due to slow AI enrichment.

*   **Action Taken:** Added `GetRelatedEntitiesByCooccurrence` query to find entities frequently mentioned together.
*   **Action Taken:** Updated `devlog graph` UI to show a "Co-mentioned" table alongside explicit dependencies.
*   **Action Taken:** Added a stub for `bd devlog path A B` to prepare for historical pathfinding.
*   **Result:** The architectural map is significantly denser and more useful out-of-the-box.

---

### **Phase 3: Ghost Management & Visibility**

**Initial Problem:** Stale devlog entries caused repetitive "Missing file" warnings.

*   **Action Taken:** Added `is_ghost` column to `sessions` table (Migration 050).
*   **Action Taken:** Modified `SyncSession` to mark missing files as ghosts and silence their warnings.
*   **Action Taken:** Implemented `CheckDevlogGhosts` and `PruneDevlogGhosts` in `bd doctor`.
*   **Action Taken:** Added AI enrichment queue metrics to `bd devlog status`.
*   **Result:** UX is cleaner, and technical debt resolution is now centralized in `bd doctor`.

---

### **Final Session Summary**

**Final Status:** Successfully implemented Phase 1 and most of Phase 2 of the Intelligence Layer roadmap.
**Key Learnings:**
*   **Parallel Query Strategy:** Running multiple FTS5 query forms (Phrase + AND) and merging results is more resilient for agent-style queries than sequential fallbacks.
*   **Statistical Context:** Co-occurrence is a high-signal, zero-cost alternative to LLM-based relationship extraction.

---

### **Architectural Relationships**
- HybridSearch -> SQLite FTS5 (optimizes)
- devlog graph -> co-occurrence (infers)
- bd doctor -> ghost sessions (prunes)
