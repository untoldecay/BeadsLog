# [feature] Relationship suggestions + manual links (BeadsLog-58r)

**Date:** 2026-08-09

## Problem
Entities that co-occur strongly in devlogs (e.g. floating-note ↔ editor ↔ slash
on Fray) look isolated because they have no EXPLICIT dependency edge — only
implicit co-occurrence. Wanted a strategy like alias suggestions to hint the
agent to promote strong co-occurrence into real edges, and a way to create edges
that stick and share.

## Work Done (mirrors the alias machinery)
- **Suggestions**: `GetLinkSuggestions` — entity pairs with >= N shared sessions
  and NO entity_deps edge either direction, excluding dismissed + ranked by
  co-occurrence. `bd devlog links suggest [--json --limit]`,
  `bd devlog links dismiss A B` (new `link_dismissals` table, migration 058).
- **Hint**: `showLinkHints` — "🔗 N relationship opportunities — review: bd devlog
  links suggest" printed after graph/search (beside the alias hint).
- **Creation**: `bd devlog link A B --relationship uses` → `AddManualLink`
  inserts an entity_deps edge with **source='user-link'** (a distinct sentinel —
  the schema default 'manual' is what UNSET extracted edges inherit, so a
  separate value is required to tell user edges apart).
- **Sticky + shareable**: manual edges export to `.beads/links.jsonl` (idempotent,
  keyed by entity NAME) alongside aliases; imported on `bd sync`/init and
  re-applied in the devlog-sync path AFTER extraction (so endpoints resolve).
  Survives DB rebuild and travels to teammates — verified with a two-clone test
  (machine B restored the link from git-synced links.jsonl).

## Validation
- Unit: TestLinkSuggestionsAndManualLink (suggest, link removes it, export filters
  to user-link, unknown-entity errors, dismissal permanent).
- Live: full loop + cross-clone durability + clean links.jsonl (only user edges).
- E2E Test 30; suite 30/30 green.

## Relation to other work
Delivers the edges-half of the shareable-graph idea (BeadsLog-5xf): links.jsonl
is a committed, per-record-mergeable graph artifact.

## Final Session Summary
**Final Status:** BeadsLog-58r done on branch dev/graph-panel-ui (batched with
the shadcn panel restyle for one release).
**Key Learnings:**
- entity_deps.source DEFAULTs to 'manual', colliding with unset extracted edges —
  user links need a distinct sentinel ('user-link') to be separable/exportable.
- Re-applying manual edges must happen AFTER entity extraction (devlog-sync path),
  or name→id resolution fails on a rebuilt DB.
