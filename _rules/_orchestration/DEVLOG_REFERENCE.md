# Devlog Commands

⚠️ **Load ONLY when bd devlog --help insufficient**

## Search
```bash
bd devlog search "nginx timeout" --preview  # Show snippets
bd devlog search "modal" --explain         # See ranking breakdown
bd devlog search "auth" --limit 10
bd devlog list --last 5
```

## Architecture
```bash
bd devlog graph "nginx"                    # Explicit + Implicit relationships
bd devlog path "AuthService" "JWT"         # Find historical chain between entities
bd devlog impact "AuthService"             # Reverse dependency check
bd devlog status                           # Show system health & AI queue
bd devlog verify --fix
```

## Maintenance
```bash
bd devlog sync        # Ingest new markdown files
bd devlog verify      # Check for missing metadata
bd devlog reset       # Clear local cache (rare)
```
