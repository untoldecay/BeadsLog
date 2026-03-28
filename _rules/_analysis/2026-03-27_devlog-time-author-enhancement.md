# Analysis: Devlog Time Precision and Author Attribution

**Date:** 2026-03-27
**Branch:** `dev/devlog-time-author-01`
**Status:** Implementation Complete & Verified

## 1. Objective
Assess the impact and dependencies of enhancing the BeadsLog devlog system to include granular time logging (`HH:MM`) and developer attribution (GitHub author). This involves modifying the generation prompt, the index structure, the database schema, and the ingestion logic.

## 2. Current State Assessment

### 2.1 Index Structure
The `_rules/_devlog/_index.md` currently uses a 4-column Markdown table:
`| Subject | Problems | Date | Devlog |`
Where `Date` is formatted as `YYYY-MM-DD`.

### 2.2 Parsing Logic
`parseIndexMD` in `cmd/bd/devlog_core.go` strictly validates the number of pipes (exactly 5 for 4 columns). It parses the `Date` column using `YYYY-MM-DD` or `Jan 2` layouts.

### 2.3 Data Model
The `sessions` table in SQLite has a `timestamp` column (DATETIME) but no `author` column.
The `sessions_fts` virtual table includes `title` and `narrative` but lacks `author`.

### 2.4 User Attribution Mechanisms
BeadsLog currently distinguishes between:
- **Actor (Name):** Retrieved via `git config user.name`. Used for audit trails (who executed).
- **Owner (Email):** Retrieved via `git config user.email`. Used for CV attribution (Decision 008 - human responsibility).
For devlogs, the **Actor (Name)** is more appropriate for display, while the **Owner (Email)** provides unique identification across repositories.

## 3. Proposed Enhancements

### 3.1 Metadata Granularity
- **Time:** Update the `Date` column in `_index.md` to support `YYYY-MM-DD HH:MM`.
- **Author:** Add an `Author` column to `_index.md`.
New format: `| Subject | Problems | Author | Date | Devlog |`

### 3.2 Automation Prompt & Command
**Shift from Manual Editing to Automated Recording:**
To avoid manual table corruption and reduce token usage, introduce `bd devlog record`:
```bash
bd devlog record --subject "[fix] bug-id" --problem "Description" --file "path/to/log.md"
```
This command will:
1. **Automatic Time:** Automatically retrieve the current system time and format it as `YYYY-MM-DD HH:MM`.
2. **Automatic Author:** Retrieve `user.name` and `user.email` from git config.
3. **Safe Append:** Format the Markdown row according to the new 5-column schema and append it to `_index.md` safely.
4. **Sync:** Trigger an immediate `bd devlog sync`.

**Mandatory Fields for `bd devlog record`:**
- `--subject`: (e.g., `[fix] issue-123`) The subject of the session.
- `--problem`: Brief description of the problem solved.
- `--file`: Relative path to the devlog markdown file.
- *Automatic:* `--author` (from `git config user.name`), `--date` (current time `YYYY-MM-DD HH:MM`).

### 3.3 CLI & Logic Updates
...
- **`parseDate`:** Add support for `2006-01-02 15:04` layout.
- **`SyncSession` Logic (ID Stability):** 
  - To prevent ID rotation when adding time to an existing log, `SyncSession` will be updated to perform a **Lookup-by-Filename** before generating a new ID. 
  - If a session with the same `filename` exists, it retains its original `sessionID` even if the `Date` or `Subject` in the index is updated.

## 4. Integration & Onboarding Flow

### 4.1 Automatic Migration during `init`
`bd init` and `bd devlog initialize` will automatically detect if an existing `_index.md` uses the legacy 4-column format. 
- If detected, they will run `bd devlog migrate --format-index` silently (or with a notice in non-quiet mode) to ensure the environment is ready before any sync occurs.

### 4.2 Agent Onboarding (`bd onboard`)
The `bd onboard` command is the entry point for AI agents. It will be updated to:
1. **Format Check:** Verify the devlog index format.
2. **Guidance:** Explicitly mention the new `bd devlog record` command as the standard way to persist session memory.
3. **Protocol Refresh:** Ensure the `_generate-devlog.md` prompt is correctly placed and matches the current configuration (Time + Author enabled).

### 4.3 Agent Protocol Updates (GEMINI.md / AGENTS.md)
The core agent protocols MUST be updated to reflect the new automated workflow:

**Updated "CLOSE" Protocol:**
```markdown
🚨 SESSION CLOSE PROTOCOL (REQUIRED) 🚨
1. Generate devlog: `cat _rules/_devlog/_generate-devlog.md`
2. Record session: `bd devlog record --subject "[fix]..." --problem "..." --file "..."`
3. git add <files> && git commit -m "..."
4. bd sync && bd devlog sync
5. git push
```

## 5. Implementation Roadmap

### 5.1 Phase 1: Core Logic
- Update `IndexRow` and `parseIndexMD` for 5 columns.
- Implement `bd devlog record` and `bd devlog migrate`.
- Update `SyncSession` with **Lookup-by-Filename** ID stability.

### 5.2 Phase 2: User Experience
- Implement `bd devlog authors`.
- Add `--author` filters to `list`, `search`, and `resume`.
- Update `printBlockerMessage` in `check_cli.go`.

### 5.3 Phase 3: Protocol & Init
- Integrate migration into `bd init`.
- Update `GEMINI.md` and `AGENTS.md` templates.

## 6. Verification & Testing Strategy

To ensure the stability of project memory and the reliability of the new automated recording workflow, the following testing tiers MUST be implemented:

### 6.1 Unit Testing (`*_test.go`)
- **Parser Stability:** `parseIndexMD` tests must verify backward compatibility with 4-column tables and strict validation of the new 5-column format.
- **Date Precision:** `parseDate` tests for `YYYY-MM-DD` (legacy) and `YYYY-MM-DD HH:MM` (new).
- **ID Stability:** Verify `hashID` logic remains stable when time metadata is added to an existing file entry (via Lookup-by-Filename).
- **Command Logic:** Unit tests for `record` and `migrate` internal logic (appending, formatting, and git detection).

### 6.2 Integration Testing
- **End-to-End Sync:** Verify that `bd devlog record` followed by `bd devlog sync` results in the correct `author`, `author_email`, and high-precision `timestamp` in the SQLite database.
- **Filtering Logic:** Verify SQL query generation for `--author` flags in `list`, `search`, and `resume`.

### 6.3 E2E Sandbox Validation (`_sandbox/devlog-tpa-e2e/`)
A dedicated sandbox script `test_enhancement.sh` will be created to perform the following:
1. **Bootstrap:** Initialize a fresh Beads repo.
2. **Legacy State:** Create a legacy 4-column index with existing devlogs.
3. **Migration:** Run `bd devlog migrate --format-index` and verify filesystem transformation.
4. **Recording:** Execute `bd devlog record` and verify the entry in `_index.md`.
5. **Database Verify:** Run `bd devlog sync` and query the database for author/time precision.
6. **Command Verify:** Run `bd devlog authors` and `bd devlog list --author` to ensure metrics and filters work.
7. **Enforcement Verify:** Trigger `bd check --hook pre-commit` without a log and verify the new instruction message.
**Status:** Implementation Complete & Verified

## 8. Implementation Notes (Final)
The implementation followed the roadmap with a few critical refinements:
1. **Author Override:** Added a `--author` flag to `bd devlog record` to handle environments where `git config` is missing (e.g., restricted agent environments).
2. **Sync Sensitivity:** Updated `SyncSession`'s `needsUpdate` logic to be sensitive to `author` and `author_email` changes, ensuring that index-only metadata updates trigger a database refresh even if the underlying markdown file hasn't changed.
3. **FTS Rebuild:** Confirmed that dropping and recreating the FTS table is the most reliable path for adding searchable metadata columns in SQLite.

## 9. Conclusion
The devlog system is now a robust, multi-user project memory with high-precision time tracking. The introduction of `bd devlog record` successfully eliminates the primary source of index corruption while significantly reducing the effort required for agents to maintain their development logs.

---
*Report by BeadsLog Agent - 2026-03-27*
