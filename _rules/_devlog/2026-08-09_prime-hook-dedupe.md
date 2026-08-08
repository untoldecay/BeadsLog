# [perf] Dedupe SessionStart protocol injection — bd prime --hook (BeadsLog-29q)

**Date:** 2026-08-09

## Problem
The A/B experiment (ABTest/SUMMARY.md rec #1) found the protocol is
double-loaded at session start: the harness loads the CLAUDE.md
`<beads_protocol>` block AND the SessionStart hook runs `prime` which re-emits
the same ~738 tokens. `bd prime --hook` also errored ("unknown flag") — the
flag the SessionStart hook already passes wasn't implemented in bd.

## Work Done (cmd/bd/prime.go)
- Added `--hook` flag. In hook mode, if an agent file already carries the
  protocol block (`agentProtocolPresent()` scans Candidates for
  `ProtocolStartTag`), prime emits the LEAN (~155 tok) reminder instead of the
  full protocol — the harness already provides the full block, so re-emitting
  it is pure duplication. If NO agent file carries it (e.g. stealth setups),
  --hook stays full so nothing is lost. Explicit `--full`/`--mcp` still win.
- Plain `bd prime` (the PreCompact hook) is unchanged — it still restores the
  full protocol after compaction truncates it.
- No `.claude/settings.json` change needed: SessionStart already calls
  `prime --hook`, so a rebuilt binary auto-dedupes.

## Measured impact (this repo)
- `bd prime --hook`: 738 → **131 tokens** (lean, protocol present).
- `bd prime` (PreCompact): 738 tokens (unchanged).
- No-agent-file fallback: non-lean (restricted/full), nothing dropped.
- ≈ **600 tokens saved per session start**, zero behavior change.

## Validation
- Unit: TestAgentProtocolPresent (false with no file / file-without-block; true
  with block).
- E2E Test 28: `prime --hook` output strictly smaller than full `prime`, and
  still contains the session-close reminder. Suite 28/28 green.

## Final Session Summary
**Final Status:** BeadsLog-29q done on branch dev/prime-hook-dedupe. This is
rec #1 (the free win) from the protocol-footprint A/B. Recs #2 (live lean
PRIME.md A/B) and #3 (make lean the default) remain open decisions.
**Key Learnings:**
- The biggest immediate token saving wasn't the lean rewrite — it was removing
  a redundant re-injection of content the harness already loads.
