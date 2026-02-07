# Agent Evaluation Mode

**Measure agent performance, protocol compliance, and token efficiency.**

BeadsLog features a specialized **Evaluation Mode** designed to analyze how effectively AI agents use the knowledge graph versus brute-force searching. This provides quantitative proof of the benefits of the "Always-On Protocol."

## 🚀 How it Works

Eval Mode attaches a unique **Session ID** to every command executed by an agent. This allows the system to segment continuous logs into distinct "test runs" for side-by-side comparison.

### 🔄 The Eval Lifecycle

1.  **Start:** `bd eval start [scenario_id]`
    - Activates high-fidelity tracing.
    - Sets the "Intent" or "Scenario" for grouping.
2.  **Iterate:** Run tasks with your agent.
3.  **Rotate:** `bd eval next`
    - Ends the current run and starts a new one (rotates the Session ID).
    - Perfect for A/B testing different models or prompts for the same task.
4.  **Report:** `bd eval report [N]`
    - Generates a comparative matrix of the last N sessions, grouped by task similarity.
5.  **Stop:** `bd eval stop`
    - Disables tracing and cleans up session state.

## 📋 Methodology & Scoring

The reporting engine groups sessions by **Task Intent** using fuzzy matching (Levenshtein distance) on prompts. It then calculates a **Weighted Efficiency Score** based on the following rubric:

### 1. Protocol Adherence Rubric

| State | Logic | Emoji | weight |
| :--- | :--- | :--- | :--- |
| **Optimal (Mastery)** | `Map` tools used FIRST + `Verify` tools used LATER | 🟢🟢🟢 | **1.5x** |
| **Shallow** | `Map` tools used + NO `Verify` tools (hallucination risk) | 🟠 | **0.8x** |
| **Disordered** | `Verify` tools used BEFORE `Map` tools (sequence break) | 🔴🟠 | **0.5x** |
| **Blind** | NO `Map` tools used (pure brute-force) | 🔴 | **0.1x** |

### 2. Efficiency Formula

Efficiency is calculated relative to the average token cost of all runs in the same task group:

`Efficiency = (Strategy Weight) * (Group Average Tokens / This Run Tokens)`

- **> 1.0x 🚀:** More efficient than the group average.
- **< 1.0x 📉:** Less efficient (wasted tokens or poor strategy).

### 💰 Token Weighting

To estimate costs accurately, tools are assigned weights based on their context consumption:
- **Brute-Force (`grep`, `search_file_content`):** ~2,500 tokens (Heavy context).
- **Inspection (`read_file`):** ~1,200 tokens (Medium context).
- **Retrieval (`bd devlog search/graph`):** ~150 tokens (Light context).

## 📊 Comparative Report Example

When you run `bd eval report 3`, you get a grouped matrix:

```text
TASK GROUP: "Authentication Logic Fix" (Matched 2 runs)
╭───────────────────┬───────────────────┬───────────────────┬──────────────────┬──────────────────┬──────────────────┬──────────────────╮
│ Run               │ Strategy          │ Tools Used        │ Tokens           │ Time             │ Status           │ Eff.             │
├───────────────────┼───────────────────┼───────────────────┼──────────────────┼──────────────────┼──────────────────┼──────────────────┤
│         1         │  Optimal (Mastery)│  bd devlog graph  │       1350       │        12s       │       PASS       │     2.5x 🟢🟢🟢  │
│ (770496985-61285) │                   │     read_file     │                  │                  │    (100/100)     │                  │
├───────────────────┼───────────────────┼───────────────────┼──────────────────┼──────────────────┼──────────────────┼──────────────────┤
│ 2                 │ Blind             │ bd search         │ 15400            │ 45s              │ FAIL             │ 0.1x 🔴          │
│ (770497002-62060) │                   │                   │                  │                  │ (0/100)          │                  │
╰───────────────────┴───────────────────┴───────────────────┴──────────────────┴──────────────────┴──────────────────┴──────────────────╯
```

## 🛠️ Commands Reference

| Command | Usage |
| :--- | :--- |
| `bd eval start [name]` | Enable eval mode and tag the current scenario |
| `bd eval next [name]` | Start a fresh run for the same (or new) scenario |
| `bd eval report [N]` | Generate comparative table for last N sessions |
| `bd eval stop` | Disable eval mode |

## 🔒 Privacy & Security

Evaluation logs are stored in `.beads/interactions.jsonl`.
- **Redaction:** Sensitive flags (API keys, tokens) are automatically redacted as `[REDACTED]` before being written to disk.
- **Cleanup:** `bd eval stop` ensures eval-specific metadata is cleared.