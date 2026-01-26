# Comprehensive Development Log: Entity Extraction Evolution and Schema Migration

**Date:** 2026-01-26

### **Objective:**
To implement the "Entity Extraction with Ollama + Regex Fallback" feature (Issue bd-byg), specifically focusing on the architectural foundation: schema updates to support extraction metadata (confidence, source) and refactoring the extraction logic into a modular pipeline.

---

### **Phase 1: Architecture Analysis and Schema Design**

**Initial Problem:** The existing entity extraction logic was hardcoded in `cmd/bd/devlog_core.go` and only supported simple regex patterns. The `entities` table lacked fields to distinguish between high-confidence matches (from LLM) and lower-confidence ones (from regex), or to track the extraction source.

*   **My Assumption/Plan #1:** I needed to modify the `entities` table to add `confidence` and `source` columns. I also needed an `extraction_log` table to track performance metrics as per the PRD.
    *   **Action Taken:** I created a new migration file `internal/storage/sqlite/migrations/045_extraction_evol_schema.go`.
    *   **Result:** The migration adds `confidence` (REAL, default 1.0) and `source` (TEXT, default 'regex') to `entities`. It also creates the `extraction_log` table.
    *   **Analysis:** This schema change is additive and backward compatible. The default values ensure existing data remains valid.

---

### **Phase 2: Refactoring Extraction Logic**

**Initial Problem:** The extraction logic was buried in `cmd/bd/devlog_core.go`, making it hard to extend with an Ollama-based extractor later.

*   **My Assumption/Plan #1:** I should create a dedicated `internal/extractor` package to house the extraction pipeline.
    *   **Action Taken:** Created `internal/extractor/types.go` (defining `Entity`, `Extractor` interface), `internal/extractor/regex.go` (porting existing logic), and `internal/extractor/pipeline.go` (orchestrator).
    *   **Result:** A clean interface `Extractor` that returns `[]Entity` with metadata. The `Pipeline` runs all configured extractors (currently just regex) and merges results.
    *   **Correction:** Initially, I copied the *old* regex patterns. During testing, I realized I needed to add the *new* patterns defined in the PRD (e.g., `nginx[\w-]*`, `use[\w]+`) to `RegexExtractor` to meet the requirements.

---

### **Phase 3: Integration and Testing**

**Initial Problem:** I needed to ensure the new pipeline integrated correctly with the existing `SyncSession` workflow and that the regex patterns worked as expected.

*   **My Assumption/Plan #1:** I tried to write an integration test `cmd/bd/devlog_extraction_test.go` that spun up a temporary SQLite DB.
    *   **Action Taken:** Wrote the test and tried to run it with `go test`.
    *   **Result:** `go test` failed because I couldn't easily include all the dependencies of `package main` (cmd/bd) in the test run.
    *   **Analysis/Correction:** I realized testing the *logic* was more important and easier than testing the *CLI integration* at this stage. I switched to writing a unit test `internal/extractor/pipeline_test.go` which tested the `Pipeline` and `RegexExtractor` in isolation.
*   **My Assumption/Plan #2:** The unit test would pass with the ported regex.
    *   **Action Taken:** Ran the unit test.
    *   **Result:** Failed. Entities like "nginx" and "useSortable" were not found.
    *   **Analysis/Correction:** This confirmed that the *old* regex patterns were insufficient. I updated `internal/extractor/regex.go` to include the expanded patterns from the PRD. The test then passed.

---

### **Final Session Summary**

**Final Status:**
*   Schema migration `045_extraction_evol_schema` created and registered.
*   `internal/extractor` package implemented with `RegexExtractor` and `Pipeline`.
*   `cmd/bd/devlog_core.go` refactored to use the new pipeline.
*   Extraction logic now supports confidence scores and source tracking.
*   Regex patterns expanded to cover more architectural entities (nginx config, hooks, services).

**Key Learnings:**
*   **Refactoring before Feature Add:** Moving the extraction logic to its own package *before* adding the complex Ollama logic was crucial. It clarified the interface and made testing the regex fallback much easier.
*   **Test Strategy:** When working in `main` packages with heavy dependencies, it's often better to extract logic to a library package (`internal/...`) and test it there, rather than trying to wrangle `go test` for the binary's main package.

---

### **Architectural Relationships**
<!-- Format: [From Entity] -> [To Entity] (relationship type) -->
- RegexExtractor -> Extractor (implements)
- Pipeline -> Extractor (uses)
- SyncSession -> Pipeline (uses)
- entities -> confidence (has column)
- entities -> source (has column)
