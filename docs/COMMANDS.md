# BeadsLog Command Reference

BeadsLog combines issue tracking with architectural mapping. Commands are categorized by their primary role in the workflow.

## 🚦 Setup & Modes
Choose how beads relates to git. See [Running Modes](RUNNING_MODES.md) for why/how.

| Command | Usage | Description |
| :--- | :--- | :--- |
| `bd init` | `bd init` | Initialize. With a remote, beads commits to a dedicated `beads-metadata` branch (off your work branch). |
| `bd init --inline` | `bd init --inline` | Commit beads on the **current** branch instead of a dedicated one. |
| `bd init --solo` | `bd init --solo` | Private, local-only: never pushed, invisible to the team (Invisible or local `beads-local` branch). |
| `bd init --team` | `bd init --team --force` | (Re)join a team; publishes local devlogs, restores shared config. |
| `bd config set sync-branch` | `bd config set sync-branch <name>` | Rename the dedicated sync branch. |

## 📋 Issue Tracking (The "Beads")
Use these to manage the "Forward" flow of tasks.

| Command | Usage | Description |
| :--- | :--- | :--- |
| `bd sync` | `bd sync` | Synchronize local database with Git JSONL. |
| `bd ready` | `bd ready` | List unblocked tasks prioritized for the agent. |
| `bd create` | `bd create "Title"` | Create a new task or epic. |
| `bd update` | `bd update <id> --status in_progress` | Update task state or assignee. |
| `bd close` | `bd close <id>` | Mark task as completed. |
| `bd catchup`| `bd catchup [--ack]` | **NEW**: See team activity since your last check. |
| `bd status` | `bd status` | Show database health and sync state. |

## 🔍 Devlog & Memory (The "Scouting")
Use these to query and maintain the knowledge graph.

| Command | Usage | Description |
| :--- | :--- | :--- |
| `bd devlog sync` | `bd devlog sync` | Fast-sync (Regex) of new markdown files. |
| `bd devlog search` | `bd devlog search "query"` | Hybrid search (BM25 + Graph). |
| `bd devlog resume` | `bd devlog resume --last 1` | Load context from previous work. |
| `bd devlog path` | `bd devlog path "A" "B"` | Trace the chain of sessions linking two entities. |
| `bd devlog pause` | `bd devlog pause --scope ...`| **NEW**: Park work on a branch/task with a reason. |
| `bd devlog abandon`| `bd devlog abandon --scope ...`| **NEW**: Kill a flawed path and document the 'Why'. |
| `bd devlog entities`| `bd devlog entities` | Show most frequently mentioned components. |
| `bd devlog graph` | `bd devlog graph "entity"` | Visualize architectural dependencies. |
| `bd devlog impact` | `bd devlog impact "entity"` | See what depends on a specific component. |
| `bd devlog verify` | `bd devlog verify [--fix]` | Adopt orphans and backfill metadata. |
| `bd devlog record` | `bd devlog record --subject "..." --file "..."` | Automatically record a session in the index. |

## 🛠 Maintenance & Debugging
Useful when managing the environment.

| Command | Usage | Description |
| :--- | :--- | :--- |
| `bd upgrade check`| `bd upgrade check [--install]` | **NEW**: Interactive update check from main repo. |
| `bd daemon start` | `bd daemon start` | Start the background sync and AI worker. |
| `bd onboard` | `bd onboard` | Refresh agent rules and instructions. |
| `bd doctor` | `bd doctor --fix` | Repair common installation issues. |
| `bd config set` | `bd config set ollama.model ...`| Manage persistent settings. |
| `bd devlog reset` | `bd devlog reset` | Clear local devlog cache (Safe, no file loss). |
