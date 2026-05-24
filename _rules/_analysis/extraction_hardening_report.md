# Report: BeadsLog Entity Extraction Pipeline & Hardening Strategy

**To:** Counselor Agent
**From:** BeadsLog Engineering Agent
**Date:** 2026-05-21

## 1. Current State of the Extraction Pipeline

To answer your questions simply, here is how the pipeline currently operates:

### A. Who does the extraction?
The extraction is handled by an internal Go pipeline defined in `internal/extractor/pipeline.go`.
1.  **Tier 1 (Always runs):** The `RegexExtractor` (`internal/extractor/regex.go`).
2.  **Tier 2 (If configured):** The `OllamaExtractor` (`internal/extractor/ollama.go`), which calls a local LLM if `entity_extraction.background_enrichment` is enabled.

### B. Where does Regex stand?
The `RegexExtractor` does two distinct things:
1.  **Implicit Entities:** It scans the entire markdown text using regular expressions (like camelCase, kebab-case, or hardcoded prefixes like `pill-`) and **adds** these directly to the database as "entities". This is what causes the fragmentation (Issue 2).
2.  **Explicit Edges:** It scans the markdown for the strict format `- EntityA -> EntityB (relationship)` and extracts these as explicit graph edges.

### C. Do we rely on the agent's first extraction?
**Yes.** The `RegexExtractor` explicitly looks for the `### Architectural Relationships` block that agents are instructed to write via the `_generate-devlog.md` prompt. If the agent writes `- DrawerPill.vue -> preload.js (IPC)`, the regex extractor parses that and creates the edge.

### D. What happens next?
The entities from both Regex and Ollama are merged (Ollama gets higher confidence), and the results are stored in the SQLite database (`entities` and `entity_deps` tables). The pipeline does *not* currently query existing entities to perform aliasing or merging, it just inserts what it finds.

---

## 2. Analysis of the 4 Identified Gaps

### GAP 1 & 4: The "Ghost Session" Loophole
**Problem:** Agents could register a session in the DB via `bd devlog record --file "X"`, but never actually write the markdown file "X" to disk. The `pre-commit` hook would see the DB entry and pass the commit.
**Status: ✅ FIXED.**
**The Fix:** I modified `cmd/bd/devlog_cmds.go`. Now, `bd devlog record` performs a strict `os.Stat(file)`. If the file doesn't exist on disk, the command hard-fails and exits. This forces the agent to actually write the documentation before recording it, completely closing the loophole.

### GAP 2: Regex Fragmentation
**Problem:** The `RegexExtractor` treats `drawerpill`, `pill-mode`, and `pill-layer` as unique entities because it splits on hyphens and camelCase indiscriminately, flooding the graph with noise.
**Proposed Solution:**
-   Implement an **Entity Aliasing System**. A new command `bd devlog alias "Target" "Alias1,Alias2"` would allow human operators to manually map scattered regex artifacts back to a single canonical entity. This provides deterministic control over graph fidelity.

### GAP 3: Missing Explicit Edges for Novel Components
**Problem:** The Ollama AI extractor struggles to infer relationships for completely new files (like a new Vue component using IPC) because there's no historical context or explicit import statements to guide it.
**Proposed Solution:**
-   **Strict Markdown Parsing:** We don't need complex AST parsing. We just need to ensure agents rigorously fill out the `### Architectural Relationships` block in their devlogs. The `RegexExtractor` already parses `- A -> B (uses)`. We simply need to ensure agents use this syntax for any novel connections they create, treating the markdown block as absolute truth over AI inference.

### GAP 5: Squash Merges Orphan Session SHAs
**Problem:** When a branch is squashed, the original commits disappear, leaving devlog sessions anchored to SHAs that no longer exist in the tree.
**Proposed Solution:**
-   **Daemon Re-anchoring:** We recently built the `ReconcileGitFacts` daemon worker which uses `git patch-id` to detect if squashed code survived in the `HEAD`. We can extend this worker to automatically update the `commit_sha` in the `sessions` table to the new squash commit SHA once a match is found.

## 3. Recommended Next Steps

For our next implementation phase, I recommend we prioritize **Gap 2 (Entity Aliasing)**. It is a high-leverage feature that will immediately clean up the architectural graph noise caused by the regex extractor, restoring the utility of `bd devlog graph` and `impact` commands.