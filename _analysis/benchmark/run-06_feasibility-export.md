# Run 06 — Feasibility: `bd devlog export --format json`

- **Date:** 2026-08-23 · Corpus: full codebase · Model: sonnet.
- **Question:** "Assess feasibility/impact/dependencies of adding `bd devlog export --format json` that dumps the full knowledge graph (entities, edges, sessions) for external tools."

## Result

| | Arm A — bd | Arm B — grep |
|---|---|---|
| Tokens | 70,105 | **22,543** (grep −68%) |
| Tool-calls | 48 | **2** |
| Judge total | **18/20** | 14/20 |

**Blind judge: bd won (+4).** grep was dramatically cheaper (2 big calls) and reached the same core conclusion, but bd scored higher on grounding + prior-art: it recalled that export overlaps **BeadsLog-yj6** (`bd devlog score`) and **BeadsLog-5xf** (shareable graph) and proposed a shared `buildFullGraphData()` helper, and it caught that the HTML exporter's `htmlGraphNode.Group/Val` are **render-artifacts** needing a richer semantic export type — not plug-and-play. grep concluded "no prior art exists" and treated the HTML structs as directly reusable.

## The feature
Both arms agreed: **~80–120 lines, JSON structs already exist** in `devlog_graph_html.go`, no schema change, and **drop `--format` — reuse the existing global `--json`** (single-value enum is over-spec'd). bd's sharper version: add an `if jsonOutput` branch to `devlogGraphCmd`, define a semantic `graphJSONExport{entities,edges,sessions}` type (not the render structs), default explicit-deps-only with `--include-cooccurrence` opt-in (O(n²) risk).

## Conclusion
Cheapest grep win of the feasibility set, but bd's report was materially better (prior-art + the render-artifact catch). A real, small feature.
