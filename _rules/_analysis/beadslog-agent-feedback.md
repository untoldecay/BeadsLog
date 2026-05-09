# BeadsLog — Agent Feedback Report (Claude Opus 4.6)

**Date:** 2026-05-09
**Context:** Extended usage across a multi-week development session on Fray (Electron + Vue 3 PWA). 440+ sessions in the devlog database, used for grid performance optimization, editor hardening, Eagle import, multi-window sync debugging.

---

## What Works Well

### Search is the killer feature
`bd devlog search` saved real time multiple times during this session. Key examples:
- Searching "caret jump" found 4 historical sessions that documented a fix from April 3 — prevented re-investigating a solved problem
- Searching "self-emit" and "isUpdatingInternally" traced the full history of the editor sync guard system across 6 sessions
- Searching "fetchAllEntities" immediately surfaced the entity refresh storm pattern and its impact chain

The hybrid BM25 + graph search is genuinely useful. The "Did you mean?" suggestions caught several near-misses.

### Entity extraction adds real value
When entities ARE populated, `bd devlog graph` and `bd devlog impact` give architectural insights that would otherwise require reading code. Example: `bd devlog impact "settaskmetadata"` immediately showed downstream effects (useTaskDragDrop, overlaySidebar, fetchAllEntities backfill loop) — this directly informed a fix.

### Low-friction recording
`bd devlog record` is lightweight enough that it doesn't feel like a chore. The `--subject` / `--problem` / `--file` pattern captures enough context without demanding a full write-up.

### The enforcement hook is good (when opted-in)
The pre-commit devlog blocker caught several commits where I would have forgotten to record. For a team that wants complete history, this is the right default. `--no-verify` is available for trivial changes — the balance feels right.

---

## What Needs Improvement

### 1. `bd devlog show` discoverability
I missed this command entirely for the first week. I kept doing `bd devlog search` → finding session IDs → then manually reading the markdown files with `cat`. The `--help` lists it clearly — this is a documentation/onboarding issue, not a tool issue.

**Suggestion:** The first time a user runs `bd devlog search` and gets results, the output could hint: "Use `bd devlog show <session-id>` to read a session."

### 2. Sync warnings are noisy
Every `bd devlog sync` and `bd devlog record` outputs:
```
Missing log session, _rules/_devlog/2026-04-27_overlay-transition-glitch-report.md : open ...
Missing log session, _rules/_devlog/2026-04-28_data-layer-performance.md : open ...
```
These are stale references to deleted files. They appear on every single operation (7 lines each time). After 50+ operations in a session, they become invisible noise.

**Suggestion:** Auto-prune orphaned references after N warnings, or add a `bd devlog cleanup` command that removes stale entries. At minimum, suppress repeated warnings for the same files.

### 3. Graph depth is often 0
Most entities show `(No known dependencies)` in impact analysis. With 442 sessions but only 55% enriched, the graph feels underpopulated. Even after partial enrichment (715 → 943 relationships), specific queries like `bd devlog graph "editor"` still showed depth 0 for most matches.

**Root causes identified:**
- Enrichment daemon seems to stall or not process certain project DBs (3 daemon instances running, queue stuck at 145 remaining)
- Regex-based entity extraction catches component names but misses architectural relationships
- The Ollama-based enrichment is slow (one session at a time) and may not be running for all projects

**Suggestion:**
- `bd devlog status` should show enrichment queue depth and last processing time
- `bd devlog extract --all` should show a progress bar
- Consider a lightweight relationship inference pass that doesn't need Ollama — e.g., if two entities appear in the same session, infer a "co-occurs" relationship

### 4. `--file` parameter confusion
`bd devlog record --file` takes a filename only, not a relative path. I got this wrong on first use (`_rules/_devlog/2026-05-01_file.md` vs just `2026-05-01_file.md`). This is documented in the project memory now but surprised me initially.

**Suggestion:** Accept both formats — strip the directory prefix if present.

### 5. Entity name normalization
Entity names are lowercased and sometimes split inconsistently:
- `isUpdatingInternally` → `updatinginternally` (lost the "is" prefix)
- `fetchAllEntities` → `fetchallentities` (lost camelCase)
- `self-emit` and `` `self-emit` `` are separate entities

This makes searching by exact function/variable name unreliable.

**Suggestion:** Preserve original casing in entity storage, normalize only for matching.

### 6. No way to read session content from search results
`bd devlog search "query"` returns session IDs like `[sess-f8e14d]` but there's no inline preview. You have to run a separate `bd devlog show sess-f8e14d` command. For rapid investigation, this adds friction.

**Suggestion:** `bd devlog search "query" --preview` could show the first 3-5 lines of each matching session.

---

## Usage Patterns That Worked

### The retrieval-led reasoning loop
The CLAUDE.md protocol says "prefer retrieval-led reasoning over training-led reasoning." This is genuinely good advice. Every time I searched devlogs before investigating code, I found relevant history faster than reading source. The protocol works.

### Effective search → graph → code workflow
1. `bd devlog search "topic"` — find relevant sessions
2. `bd devlog graph "entity"` — understand dependencies
3. `bd devlog impact "entity"` — find what might break
4. Only then read the actual code

This order is correct. The tool supports it well when entities are populated.

### Recording as commit prerequisite
The pre-commit hook ensures every code change has a corresponding devlog entry. Over 440+ sessions, this creates a searchable decision history that's invaluable for debugging regressions. "Why was this changed?" is always answerable.

---

## Usage Patterns That Didn't Work

### Issue tracking was unused
`bd ready`, `bd list`, `bd create` — I never used these during the session. All task tracking happened through Claude Code's built-in TaskCreate/TaskUpdate or through conversation context. The devlog side dominated completely.

**Why:** Issues feel like a separate workflow from the devlog flow. Devlog entries are created naturally alongside commits. Issues require proactive planning. In a fast iteration session, the devlog captures everything — issues feel redundant.

**Possible improvement:** Auto-create issues from devlog entries that mention "TODO", "DEFERRED", or "KNOWN ISSUE". This would bridge the gap without requiring manual issue creation.

### Graph queries during active debugging
When actively debugging (e.g., the task caret jump investigation), I needed answers fast. `bd devlog graph` results with depth 0 weren't helpful — I fell back to grep/search. The graph is better for architectural planning than for urgent debugging.

---

## Feature Requests (Prioritized)

### High priority
1. **`bd devlog search --preview`** — Show first few lines of matching sessions inline
2. **Stale reference auto-cleanup** — Stop warning about the same deleted files on every operation
3. **Enrichment queue visibility** — `bd devlog status` should show: queue depth, last processed session, estimated time remaining
4. **Accept both filename and path in `--file`** — Strip directory prefix if present

### Medium priority
5. **Entity name preservation** — Store original casing, normalize only for matching
6. **`bd devlog timeline`** — Chronological view with 1-line summaries per session
7. **Co-occurrence relationships** — If entities appear in the same session, infer a lightweight relationship without AI
8. **Auto-issue from devlog** — Extract "DEFERRED" / "TODO" / "KNOWN ISSUE" patterns into issues

### Low priority
9. **`bd devlog diff sess-a sess-b`** — Compare what changed between sessions
10. **`bd devlog link entity-a entity-b`** — Manual relationship creation for cases the AI misses

---

## Overall Assessment

BeadsLog's devlog search is the standout feature — it turns commit history into searchable project memory. The 440-session corpus became a genuine knowledge base that informed real debugging decisions. The enforcement hook ensures completeness without being oppressive.

The graph/impact system has the right architecture but needs more data (enrichment coverage) and better discoverability to reach its potential. When it works (populated entities with relationships), it provides insights that no other tool gives. When it doesn't (depth 0, no dependencies), it feels like a dead end.

**Would I install it from a store?** If the README led with the search use case ("find a decision from 2 months ago in 3 seconds"), yes. If it led with the graph/entity system, probably not — that feature needs the enrichment pipeline working reliably to deliver on its promise.

**Bottom line:** The devlog side is production-ready and valuable today. The graph side is architecturally sound but needs enrichment reliability and entity quality improvements to match.
