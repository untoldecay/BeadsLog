# [perf] Graph viewer smooth on large graphs + min-degree filter (BeadsLog-dn5)

**Date:** 2026-08-09

_No architectural changes_

## Problem
On a large graph (Fray project: 444 source nodes) dragging a node lagged/broke
the render. Root causes, partly a regression from the prior viewer round:
- `autoPauseRedraw(false)` repainted every frame while the force sim was live.
- Neighbor-highlight used a per-frame O(nodes×degree) linear scan.
- Dragging reheated the whole n-body simulation.
Separately, generic prose-noise entities (config/component/service) dominate.

## Work Done (cmd/bd/devlog_graph_html.go — template only)
- **O(1) highlight:** precompute the active node's neighbor SET once on
  hover/click change (recomputeActive) instead of scanning every frame.
- **Freeze on settle:** onEngineStop pins every node (fx/fy). Dragging one node
  then moves only that node — no full-sim reheat. onNodeDrag keeps it pinned.
  cooldownTime(6000) settles faster. autoPauseRedraw(false) kept, but now cheap
  (draw-only; physics idle) so highlight stays reliable.
- **min-connections slider:** a degree threshold (nodeVisibility + linkVisibility)
  hides the long tail of leaf-noise, cutting clutter AND render cost on huge
  graphs. Live, no relayout.

## Validation
- Regenerated (JS syntax OK); e2e Test 29 still asserts the viewer controls,
  suite 29/29 green.
- Manual: freeze + O(1) highlight makes drag smooth; slider hides low-degree
  nodes instantly.

## Still open (follow-up)
Generic prose-noise HUBS (config/component/service) have high degree so a
min-degree filter keeps them — they need an extraction-side stoplist or a viewer
name-stoplist. Tracked separately.

## Final Session Summary
**Final Status:** BeadsLog-dn5 done on branch dev/graph-viewer-perf.
**Key Learnings:**
- The prior round's autoPauseRedraw(false) + per-frame neighbor scan was the
  large-graph regression; freezing physics after layout is the key to smooth
  drag on big graphs.
