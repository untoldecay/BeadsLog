# [enhance] Alias suggestions: signal quality + count hint + review command (BeadsLog-819)

**Date:** 2026-08-10

## Problem
`bd devlog graph`/`search` appended EVERY alias suggestion pair to output —
3,152 pairs on this repo (~9,500 lines), 70% of them single-session noise
(two entities co-occurring once = 100% Jaccard). No LIMIT in SQL, no cap in
showAliasHints, five call sites. Worse: the `ui.IsTerminal()` gate hid hints
from agents in pipes while the embedded protocol told agents to verify them.

## Work Done
### Signal (internal/queries/entities.go)
- **Session floor**: both entities need >= 2 sessions (SQL) — identity can't
  be claimed from one observation.
- **Name similarity gate** (`nameSimilarity`, min 0.55): suggestions need to
  LOOK alike, not just co-occur — exact normalized = 1.0, substring
  containment (>=4 chars) = 0.7, else Levenshtein ratio (reuses
  `utils.ComputeDistance`). Kills the "cut-release ↔ bump-version" class.
- **Dismissals**: new `alias_dismissals` table (migration 057), pairs stored
  as sorted normalized names; `DismissAliasPair` + exclusion in
  `GetAliasSuggestions` so rejected pairs never resurface — review converges.
- Ranked by combined score (overlap × name), `limit` param.
- **Result on this repo's real DB: 3,152 → 16 suggestions, all genuine**
  (truncation artifacts like "ulti-agent"↔"multi-agent", real variants).

### UX (cmd/bd/devlog_cmds.go)
- `showAliasHints` → one line: "💡 N alias opportunities — review: bd devlog
  aliases suggest". IsTerminal gate REMOVED so agents in pipes see it.
- New `bd devlog aliases` group: `suggest` (--limit default 10, --json with
  overlap/name/score fields, per-row merge & reject commands) and
  `dismiss <a> <b>` (order-insensitive).

## Validation
- Unit: TestGetAliasSuggestions rewritten for the new contract (similar-name
  suggested / dissimilar filtered / single-session floored / dismissal honored).
- E2E Test 23 in `_sandbox/run_e2e_tests.sh`: no OPPORTUNITY flood, hint and
  suggest consistent, dismiss removes pair.
- Live on repo DB: suggest/dismiss/--json all verified.

## Gotchas
- Nested queries on one *sql.DB grab a second pooled connection — breaks
  `:memory:` test DBs (each connection = separate empty DB). Load lookup maps
  BEFORE opening the main result set.
- Removing the IsTerminal gate re-exposed the pipefail+`grep -q` SIGPIPE race
  in e2e Tests 1/8/13 (hint prints after the match, grep exits early). Fixed
  by capture-then-grep — same fix as Test 15 previously.

## Final Session Summary
**Final Status:** Shipped on dev/solo-mode-paj; BeadsLog-819 closed. Also
filed BeadsLog-a2l (init --solo should offer existing sync-branch reuse).
**Key Learnings:**
- Co-occurrence measures "related", not "identical" — any identity heuristic
  needs an orthogonal signal (name similarity) and a floor on evidence count.
- A review queue without dismissals never converges for agents.
