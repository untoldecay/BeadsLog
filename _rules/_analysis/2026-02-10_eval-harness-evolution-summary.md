# Analysis: Evolution of the OpenCode Evaluation Harness

**Date:** 2026-02-10
**Context:** This document consolidates the key narratives and technical pivots from the last five development sessions (Feb 8 - Feb 10), documenting the resolution of the "infernal loop" of file reversions and the stabilization of the evaluation framework.

---

## 1. The Pivot: From Gemini CLI to OpenCode (Feb 9)
The initial attempts to use Gemini CLI for headless evaluations were plagued by infinite loops (specifically the agent's addiction to `bd prime`) and authentication hangs. 

**Key Pivot:** We replaced the Gemini CLI wrapper with a native integration of the OpenCode CLI.
- **Why:** OpenCode supports native MCP discovery via `opencode.json`, eliminating the need for complex instruction injection to find tools.
- **Outcome:** Achieved stable, structured JSONL tracing of agent tool calls.

## 2. Conquering the "Infernal Loop" (Feb 10)
During stabilization, we encountered a critical failure mode: agent modifications within the sandbox were being silently reverted.

**Root Causes:**
1. **Index Conflict:** Git worktrees were sharing or conflicting with the main repository's state.
2. **Git Stash Interaction:** The `eval` tool's automatic stashing was reverting uncommitted harness code *on the developer's disk* before the test ran, causing the harness to run with "old" logic.

**Solutions:**
- **Detached HEAD Isolation:** Switched to `git worktree add -d` and implemented a "Fix 1 + 3" strategy: clearing the cached index and forcing a fresh checkout from HEAD.
- **Main Repo Commitment:** Explicitly committed the harness code to protect it from the `git stash` lifecycle.
- **Configuration Safeguards:** Updated `opencode.json` to disable Git auto-management, preventing the agent from triggering its own reset loops.

## 3. Real-Time Visibility & Token Accuracy (Feb 9-10)
Evaluations were initially "blind," providing no feedback during the long execution window and reporting 0 tokens.

**Enhancements:**
- **Scrolling Viewport:** Implemented a 10-line live stream window showing real-time `tool_use(args)` events.
- **Nested Usage Parsing:** Identified that OpenCode reports tokens within `step_finish` events. Updated `aggregateEvent` to accurately sum input, output, and total tokens.

## 4. Behavioral Enforcement (Feb 8)
Before the harness was stable, we addressed agent "panic" and "downgrading" behaviors.
- **Anti-Panic Protocol:** Added guidance footers to CLI errors to prevent agents from falling back to `grep` when a `bd` command fails.
- **Idempotency Fix:** Resolved a bug where `bd onboard` would downgrade a finalized "Ready" protocol back to a setup trap.

---

## Consolidated Devlog History (Reference)

### [2026-02-10] Finalize OpenCode Eval Harness
Implemented Detached HEAD isolation, forced index resets, and nested token parsing. Added automatic stashing and an interactive `huh` post-run menu.

### [2026-02-09] Pivot Eval Harness to OpenCode CLI
Switched from Gemini CLI to OpenCode to resolve loop and auth issues. Implemented initial stream parsing and isolated the harness in `eval_task_opencode.go`.

### [2026-02-08] Validate Worktree-Based Eval System
Validated the feasibility of using `git worktree` for isolation, identifying the necessity of the `BEADS_DIR` environment override.

### [2026-02-08] Anti-Panic Protocol Enforcement
Injected recovery guidance into CLI error outputs to maintain protocol compliance during failures.

### [2026-02-08] Add Help Tips to Command Outputs
Added contextual `💡 Tip: Use --help` reminders to improve tool discoverability for both humans and agents.
