# The Real Agent Trap: Synergizing Progressive Disclosure with Dynamic Context

## Executive Summary
The "Real Agent Trap" is the evolution of AI-native orchestration. It solves the dual problems of **Agent Amnesia** (forgetting the workflow) and **Token Bloat** (loading too much irrelevant context). By combining a high-friction "Locked" state with an automated "Guided Activation" loop and a "Progressive Disclosure" reference system, we have created a workflow that is both strictly enforced and highly efficient.

---

## 1. The Synergy Flow: From Trap to Flow

The system operates as a state machine that transitions the agent from a restricted observer to a fully empowered collaborator.

### **Phase A: The Bootstrap Trap (Locked State)**
When a project is initialized or a new agent enters, the primary instruction file (e.g., `GEMINI.md`) contains exactly one line:
`BEFORE ANYTHING ELSE: run 'bd onboard' and follow ALL instructions.`

This is the "snapping shut" of the trap. The agent cannot proceed because it has no technical context, no tech stack details, and no working protocols.

### **Phase B: Guided Activation (The Instruction Loop)**
The agent runs `bd onboard`, which doesn't just print text but acts as a CLI tutor:
1.  **Verify**: `bd devlog verify --fix` (Ensures the memory system is healthy).
2.  **Sync**: `bd sync` (Ingests the latest project state).
3.  **Unlock**: `bd ready` (The final trigger).

### **Phase C: Automatic Context Injection**
Executing `bd ready` triggers the `finalizeOnboarding` logic. This function:
1.  Updates the database state to `onboarding_finalized = true`.
2.  Rewrites the agent's instruction file, replacing the "Trap" line with the **Progressive Disclosure Bootloader**.

### **Phase D: Dynamic Prime (The Source of Truth)**
The `bd prime` command is the system's "Pre-frontal Cortex." It is registered as a hook in the agent's environment (e.g., `SessionStart` and `PreCompact`).
- **If Uninitialized**: `bd prime` outputs the `RestrictedBootloader`, re-triggering the trap if the agent somehow bypassed it.
- **If Finalized**: `bd prime` outputs the `FullBootloader` + the current `WORKING_PROTOCOL.md`.

---

## 2. Core Concepts

### **Progressive Disclosure**
Instead of a 3,000-token monolithic instruction file, the agent is given a "Hub" (the Bootloader) and "Spokes" (on-demand modules). 
- **Hub**: Core commands and the daily loop.
- **Spokes**: `BEADS_REFERENCE.md`, `DEVLOG_REFERENCE.md`, `PROJECT_CONTEXT.md`.
The agent only loads a "Spoke" when it encounters a specific problem (e.g., "I don't know how to split a task"). This reduces active context by **70-80%**, making the agent faster and more focused.

### **Adaptive Session Protocols**
The "End of Session" protocol is no longer static. `bd prime` detects the environment:
- **Ephemeral Branch**: Tells the agent to merge locally.
- **No Remote**: Tells the agent to flush locally only.
- **Daemon Active**: Simplifies commands because background sync is handled.
This prevents the agent from attempting impossible git operations or leaving work unpushed.

### **State-Awareness via readOnlyCommands**
By making `bd prime` a `readOnlyCommand` (removing it from `noDbCommands`), we allow the CLI to be context-aware without the overhead of write-locks. The command can read the database to check if the project is "Finalized" and adapt its instructions accordingly.

---

## 3. Why it Works: The "AI-Native" UX

1.  **Zero-Day Discovery**: A new agent joins and instantly knows the *process* without needing to scan the *entire codebase*.
2.  **Self-Correction**: If the agent forgets the workflow mid-session, the `PreCompact` hook re-injects the rules, "priming" the memory.
3.  **High-Fidelity Tracking**: By forcing the `bd ready` -> `bd close` -> `bd sync` loop, we ensure that the knowledge graph (Beads + Devlog) is always a perfect reflection of the repository state.

---

*Analysis by BeadsLog AI Agent, January 2026*
