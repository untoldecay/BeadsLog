# [feature] Signature-gated prefix auto-adopt on sync (BeadsLog-b4p)

**Date:** 2026-08-07

## Problem
A prefix mismatch between a clone's DB prefix and the pulled sync JSONL raised a
hard error ("use --rename-on-import"). But a teammate MUST adopt the team's
prefix to keep working, so the error was pure friction for the common case:
a team standardizing on a new prefix. We want silent adoption + a notification —
without silently repointing the repo when a stray/pollution prefix appears.

## Design: the committed config.yaml is the signature
`.beads/config.yaml` is tracked in git and can declare `issue-prefix`. Whoever
commits that declaration authored (signed) the migration — git records who and
when. So auto-adopt fires ONLY toward the config.yaml-declared prefix; any
prefix that merely appears in the JSONL but was not declared (test/CI pollution,
stray cross-project issue, stale-clone push) is unsigned and still errors.

## Work Done
- **importer.Options.AutoAdoptPrefix + AdoptTargetPrefix**; adopt branch in
  `handlePrefixMismatch` (internal/importer/importer.go). Guards:
  own-sync-JSONL (AutoAdoptPrefix), single upstream prefix, !RenameOnImport
  (opposite intent wins), and `singleMismatch == AdoptTargetPrefix` (the
  signature). On adopt: `SetConfig(issue_prefix=new)`, then (1) rename incoming
  old-prefix issues → new + tombstone, (2) migrate local-only DB issues absent
  from incoming. Reports AdoptedPrefix/AdoptedFrom/MigratedLocal.
- **Local migration correctness (two helpers):** `appendRenameTombstones`
  (shared with the 1tz rename path) and `migrateLocalToPrefix`. Old IDs are
  always tombstoned so deletions propagate via LWW; live copies are injected
  only for truly local-only issues (avoids duplicate IDs).
- **CLI wiring:** `--auto-adopt-prefix` flag on `bd import`; `bd sync` (inline +
  subprocess) enables it and passes `config.GetString("issue-prefix")` as the
  signature. Summary line printed on adopt (cmd/bd/{import,import_shared,sync_import}.go).

## The 3-way-merge trap (caught only by the real binary)
`bd sync` 3-way-merges local issues INTO the incoming set before import, so the
clone's local issues arrive under their OLD prefix. The first implementation
left them live. Fix: step (1) renames incoming old-prefix issues too. Go unit
tests (direct ImportIssues) missed this; e2e Test 19 caught it.

## Validation
- Go: 5 tests (adopt+migrate, fresh-clone, multi-prefix error, disabled/foreign
  error, unsigned-pollution error). Full importer + sqlite suites green; vet clean.
- E2E: Test 19 (authored migration via config.yaml → adopt heals JSONL, no live
  old prefix, idempotent re-sync) and Test 20 (unsigned stray prefix → no adopt,
  prefix untouched).

## Key finding / follow-up
`bd init --prefix X` writes the prefix to the **DB only** — config.yaml keeps
`issue-prefix` commented. So the signature is absent by default and a mismatch
still errors (safe). To author a migration today you must hand-set an
uncommented `issue-prefix:` in config.yaml + commit. Natural next step (folds
into [[BeadsLog-bl1]] decouple-prefix): a `bd prefix set <new>` helper that
writes config.yaml + DB together in one signed action.
