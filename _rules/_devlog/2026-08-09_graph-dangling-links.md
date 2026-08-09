# [fix] Graph viewer crash: dangling links from arrow-in-name entities (BeadsLog-jip)

**Date:** 2026-08-09

_No architectural changes_

## Problem
On the Fray graph, force-graph threw "node not found: noteid" and a TypeError.
Root cause: a garbage entity name contained the relationship arrow —
"noteid → relative index for --stagger-index css variable". writeGraphHTML
derives link parents by splitting n.Path on ' → ', so an entity name containing
that delimiter produced a phantom parent id ("relative index...") → a link to a
node that was never emitted → crash.

## Work Done
- **Crash guard (cmd/bd/devlog_graph_html.go):** after building nodes+links,
  drop any link whose source or target isn't an emitted node. Makes the viewer
  robust to ANY malformed data (fixes existing projects without needing a prune
  first). Verified: reproduced the exact shape (arrow-named source with an edge)
  → 0 dangling links in the export.
- **Prevent the garbage (internal/extractor/noise.go):** IsNoise now rejects
  names containing '→'/'->' (an entity name never holds the arrow) and names
  with 5+ space-separated tokens (sentence fragments; file paths/commands are
  ≤4 tokens, no spaces). So these never become entities and are auto-pruned on
  sync. Real names (bd devlog enrich, internal/storage/... paths) still survive.

## Validation
- Unit: TestIsNoiseArrowAndFragments (arrow/fragments filtered; legit names pass).
- Live repro: arrow-garbage entity + edge → export has 0 dangling links (no
  crash); 'bd devlog sync' auto-pruned the garbage (arrow entities left: 0).
- Full e2e suite 29/29 green.

## Final Session Summary
**Final Status:** BeadsLog-jip done on branch dev/graph-dangling-links.
**Key Learnings:**
- Two-layer fix: the exporter must be self-consistent (drop dangling links) AND
  the data cleaner (reject arrow/fragment names). The first fixes existing
  broken graphs immediately; the second stops new ones.
