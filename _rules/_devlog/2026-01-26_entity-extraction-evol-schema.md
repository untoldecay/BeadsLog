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

---

### **Phase 4: Ollama Integration**

**Initial Problem:** With the architecture in place, I needed to implement the actual LLM-based extraction using Ollama to fulfill the "Tier 2" requirement.

*   **My Assumption/Plan #1:** I would use the `github.com/ollama/ollama/api` client to connect to a local Ollama instance.
    *   **Action Taken:** Added the dependency via `go get`. Created `internal/extractor/ollama.go`.
    *   **Challenge:** The `GenerateRequest.Format` field expects a `json.RawMessage` (byte slice) but I initially tried to pass a string "json".
    *   **Correction:** Updated the code to use `json.RawMessage(`"json"`)`.
*   **My Assumption/Plan #2:** I needed to update the `Pipeline` to support multiple extractors and merging.
    *   **Action Taken:** Refactored `NewPipeline` to accept an optional `ollamaModel` string. If provided, it initializes the `OllamaExtractor` and adds it to the list.
    *   **Logic:** The pipeline runs all extractors. Since `RegexExtractor` runs first, its results populate the map. When `OllamaExtractor` runs (if successful), it updates the map. Since Ollama entities have higher confidence (1.0 vs 0.8), they naturally override regex matches, while unique regex matches (fallback) are preserved.
*   **My Assumption/Plan #3:** Configuration should drive the enablement of Ollama.
    *   **Action Taken:** Updated `internal/config/config.go` with default settings for `entity_extraction` and `ollama`. Updated `cmd/bd/devlog_core.go` to read these configs and pass the model to the pipeline.

**Result:** The system now supports a hybrid extraction pipeline. If Ollama is configured and running, it enhances extraction quality. If not (or if configured off), it gracefully degrades to the robust Regex fallback.

---

### **Phase 5: Config Fix and Test**

**Initial Problem:** \`bd config set\` wasn't persisting Ollama settings to \`config.yaml\`, causing \`bd devlog sync\` to default to Regex-only mode.

*   **My Assumption/Plan #1:** Keys must be registered in \`YamlOnlyKeys\` to be written to YAML.
    *   **Action Taken:** Added \`ollama.\` and \`entity_extraction.\` prefixes to \`YamlOnlyKeys\` in \`internal/config/yaml_config.go\`.
    *   **Result:** \`bd config set\` now correctly updates \`config.yaml\`.

**Initial Problem:** Entity source wasn't updating when Ollama boosted confidence.
*   **Analysis:** The SQL \`ON CONFLICT\` clause only updated \`confidence\`, leaving \`source\` as 'regex' (default) even if confidence rose to 1.0.
    *   **Action Taken:** Updated SQL in \`cmd/bd/devlog_core.go\` to conditionally update \`source\` when \`excluded.confidence > confidence\`.

**Testing:**
*   Created dummy session \`2026-01-26_ollama-test.md\` with content about Redis/Memcached.
*   Synced with \`ministral-3:3b\`.
**Result:** 16 entities extracted (vs 3 with regex), capturing "redis", "memcached", "pgvector".

---

### **Phase 6: Retrofit & Repair Workflow**

**Initial Problem:** Existing devlogs were missing from `_index.md` (orphans) or lacked the new entity metadata. `bd devlog sync` only processes changed files, leaving old sessions outdated.

*   **My Assumption/Plan #1:** `bd devlog verify --fix` should be the active repair tool.
    *   **Action Taken:** Updated `cmd/bd/devlog_cmds.go`:
        1.  **Orphan Detection:** Scans the `_devlog` directory for `.md` files missing from `_index.md`.
        2.  **Adoption:** Automatically appends orphans to `_index.md` with parsed date/title.
        3.  **Backfill:** Iterates over sessions missing metadata and force-runs `extractAndLinkEntities` using the original narrative.
    *   **Challenge:** The `OllamaExtractor` parser failed on real data from `ministral-3:3b` because the model output inconsistent JSON (sometimes arrays for names, sometimes headers).
    *   **Correction:** Refined the prompt in `internal/extractor/ollama.go` to enforce a strict schema and updated the parser to handle `json.RawMessage` for robust unmarshaling.
    *   **Challenge:** `hashID` panicked on short strings (length < 6).
    *   **Correction:** Updated `hashID` in `cmd/bd/devlog_core.go` to use `%06x` padding.

**Result:** `bd devlog verify --fix` now successfully adopts orphaned files and enriches old sessions with high-confidence entities from Ollama.

---

### **Phase 7: Verify UX Enhancements**

**Initial Problem:** Backfilling large histories with AI is slow (14s/session -> ~8m for 34 sessions), potentially locking the user. Also, users couldn't target specific sessions for repair.

*   **My Assumption/Plan #1:** Add a "Fast-Path" flag and targeting support.
    *   **Action Taken:** Updated `extractAndLinkEntities` to accept `ExtractionOptions` (supporting `ForceRegex`).
    *   **Action Taken:** Updated `bd devlog verify`:
        1.  Added `--fix-regex` flag to bypass AI.
        2.  Added argument support `[target]` to filter sessions by ID/filename.
        3.  Added a UX disclaimer when running AI backfill on multiple files.

**Result:** Users can now perform instant repairs with regex (`--fix-regex`) or surgically repair specific sessions with AI (`verify sess-123 --fix`).

---

### **Phase 8: Documentation Polish**

**Initial Problem:** The help text for `bd devlog verify --fix` was outdated ("Generate re-investigation directive") and didn't reflect its new active repair capabilities.

*   **Action Taken:** Updated the flag usage description in `cmd/bd/devlog_cmds.go`.

**Result:** `bd devlog verify --help` now correctly describes `--fix` as "Adopt orphans and backfill missing metadata".

---

### **Phase 9: Semantic Relationship Extraction**

**Initial Problem:** Entities were being extracted, but the "Architectural Knowledge Graph" was limited to explicit `A -> B` arrows in the text. The LLM had the context to infer relationships (e.g., "uses", "configures") but wasn't asked for them.

*   **My Assumption/Plan #1:** Upgrade the Ollama prompt to request a `relationships` array.
    *   **Action Taken:** Updated `internal/extractor/ollama.go` prompt to ask for `{"from": "A", "to": "B", "type": "rel"}`.
    *   **Action Taken:** Updated `Extractor` interface and `Pipeline` to propagate these relationships.
    *   **Action Taken:** Updated `RegexExtractor` to match the new interface (wrapping its existing logic).

**Result:** `bd devlog sync` now extracts semantic relationships. Verified with `sess-d4fa27` where it correctly inferred `extraction-pipeline -> regex-extractor (uses)` from code snippets.






