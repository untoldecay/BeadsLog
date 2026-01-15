# Agent Instructions

This project uses **bd** (beads) for issue tracking. Run `bd onboard` to get started.

## Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>         # Complete work
bd sync               # Sync with git
```

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds


## Devlog Protocol (MANDATORY)

This repo uses **Beads Devlog** to persist context and prevent regressions. You must follow this protocol:

### 1. 🟢 Start of Session
- **Resume Context:** Run `bd devlog resume --last 1` to load the immediate history.
- **Goal:** Understand what happened last time to avoid repeating mistakes.

### 2. 🟡 During Work
- **Bug Encounter:** If you hit a bug, run `bd devlog search "error message"` to see if it's a known issue.
- **Information Request:** If asking about a component, run `bd devlog impact <component>` to understand dependencies.
- **Planning:** If making a plan, check `bd devlog graph <entity>` to verify architectural assumptions.

### 3. 🔴 End of Session
- **Log & Sync:**
  1. Generate the session log using `_rules/_devlog/_generate-devlog.md`.
  2. Run `bd devlog sync` to persist it to the database.

### 4. 🔎 Metadata Audit
- **Verification:** Occasionally run `bd devlog verify`.
- **Self-Fix:** If sessions are missing metadata, run `bd devlog verify --fix` and follow the instructions to re-analyze them.

### 5. 🩹 Self-Correction
- **Directives:** If a command output starts with `🚀 **AI ... DIRECTIVE**`, you MUST prioritize fixing the reported issue and then **RE-RUN** the failed command immediately.
