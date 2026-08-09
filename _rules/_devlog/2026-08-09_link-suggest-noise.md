# [fix] Link suggestions polluted by noise hubs (BeadsLog-rob)

**Date:** 2026-08-09

_No architectural changes_

## Problem
On Fray, 'bd devlog links suggest' was dominated by junk: "user ↔ uses"
(uses = a verb that became an entity) buried the real pairs
(floating-note↔editor↔slash). Three causes: (1) 'uses'/'user'/'users' weren't
in the noise stoplist, so auto-prune left them; (2) GetLinkSuggestions didn't
exclude noise-name entities; (3) raw co-occurrence-count ranking surfaced
hub-with-everything over tight architectural pairs.

## Work Done
- **Stoplist** (noise.go): added user/users/uses/app/application/data/value(s)/
  item(s). Auto-prune now removes them; extraction never re-adds them.
- **Defensive filter** (cmd/bd): filterNoiseLinks drops suggestions where either
  endpoint IsNoise — a safety net for DBs that predate the stoplist and haven't
  re-pruned. Applied in both the suggest command and the 🔗 hint; limit applied
  AFTER filtering so the list isn't padded with junk.
- **Ranking** (GetLinkSuggestions): ORDER BY Jaccard overlap DESC, co DESC — tight
  pairs (appear together and rarely apart) now beat loose hubs.

## Validation
- Repro: user/uses noise hub + tight floating-note↔editor + loose auth-service
  hub -> suggest returns floating-note↔editor FIRST (100% overlap), auth-service
  pairs below (33%), user/uses gone.
- Unit: noise test extended (user/uses/app/data filtered). E2E Test 30; suite 30/30.

## Final Session Summary
**Final Status:** BeadsLog-rob done on branch dev/link-suggest-noise.
**Key Learnings:**
- Co-occurrence link suggestions need BOTH noise exclusion and overlap ranking —
  raw co-count alone surfaces ubiquitous hubs, not architecture.
