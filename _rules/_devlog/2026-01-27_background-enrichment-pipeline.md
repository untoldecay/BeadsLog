# Comprehensive Development Log: Background AI Enrichment Pipeline

**Date:** 2026-01-27

### **Objective:**
To decouple slow Ollama entity extraction (15s+) from the critical user path (`bd devlog sync`, `git commit`). Implement a background worker in the daemon that processes sessions asynchronously, enriching them with AI metadata and crystallizing relationships back to disk without blocking the CLI.

---

### **Phase 1: Database and RPC Infrastructure**

**Initial Problem:** We needed a way to track which sessions were only processed by Regex and which were enriched by AI.

*   **My Assumption/Plan #1:** Add an `enrichment_status` column to the `sessions` table.
    *   **Action Taken:** Created migration `046_devlog_enrichment_status.go`. Status codes: 0 (Pending), 1 (Regex Done), 2 (AI Done), 3 (Failed).
    *   **Action Taken:** Updated `internal/rpc/protocol.go` and `internal/rpc/server.go` to expose `queue_length` via a new `get_enrichment_stats` operation.
    *   **Result:** The daemon can now report how many sessions are waiting for AI.

---

### **Phase 2: Sync Decoupling**

**Initial Problem:** `bd devlog sync` was slow because it called Ollama synchronously.

*   **My Assumption/Plan #1:** Make the CLI sync always use `ForceRegex: true`.
    *   **Action Taken:** Refactored `SyncSession` in `cmd/bd/devlog_core.go`. It now completes in milliseconds, marks the session as `status=1`, and returns control to the user.
    *   **Result:** `git commit` hooks are fast again.

---

### **Phase 3: Daemon Worker Implementation**

**Initial Problem:** We needed a reliable way to run Ollama in the background.

*   **My Assumption/Plan #1:** Implement a serial queue in the daemon loop.
    *   **Action Taken:** Created `cmd/bd/devlog_enrichment.go` with `ProcessEnrichmentQueue`. It picks one session, runs the full pipeline (Ollama + Regex), and updates the DB.
    *   **Challenge:** Crystallization (writing back to disk) changes the file hash, which would normally trigger a re-sync loop.
    *   **Correction:** Updated the worker to re-calculate and save the new file hash to the database immediately after crystallization.
    *   **Challenge:** Placing the worker in the `default` case of the event-driven loop caused starvation due to frequent "Import complete" messages.
    *   **Correction:** Moved the worker to a dedicated goroutine within the daemon, ensuring it runs independently of RPC activity.

---

### **Final Session Summary**

**Final Status:** **Successfully Implemented.** The background pipeline is active.
**Key Learnings:**
*   **Asynchronous UX:** Decoupling slow AI tasks is essential for tool adoption. Users/Agents should never wait for an LLM unless they explicitly asked for a result *now*.
*   **Hash Staleness:** In a system where the tool modifies its own source files (Crystallization), the database MUST proactively update its internal hash cache to prevent infinite sync loops.

---

### **Architectural Relationships**
<!-- Format: [From Entity] -> [To Entity] (relationship type) -->
- daemon -> ProcessEnrichmentQueue (runs in goroutine)
- ProcessEnrichmentQueue -> extractAndLinkEntities (uses AI mode)
- ProcessEnrichmentQueue -> sessions (updates enrichment_status)
- bd status -> get_enrichment_stats (queries queue length)

### Architectural Relationships
<!-- Format: [From Entity] -> [To Entity] (relationship type) -->
- sessions -> enrichment_status (has)
- syncsession -> sessions (updates)
- processenrichmentqueue -> ollama (executes)
- processenrichmentqueue -> regex (executes)
- daemon -> get_enrichment_stats (exposes)
- crystallization -> file hash (updates)
- processenrichmentqueue -> default goroutine (runs_in)
- rpc -> get_enrichment_stats (provides)
