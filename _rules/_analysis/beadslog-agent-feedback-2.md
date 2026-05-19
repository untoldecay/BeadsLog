# BD Devlog Feature Feedback — Agent Perspective
**Date:** 2026-05-10
**Context:** Used bd devlog extensively during "Send to Note" feature implementation and "Active Note Instant Entity Refresh" investigation across a multi-day session.

## Features Used

### Search (with inline previews)
**Rating: Strong improvement**

The new search results format showing `session ID + subject + snippet` inline is a major upgrade. Queries like `bd devlog search "entity refresh storm cooldown" --strict` returned exactly 1 hit with enough context to triage before reading the full session. Previously would have required `show` on each session to determine relevance.

The `Related Entities (Matched via FTS)` box that appears when no exact match is found is particularly valuable — it acts as a discovery mechanism, pointing toward entity names I didn't know existed. Example: searching "active note" surfaced `activenotes`, `set-active`, `noteeditorview`.

**Suggestions:**
- Add `--limit N` to cap results for broad queries (e.g., "corruption" returned 22 sessions — top 5 would suffice in most cases)
- Show the session date in the inline preview (currently only visible in `show`)

### Graph
**Rating: High value, used most frequently**

`bd devlog graph "TaskOverlay"` instantly mapped the component dependency tree (3 depth levels, 15+ entities) without reading any source files. Used it to understand architecture before writing code for: `notesstore`, `editorcontextmenus`, `useTaskActions`, `moremenu`, `universalpicker`, `extractorservice`, `contextualmenu`.

Each graph call saved 2-3 Grep/Read cycles and provided the "shape" of a subsystem in one shot.

**Suggestions:**
- Allow filtering by relationship type (e.g., `--type uses` vs `--type contains`)
- The fuzzy matching is good — `"TaskOverlay"` matched 3 variants. Keep this behavior.

### Impact
**Rating: Useful but underused**

`bd devlog impact "editorcontextmenus"` → `tsstackwysiwygeditor (uses)` — one call, instant answer to "what depends on this?". I used it less frequently than `graph` because `graph` already shows dependencies. Impact is most valuable when assessing blast radius of a change.

### Show
**Rating: Essential, but one data quality issue**

`bd devlog show sess-302c5a` gave me the exact historical context for why `fetchAllEntities` was added after `updateNote` (tasks not appearing in tray after paste). This decision provenance was critical for understanding whether I could safely bypass the batched path.

**Issue:** `sess-79bd3f` was tagged as "[perf] skip redundant fetchAllEntities on cache hit, batch entity refreshes" but `show` returned "Search UX Overhaul" content. The session metadata and content didn't match — possible indexing issue where two sessions share a log file or the enrichment overwrote the wrong session.

### Path (new)
**Rating: Promising but blocked by entity resolution**

Tried `bd devlog path "refreshnoteentities" "tray"` → "One or both entities not found." The entity names need exact matches from the graph DB, but there's no fuzzy fallback like `graph` has. The concept is powerful — tracing how two architectural concepts are connected through historical sessions — but it needs the same fuzzy matching that makes `graph` usable.

**Suggestions:**
- Add fuzzy matching (same as `graph`)
- Show "Did you mean?" suggestions (like `graph` already does for some queries)
- Consider allowing session-based paths, not just entity-based

### Entities
**Rating: Low signal for discovery**

Top entities list was dominated by generic terms (`service: 5630`, `used: 4500`, `border-radius: 4703`) rather than architectural components. The list reads more like a word frequency counter than an architecture index.

**Suggestions:**
- Filter by minimum relationship count (not just mentions)
- Exclude CSS properties and common English words
- Add `--type component|service|store|composable` filtering
- Show entities with the most *relationships*, not just mentions

### Resume
**Rating: Useful for session continuity**

`bd devlog resume --last 1` loaded the previous session's context in one call. Good for cold starts.

### Record + Sync
**Rating: Smooth workflow**

`bd devlog record --subject "..." --problem "..." --file "..."` + automatic sync worked reliably. The git hook enforcement ensures devlog discipline. The `bd devlog sync` at the end of `record` is a nice touch — no manual step needed.

## Quantitative Impact

For the "active note instant entity refresh" investigation:

| Metric | With bd devlog | Without bd devlog (estimated) |
|--------|---------------|-------------------------------|
| Queries/tool calls | ~8 (search + graph + show) | ~20 (grep + read + trace) |
| Time | ~15 min | ~30 min |
| Tokens (tool I/O) | ~10K | ~35-40K |
| Historical context | 4 decision points recovered | 0 (code comments only) |
| Confidence in fix | High (understood constraints) | Medium (would be more conservative) |

## Key Insight

The devlog's biggest value isn't speed — it's **decision provenance**. Knowing *why* the 2s cooldown exists (prevent storms in 5000-note libraries) meant I could confidently bypass it for active notes without re-introducing the original problem. Without that context, I'd have proposed a less effective fix or added unnecessary guards.

## Summary

| Feature | Value | Maturity |
|---------|-------|----------|
| Search (inline previews) | High | Production-ready |
| Graph | High | Production-ready |
| Impact | Medium | Production-ready |
| Show | High | Data quality issue on some sessions |
| Path | High potential | Needs fuzzy entity matching |
| Entities | Low | Needs filtering/ranking improvements |
| Resume | Medium | Production-ready |
| Record + Sync | High | Production-ready |
