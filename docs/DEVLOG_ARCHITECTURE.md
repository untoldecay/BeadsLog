# Devlog Architecture

This document describes the technical architecture of the BeadsLog devlog system—the engine that transforms work narratives into a permanent architectural knowledge graph.

## The Knowledge Pipeline

BeadsLog uses a 2-tier, asynchronous pipeline to process "Beads" (work sessions). The goal is to keep the user's workflow instant while ensuring the graph is enriched with deep AI intelligence.

```text
┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
│  Markdown File  │─────▶│   bd devlog sync  │─────▶│ SQLite Database  │
│  (_rules/_devlog)│      │   (Regex Only)  │      │  (Knowledge Index)│
└─────────────────┘      └────────┬────────┘      └────────┬────────┘
                                  │                        │
                           Sets Status: 1                  │
                           (Regex Done)                    │
                                  │                        │
                                  v                        │
                         ┌─────────────────┐               │
                         │   bd daemon     │◀──────────────┘
                         │ (Background AI) │
                         └────────┬────────┘
                                  │
                          1. Ollama Extract
                          2. Crystallize (Write-back)
                          3. Update Hash & Status: 2
                                  │
                                  v
                         ┌─────────────────┐
                         │  Markdown File  │
                         │ (With Arrows!)  │
                         └─────────────────┘
```

## Tiered Extraction Engine

The extraction logic is encapsulated in the `internal/extractor` package and follows a "Fast-First, Smart-Later" strategy.

### Tier 1: Regex Extractor (Instant)
- **Role:** Synchronous fallback.
- **Speed:** <1ms.
- **Logic:** Scans for explicit `A -> B` arrow patterns and known technical keywords (e.g., `AuthService`, `nginx.conf`).
- **Confidence:** 0.8.

### Tier 2: Ollama Extractor (Semantic)
- **Role:** Asynchronous enrichment.
- **Speed:** 10s - 30s per session.
- **Logic:** Uses a local LLM (e.g., `llama3.2:1b`) to read the prose and infer relationships that aren't explicitly typed as arrows.
- **Confidence:** 1.0 (Boosted).

## Background Enrichment Worker

To prevent the slow LLM from blocking the CLI, the `bd daemon` runs a dedicated **Enrichment Worker**.

1. **Queueing:** Every time a session is synced, it is marked with `enrichment_status = 1`.
2. **Serial Processing:** The worker pulls the most recent "Regex-only" session from the database.
3. **Extraction:** It runs the full AI pipeline.
4. **Crystallization:** Discovered relationships are formatted into Markdown and appended to the original file.
5. **Hash Synchronization:** After writing to disk, the worker re-calculates the file's content hash and updates the SQLite record. This prevents the next `sync` from thinking the file was "manually" changed.

## Data Schema

Devlog data is stored in four primary SQLite tables:

| Table | Purpose |
| :--- | :--- |
| `sessions` | Stores the narrative, file hash, `author`, `agent`, and `enrichment_status`. |
| `entities` | The unique components (Services, Technologies) with confidence scores. |
| `session_entities` | A join table linking sessions to the entities they mention. |
| `entity_deps` | The directed graph of relationships (the "Edges"). |

### ID Stability (Lookup-by-Filename-and-Title)
To prevent duplicate sessions when metadata (like Time or Author) is updated in the index, BeadsLog uses a **Lookup-by-Filename-and-Title** strategy. Before generating a new hash-based ID, the system checks if a session with the same filename and subject already exists in the database. If found, it preserves the original ID while updating the metadata.

### 6-Column Index Standard
As of v0.35+, the `_index.md` uses a 6-column Markdown table:
`| Subject | Problems | Author | Agent | Date | Devlog |`
This allows for precise tracking of both the human responsible for the work and the AI agent that executed it. Legacy 4 or 5-column indexes are automatically migrated during `bd init` or `bd devlog initialize`.

Most AI tools store "memory" in a hidden database. BeadsLog prioritizes **Crystallization**:
- **Portability:** Your architectural graph is version-controlled in Git.
- **Transparency:** You can see and edit the relationships the AI discovered.
- **Resilience:** If the SQLite database is deleted, the graph is instantly rebuilt from the Markdown files using the fast Regex extractor.

## Performance & Scaling
By offloading AI to the daemon, the system can scale to hundreds of devlog sessions without impacting CLI responsiveness. The daemon uses a debounced flush mechanism to ensure the Git-tracked JSONL remains in sync with the AI's discoveries.
