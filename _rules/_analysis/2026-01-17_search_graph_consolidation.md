# Roadmap: Unified Search & Graph Optimization (BM25 + Fuzzy UX)

**Date:** 2026-01-17
**Status:** Approved Roadmap
**Based on:**
- `2026-01-16_fuzzy-graph-optimization-plan.md` (UX/Fuzzy Logic)
- `2026-01-16_confront-search-optimisation.md` (BM25/FTS5 Architecture)

---

## 1. Executive Summary

We will replace the current naive substring search with a **Hybrid Search Engine** combining **SQLite FTS5 (BM25)** for text relevance and **Entity Graph Expansion** for architectural context. Simultaneously, we will upgrade `graph` and `impact` commands to be "human-friendly" by defaulting to fuzzy matching with grouped outputs and "Did you mean?" suggestions.

## 2. Technical Architecture

### 2.1 Database Schema (FTS5)
We will introduce Virtual Tables for full-text search.

```sql
-- Full-text search for sessions
CREATE VIRTUAL TABLE sessions_fts USING fts5(
    title,
    narrative,
    problems,      -- Extracted problem descriptions
    solutions,     -- Extracted solution descriptions
    content='sessions',
    content_rowid='rowid',
    tokenize='porter unicode61'
);

-- Full-text search for entities
CREATE VIRTUAL TABLE entities_fts USING fts5(
    name,
    description,
    content='entities',
    content_rowid='rowid'
);
```

**Sync Strategy:** Triggers (`AFTER INSERT/UPDATE/DELETE`) on `sessions` and `entities` tables to keep FTS tables in sync automatically.

### 2.2 Hybrid Search Algorithm
1.  **Phase 1: Text Relevance (BM25):** Query `sessions_fts` to find sessions matching the user's terms, ranked by BM25 score.
2.  **Phase 2: Entity Identification:** Query `entities_fts` to find entities matching the user's terms.
3.  **Phase 3: Graph Expansion:** For identified entities, find their 1-hop neighbors (parents/children).
4.  **Phase 4: Context Injection:** Boost or include sessions that heavily reference the identified/expanded entities, even if the exact search term is missing (optional but powerful).
5.  **Phase 5: Result Merging:** Combine results, deduplicate, and sort by a hybrid score.

---

## 3. UX Specification

### 3.1 Standardized Flags
All `devlog` commands (`search`, `graph`, `impact`) will support:
- `--strict`: Disable fuzzy/stemming/expansion (Exact match only).
- `--limit N`: Cap results (Default: 25).
- `--json`: Machine-readable output.

### 3.2 Command Behavior

#### `bd devlog search <query>`
- **Default:** Hybrid BM25 + Entity Expansion.
- **Output:** Ranked list with relevance scores (debug mode) or simple ordered list.
- **Flags:**
    - `--text-only`: Skip entity expansion (pure BM25).
    - `--entity`: Force entity-first lookup.

#### `bd devlog impact <term>`
- **Default:** Fuzzy match (substring/FTS) -> Grouped Output.
- **Output:**
  ```text
  Impact of 'modal' (2 matches):

  [AddColumnModal]
  - depends on: api-client

  [ManageColumnsModal]
  - depends on: rowdetailmodal
  ```
- **Zero Results:** Show "Did you mean?" if close matches exist.

#### `bd devlog graph <term>`
- **Default:** Fuzzy match -> Visual Separation.
- **Output:**
  ```text
  Graph for 'modal' (Matches: 3):

  === AddColumnModal ===
  ... tree ...

  === ManageColumnsModal ===
  ... tree ...
  ```
- **Zero Results:** "No entity found for 'xyz'. Did you mean: X, Y, Z?"

---

## 4. Implementation Plan (Beads Issues)

### Phase 1: Foundation (FTS5)
- **Task:** Update `internal/storage/schema.go` to include `sessions_fts` and `entities_fts` tables and triggers.
- **Task:** Create migration/init logic to populate FTS tables for existing data.
- **Task:** Verify FTS5 availability in the build.

### Phase 2: Core Logic
- **Task:** Implement `internal/queries/search.go` with BM25 query support.
- **Task:** Implement `internal/queries/fuzzy.go` for "Did you mean?" logic (using FTS or Levenshtein if feasible, else simple `LIKE`).

### Phase 3: Command Refactor - Search
- **Task:** Refactor `cmd/bd/devlog_search.go` to use the new Hybrid Search.
- **Task:** Add `--text-only` and `--strict` flags.

### Phase 4: Command Refactor - Graph/Impact
- **Task:** Refactor `devlogImpactCmd` to support grouped fuzzy output.
- **Task:** Refactor `devlogGraphCmd` to support multi-root visualization.
- **Task:** Implement "Did you mean?" fallback for zero results.

### Phase 5: Cleanup & Polish
- **Task:** Update help text and documentation.
- **Task:** Remove old naive search code.

---

## 5. Success Metrics
- **Relevance:** `search "modal"` puts specific "modal" implementation tasks at the top, not generic mentions.
- **Clarity:** `graph "modal"` shows clear, separated trees instead of a confusing mesh or empty result.
- **Helpfulness:** Zero-result queries provide actionable alternatives 100% of the time if similar terms exist.
