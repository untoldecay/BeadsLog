# The BeadsLog Daemon

The daemon (`bd daemon`) is the invisible engine that powers BeadsLog's automation. It ensures your project's memory stays in sync and grows smarter without ever blocking your creative flow.

## Core Responsibilities

### 1. Zero-Latency Synchronization
The daemon monitors your local database and the Git repository. When a commit happens, or when remote changes are pulled, the daemon automatically synchronizes the state in the background. This ensures that commands like `bd list` or `bd ready` always show the latest data without requiring a manual `bd sync`.

### 2. Background AI Enrichment
AI processing (Ollama) is powerful but slow. The daemon solves this by managing an **Enrichment Queue**.
- When you save a devlog, the CLI performs a near-instant Regex extraction.
- The daemon identifies the new session and schedules it for "Tier 2" AI processing.
- Ollama runs in a separate thread, so your terminal remains responsive.

### 3. Crystallization (The Write-Back)
Once the AI discovers new architectural relationships (e.g., `ServiceA -> DatabaseB`), the daemon performs **Crystallization**. It writes these discoveries directly back into your Markdown files.
- **Persistent:** The knowledge is now version-controlled in Git.
- **Self-Healing:** Your documentation improves automatically as you build.
- **Safe:** The daemon uses file locking and re-hashes files after writing to ensure no data is lost or duplicated.

## Importance for Agents
For AI coding agents, the daemon is critical. It allows the agent to move from one task to the next in milliseconds, while the "heavy lifting" of mapping the project's architecture happens silently in the background.

## Commands
- `bd daemon start`: Starts the background worker.
- `bd status`: Shows the current queue length and worker health.
- `bd daemon stop`: Gracefully shuts down the worker.