# [feature] bd refresh — one-command post-update (BeadsLog-q9o.1)

**Date:** 2026-08-08

## Problem
After a binary update, users had to remember a scatter of steps (daemon
restart, doctor, sync, namespace adoption). Ship one command — `bd refresh` —
that runs the mandatory ones and prints a summary, so nobody wonders what to
run after `go install`.

## Work Done
`cmd/bd/refresh.go` — orchestrator, mostly reusing existing pieces:
- **Version/changelog**: reports previousVersion → Version from the per-machine
  detection globals; records changelog-seen (per-machine) at the end.
- **Migrations**: already run on DB open (idempotent) — reported, not re-run.
- **Daemon**: `restartDaemonForVersionMismatch()` when a bump was detected.
- **Doctor**: curated update-relevant checks (installation, git hooks, schema
  compat, db version) summarized as "✓ N passed ○ M warning(s) — first fix".
  `--fix` execs `bd doctor --fix`.
- **Namespace**: gated on `sync-mode != local-only` AND a remote existing (mode,
  not remote-presence — a solo user may keep an old remote). Reuses the
  signature-gated adopt by exec'ing `bd sync --no-push` — **fetch only, never a
  push**. Degrades gracefully to "run 'bd sync' manually" on error.
- `--devlog` opt-in adds `devlog sync` + `verify --fix` (slow).
- devlog-status-style summary; collapses to one line when nothing changed.

Entry points wired:
- `bd onboard` gains Step 0 ("bd was upgraded — run bd refresh first") when an
  upgrade is detected.
- `bd upgrade` install tail now points to `bd refresh`.
- The 🔄 post-list/ready banner points to `bd refresh` (done in q9o.2).

## Validation
- E2E Test 27: solo → "Namespace: skipped (solo / local-only mode)";
  no-remote → "skipped (no git remote)"; summary block present. Deterministic,
  no network. Adoption itself stays covered by Tests 18-22 (the reused path).
- Live: solo/no-remote gating correct; team path invokes sync --no-push and
  degrades gracefully when the remote is malformed. Suite 27/27 green.

## Final Session Summary
**Final Status:** q9o.1 done → epic q9o COMPLETE (q9o.1/.2/.3 all closed) on
branch dev/refresh-postupdate.
**Key Learnings:**
- Most post-update work was already automatic; refresh is orchestration +
  honest reporting, reusing restartDaemonForVersionMismatch, doctor checks, and
  the sync adopt path rather than reimplementing any of them.
