# Devlog Commands

⚠️ **Load ONLY when bd devlog --help insufficient**

## Search
```bash
bd devlog search "nginx timeout"
bd devlog search "modal" --type fix
bd devlog search "auth" --since 2026-01
bd devlog list --last 5
```

## Architecture
```bash
bd devlog graph "nginx"
bd devlog impact "AuthService"
bd devlog status
bd devlog verify --fix
```

## Maintenance
```bash
bd devlog sync        # Ingest new markdown files
bd devlog verify      # Check for missing metadata
bd devlog reset       # Clear local cache (rare)
```
