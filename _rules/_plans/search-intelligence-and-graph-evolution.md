# Plan: Search Intelligence & Graph Evolution

**Status:** In-Planning (P1: Search, P2: Graph/UX)
**Objective:** Transform BeadsLog from a storage-first memory system into an intelligence-led retrieval layer by implementing recursive search fallbacks, lightweight relationship inference, and proactive system health management.

---

## **Current State**
1.  **Search:** Uses SQLite FTS5 for BM25 ranked text search. It is strictly limited to the user's input string. Zero results lead to a dead end.
2.  **Graph:** Relies on explicit AI-extracted relationships (`A -> B`). Coverage is low (~55%) because heavy Ollama-based enrichment is slow or often offline.
3.  **Maintenance:** `bd devlog sync` is noisy, repeating warnings about missing files on every operation.

---

## **Proposed Solution**

### **1. Hybrid Ranking Engine (P1)**
Implement an intentional scoring policy using SQLite FTS5:
- **Scoring Formula:** `final_score = (-bm25) + 3.0(phrase) + 1.5(near_5) + 0.75(entity_pair) + recency_bonus`.
- **Multi-Form Queries:** Generate Phrase, NEAR, and AND logic variants from user input.
- **Foothold Guarantee:** Fall back to broad OR matches only if high-confidence results are absent, ensuring agents always have a foothold.
- **Explainability:** Add `bd devlog search --explain` to visualize the ranking weights (BM25 vs Bonuses).

### **2. Weighted Graphing & Semantic Edges (P2)**
Move beyond "dumb" co-occurrence to intentional relationship inference:
- **Proximity Weighting:** 
    - Same sentence = +5 weight.
    - Same session = +1 weight.
- **Typed Semantics:** Use regex to identify "Intentional Edges" (e.g., "fix X", "depends on Y", "refactor Z").
- **Canonical Normalization:** Implement an Alias Table to merge noisy entity names (e.g., `AuthService` vs `auth-service`) using simple string similarity heuristics.

### **3. Path Search (P2)**
Implement the "Killer Forensic Feature":
- **Command:** `bd devlog path <EntityA> <EntityB>`.
- **Logic:** Perform a shortest-path traversal on the `entity_deps` table to show how disparate components are historically linked through devlog sessions.

### **4. "Ghost" File & Memory Health (P2)**
- **Ghost Tracking:** Mark missing files as `is_ghost = 1` and suppress warnings in daily commands.
- **Memory Status:** `bd devlog status` will show queue depth and a "Memory Health" percentage (Optimized vs Total).

---

## **Implementation Steps**

### **Phase 1: Hybrid Search (P1)**
- [ ] **Query Generator:** Implement a helper to transform raw strings into FTS5-compatible Boolean/NEAR/Phrase expressions.
- [ ] **SQL Harness:** Refactor `internal/storage/sqlite/search.go` to use a Common Table Expression (CTE) for multi-tier scoring and joining metadata.
- [ ] **Explain UI:** Update `devlogSearchCmd` to handle fallback logic and display score breakdowns in `--explain` mode.
- [ ] **Preview Mode:** Add `--preview` flag to show the first 3 lines of each matching session.

### **Phase 2: Statistical Graphing (P2)**
- [ ] Update schema to track co-occurrence counts (or query `session_entities` dynamically).
- [ ] Modify `devlogGraphCmd` to surface these implicit links.

### **Phase 3: UX & Visibility (P2)**
- [ ] Update `bd devlog status` with queue metrics.
- [ ] Integrate "Ghost" file detection into `bd sync` and pruning into `bd doctor`.

---

## **Key Learnings & Risks**
- **Search Precision:** Tier 3 (OR) can be noisy. Results must be clearly labeled as "partial matches".
- **Graph Noise:** Co-occurrence isn't always a relationship. Use a threshold (e.g., must appear together in >2 sessions) to filter noise.
