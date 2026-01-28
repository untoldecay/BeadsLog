# Git Hooks & Automation

BeadsLog uses Git Hooks to ensure that your project's memory is updated automatically as part of your natural Git workflow.

## Installed Hooks

### 1. `pre-commit` (The Enforcer)
The `pre-commit` hook ensures that every code change is documented. If you try to commit code without a corresponding devlog entry, the commit is blocked.
- **Benefit:** Prevents "knowledge decay" by ensuring no decision goes unrecorded.
- **Override:** Can be bypassed with `git commit --no-verify` (not recommended).

### 2. `post-commit` & `post-merge` (The Triggers)
These hooks trigger the daemon to perform a background sync immediately after a commit or a pull.
- **Benefit:** Ensures your local database and architectural graph are always in sync with the latest code state.

### 3. `prepare-commit-msg` (The Trailer)
This hook automatically appends "Agent Identity" trailers to your commit messages if you are working in an agent-assisted environment. This provides an audit trail of which agent performed which task.

## Manual Management
You can manage your hooks using the following commands:
- `bd hooks install`: Automatically installs and wires all necessary hooks.
- `bd hooks status`: Verifies that hooks are active and correctly configured.

## Architecture
The hooks are designed to be "Thin Clients." They don't perform the heavy synchronization logic themselves; instead, they send a fast signal to the **Daemon**, which then handles the task in the background. This ensures that `git commit` remains fast even in large repositories.
