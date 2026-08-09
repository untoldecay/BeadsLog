# [feature] Interactive graph viewer v1 (BeadsLog-0gg)

**Date:** 2026-08-09

_No architectural changes_

## Problem
The --html graph dumped all node labels at once ("bleeding" on a 300-node
graph) and had no way to inspect a node or navigate. User asked for
non-intrusive comprehension quickwins.

## Work Done (cmd/bd/devlog_graph_html.go — template only, no Go data changes)
All client-side in the self-contained HTML (force-graph@1 callbacks); degree +
neighbors computed in-browser from the links array:
- **Declutter:** only the top-12 hubs by degree stay labeled at rest; hover a
  node to focus (reveal its label + neighbors, dim/fade the rest); hover
  highlights incident links.
- **Click panel:** slide-in side panel with the node's connection count and
  neighbors grouped by relationship, sorted by degree; click a neighbor to
  center+zoom to it. Click background to close.
- **Size by degree:** hubs are visibly larger.
- **Navigation:** search box (centers/zooms to first match), Fit button
  (zoomToFit), pin-on-drag, stats readout (N nodes · M links).
- **Filter:** co-occurrence (dashed) links on/off toggle.
- Kept the JS backtick-free so it stays inside the Go raw-string template.

## Validation
- Exported this repo's graph (300 nodes / 311 links); all controls present;
  data intact; placeholders replaced; `node --check` on the viewer JS = OK.
- E2E Test 29 extended to assert the viewer controls (search/panel/coToggle/
  onNodeClick/showPanel) exist in the export. Suite 29/29 green.

## Final Session Summary
**Final Status:** BeadsLog-0gg done on branch dev/graph-full (with the
whole-graph + --open features).
**Key Learnings:**
- Degree/neighbor data needed no Go changes — the links array already carries
  it; compute in-browser. Kept the whole enhancement to one template edit.
- Anchor-label only the top-N hubs to keep large graphs readable; everything
  else appears on hover/zoom/search.
