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

### **Final Session Summary**

**Final Status:** **Success.** The evaluation harness is fully operational. It successfully executes the OpenCode agent in isolated worktrees, captures tool-call sequences accurately, and identifies "Optimal (Mastery)" strategies in live runs.
**Key Learnings:**
*   **Agent Compliance:** OpenCode CLI is significantly more compliant with local instruction files (`AGENT.md`) than the current Gemini CLI in headless mode.
*   **JSONL Robustness:** When wrapping external AI CLIs, always assume a JSONL stream with interspersed noise rather than a single JSON object.
*   **Intent over Tool Name:** In tool-use traces, the name of the tool (e.g., `bash`) is often less important than the specific arguments (the command intent) for measuring protocol compliance.

---

### **Architectural Relationships**
- bd eval task -> OpenCode CLI (wraps)
- OpenCode CLI -> bd mcp (uses for retrieval)
- aggregateEvent -> OpenCode Events (parses)
- scoreTrace -> Tool Intents (evaluates)
