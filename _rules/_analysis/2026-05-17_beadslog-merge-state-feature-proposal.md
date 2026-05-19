# BeadsLog Feature Proposal: Session Lifecycle Status
**Date:** 2026-05-17
**Triggered by:** Architecture analysis where agent treated archived-branch work as merged baseline

---

## The Problem

An agent ran `bd devlog search "modularization"` and found sessions describing completed work:
- sess-b0e613: "WYSIWYG Editor modularization (Phases 1-3)" — recorded 2026-04-27
- sess-66309e: "Complete Phase 4 of WYSIWYG Editor modularization" — recorded 2026-05-05
- sess-25f818: "modularize UnifiedContentView task mode" — recorded 2026-04-08

The agent concluded all three were in the codebase and built recommendations on top of them ("continue Phase 6", "the orchestration core remains after extracting 4,800 lines").

**Reality:** The WYSIWYG modularizations were never merged to develop. The UCV Phase 1 *was* merged (via sess-f334e5). The agent couldn't distinguish the two cases. It said index.vue "grew back by 1,350 lines" — but the extraction simply never landed.

### Note on branch tracking

The `_index.md` Branch column already exists and newer sessions populate it automatically (e.g. `fix/ordered-list-nan (local)` on sess-c68842). Older sessions show `N/A` because they predate the feature. Branch tracking is **solved** — what's missing is a status dimension that tells agents whether the work described in a session is actionable, landed, paused, or dead.

---

## Proposed Feature: Session Lifecycle Status

### The five states

| Status | Meaning | When to set |
|---|---|---|
| **active** | Work in progress on a living branch | Default on `bd devlog record` |
| **merged** | Code landed on main/develop | On merge session or via git hook |
| **paused** | Deliberately shelved — will resume later | Manual, when parking a branch |
| **archived** | Kept for reference, not planned to merge | Manual, when a branch is kept but abandoned as a merge candidate |
| **abandoned** | Deliberately dropped | Manual, when an approach is rejected |

### Why `reason` is mandatory for non-active states

The status alone is useless for agents. "Abandoned" without a reason means the agent will just re-propose the same approach. The reason is the actual signal:

| Status + Reason | What the agent learns |
|---|---|
| **paused** — "Waiting on TextKit 2 decision before merging editor changes" | Don't touch the editor sync loop yet |
| **archived** — "Caused scroll jitter regression under 5000-note libraries" | This approach has a perf constraint to solve first |
| **abandoned** — "Replaced by provide/inject approach, see sess-XXXXX" | Follow the link to the replacement, don't re-propose |
| **paused** — "Depends on sess-b0e613 which is paused" | Dependency chain — resolve the root pause first |

Without reasons, agents will:
- Re-propose abandoned approaches (wasting cycles)
- Build on paused work without knowing the blocker
- Miss that archived work failed for a specific reason worth avoiding

### Linked sessions

Non-active statuses often reference other sessions:
- **paused** → may depend on another session being completed first
- **abandoned** → may point to a replacement session
- **merged** → should reference the merge session

These links should be structural (queryable), not just prose in the reason string.

---

## CLI Design

### Setting status

```bash
# Pause work with a reason
bd devlog status sess-b0e613 --set paused \
  --reason "Waiting on TextKit 2 decision before merging editor changes"

# Archive with dependency
bd devlog status sess-66309e --set archived \
  --reason "Depends on sess-b0e613 (paused)" \
  --refs sess-b0e613

# Mark as merged with link to merge session
bd devlog status sess-25f818 --set merged \
  --ref sess-f334e5

# Abandon with replacement pointer
bd devlog status sess-XXXXX --set abandoned \
  --reason "Replaced by provide/inject approach" \
  --refs sess-YYYYY
```

### Querying by status

```bash
# Filter search results
bd devlog search "modularization" --status merged
bd devlog search "modularization" --status paused,archived
bd devlog search "modularization" --status active,merged   # "what can I build on?"

# List by status
bd devlog list --status paused        # what's parked?
bd devlog list --status abandoned     # what was dropped and why?
```

### Display in search results

```
📄 Found 3 sessions:
1. [sess-66309e] [⏸ PAUSED]   Complete Phase 4 of WYSIWYG Editor modularization
   Reason: Depends on sess-b0e613 (paused)
2. [sess-25f818] [🟢 MERGED]   modularize UnifiedContentView task mode
   Merged via: sess-f334e5
3. [sess-b0e613] [⏸ PAUSED]   WYSIWYG Editor modularization (Phases 1-3)
   Reason: Waiting on TextKit 2 decision
```

### Resume warnings

```bash
bd devlog resume --last 1
# If loading a non-active session:
⚠️  Session sess-b0e613 is PAUSED: "Waiting on TextKit 2 decision"
    Recorded on branch: worktree/editor-modularization (not merged to develop)
    Code changes are NOT in the working tree.
```

---

## How Agents Should Consume This

### Decision tree for agents

```
When reading a devlog session that describes code changes:

1. Check session status (visible in search results)

2. If ACTIVE or MERGED:
   → Code should be in the working tree
   → Still verify with file reads (code may have changed since)
   → Safe to build on

3. If PAUSED:
   → Code is NOT in the working tree
   → Decisions and learnings ARE valid context
   → Read the reason — it tells you what's blocking
   → Do NOT resume or build on it without asking the user
   → Follow --refs to understand the dependency chain

4. If ARCHIVED:
   → Code exists on a branch but won't be merged as-is
   → May be cherry-pickable or worth re-doing differently
   → Read the reason — it may reveal constraints (perf, regressions)

5. If ABANDONED:
   → Read the reason FIRST — it's the most valuable signal
   → Do NOT re-propose the same approach without addressing the reason
   → Follow --refs to find the replacement (if any)
```

### Updated `docs/agents/domain.md` protocol addition

```
Step 2b — For sessions describing code changes, check lifecycle status:

  Status is shown in bd devlog search results: [🟢 MERGED], [⏸ PAUSED], etc.
  If no status shown, the session predates this feature — verify manually
  by checking if the branch exists and whether it's merged.

  PAUSED/ARCHIVED/ABANDONED sessions:
  - Decisions and learnings → valid context
  - Code changes → NOT in working tree, do not assume as baseline
  - Reason → read it, it's the key signal
  - Refs → follow linked sessions for dependency chains or replacements
```

---

## Implementation Phases

### Phase 1: Status field + set command
- Add `status` and `status_reason` fields to session DB schema
- `bd devlog status <sess-id> --set <state> --reason "..."` writes them
- Default status: `active` on new sessions
- Display status in `bd devlog list` and `bd devlog search` output

### Phase 2: Query filtering
- `--status` filter on `bd devlog search` and `bd devlog list`
- `bd devlog resume` warns when loading non-active sessions

### Phase 3: Linked references
- `--refs sess-a,sess-b` on status command stores structural links
- `bd devlog show` displays linked sessions
- `bd devlog landed <sess-id>` summarizes: status + reason + branch + refs

### Phase 4: Auto-detection (optional)
- Git merge hook auto-sets `merged` status when a branch is merged
- `bd devlog doctor` detects sessions on deleted/unmerged branches and prompts for status

---

## Evidence: What This Would Have Prevented

| Session | Agent assumed | Reality | With lifecycle status |
|---|---|---|---|
| sess-25f818 (UCV Phase 1) | Merged ✅ | Merged via sess-f334e5 | `[🟢 MERGED]` — correct |
| sess-b0e613 (WYSIWYG Phases 1-3) | Merged ❌ | Never merged, on archived branch | `[⏸ PAUSED]` + reason → agent wouldn't build on it |
| sess-66309e (WYSIWYG Phase 4) | Merged ❌ | Never merged, depends on sess-b0e613 | `[⏸ PAUSED]` + ref to sess-b0e613 → agent sees dependency chain |
| sess-d59501 (events.js 2025) | Merged ✅ | Merged (eventModules/ in working tree) | `[🟢 MERGED]` — correct |

The agent got 2/4 right by luck (verifying line counts). With lifecycle status, it would get 4/4 right by data.

---

## Related files
- `_rules/_analysis/2026-05-17_architecture-deepening-opportunities.md` — the analysis that exposed this gap
- `_rules/_analysis/2026-05-17_devlog-protocol-comparison-report.md` — Run 1 vs Run 2 comparison
- `docs/agents/domain.md` — current devlog query protocol for agents
