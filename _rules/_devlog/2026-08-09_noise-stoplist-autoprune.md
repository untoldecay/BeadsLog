# [enhance] Generic prose-noise: extraction stoplist + auto-prune on sync (BeadsLog-4qu)

**Date:** 2026-08-09

## Problem
Generic terms (config, component, service, services, technologies, system,
module, feature, state) leaked into the graph as high-degree HUB noise (seen on
the Fray project: config 104, component 60, service 54…). A min-degree filter
can't help — they're hubs. Root cause: the extraction noise stoplist
(commonWords) lacked these nouns, so IsNoise passed them; the regex extractor
happened not to emit them (so e2e Test 3 passed) but the AI extractor does.

## Work Done
- **#1 Stoplist (internal/extractor/noise.go):** added the generic
  architecture-prose nouns to commonWords. A bare word now filters; specific
  compound names survive because they carry an uncommon token — verified
  auth-service, config.yaml, UserService, drawerpanelview all pass.
- **#2 Auto-prune on sync (cmd/bd/devlog_cmds.go):** extracted the
  `prune --noise` removal into `pruneNoiseEntities(ctx, db)` and call it in the
  sync path right after AutoAliasDuplicates. So every `bd devlog sync` (which
  `bd devlog record` chains) sweeps out noise entities — the backlog is cleaned
  incrementally and can't accumulate. Self-limiting: no-op once clean. Logs
  "🧹 Removed N noise entity(ies)" only when it acts.

## Validation
- Unit: TestIsNoiseGenericProseNouns (generics filtered; specific names survive).
- Live: injected config/component/service + auth-service/drawerpanelview,
  ran sync → removed the 3 generics, kept the 2 real ones.
- Full e2e suite 29/29 green (Test 3's component/module/feature/system/state
  now also stoplist-filtered; auto-prune doesn't eat legit entities).

## Final Session Summary
**Final Status:** BeadsLog-4qu done on branch dev/graph-viewer-perf (ships with
the viewer perf fix as the Fray-review batch).
**Key Learnings:**
- The stoplist filters at IsNoise so it works for BOTH extractors (regex + AI) —
  the regex-only non-emission was masking the gap.
- Auto-prune belongs on sync (not per-record) and is safe because it's
  self-limiting and only touches IsNoise matches.
