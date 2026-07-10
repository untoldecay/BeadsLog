# [feature] Batch 2: Graph Trust & Interactive HTML Graph

**Date:** 2026-07-10

## Problem
The entity graph — the backbone of the agent MAP step — was polluted at the source: the kebab-case regex matched mid-word ("Multi-layered" → `ulti-layered`), phrase fragments like `map it step` entered via manual arrows, separator variants (`ollama-extractor` vs `ollamaextractor`) fragmented the graph, the search entity bonus was hardcoded to 0, and humans had no visual way to explore the graph.

### **Phase 1: Extraction-Time Noise Filtering**

*   **Action Taken:** Word-bounded the kebab pattern in internal/extractor/regex.go so mid-word fragments can't match.
*   **Action Taken:** Added `IsNoise()` (internal/extractor/noise.go): rejects sub-3-char names, pure numbers, malformed edges, and multi-word names made only of common English words. Applied in the pipeline merge so regex, Ollama, and manual-arrow sources are all filtered.
*   **Action Taken:** Added `bd devlog prune --noise` to purge pre-existing junk entities from the DB (removed 6 on this repo).

### **Phase 2: Auto-Alias Separator Variants**

*   **Action Taken:** `AutoAliasDuplicates()` (internal/queries/entities.go) merges entities whose names are identical after separator normalization, canonical = highest mention_count, via the existing transactional `AliasEntities()`. Variant names persist in the `entity_aliases` registry so future extractions resolve automatically; `unalias` undoes.
*   **Action Taken:** Hooked into `devlog sync`; on this repo the first run merged 33 duplicate variants.
*   **Safety Decision:** Only exact normalized matches auto-merge. Near matches (`ollama` vs `ollamaextractor`) remain suggestions — Jaccard co-occurrence alone cannot distinguish "same component" from "components that always ship together".

### **Phase 3: Entity Bonus Wired**

*   **Action Taken:** `sessionsLinkedToEntities()` pre-fetches sessions graph-linked to entities matching query tokens (one IN-query, O(1) lookup per row); `EntityBonus = 0.75` replaces the hardcoded 0 in internal/queries/search.go.
*   **Result:** Graph-linked sessions now outrank text-only matches; visible via `search --explain`.

### **Phase 4: Obsidian-Style HTML Graph**

*   **Action Taken:** `bd devlog graph <entity> --html <path>` (cmd/bd/devlog_graph_html.go) exports a standalone interactive force-directed graph (force-graph via CDN): drag/zoom/hover, explicit deps solid with arrows, co-occurrence dashed, depth-colored nodes, dark theme.
*   **Note:** Dropped cobra NoOptDefVal after testing — `--html path` (space syntax) silently became a positional arg; explicit path argument is agent-safer.

### **Phase 5: Tests**

*   **Action Taken:** internal/extractor/noise_test.go (IsNoise cases, mid-word truncation regression, pipeline filtering), TestAutoAliasDuplicates (internal/queries), and 4 CLI tests in cmd/bd/cli_devlog_batch2_test.go (sync auto-merge, entity bonus via --json, prune --noise, graph --html).
*   **Result:** All green. Short-suite deltas (TestInitCommand, TestOutputContextFunction) verified pre-existing on pristine v0.54.0 via baseline worktree.

## Final Session Summary

**Final Status:** Batch 2 shipped on `feature/enhance-fable` (issue BeadsLog-6bq). Real-repo impact: 33 variants merged, 6 junk entities purged, top-entities list now shows clean canonical names with proper casing.
**Key Learnings:**
*   **Filter at extraction, not display:** display-time SQL filters left junk in `session_entities`/`entity_deps` where every graph query inherited it.
*   **Name identity beats co-occurrence for auto-merge:** normalized-equal names are proof of duplication; co-occurrence is only correlation.

### Architectural Relationships
- RegexExtractor -> IsNoise (filters)
- pipeline -> IsNoise (uses)
- devlog sync -> AutoAliasDuplicates (triggers)
- AutoAliasDuplicates -> AliasEntities (uses)
- HybridSearch -> sessionsLinkedToEntities (scores)
- devlog graph -> writeGraphHTML (exports)
