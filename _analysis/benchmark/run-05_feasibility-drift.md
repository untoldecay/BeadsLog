# Run 05 — Feasibility: reproducible graph / cross-machine drift

- **Date:** 2026-08-23 · Corpus: full BeadsLog codebase + devlogs · Model: sonnet (arms + judge).
- **Question:** "Agents on different machines build slightly different knowledge graphs from the same devlogs. Assess the root cause, what the project already considered/tried, and the feasibility/impact/dependencies of making graph construction reproducible."

## Result

| | Arm A — bd | Arm B — grep |
|---|---|---|
| Tokens | 88,429 | 115,479 (bd −23%) |
| Tool-calls | 58 | 84 (bd −31%) |
| Judge total | 16/20 | **17/20** |

**Blind judge: grep won (+1).** Both correctly diagnosed the root causes (Ollama `temperature=0.8`/`seed=-1` with an unimplemented `ollama.go:124` "set options for deterministic output" comment; Go map-iteration order in `pipeline.go`; `AutoAliasDuplicates` tie-break) and both recalled the prior art (`BeadsLog-5xf` graph.jsonl design, `BeadsLog-yj6` fingerprint). The judge preferred grep because it cited **verifiable `issues.jsonl:40/:168` line numbers**, while bd cited **devlog session IDs (`sess-7855d3`, …) that can't be checked against committed files** — a real traceability weakness (→ product finding: cite filenames, not sess-IDs).

## The genuinely-good feature this surfaced
An **acknowledged, unimplemented gap**: `internal/extractor/ollama.go:124` says "Set options for deterministic output if possible" but sets none.
- **~90% fix, immediate:** `Options:{temperature:0, seed:42}` in the Ollama request + sort `pipeline.go`'s map output before returning + break `AutoAliasDuplicates` ties on name.
- **100% fix:** the `graph.jsonl` shareable-graph design already written in **BeadsLog-5xf** (open), measured by **BeadsLog-yj6** `bd devlog score` (open). Follows the existing `aliases.jsonl`/`links.jsonl` committed-artifact + LWW pattern. Import order must be aliases → graph → links (user-link edges win).

## Conclusion
bd was cheaper (−23% tok, −31% calls) but lost quality by 1 on the sess-ID traceability issue. Both reports are strong and buildable. This is the highest-value feature find of the benchmark.
