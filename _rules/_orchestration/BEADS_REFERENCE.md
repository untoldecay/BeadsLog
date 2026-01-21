# Beads Commands

⚠️ **Load ONLY when bd --help insufficient**

## Issue Lifecycle
```bash
bd new "Title" --type bug --priority high
bd ready                           # P0 issues
bd update <id> --status in-progress
bd update <id> --assign @teammate
bd close <id>
bd split <id> "Sub-task"
bd block <current> <blocking>
```

## Query
```bash
bd list --status open
bd show <id>
bd search "keywords"
```

## Sync & Versioning
```bash
bd sync               # Sync local DB <-> JSONL <-> Git
bd deploy staging     # Deploy to staging environment
bd deploy production  # Deploy to production environment
```
