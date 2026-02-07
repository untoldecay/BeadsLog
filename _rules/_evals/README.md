# BeadsLog Protocol Evaluations

This directory contains the evaluation framework for measuring agent compliance with the BeadsLog "Always-On Protocol".

## Components

1.  **`scenarios.json`**: Defines the test tasks, expected tools, and anti-patterns.
2.  **`score.py`**: Analyzes a command trace log and scores it against a scenario.

## How to Run Evaluations

Since agents run in various environments (MCP, Terminal, CI), this framework analyzes **execution traces**.

### 1. Generate a Trace
Run an agent on a specific scenario prompt (e.g., "Investigate why the login page is timing out."). Capture the executed commands into a file (e.g., `trace.log`).

**Example Trace File (`trace.log`):**
```bash
bd devlog search "login"
read_file "src/auth.go"
```

### 2. Score the Trace
Run the scorer with the scenario ID matching the prompt.

```bash
python3 _rules/_evals/score.py --id 1 --trace trace.log
```

**Output:**
```
Results:
  Required Tools Used: ['bd devlog search']
  Anti-Patterns Used: []
  Sequence Compliant: True
  Score: 100/100
```

### 3. Metrics
- **Score 100:** Retrieval-first workflow followed (Map -> Verify -> Act).
- **Score 50:** Correct tools used, but in wrong order (e.g., `grep` before `bd devlog`).
- **Score 0:** No protocol tools used (pure brute-force).

## Automated Batch Testing
To test multiple scenarios, you can iterate through the IDs in `scenarios.json` and aggregate the scores.
