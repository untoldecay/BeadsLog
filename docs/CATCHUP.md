# Project Catchup (`bd catchup`)

**Problem:** In fast-moving projects with multiple contributors (human or AI), it is easy for developers to lose track of what was completed on other branches or what decisions were made during their absence.

**Solution:** `bd catchup` provides a localized, project-specific activity feed. It identifies all sessions recorded, issues closed, and state changes (pauses/abandonments) that have occurred since you last checked.

## 🚀 How to Use

### 1. View New Activity
```bash
bd catchup
```
This prints a raw delta of project progress since your last "acknowledged" catchup.

### 2. The Feature Digest (recommended for humans)
```bash
bd catchup --digest
```
Groups the same activity **by feature arc** — sessions sharing a work branch form an arc; trunk work is grouped by its dominant entity. Each arc opens with a 2-3 line narrative extracted from real session content (the first session's Problem and the latest Final Status), followed by its sessions, key entities, and any lifecycle changes that hit that arc since you last looked ("⚠ PAUSED since you last looked — ..."). Add `--json` for machine-readable output.

### 3. Acknowledge and Reset
```bash
bd catchup --ack
```
Once you are up to speed, use the `--ack` flag to update your local "seen" timestamp.

### 4. Agent-Assisted Summaries
For a high-signal brief, ask your agent to use the catchup prompt:
> *"I've run bd catchup. Please use `_rules/_devlog/_generate-catchup.md` to summarize this activity for me, focusing on architectural drifts and blocked paths."*

## 🛠 How it Works
BeadsLog stores a `last_catchup_time` in the SQLite `metadata` table. Because this database is synced via Git (`bd sync`), your catchup status follows you across different machines and branches.

## 👥 Benefits for Teams
- **Zero Noise:** Only shows what happened *since your last check*.
- **Attribution:** Every item in the feed includes the Author/Actor who performed the work.
- **Contextual Alignment:** Ensures that agents don't re-propose solutions that were just implemented by someone else.
