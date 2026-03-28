# BeadsLog Command Reference

BeadsLog combines issue tracking with architectural mapping. Commands are categorized by their primary role in the workflow.

## 📋 Issue Tracking (The "Beads")
Use these to manage the "Forward" flow of tasks.

| Command | Usage | Description |
| :--- | :--- | :--- |
| `bd sync` | `bd sync` | Synchronize local database with Git JSONL. |
| `bd ready` | `bd ready` | List unblocked tasks prioritized for the agent. |
| `bd create` | `bd create "Title"` | Create a new task or epic. |
| `bd update` | `bd update <id> --status in_progress` | Update task state or assignee. |
| `bd close` | `bd close <id>` | Mark task as completed. |
| `bd status` | `bd status` | Show database health and sync state. |

## 🔍 Devlog & Memory (The "Scouting")
Use these to query and maintain the knowledge graph.

| Command | Usage | Description |
| :--- | :--- | :--- |
| `bd devlog sync` | `bd devlog sync` | Fast-sync (Regex) of new markdown files. |
| `bd devlog search` | `bd devlog search "query"` | Hybrid search (BM25 + Graph). |
| `bd devlog resume` | `bd devlog resume --last 1` | Load context from previous work. |
| `bd devlog entities`| `bd devlog entities` | Show most frequently mentioned components. |
| `bd devlog graph` | `bd devlog graph "entity"` | Visualize architectural dependencies. |
| `bd devlog impact` | `bd devlog impact "entity"` | See what depends on a specific component. |
| `bd devlog verify` | `bd devlog verify [--fix]` | Adopt orphans and backfill metadata. |
| `bd devlog record` | `bd devlog record --subject "..." --problem "..." --file "..."` | Automatically record a session in the index with dual attribution. |
| `bd devlog migrate`| `bd devlog migrate --format-index` | Upgrade legacy 4/5-column indexes to the 6-column standard. |
| `bd devlog authors`| `bd devlog authors` | List contributors (Human + Agent) and their session metrics. |
| `bd devlog enrich` | `bd devlog enrich --all` | Schedule sessions for background AI update. |
| `bd devlog extract`| `bd devlog extract [target]`| Foreground AI regeneration for a session. |

## 🛠 Maintenance & Debugging
Useful when managing the environment.

| Command | Usage | Description |
| :--- | :--- | :--- |
| `bd daemon start` | `bd daemon start` | Start the background sync and AI worker. |
| `bd onboard` | `bd onboard` | Refresh agent rules and instructions. |
| `bd doctor` | `bd doctor --fix` | Repair common installation issues. |
| `bd config set` | `bd config set ollama.model ...`| Manage persistent settings. |
| `bd devlog reset` | `bd devlog reset` | Clear local devlog cache (Safe, no file loss). |
