# BeadsLog: what it is, why it doesn't fight Git, and what it's for

*A 5-minute read for the team — technical but no prerequisites.*

---

## What it is, in one paragraph

BeadsLog is a **project memory for AI agents**, stored inside the repo and versioned by Git. It has two halves: an **issue tracker** (`bd` — tasks, bugs, dependencies, all in the terminal) and a **devlog** (a searchable journal of every working session: what was decided, what was tried, what failed and why). Agents read it before coding and write to it after. Humans can too, but the primary reader is the agent you'll open tomorrow morning.

---

## The problem it solves (why vibecoding needs this)

An AI agent has no memory between sessions. Every new session starts from zero and rediscovers the project by grepping around. Three failure modes follow, and everyone who has vibecoded has met them:

1. **Amnesia.** The agent re-makes a decision that was already made last week — sometimes the opposite one.
2. **Zombie solutions.** The agent re-tries an approach that was already tested and abandoned, because nothing recorded *why* it was abandoned.
3. **Local vision.** The agent sees the 3 files in front of it, not the system. It "fixes" something and breaks two components that depended on it.

Cursor rules don't solve this — more on that below. What solves it is **accumulated, queryable history**:

- `bd devlog resume` — agent loads the last session's context in one command.
- `bd devlog search "auth timeout"` — full-text search over every past session. Finds the fix from three weeks ago instead of re-deriving it (faster, and far fewer tokens than grepping the codebase).
- `bd devlog graph "auth-service"` / `bd devlog impact` — an entity graph built from the logs: which components depend on which, so the agent can estimate blast radius *before* touching something.
- Sessions carry a lifecycle status (`VALIDATED`, `PAUSED`, `ABANDONED` + reason) — so an agent that finds an abandoned experiment knows it's an abandoned experiment, not baseline truth.

It's a **memory plus a search engine** for the project. It doesn't replace rules or dev hygiene; it removes the specific failure modes of working with agents at high frequency.

---

## "It's one big JSON file, it'll conflict constantly" — no, and here's why

This is the core misunderstanding to clear up. Three facts:

### 1. It's JSONL, not JSON

A JSON file is **one single object**: opening brace, everything nested inside, closing brace. Two people touching different tasks still edit the *same structure* — Git sees overlapping changes and conflicts. That objection would be 100% valid for a JSON file.

JSONL (**JSON Lines**) is a different format: **one complete, independent JSON object per line**. One line = one issue. The file is not a structure; it's a list of records, like a CSV.

```jsonl
{"id":"proj-a1","title":"Fix login timeout","status":"closed", ...}
{"id":"proj-b2","title":"Add onboarding flow","status":"open", ...}
{"id":"proj-c3","title":"Refactor API client","status":"in_progress", ...}
```

If I edit issue `a1` and you edit issue `c3`, we changed **different lines**. Git merges different-line changes automatically — same as when we edit different functions in the same source file. Conflicts can only even be *attempted* when we both touch the **same issue**.

### 2. And even same-issue edits don't produce Git conflicts

The repo declares a **custom merge driver** for this file (`.gitattributes`: `.beads/issues.jsonl merge=beads`). When Git would normally show conflict markers, it instead hands the three versions (base, mine, yours) to `bd merge`, which merges **per issue, per field**, using **last-write-wins on the `updated_at` timestamp**. Deletions are handled with tombstones (soft-delete markers), so a deletion on one branch and an edit on another don't corrupt anything.

Translation: you never see `<<<<<<< HEAD` in this file. The worst realistic case is "we both edited the same field of the same issue within the same sync window, and the newer edit won" — which is exactly how Linear, Jira, and Notion resolve the same situation. They just do it on their servers; BeadsLog does it in Git.

### 3. The database is NOT in Git at all

Day-to-day, `bd` doesn't even work against the JSONL. Each machine has a **local SQLite database** (fast queries, full-text search, the graph). That database is **gitignored** — never committed, never shared, never conflicting. The JSONL is only the **interchange format**: `bd sync` exports DB → JSONL, commits, pulls, merges, imports back. So:

| Layer | Where | In Git? | Conflict risk |
|---|---|---|---|
| SQLite DB (queries, search, graph) | Each machine | ❌ ignored | None — local only |
| `issues.jsonl` (one issue per line) | Repo | ✅ | Auto-merged by the `bd` merge driver |
| Devlogs (plain Markdown, one file per session) | Repo | ✅ | Append-only files, each session writes a *new* file — near-zero overlap |

This three-layer design (local DB for speed, line-based text file for Git, custom merge driver for the edge cases) is the same pattern used by [git-bug] and other Git-native trackers. It's a solved problem, not an experiment.

---

## "Then just .gitignore it?" — that inverts the design

Two things wrong with it:

- The noisy parts **already are gitignored** (the SQLite DB, daemon runtime files, sync temp files, lock files). What's committed is precisely the small, mergeable, meant-to-be-shared part.
- Gitignoring the JSONL and devlogs would keep the tool but delete its value: memory would become **per-machine**. Your agent would know things mine doesn't. The whole point is that when you `git pull`, you pull *the team's memory along with the team's code* — same branch, same PR, same review. Issue state travels with the commit that closed the issue.

---

## "A well-structured project with Cursor rules does the same thing" — they're different layers

Rules and devlogs answer different questions:

- **Rules** = *how to work*: conventions, stack, style, do's and don'ts. Static. Written by humans, read by agents. BeadsLog projects have these too (`PROJECT_CONTEXT.md`, agent protocol files) — it doesn't compete with them, it assumes them.
- **Devlog** = *what happened and why*: decisions with their reasons, hypotheses tested, dead ends with the reason they're dead, fixes with their context. **Accumulates automatically as a side effect of working** (recorded before commits by default; can be disabled).

A rule can't tell an agent "we already tried optimistic locking here and reverted it — see session of May 12 for why." Only history can. Rules are the constitution; the devlog is the case law. A well-run project needs both, and no amount of rule-writing produces the second one — especially at vibecoding frequency, where the human doesn't always read every diff and the *reasoning* would otherwise evaporate when the session ends.

There's also a scale problem: rules files grow stale and bloat the context window of every single session. The devlog is **queried on demand** — the agent pulls only the 2-3 relevant sessions for the task at hand.

---

## Security note

Nothing leaves the repo. `bd` is a local binary working on local files; the "database" is a local SQLite file; sync is `git push` to our own remote. There's no third-party service, no telemetry endpoint for project data, no new attack surface beyond "a text file in the repo." If we can commit code to this repo, we can commit its memory.

---

## Honest limitations (so nobody discovers them mid-argument)

- **Same-field, same-window edits resolve by timestamp** (last write wins). Someone's edit can lose. Identical trade-off to every hosted tracker; frequent syncs make the window small.
- **It adds process**: agents must actually record sessions and sync. The default hooks automate this, but a bypassed workflow writes no memory.
- **The devlog's value compounds** — week one it looks like overhead; the payoff shows up the first time an agent (or a new teammate) resumes cold on a months-old decision.
- It's a young tool. The core (JSONL + merge driver + SQLite) is boring, proven tech; the intelligence layer (graph, search ranking) is actively evolving.

---

## TL;DR for the meeting

1. The "single JSON file" doesn't exist: it's **JSONL** (one issue per line), Git merges lines independently, and a **custom merge driver** auto-resolves the rest per-issue. The database itself is gitignored and local.
2. `.gitignore`-ing the shared files would keep the cost and delete the benefit: memory that doesn't travel with `git pull` isn't team memory.
3. Cursor rules and BeadsLog are complementary layers: rules say *how to work*, the devlog records *what happened and why*. Only the second one prevents agents from re-deciding, re-trying, and re-breaking things — which is the actual pain of multiplayer vibecoding.
