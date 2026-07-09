# [feature] Trust Layer Hardening: Ghost-Free Reads & Honest Health

**Date:** 2026-07-10

## Problem
The devlog memory served dead data: `resume`/`search`/`list` returned ghost sessions (index entries whose files were deleted), `show` crashed with a raw OS error on ghosts, `status` claimed "All memory optimized" while 80/148 sessions were incomplete, and sessions removed from `_index.md` could never be ghost-marked by sync (stale `is_ghost=0` forever). Analysis: `_rules/_analysis/2026-07-10_trust-layer-report.md`.

### **Phase 1: Ghost Exclusion on Read Paths**

*   **Action Taken:** Added `is_ghost = 0` filters to `devlog resume`, `devlog list` (cmd/bd/devlog_cmds.go), hybrid search (internal/queries/search.go), and catchup (internal/queries/catchup.go).
*   **Result:** Agents can no longer resume or search into sessions whose files are gone. Ghost visibility lives in `status` counts.

### **Phase 2: Ghost-Aware Show & Honest Status**

*   **Action Taken:** `devlog show` on a ghost now explains the situation and points at `bd devlog prune` instead of leaking `open ...: no such file or directory`.
*   **Action Taken:** `devlog status` now counts unenriched sessions (`enrichment_status` 0/NULL), only claims "All memory optimized" when pending/failed/unenriched/incomplete/ghosts are all zero, and prints a memory-health warning line otherwise.

### **Phase 3: Sync Reconciliation**

*   **Action Taken:** `devlog sync` now ends with a health summary (ghost count → prune hint, incomplete count → verify hint) via a shared `devlogHealthCounts` helper.
*   **Action Taken:** Root-cause fix: sync now ghost-marks sessions whose index row was removed AND whose file is missing — previously these were invisible to the row loop and stayed `is_ghost=0` forever (found live: `sess-d89ed1` and 15 GSD sessions).

### **Phase 4: Machine-First Output (--json)**

*   **Action Taken:** Added `--json` support (reusing the global flag + `outputJSON` helper) to `devlog list`, `show`, `resume`, `status`, `entities`, and `impact`. Search already had it; graph deferred until a consumer exists.

### **Phase 5: Tests**

*   **Action Taken:** New `cmd/bd/cli_devlog_trust_test.go` with 6 integration tests: ghost exclusion across read commands, ghost-aware show (exec-based, os.Exit path), sync summary, de-indexed session reconciliation, status honesty, and JSON validity for all six commands.
*   **Result:** All pass; `go test -short ./...` green. Note: the full `-tags=integration` suite fails on pristine HEAD too (pre-existing test-pollution issue in cmd/bd, unrelated to this change).

## Final Session Summary

**Final Status:** Trust layer shipped on `feature/enhance-fable` (issue BeadsLog-czn). On the real repo, `resume --last 4` now returns the 4 real sessions instead of 4 ghosts, and sync surfaced 16 previously invisible stale sessions (7 → 23 ghosts).
**Key Learnings:**
*   **Reads must be filtered at the query, not repaired by tooling:** a repair command that must be remembered (`prune`, `verify --fix`) is the same trap as a two-step record; the read path itself has to refuse dead data.
*   **Reconciliation must cover the DB-only case:** any sync that iterates the index alone can never fix records the index no longer references.

### Architectural Relationships
- devlog resume -> is_ghost filter (implements)
- devlog sync -> devlogHealthCounts (uses)
- devlog sync -> ghost reconciliation (implements)
- devlog show -> ghost error UX (refines)
- devlog status -> memory health (refines)
- devlog list -> json output (implements)
