# BeadsLog Feature Plan: GitHub Issues & Notion Sync Brick
**Date:** 2026-08-06
**Status:** Planned (paused — reactivate via bd issue `BeadsLog-3c2`)

## 1. Goal
New integration "brick" mirroring bd issues to GitHub Issues (and later Notion),
while **BeadsLog stays the source of truth**. Team/tools see the project board on
GitHub; status changes made there flow back; Notion gets read-only visibility
for free via its native GitHub synced databases.

## 2. Feasibility Findings (2026-08-04 research)

### Already in the codebase (the gap is small)
- `Issue.ExternalRef` (`internal/types/types.go:49`, indexed in SQLite) — ready to
  store `gh-123` or a Notion page ID. **No schema migration needed.**
- `Issue.SourceSystem` (`types.go:50`) — federation origin marker.
- **Linear integration is the template**: `cmd/bd/linear.go` + `internal/linear/`
  (client, mapping, types) already solved the whole shape: `--pull`/`--push`/
  `--dry-run`, `--prefer-local` conflict resolution, incremental sync by
  timestamp, config via `bd config set linear.api_key`. Partial Jira scaffolding
  also exists (`cmd/bd/jira.go`).
- LWW conflict resolution + tombstones already proven in the `bd sync` pipeline.
- Per-repo config: `.beads/config.yaml` via Viper, env-var overrides.

### External research
- **Upstream beads (untoldecay/BeadsLog)**: no GitHub/Notion integration, none
  planned; upstream migrated to Dolt storage so our SQLite+JSONL fork has
  diverged — nothing to wait for, no conflict. No community fork does this.
- **GitHub**: `google/go-github` v89, actively maintained. Auth chain:
  env token → `gh auth token` subprocess → PAT in config. Remote-change
  detection by polling `GET /issues?since=<last_sync>` (one request per sync,
  no webhooks/server needed). Watch: secondary content-creation limit
  (~80 creates/min — throttle initial export ~1/sec) and echo loops
  (compare `updated_at`, skip self-originated changes).
- **Notion**: API 2025-09-03 split databases→data sources; **no up-to-date Go
  SDK** (jomei/notionapi targets 2022-06-28) — would need a thin hand-rolled
  client (~5 endpoints). 3 req/s, no bulk writes, 2,000-char property cap.
  **YAGNI for v1**: Notion natively mirrors GitHub Issues via synced databases
  (read-only), so shipping the GitHub brick gives Notion visibility for free.
- **Prior art** (git-bug bridges, GitHub Actions, Unito): dominant pattern is
  one-way mirror + incremental status pull, LWW on timestamp, local tracker as
  source of truth, beads ID embedded in the remote issue body
  (`<!-- beads:BeadsLog-xxx -->`) for recoverability. Nobody small does full
  bidirectional — that's where the complexity lives.

## 3. The Bidirectionality Ladder (chosen scope: Level 2)

| Level | What | Risk |
|---|---|---|
| 1. Mirror (push only) | bd → GitHub; remote edits overwritten | None |
| **2. Mirror + status pull-back (v1)** | Level 1 + pull only close/reopen events | Low — one field, trivially comparable, dodges echo loops / field mapping / delete semantics |
| 3. Selective bidirectional (v2) | Also import GitHub-*born* issues (conflict-free by construction), maybe comments | Medium |
| 4. Full bidirectional | Any field, anywhere, merged | The swamp — explicitly out of scope |

## 4. Implementation Roadmap (v1: `bd github sync`, ~4-6 days)

### Phase 1: Client & mapping
- `internal/github/client.go` — go-github v89 wrapper; auth chain
  (env → `gh auth token` → config PAT).
- `internal/github/mapping.go` — Issue↔GitHub conversion, copied from
  `internal/linear/mapping.go` pattern (title/body/state/labels; priority &
  type flattened to labels on push).
- Mapping stored in `ExternalRef` (`gh-123`) + beads ID as HTML comment in the
  GitHub issue body.

### Phase 2: Command
- `cmd/bd/github.go` — `bd github sync` with `--push`, `--pull` (status-only,
  via `since=`), `--dry-run`, `--prefer-local` default; registered in the
  "advanced" cobra group like `linear`.
- Config: `bd config set github.repo owner/name`, `github.token` optional.
- Throttle initial bulk export (~1 create/sec), honor `Retry-After`.
- Track `last_sync` timestamp for incremental `since=` polling.

### Phase 3: Tests & hardening
- Unit tests on mapping (both directions), echo-loop suppression test
  (push → pull sees own change → no-op), dry-run assertions.
- Sandbox e2e script against a scratch GitHub repo (pattern:
  `sandbox/test_no_arch_marker.sh`).

### v1.5 (validate before building): Notion
- First try Notion native GitHub synced database on top of v1.
- Only build `bd notion sync` (thin client, 2025-09-03, one-way) if read-only
  proves insufficient.

### Explicitly deferred
Comment sync, bidirectional field editing, webhooks/servers, GitHub App auth.
