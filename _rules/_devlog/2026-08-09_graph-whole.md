# [feature] bd devlog graph with no entity — whole graph (BeadsLog-xuv)

**Date:** 2026-08-09

## Problem
`bd devlog graph` required exactly one entity (`cobra.ExactArgs(1)`). No way to
load the whole graph. Question raised: is `--html` really needed for that?

## Answer / design
Yes for the visual, no to run the command:
- The graph here is 1031 entities / 319 edges — a full ASCII tree in the
  terminal would be unreadable, so `--html` (interactive force-graph) is the
  right medium for the WHOLE graph.
- But the command shouldn't be useless without it, so no-arg terminal prints a
  compact SUMMARY (entity/edge counts + top-15 most-connected by out-degree).

## Work Done
- `Args: cobra.ExactArgs(1)` → `MaximumNArgs(1)`; no-arg dispatches to
  `runFullGraph` (cmd/bd/devlog_graph_html.go).
- `runFullGraph`:
  - `--html`: every edge-source entity becomes a root at depth 1; reuses
    `GetEntityGraphExact` + `writeGraphHTML` (which de-dupes into the union =
    full graph). Zero new graph-query code.
  - no `--html`: prints the summary (counts + hub list) + a tip to use `--html`.
- Updated command help/Long with the two whole-graph forms.

## Validation
- Live on this repo: no-arg summary shows 1031 entities / 319 edges + top hubs;
  `--html` exported the full graph (300 nodes / 311 links — the connected
  subgraph).
- E2E Test 29: no-arg prints "Whole Graph" + counts; `--html` yields a
  non-empty file with a nodes array. Suite 29/29 green.

## Final Session Summary
**Final Status:** BeadsLog-xuv done on branch dev/graph-full.
**Key Learnings:**
- Full-graph HTML needed no new query layer — iterating edge-sources through the
  existing per-entity path + writeGraphHTML's de-dup yields the union graph.
- Terminal can't usefully render 1000 nodes; the right no-arg default is a
  summary, with --html reserved for the actual visualization.
