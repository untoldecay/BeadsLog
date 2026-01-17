# Comprehensive Development Log: Debugging Fuzzy Search & Graph Resolution

**Date:** 2026-01-17

### **Objective:**
To resolve persistent "No entity found matching 'X'. Did you mean? - X" errors in `bd devlog graph` and `bd devlog impact` commands, despite the underlying `ResolveEntities` logic appearing sound in isolated tests. This session focused on diagnosing and fixing the discrepancy between expected and actual command behavior.

---

### **Phase 1: Initial Diagnosis & Test Environment Flaws**

**Initial Problem:**
`bd devlog graph [entity]` and `bd devlog impact [entity]` consistently failed, reporting "No entity found matching 'X'. Did you mean? - X", where 'X' was a known entity name. Simultaneously, `bd devlog search` and the "Did you mean?" suggestions from `SuggestEntities` (which uses the same `ResolveEntities` function internally) were working correctly.

*   **Hypothesis:** Bug in `queries.ResolveEntities` or database state.
    *   **Action Taken:**
        *   Added debug print statements to `queries.ResolveEntities` (`internal/queries/fuzzy.go`).
        *   Rebuilt and re-ran CLI tests.
    *   **Result:** Discovered `sqlite3: expected 2 destination arguments in Scan, not 1` error due to selecting `rowid` and `name` from `entities_fts` when only `name` was present (FTS virtual table selection quirk).
    *   **Correction:** Modified `queries.ResolveEntities` to `SELECT name FROM entities_fts` only.
    *   **Result:** Error changed to `No entity found... Did you mean? - X`, indicating `ResolveEntities` was still returning empty, but without a fatal DB error.

*   **Hypothesis:** Corrupted or stale database/FTS index.
    *   **Action Taken:** Performed `bd devlog reset` followed by `bd devlog sync` to rebuild database from scratch.
    *   **Result:** Problem persisted.

---

### **Phase 2: Isolating `ResolveEntities` in a Controlled Test Environment**

**Initial Problem:** The core `ResolveEntities` function still returned empty slices in the CLI context, despite initial debugging indicating FTS matches.

*   **Hypothesis:** The test setup (`setupTestDB`) in `internal/queries/fuzzy_test.go` itself was flawed, or the `go-sqlite3` driver was behaving erratically with in-memory DBs.
    *   **Action Taken:**
        *   Created `internal/queries/fuzzy_test.go` with `TestResolveEntities` and `TestSuggestEntities` for isolated testing.
        *   Initially, `fuzzy_test.go` failed with `sqlite3: SQL logic error: no such table: entities`.
        *   **Correction:** Diagnosed that `file::memory:?cache=shared` with `ncruces/go-sqlite3/embed` was causing inconsistent visibility of tables. Modified `setupTestDB` to use a **temporary file-based database** (`os.CreateTemp`).
    *   **Result:** All `TestResolveEntities` and `TestSuggestEntities` in `fuzzy_test.go` **passed successfully**. This proved `ResolveEntities` was logically correct when given a stable database.

---

### **Phase 3: Pinpointing the CLI Discrepancy**

**Initial Problem:** `ResolveEntities` worked in isolated tests, and when called by `SuggestEntities`, but *not* when called directly by `devlogGraphCmd` or `devlogImpactCmd`.

*   **Hypothesis:** A parameter discrepancy or subtle bug in the CLI command's invocation of `ResolveEntities`.
    *   **Action Taken:**
        *   Added explicit debug prints to `devlogGraphCmd` and `devlogImpactCmd` to trace the return values of `ResolveEntities` and `SuggestEntities`.
        *   Rebuilt and re-ran CLI tests.
    *   **Result:** Debug prints revealed: `ResolveEntities(...) returned 0 targets (error: <nil>)` from the direct command, but `SuggestEntities(...) returned N suggestions (error: <nil>)` for the same term. This confirmed the discrepancy.

*   **Root Cause Identified:** The `limit, _ := cmd.Flags().GetInt("limit")` was being used in `devlogGraphCmd` and `devlogImpactCmd`. However, the `--limit` flag was **never defined** for these two commands in the `init()` function of `devlog_cmds.go`.
    *   **Consequence:** `cmd.Flags().GetInt("limit")` returned the zero value for int, which is `0`.
    *   **Impact on SQL:** `LIMIT 0` in an SQL query (used by `ResolveEntities`) causes the query to return **zero rows**. This perfectly explained why `ResolveEntities` was returning an empty slice (`0 targets`).
    *   **Why SuggestEntities worked:** `SuggestEntities` hardcodes `limit=5`, bypassing this bug.

---

### **Phase 4: Resolution & Final Verification**

**Resolution:**
*   **Action Taken:** Modified `cmd/bd/devlog_cmds.go:init()` to explicitly add `devlogGraphCmd.Flags().Int("limit", 25, "...")` and `devlogImpactCmd.Flags().Int("limit", 25, "...")`. Also added the `--strict` flag to these commands for consistency as per roadmap.
*   **Action Taken:** Rebuilt the `bd` binary.
*   **Action Taken:** Re-ran all CLI tests with `quickstart`, `agentrules`, `verifycmd`, and typo `agntrules`.
*   **Result:** All `bd devlog graph`, `bd devlog impact`, and `bd devlog search` commands now function correctly, resolving entities and displaying expected output.

**Final Status:**
All fuzzy search and graph resolution features are fully functional. The implementation is robust, and the debugging process has significantly increased confidence in the underlying `ResolveEntities` function.

**Key Learnings:**
*   **Flag Definition Criticality:** Undefined Cobra flags silently return zero values, leading to subtle bugs in downstream logic (e.g., `LIMIT 0`).
*   **Test Environment Stability:** In-memory SQLite (`file::memory:?cache=shared`) can be unreliable for complex schema testing; temporary file-based DBs offer better isolation and robustness for tests.
*   **Debugging Layers:** The problem crossed multiple layers (CLI flags -> core resolver logic -> database interaction), requiring a methodical approach with debug prints and isolated testing to pinpoint the root cause.

---

### **Architectural Relationships**
- `bd devlog graph/impact` -> `ResolveEntities` (calls)
- `ResolveEntities` -> `entities_fts` (queries)
- `ResolveEntities` -> `entities` (queries)
- `ResolveEntities` -> `cmd.Flags().GetInt("limit")` (uses CLI flag)
- `SuggestEntities` -> `ResolveEntities` (calls)
