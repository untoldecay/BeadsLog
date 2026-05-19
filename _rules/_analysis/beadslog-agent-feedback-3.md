# Devlog-First Protocol: Run Comparison Report
**Date:** 2026-05-17
**Context:** Testing mattpocock/skills `improve-codebase-architecture` with and without bd devlog protocol

---

## Run 1: No devlog protocol (Explore agent)

### bd commands used: 2
- `bd devlog entities` (1×)
- `bd devlog resume --last 1` (1×)
- Immediately fell back to Explore agent → glob, grep, file reads across 100+ files

### What it produced
A generic line-count audit. "UnifiedContentView is 3,599 lines, split it." No awareness of prior modularization work, deferred decisions, or existing architecture PRDs.

### Critical misses
- Proposed decomposing TsStackWysiwygEditor into MarkdownParser + FormattingEngine — **already done** (Phases 1-4, eventModules/, formattingRules/)
- Proposed extracting TaskRowItem, TrayTaskList from UnifiedContentView — **already done** (Phase 1, sess-25f818)
- Proposed "Repository pattern for database.js" — no awareness of the sync/iCloud retry logic
- Proposed extracting MentionAutocomplete from editor — already in `useMentions` composable
- Zero references to any prior session, decision, or architectural rationale

---

## Run 2: Devlog-first protocol

### bd commands used: ~22
| Command | Count | Purpose |
|---|---|---|
| `bd devlog entities` | 2× | By significance + by mentions |
| `bd devlog graph` | 4× | UCV, editor, notesstore, drawerpanelview, calendartray |
| `bd devlog impact` | 4× | UCV, editor, notesstore, calendartray |
| `bd devlog search` | 8× | architecture, refactor, modularization, decompose, coupling, WYSIWYG phases, events.js, database.js |
| `bd devlog resume --last 3` | 1× | 3 sessions, not 1 |
| `bd devlog path` | 2× | UCV↔notesstore, drawerpanelview↔editor |
| `bd devlog show` | 1× | Attempted, file missing from worktree |
| **Fallback** | — | Targeted `wc -l` and `ls` to verify current state against devlog claims |

### What it produced
Contextual analysis that knows:
- What was already extracted and what remains
- Why UCV grew back (Phase 6 was explicitly deferred)
- The editor's Facade pattern is an intentional decision, not tech debt
- The TextKit 2 PRD shapes what sync loop extraction should look like
- Props passthrough was already documented as debt in the devlog graph
- `setTaskMetadata` has exactly 3 callers (from impact analysis, not grep)

---

## Comparison

| Dimension | Run 1 | Run 2 |
|---|---|---|
| bd commands | 2 | ~22 |
| Prior work awareness | None — re-proposed completed work | Full — references 7 prior sessions |
| Deferred decisions | Unaware | Cited "Phase 6 deferred" decision |
| Architecture PRDs | Unaware | Aligned with TextKit 2 PRD |
| Dependency data source | Import-grep (manual) | Devlog graph (pre-mapped) |
| False positives | ~3 (proposed already-done extractions) | 0 |
| Actionability | "Split this big file" | "Continue Phase 6, extract sync loop aligned with PRD" |

---

## The Merge-State Blindspot

**Critical finding:** Run 2 referenced prior modularization sessions (sess-b0e613, sess-66309e, sess-25f818) as "complete" — but these extractions live on **archived branches that were never merged to develop**. The devlog records them as completed work, but the code on the active branch still contains the monolithic versions.

This means:
- The devlog said "Phase 1-4 modularization complete" → agent assumed it was in the codebase
- The agent verified line counts (4,305 lines in index.vue) and noted "it grew back" — but the real explanation is simpler: **the extractions were never merged**
- Recommendations like "continue Phase 6" assume Phases 1-5 are in place — they aren't on the active branch

### Impact on analysis quality
Run 2 was still significantly better than Run 1 (it knew the work existed, knew the architectural rationale, aligned with PRDs). But it drew a wrong conclusion about *why* the monolith persists: it's not that UCV "grew back by 1,350 lines" — it's that the extraction branch was archived before merge.

### What beadslog needs
A **merge-state flag** on devlog sessions so agents can distinguish:
- Work that was completed AND merged (safe to build on)
- Work that was completed but NOT merged (exists only on archived/feature branches)
- Work that was planned but never started

See: `2026-05-17_beadslog-merge-state-proposal.md` for the full feature proposal.

---

## Conclusion

The devlog-first protocol is a clear improvement over blind exploration. But devlog sessions currently lack branch/merge metadata, which can lead agents to treat archived experimental work as established baseline. Adding merge-state awareness to beadslog would close this gap.
