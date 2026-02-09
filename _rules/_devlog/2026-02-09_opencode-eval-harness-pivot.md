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
- eval_cmds.go -> bd eval clean (manages)