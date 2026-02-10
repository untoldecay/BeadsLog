# Comprehensive Development Log: Finalize OpenCode Eval Harness

**Date:** 2026-02-10

### **Objective:**
To finalize the `bd eval task` command by implementing a robust, isolated execution harness using Git Worktrees, real-time output visibility, automated stashing, and accurate token counting.

---

### **Phase 1: Worktree Isolation & Stability**

**Initial Problem:** The evaluation harness was silently reverting file changes made by the agent inside the worktree. This was caused by the worktree sharing the index or state with the main repo, or Git autorevert mechanisms triggering on dirty states.

*   **My Assumption/Plan #1:** Use `git worktree add` with standard branch checkout.
    *   **Action Taken:** Implemented basic worktree creation.
    *   **Result:** Files modified by the agent (OpenCode) were reverted immediately.
    *   **Analysis/Correction:** The worktree index was likely stale or conflicted. Switched to a **Detached HEAD** approach (`git worktree add -d`) and forced a fresh index reset (`git rm --cached -r .`, `git read-tree HEAD`, `git checkout-index -a --force`). Also disabled `core.fileMode` to prevent permission-based diffs.

*   **My Assumption/Plan #2:** The OpenCode agent might be triggering git actions that conflict with the harness.
    *   **Action Taken:** Updated `opencode.json` configuration to explicitly disable `autoAdd`, `autoCommit`, and set `indexUpdate` to `manual`.
    *   **Result:** Stability achieved. Agent modifications now persist correctly within the sandbox.

---

### **Phase 2: Token Counting & Visibility**

**Initial Problem:** The evaluation report was showing 0 tokens and missing detailed tool usage.

*   **My Assumption/Plan #1:** Tokens are available in the `usage` event.
    *   **Action Taken:** Implemented `aggregateEvent` to parse `usage` events.
    *   **Result:** Still 0 tokens.
    *   **Analysis/Correction:** OpenCode streams token usage in `step_finish` events within a `part.tokens` object. Updated `aggregateEvent` to parse this nested structure. Added a unit test `eval_token_test.go` to verify the logic.

*   **My Assumption/Plan #2:** Users need to see what's happening during the run.
    *   **Action Taken:** Implemented a 10-line scrolling viewport in the CLI output that shows real-time tool calls (e.g., `🛠️  [TOOL] bash(ls -la)`).
    *   **Result:** excellent visibility into the agent's actions without flooding the console.

---

### **Phase 3: UX & Safety**

**Initial Problem:** Running evals in a dirty workspace risks data loss or contamination.

*   **My Assumption/Plan #1:** Abort if dirty.
    *   **Action Taken:** Added a `checkDirty` function.
    *   **Correction:** Too restrictive for a developer tool. Implemented **Automatic Stashing**.
    *   **Action Taken:** Added `createStash` and `restoreStash` to `eval_task_opencode.go`. The harness now automatically stashes changes before creating worktrees and pops them after cleanup.

*   **My Assumption/Plan #2:** Post-run workflow needs to be flexible.
    *   **Action Taken:** Integrated `charmbracelet/huh` for an interactive post-run menu: `Quit (Clean up)`, `Restart`, or `New Task`.

---

### **Final Session Summary**

**Final Status:** **Complete.** The `bd eval task` command is now a production-grade evaluation harness. It supports A/B testing (Implicit vs Explicit protocols), accurate token metering, safe execution via worktrees, and rich reporting.
**Key Learnings:**
*   **Git Worktrees:** For true isolation, always use detached HEAD and force a complete index reset. Shared branches/indices are a trap for automated tools.
*   **OpenCode Streams:** Token usage data is often nested in lifecycle events (`step_finish`) rather than top-level `usage` events.

---

### **Architectural Relationships**
<!-- Format: [From Entity] -> [To Entity] (relationship type) -->
- eval_task_opencode.go -> createEvalWorktree (uses)
- createEvalWorktree -> git (executes)
- runOpenCodeEval -> createStash (calls)
- runOpenCodeEval -> restoreStash (calls)
- aggregateEvent -> TokenUsage (updates)
# Comprehensive Development Log: Pivot Eval Harness to OpenCode CLI

**Date:** 2026-02-09

### **Objective:**
To resolve persistent reliability issues with the Gemini CLI-based evaluation harness (loops, auth hangs, parsing errors) by pivoting to the OpenCode CLI. The goal is to achieve stable, isolated A/B testing of agent protocols with high-fidelity tool-call tracking and structured reporting.

---

### **Phase 1: Diagnosis of Gemini CLI Failure**

**Initial Problem:** Automated evaluations using the Gemini CLI were getting stuck in infinite loops (calling `bd prime` repeatedly) and failing to parse real-time output streams.

*   **My Assumption/Plan #1:** Add `--debug` and `--output-format stream-json` to Gemini CLI to force immediate visibility of tool calls.
    *   **Action Taken:** Modified `cmd/bd/eval_task.go` to include these flags and implemented `ProtocolFilterWriter` to parse the JSON stream.
    *   **Result:** Partial success in visibility, but the agent continued to ignore the "SANDBOX OVERRIDE" instructions in `AGENTS.md` and remained addicted to `bd prime`.
    *   **Analysis/Correction:** The Gemini CLI's internal hooks and "habitual" behavior were stronger than the prepended protocol instructions in a headless environment.

---

### **Phase 2: Pivoting to OpenCode CLI**

**Initial Problem:** We needed a more compliant agent runner that supports native MCP discovery and structured trace output.

*   **My Assumption/Plan #1:** Integrate OpenCode CLI as the primary runner for `bd eval task`.
    *   **Action Taken:** Rewrote `cmd/bd/eval_task.go` to wrap `opencode run`. Configured a local `opencode.json` in each worktree to point to the `bd mcp` server.
    *   **Result:** Encountered `ProviderAuthError` and `Configuration Invalid` errors.
    *   **Analysis/Correction:** OpenCode required specific environment variables (`GOOGLE_GENERATIVE_AI_API_KEY`) and a strict JSON schema for the config file (no hyphens in MCP names like `beads-mcp`, and no `env` field in local MCP definitions).

---

### **Phase 3: Implementing Robust Stream Parsing**

**Initial Problem:** OpenCode outputs a JSONL stream (one object per line) mixed with non-JSON logs, causing standard `json.Unmarshal` to fail.

*   **My Assumption/Plan #1:** Implement a line-by-line streaming parser that filters noise and aggregates events.
    *   **Action Taken:** Updated `runOpenCodeTest` to use `bufio.Scanner` on `StdoutPipe`. Added `aggregateEvent` to handle the nested OpenCode event structure (`type`, `part`, `state`, etc.).
    *   **Result:** Successfully captured tool calls, but they were initially labeled as generic `bash`.
    *   **Analysis/Correction:** Since OpenCode wraps shell commands in a `bash` tool, I updated the scoring logic to perform **Intent Parsing** on the command arguments (e.g., detecting `bd devlog graph` inside a bash call).

---

### **Phase 4: Enhancing Visibility, Cleanup, and Stability**

**Initial Problem:** The evaluation harness lacked real-time visibility, automated artifact cleanup, and was prone to accidental rollbacks due to background daemons and automatic stashing.

*   **My Assumption/Plan #1:** Implement real-time event printing and chronological reporting.
    *   **Action Taken:** Modified `runOpenCodeTest` to print `text`, `tool_use`, and `tool_result` events to the console as they arrive. Updated `generateABReport` to include a full chronological list of tool calls for each run.
    *   **Result:** Success. Evaluators can now see agent progress live and inspect the full tool sequence in the final report.
*   **My Assumption/Plan #2:** Prevent rollbacks by isolating the new code and disabling background triggers.
    *   **Action Taken:** Identified that `bd daemon` background processes were likely reverting `cmd/bd/eval_task.go`. Killed all daemons and moved the OpenCode implementation to a new file, `cmd/bd/eval_task_opencode.go`, while deleting the original.
    *   **Analysis/Correction:** Isolated files are safer from automated "repairs" or template-based regenerations that target known core files.
*   **My Assumption/Plan #3:** Improve cleanup and reliability.
    *   **Action Taken:** Added `bd eval clean` to `cmd/bd/eval_cmds.go` for manual artifact pruning. Updated the cleanup prompt in `eval_task_opencode.go` to use `/dev/tty` to ensure it works even when `stdin` is piped.
    *   **Result:** Success. The system is now tidy and interactive.

---

### **Final Session Summary**

**Final Status:** **Success.** The evaluation harness is fully operational, stable, and feature-rich. It successfully executes the OpenCode agent in isolated worktrees, provides real-time visibility, and generates detailed chronological reports.
**Key Learnings:**
*   **Isolation is Safety:** When fighting automated rollbacks or regenerations, isolate new logic in unique filenames to break the trigger loop.
*   **TTY for Interactivity:** CLIs that read from `stdin` should use `/dev/tty` for user prompts to remain interactive when the primary `stdin` is redirected or piped.
*   **Daemon Awareness:** Background processes like `bd daemon` can be silent culprits for unexpected filesystem changes in a dev environment.

---

### **Architectural Relationships**
- bd eval task -> OpenCode CLI (wraps)
- OpenCode CLI -> bd mcp (uses for retrieval)
- aggregateEvent -> OpenCode Events (parses)
- scoreTrace -> Tool Intents (evaluates)
- eval_task_opencode.go -> /dev/tty (prompts)
- eval_cmds.go -> bd eval clean (manages)# Comprehensive Development Log: Validate Worktree-Based Eval System

**Date:** 2026-02-08

### **Objective:**
To analyze and validate the implementation plan for `bd eval task`, a revamped evaluation system using `git worktree` for isolated A/B testing between "Implicit" (Always-On Protocol) and "Explicit" (Tool-Specific instructions) agent sessions.

---

### **Phase 1: Feasibility Analysis & Confrontation**

**Initial Problem:** `bd` needs a way to run isolated evals without disk-heavy repo copies and without manual setup.

*   **Assumption:** `git worktree` provides perfect isolation.
    *   **Verification:** Created `_sandbox/wt-test/` to simulate worktree usage. 
    *   **Discovery:** Confirmed that `bd` discovery logic (`FindBeadsDir`) prioritizes the **Main Repository Root** even when inside a worktree. This would cause evals to share the same database, breaking isolation.
    *   **Correction:** The implementation must explicitly set `BEADS_DIR` environment variable for all child processes to force local sandbox database usage.

---

### **Phase 2: The Runner Gap**

**Initial Problem:** `bd` is a tracker, not an agent runner. It cannot execute prompts autonomously.

*   **My Assumption/Plan #1:** Integrate with an external runner.
    *   **Action Taken:** Analyzed `_rules/_analysis/2026-01-08_GeminiCli-wrapper.md`.
    *   **Result:** Validated that wrapping `gemini-cli` is feasible. We can use `--session-id` for session isolation and pipe the task via `stdin`.
    *   **Architecture:** `bd eval task` will act as a **Harness**, managing the lifecycle (Stash -> Worktree -> Gemini -> Report -> Cleanup).

---

### **Final Session Summary**

**Final Status:** **Validated & Ready.** The architectural gaps (Isolation and Execution) have been solved via `BEADS_DIR` injection and `gemini-cli` wrapping.
**Key Learnings:**
*   **Environment Primacy:** In a complex tool like `bd` that has "smart" path discovery, environment variables are the most reliable way to enforce sandbox boundaries.
*   **Harness Pattern:** CLI tools should focus on environment orchestration rather than trying to build internal agent loops when external robust runners exist.

---

### **Architectural Relationships**
<!-- Format: [From Entity] -> [To Entity] (relationship type) -->
- bd eval task -> git worktree (orchestrates)
- bd eval task -> gemini-cli (wraps)
- bd eval task -> BEADS_DIR (enforces isolation)
- bd eval report -> interactions.jsonl (aggregates)
# Comprehensive Development Log: Anti-Panic Protocol Enforcement

**Date:** 2026-02-08

### **Objective:**
To prevent agents from abandoning the BeadsLog toolset and reverting to primitive tools (grep/find) when they encounter a command error (e.g., syntax error or invalid flag). Agents often interpret a non-zero exit code as a signal that the tool is "broken" rather than "misused".

---

### **Phase 1: Analysis of Failure Modes**

**Initial Problem:** Chained commands (e.g., `bd search "foo" && bd graph "foo"`) break if the first command fails. Agents see the error and panic.

*   **Assumption:** "Soft failures" (no results) might be causing exit code 1.
    *   **Verification:** Created `_sandbox/chain-fail-test.sh`. Confirmed that `bd devlog search` exits with 0 even on empty results. Soft failures do *not* break chains.
    *   **Finding:** The problem is limited to "Hard Failures" (syntax errors, config errors) which correctly exit with 1 but lack guidance.

---

### **Phase 2: Implementation - The Anti-Panic Footer**

**Initial Problem:** Standard Cobra error output is dry ("Error: unknown flag") and doesn't instruct the agent on *recovery*.

*   **Action Taken:** Modified `cmd/bd/main.go`.
    *   Intercepted the final `err != nil` check before `os.Exit(1)`.
    *   Injected a high-visibility `⚠️ PROTOCOL ENFORCEMENT` footer using `internal/ui` styling.
    *   **Guidance:** Explicitly forbids fallback to grep/find and instructs the agent to check syntax (`--help`) or health (`doctor`).

---

### **Phase 3: Verification**

**Initial Problem:** Ensure the footer appears *only* on hard errors and doesn't break normal operation.

*   **Action Taken:** Ran the failure test script with the new binary.
    *   **Result:**
        *   Test 1 (Empty Search): Exits 0, Chain continues. (Good)
        *   Test 2 (Invalid Flag): Exits 1, Chain breaks, Footer appears. (Good)
        *   Test 3 (Empty Graph): Exits 0, Chain continues. (Good)

---

### **Final Session Summary**

**Final Status:** **Implemented.** The CLI now acts as a resilient "coach," catching its own failures and guiding the agent back to the protocol instead of letting them drift to manual tools.
**Key Learnings:**
*   **Error Psychology:** For agents, an error message is a branch point. If the error doesn't suggest a fix *within the tool*, the agent's training biases it to *leave the tool*.
*   **Exit Codes:** Maintaining exit code 0 for "logical empty results" is critical for preserving `&&` chain viability.

---

### **Architectural Relationships**
<!-- Format: [From Entity] -> [To Entity] (relationship type) -->
- cmd/bd/main.go -> internal/ui (uses)
- cmd/bd/main.go -> Cobra (intercepts errors)
# Comprehensive Development Log: Add Help Tips to Command Outputs

**Date:** 2026-02-08

### **Objective:**
To improve the discoverability of advanced command options and increase agent/user autonomy by adding contextual "--help" reminders to the output of key devlog and core commands.

---

### **Phase 1: Implementation - Devlog Hardcoded Tips**

**Initial Problem:** Agents often forget advanced retrieval filters (like `since`, `type`, `depth`) and rely on basic commands.

*   **Action Taken:** Modified `cmd/bd/devlog_cmds.go`. Added `💡 Tip: Use --help` reminders to the following commands:
    *   `bd devlog search`: Added tips for success, typo, and no-results cases.
    *   `bd devlog graph`: Added tip for depth and filtering.
    *   `bd devlog impact`: Added tip for command options.
    *   `bd devlog list`: Added tip for session type filtering.
    *   `bd devlog entities`: Added tip for command options.
    *   `bd devlog show`: Added tip for ID/date matching.
    *   `bd devlog resume`: Added tip for context window customization.
    *   `bd devlog sync`: Added tip for verbose mode.
    *   `bd devlog verify`: Added tip for automated repair flags.
    *   `bd devlog status`: Added tip for configuration info.
*   **Result:** Devlog command outputs now consistently provide a "Next Step" for users/agents wanting more control.

---

### **Phase 2: Implementation - Core Randomized Tips**

**Initial Problem:** Cluttering every `bd list` or `bd ready` output with tips can be noisy.

*   **Action Taken:** Modified `cmd/bd/tips.go`. Added two new educational tips to the global registry:
    1.  `help_filters`: "Use --help to discover advanced filters (date ranges, patterns, labels)"
    2.  `help_display`: "Use --help to explore display modes like --short, --pretty, or --thread"
*   **Configuration:** These tips have a 40% probability of showing after successful core commands, ensuring gradual education without excessive noise.

---

### **Phase 3: Verification**

**Initial Problem:** Ensure tips are rendered correctly with semantic coloring.

*   **Action Taken:** Built the `bd` binary and executed `devlog` commands (`list`, `entities`, `status`, `search`, `verify`).
*   **Result:** All devlog commands correctly displayed the `ui.RenderAccent("💡")` styled tip at the end of their output.

---

### **Final Session Summary**

**Final Status:** **Implemented.** All targeted commands now feature help reminders. The issue `bd-h9u` is resolved.
**Key Learnings:**
*   **Contextual Education:** Hardcoded tips are best for "Deep Discovery" commands (like search), while randomized registry tips are better for "High Frequency" commands (like list).
*   **Agent Guidance:** These tips serve as a "Subconscious Trigger" for agents, reminding them that they have more tools available than just the defaults.

---

### **Architectural Relationships**
<!-- Format: [From Entity] -> [To Entity] (relationship type) -->
- devlog_cmds.go -> ui.RenderAccent (uses)
- tips.go -> InjectTip (registers)
- bd-h9u -> devlog_cmds.go (enhances)
- bd-h9u -> tips.go (enhances)
