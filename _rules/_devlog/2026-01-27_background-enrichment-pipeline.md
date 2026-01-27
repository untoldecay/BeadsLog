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

---

### **Phase 11: Regeneration & Enrichment Controls**

**Initial Problem:** Users needed a way to "upgrade" existing devlogs with new AI extraction logic (e.g., when improving prompts or switching models) without waiting for a file change.

*   **Action Taken:** Implemented `bd devlog extract [target]`. This is a foreground command that force-runs the full AI + Regex pipeline on any session, even if it's already "complete."
*   **Action Taken:** Implemented `bd devlog enrich [target] [--all]`. This command resets the `enrichment_status` to 1, effectively "re-scheduling" sessions for background processing by the daemon.
*   **Action Taken:** Refined **Crystallization** logic to support merging. It now identifies existing `### Architectural Relationships` blocks and appends new discoveries without duplicating headers or existing arrows.

**Result:** The "Retrofit" workflow is now complete. Users can surgically upgrade a single session in the foreground or trigger a full project-wide background enrichment.


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

---

### **Phase 10: Protocol Hardening**

**Initial Problem:** Agents were bypassing the Devlog system and jumping straight to manual file searching, missing the architectural context provided by the graph.

*   **Action Taken:** Updated `FullBootloader` and `ProtocolMdTemplate` in `cmd/bd/init_templates.go` to explicitly list mandatory starting commands (`bd sync`, `bd devlog sync`, `bd devlog verify --fix`).
*   **Action Taken:** Updated `cmd/bd/onboard.go` activation guide to enforce a "Memory First" policy, explicitly telling agents to STOP and map the landscape before using `ls` or `grep`.
*   **Action Taken:** Standardized `_rules/_orchestration/PROTOCOL.md` and `WORKING_PROTOCOL.md` to mandate the "Map It" step at the beginning of every task.

**Result:** The agent instructions now programmatically enforce the intended workflow, ensuring the knowledge graph is used as the primary source of truth.

- fullbootloader -> protocolmdtemplate (uses)
- protocolmdtemplate -> bd sync (mandates)
- protocolmdtemplate -> bd devlog sync (mandates)
- protocolmdtemplate -> bd devlog verify --fix (mandates)
- onboard.go -> memory first policy (enforces)
- memory first policy -> map it step (requires)
- working_protocol.md -> map it step (mandates)
