# [fix] rename-on-import heals shared JSONL via tombstones (BeadsLog-1tz)

**Date:** 2026-08-07

## Problem
`bd sync --rename-on-import` renamed issue prefixes only on the DB side. The
stale old-prefix lines in `issues.jsonl` were never neutralized, so the prefix
mismatch recurred on every subsequent import/checkout, and — worse — teammates
resurrected the old prefix via LWW merge on every clone that still exported it.
Evidence: after two rename-on-import runs, DB had 0 old-prefix issues but the
worktree JSONL still had 30 old-prefix lines (89 in committed HEAD).

## Work Done
- **Tombstone every old ID** on rename (internal/importer/importer.go,
  `handlePrefixMismatch`): the export/merge pipeline replaces the stale live
  lines in the shared JSONL and clones delete their copies instead of
  resurrecting. Old-prefix tombstones are already treated as benign, so
  subsequent imports stop erroring and TTL compaction purges them.
- **Bump `UpdatedAt` on renamed issues** so LWW prefers the new-ID version over
  the stale old-ID line; otherwise content-hash matching resolves toward the old
  ID, which then loses to its own tombstone and the issue is silently dropped.
- **Adopt the canonical ID in `upsertIssues`** when a cross-prefix content match
  carries the DB's configured prefix (prefix-migration rename) instead of
  skipping it — the skip was silent data loss.
- **`RenameImportedIssuePrefixes` returns the oldID→newID mapping** (utils.go) so
  callers can persist the rename beyond the DB.
- **`EnsureIDs` exempts tombstones from prefix validation** (sqlite/ids.go):
  old-prefix tombstones are the propagation mechanism for renames.

## Validation
- Go: `TestRenameOnImport_EmitsTombstonesForOldIDs` (end-to-end via ImportIssues)
  and `TestHandlePrefixMismatch_BumpsUpdatedAtOnRename` (LWW timestamp bump).
- E2E: Test 18 in `_sandbox/run_e2e_tests.sh` — renamed issue survives under new
  prefix, old ID tombstoned in JSONL, plain re-sync no longer raises the mismatch.
- Full importer + sqlite suites green; `go vet` clean.

## Design Notes / Follow-ups
- **Prefix is per-clone (DB config).** `config.yaml issue-prefix` only seeds
  FRESH clones; existing clones with a different prefix need a one-time
  `--rename-on-import`. Chain: set prefix in committed config.yaml → teammate's
  first `bd init` adopts it; an already-inited teammate must migrate manually.
- **Prefix decouple:** derive-from-repo/dir-name is the sharp edge (renames
  cause churn — this bug's root). Recommendation: pin a short stable
  `issue-prefix` in shared config.yaml, distinct enough to survive cross-repo
  import.
- **Silent-adopt idea:** a hard mismatch error is just friction since the
  teammate must adopt anyway; prefer a notification ("upstream prefix adopted").
  Guard: only auto-adopt from the repo's own sync JSONL with exactly one
  upstream prefix — NOT arbitrary `bd import -i <foreign-file>` (legit
  cross-project) or ambiguous multi-prefix pulls.
- **Solo↔team switch (data-loss trap):** `createSyncBranch` checks only LOCAL
  refs and creates from HEAD, so a locally-deleted-but-remote-alive sync branch
  diverges from a code tip on team re-init. Fix direction: detect local vs
  remote branch, ask merge-into-team vs stack-on-team, never auto-create from a
  code tip.
