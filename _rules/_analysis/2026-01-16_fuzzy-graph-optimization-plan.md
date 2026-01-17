# Analysis: Fuzzy Logic & Graph Optimization Plan

**Date:** 2026-01-16
**Status:** Plan Definition
**Objective:** Enhance `bd devlog` commands (`graph`, `impact`, `search`) to be more "human-friendly" and "agent-forgiving" through fuzzy matching, better grouping, and intelligent feedback loops.

---

## 1. Problem Statement

Current `bd devlog` commands are too strict, leading to "false negative" experiences where users/agents think data is missing when it's just a naming mismatch.

### **The Gaps**
1.  **Strict vs. Fuzzy Inconsistency:**
    *   `search "modal"` finds `AddColumnModal` (Fuzzy).
    *   `impact "modal"` returns 0 results because it looks for exact match `modal` (Strict).
    *   `graph "modal"` returns multiple roots but looks like a broken tree (Confusing).
2.  **Zero-Result Dead Ends:**
    *   If a user types `graph auth-service` but the entity is `AuthService` (case) or `authentication-service` (synonym), they get "No entities found" with no further guidance.
3.  **Visual Noise vs. Silence:**
    *   `graph` output with multiple roots is unstructured.
    *   `impact` output is flat.

---

## 2. Proposed Enhancements

### **2.1 Fuzzy Impact (The "DWIM" Approach)**
*   **Change:** `bd devlog impact <term>` defaults to substring matching (`LIKE %term%`).
*   **Behavior:** `impact "modal"` will automatically show impact for `AddColumnModal`, `ManageColumnsModal`, etc.
*   **Grouping:** Output must be grouped by matched entity to avoid confusion.
    ```text
    Impact of 'modal' (fuzzy match):

    [AddColumnModal]
    - depends on: api-client

    [ManageColumnsModal]
    - depends on: rowdetailmodal
    ```
*   **Control:** Add `--strict` flag for exact matching behavior.

### **2.2 Graph Visualization Improvements**
*   **Visual Separation:** If multiple roots are found (fuzzy match), explicitly separate them.
    ```text
    Graph for 'modal' (3 matches):

    === AddColumnModal ===
    AddColumnModal (0)
    └── api-client (1)

    === ManageColumnsModal ===
    ManageColumnsModal (0)
    └── rowdetailmodal (1)
    ```
*   **Limits:** Default to top N (e.g., 10) matches to prevent terminal flooding.

### **2.3 "Did You Mean?" Feedback Loop**
*   **Scenario:** User runs `graph "auth"` (strict/exact intent) or a query that returns 0 structural results but has text matches.
*   **Action:** If `graph` finds 0 nodes but `search` finds entities, print:
    ```text
    No exact graph found for 'auth'.
    Did you mean one of these?
    - AuthenticationService
    - UserAuth
    - OAuthProvider
    ```
*   **Value:** This prevents agents from hallucinating that a component doesn't exist.

### **2.4 Noise Control (Top-N + Hint)**
Instead of interactive pagination (which breaks agent streams), use "Head + Hint".
*   **Logic:** Show top 25 results (sorted by relevance/mention count).
*   **Hint:** If more exist, print:
    `... and 42 more matches. Use --limit 50 or --strict to filter.`

---

## 3. Implementation Plan

### **Phase 1: Database & Query Layer**
*   Update `internal/queries` to support `LIKE` pattern matching for relationships.
*   Ensure case-insensitive matching is standard (SQLite `NOCASE` or `LOWER()`).

### **Phase 2: CLI Command Logic**
*   **Refactor `devlogImpactCmd`:**
    *   Implement fuzzy search loop.
    *   Group output.
    *   Add `--strict` flag.
*   **Refactor `devlogGraphCmd`:**
    *   Implement "Did you mean?" check (run lightweight entity search if graph is empty).
    *   Add separators for multi-root outputs.

### **Phase 3: Testing**
*   Verify `impact "modal"` finds `AddColumnModal`.
*   Verify `graph "unknown"` suggests `KnownEntity`.
*   Verify `--strict` restores original behavior.

---

## 4. Success Metrics
*   **Zero "False Empty" Reports:** Users/Agents should rarely see "No dependencies found" if relevant entities exist in the database.
*   **Discovery Speed:** Agents can find the right component name (`AuthService` vs `auth-service`) in 1 turn (using hints) instead of 3 (search -> list -> graph).
