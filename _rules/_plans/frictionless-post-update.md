# BeadsLog Feature Plan: Frictionless Post-Update (`bd refresh`)
**Date:** 2026-08-08
**Status:** Planned — implementation on branch `dev/refresh-postupdate`

## 1. Goal
After a binary update, the user (or agent) should run **one** command and never
wonder what else to do. Ship `bd refresh` as that command, consolidate the
misleading `bd upgrade status/review/ack` trio into a single `bd upgrade`, and
fix a real cross-machine bug where "changelog seen" state leaks between clones.

## 2. Key finding: most mandatory work is ALREADY automatic
This is orchestration + wiring, not build-from-scratch.

| Already automatic | Trigger |
|---|---|
| Schema migrations (idempotent, EXCLUSIVE-locked) | every DB open — `internal/storage/sqlite/store.go:187` → `migrations.go:167` |
| Daemon restart on version mismatch | next command that contacts daemon — `cmd/bd/daemon_autostart.go:70` |
| Upgrade DETECTION (🔄 banner, auto-migrate) | per-machine `.local_version` — `cmd/bd/version_tracking.go:19` |
| Prefix auto-adopt (signature-gated) | during `bd sync` import — `cmd/bd/sync_import.go:116` |

## 3. Confirmed behaviors (verified in code)
- **A plain `git commit` never pulls.** The pre-commit hook runs
  `bd sync --flush-only` (export only) — `cmd/bd/init_git_hooks.go:305`. No
  foreign namespace enters at commit time.
- **`bd devlog record` auto-chains `bd devlog sync`** (local DB indexing only,
  no git) — `devlogRecordCmd`, "→ Syncing devlog…". The git `bd sync`
  (pull/push) lives only in the session-close protocol, run by the agent.
- **Namespace conflict outcomes on an explicit `bd sync`:** signed migration →
  auto-adopts & migrates local, tombstones old IDs; unsigned mismatch →
  **errors and blocks** (never silently merges or loses data).

## 4. Cross-machine ack bug (the reason ack must be per-machine)
Two separate version states exist:
- `.local_version` — **gitignored, per-machine**. Drives detection. Correct. ✓
- `last-seen-changelog-version` (in `config.yaml`) **and** `bd upgrade ack` →
  `LastBdVersion` (in `metadata.json`) — **committed / shared**. ✗

Consequence: acking the changelog on machine A commits the state; machine B
pulls and never sees the "what's new". `LastBdVersion` is already marked
deprecated (`internal/configfile/configfile.go:19`). Fix: move all
"seen this version" state onto per-machine `.local_version`, making ack
transparent and automatic — each machine gets its own personalised report.

---

## 5. Work items

### Item 1 — `bd refresh` (the one command)
Runs in BOTH modes; only the namespace step touches the network (fetch only,
never push). Gated on **mode**, not remote presence.

| Step | Solo | Team | Network |
|---|---|---|---|
| Report binary version + changelog-since-last | ✓ | ✓ | no |
| Migrations (already auto on DB open — just report the count) | ✓ | ✓ | no |
| Daemon restart onto new binary | ✓ | ✓ | no |
| Doctor (report-only; `--fix` opt-in) | ✓ | ✓ | no |
| Namespace probe + adopt (reuse signature-gated import) | — skipped | ✓ | **fetch only** |
| Summary in `devlog status` style | ✓ | ✓ | — |

- **Namespace gate:** run only when `sync-mode != local-only` AND a remote
  exists. Solo prints `Namespace: skipped (solo / local-only mode)`. Do NOT
  gate on remote-presence alone — a solo repo may retain a remote from before
  going solo.
- **Reuse, don't rewrite:** the fetch + signature-gated adopt already exists in
  the import path; `bd refresh` invokes it. `bd sync`'s own logic is untouched.
- **No push, ever.** No `git push`, no issue sync-out.
- Opt-in flags: `--fix` (doctor fixes), `--devlog` (adds `devlog sync` +
  `verify --fix`, both slow/AI — off by default).

**Summary format (drafted):**
```
bd refresh
==========
Binary:     v0.55.2 → v0.58.0
Migrations: 3 applied (schema now at 057)
Daemon:     restarted on v0.58.0
Doctor:     ✓ 7 checks passed   ○ 1 warning (git hooks stale — run 'bd hooks install')
Namespace:  adopted 'BeadsLog' from remote (was 'bd') · 30 issues migrated, old IDs tombstoned

✨ Up to date. No further action needed.
```
Nothing-changed collapse: `✨ Already current (v0.58.0) — nothing to do.`

**Three entry points, one command:**
- User, after `go install`: runs `bd refresh` in the repo.
- Agent protocol: `bd onboard` Step 0 — "if upgraded, run `bd refresh` first".
- After self-update: `upgrade_check.go` `--install` path ends by running/suggesting `bd refresh`.

### Item 2 — Consolidate `bd upgrade`
Replace the `status`/`review`/`ack` subcommands with a single interactive
`bd upgrade`:
1. show current version,
2. check remote for newer (existing `upgrade_check.go` logic),
3. if newer, print changelog-since-yours,
4. prompt to install.

Drop `ack` entirely (superseded by Item 3). Keep `--json` for the status shape.

### Item 3 — Per-machine version-seen state (kills the ack leak)
- Move `last-seen-changelog-version` and any ack state OFF committed files
  (`config.yaml`, `metadata.json`) and ONTO per-machine `.local_version`.
- Changelog "what's new" gating reads/writes `.local_version`, so ack becomes
  automatic and transparent per machine.
- Remove the deprecated `LastBdVersion` write path.
- Migration note: existing committed `last-seen-changelog-version` becomes inert
  (ignored); first run on each machine re-derives from `.local_version`.

## 6. Explicitly out of scope
- No changes to `bd sync` merge/adopt logic (reuse only).
- No `git push` from `bd refresh`.
- No default devlog re-index / AI enrichment in `bd refresh`.
