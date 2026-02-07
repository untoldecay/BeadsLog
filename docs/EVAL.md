# Agent Evaluation Mode

**Measure agent performance, protocol compliance, and token efficiency.**

BeadsLog features a specialized **Evaluation Mode** designed to analyze how effectively AI agents use the knowledge graph versus brute-force searching. This provides quantitative proof of the benefits of the "Always-On Protocol."

## 🚀 How it Works

Eval Mode attaches a unique **Session ID** to every command executed by an agent. This allows the system to segment continuous logs into distinct "test runs" for side-by-side comparison.

### 🔄 The Eval Lifecycle

1.  **Start:** `bd eval start [scenario_id]`
    - Activates high-fidelity tracing.
    - Generates a unique Session ID.
2.  **Iterate:** Run tasks with your agent.
3.  **Rotate:** `bd eval next`
    - Ends the current run and starts a new one (rotates the Session ID).
    - Perfect for A/B testing different models or prompts.
4.  **Report:** `bd eval report [N]`
    - Generates a beautiful, Lipgloss-based comparative matrix of the last N runs.
5.  **Stop:** `bd eval stop`
    - Disables tracing and cleans up session state.

## 📋 Methodology & Scoring

The reporting engine automatically classifies each session into one of three strategic categories:

| Category | Methodology | Primary Focus | Best Used For |
| :--- | :--- | :--- | :--- |
| **Retrieval-First** | **Metadata Only** (bd search, graph) | The "Why" & "Where" (Intent, Deps) | 🟢 Onboarding, Architecture Review |
| **Brute-Force** | **Code Only** (grep, ls, cat) | The "What" & "How" (Logic, Syntax) | 🟢 Refactoring, Debugging lines |
| **Hybrid** | **Indexed Verification** (bd Map $	o$ read) | **Synthesis** (Intent + Reality) | 👑 Master Reports, Complex Fixes |

### 💰 Token Impact Calculation

The system assigns a "Token Weight" to every tool used:
- **Brute-Force Tools (`grep`, `ls -R`):** High cost (~2,500 tokens) due to context inflation.
- **Retrieval Tools (`bd devlog`):** Low cost (~150 tokens) due to precise metadata.
- **Savings:** Calculated by comparing the total run cost against a "Vanilla Baseline" (~4,500 tokens).

## 📊 Example Report

When you run `bd eval report 3`, you get a comparative matrix:

```text
┌──────────────────┬─────────────────────────────────────┬────────────────────────────────────────┬─────────────────────────────────────┐
│ Attribute        │ Run 1                               │ Run 2                                  │ Run 3                               │
├──────────────────┼─────────────────────────────────────┼────────────────────────────────────────┼─────────────────────────────────────┤
│   Methodology    │            Metadata Only            │            Brute Force Code            │        Indexed Verification         │
│ Tools            │ bd devlog graph                     │ bd search                              │ bd search                           │
│                  │ bd devlog search                    │ bd list                                │ bd devlog graph                     │
│ Token Cost       │ 300                                 │ 600                                    │ 450                                 │
│ Primary Focus    │ The "Why" & "Where"                 │ The "What" & "How"                     │ Synthesis                           │
│ Efficiency Ratio │ High Context / Low Detail           │ High Detail / High Cost                │ Maximum Efficiency                  │
│ Status           │ PASS (100/100)                      │ FAIL (0/100)                           │ PASS (85/100)                       │
└──────────────────┴─────────────────────────────────────┴────────────────────────────────────────┴─────────────────────────────────────┘
```

## 🛠️ Commands Reference

| Command | Usage |
| :--- | :--- |
| `bd eval start` | Enable eval mode and start first session |
| `bd eval next` | Rotate to a new session ID for comparison |
| `bd eval report [N]` | Generate comparative table for last N sessions |
| `bd eval stop` | Disable eval mode |

## 🔒 Privacy & Security

Evaluation logs are stored in `.beads/interactions.jsonl`.
- **Redaction:** Sensitive flags (API keys, tokens) are automatically redacted as `[REDACTED]` before being written to disk.
- **Cleanup:** `bd eval stop` ensures eval-specific metadata is cleared.
