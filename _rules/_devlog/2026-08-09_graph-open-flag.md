# [feature] bd devlog graph --open: launch exported HTML in browser (BeadsLog-uoe)

**Date:** 2026-08-09

_No architectural changes_

## Problem
`bd devlog graph --html <path>` wrote the file and printed the path but did NOT
open it. User wanted it to launch the browser directly.

## Work Done (cmd/bd/devlog_graph_html.go, devlog_cmds.go)
- Added `--open` flag to `bd devlog graph`. After a `--html` export (both
  single-entity and whole-graph paths), opens the file in the default handler.
- `browserOpenCommand(goos, path)` — pure, unit-tested argv selection:
  darwin→`open`, windows→`rundll32 url.dll,FileProtocolHandler`, else→`xdg-open`.
- `openInBrowser` runs it via exec.Start(); `openExportedGraph` wraps it so a
  failure (e.g. headless server) prints a non-fatal hint instead of erroring —
  the export already succeeded. Opt-in only; default stays file-only for CI.

## Validation
- Unit: TestBrowserOpenCommand pins the opener argv per OS (no browser spawned).
- `--open` registered in help; full e2e suite 29/29 green (e2e does NOT pass
  --open, to avoid spawning browsers in the harness).
- Live launch intentionally not triggered (would pop an unprompted window);
  correctness rests on the unit test + standard platform openers.

## Final Session Summary
**Final Status:** BeadsLog-uoe done on branch dev/graph-full (ships with the
whole-graph feature, BeadsLog-xuv).
**Key Learnings:**
- Keep the platform-specific choice in a pure function so it's testable without
  side effects; keep the actual exec in a thin best-effort wrapper.
