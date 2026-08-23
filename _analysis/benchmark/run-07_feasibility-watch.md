# Run 07 — Feasibility: `bd devlog watch`

- **Date:** 2026-08-23 · Corpus: full codebase · Model: sonnet.
- **Question:** "Assess feasibility/impact/dependencies of a `bd devlog watch` command that live-tails newly recorded devlog sessions and re-indexes them incrementally, reusing the daemon's file-watcher and debounce logic."

## Result

| | Arm A — bd | Arm B — grep |
|---|---|---|
| Tokens | 81,886 | 95,497 (bd −14%) |
| Tool-calls | 40 | 45 |
| Judge total | **16/20** | 15/20 |

**Blind judge: bd won (+1).** Near-parity — both correctly identified that `FileWatcher`+`Debouncer` are drop-in, that the daemon's watcher is scoped to `issues.jsonl` (so "reuse" = instantiate the struct, not attach to the daemon), and that `SyncSession` is the incremental primitive. bd edged it by naming the **crystallization write-back re-trigger loop** (the subtlest correctness risk) and noting `SyncSession` already prevents it via hash-update. grep countered with a precise detail bd hedged on — `SetupStderrLogger` (a daemon-free logger) — and stayed in `devlog_cmds.go` (leaner than bd's proposed new file); grep's 500ms debounce also matched the existing default vs bd's arbitrary 800ms.

## The feature
Both agreed: **~80 lines, one new `cobra.Command`, standalone foreground** (not daemon-integrated). Wire `NewFileWatcher(devlogDir/_index.md)` → `NewDebouncer` → `parseIndexMD` + `SyncSession` loop; use `SetupStderrLogger`; block on SIGINT. One real design choice: watch `_index.md` only (new records) vs the whole devlog dir (also catches edits to existing session files — `FileWatcher`'s parent-dir watching supports both).

## Conclusion
bd slightly cheaper and slightly higher quality. A clean, small feature reusing existing infra.
