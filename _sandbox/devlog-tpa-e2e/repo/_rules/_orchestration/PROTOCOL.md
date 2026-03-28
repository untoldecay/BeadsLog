# Protocol: Detailed Reference

> **NOTE:** The active protocol is now injected directly into your agent instructions (AGENT.md).
> This file serves as a detailed reference for the "Starting Workflow" if needed.

## 1. 🟢 Initialize Memory (Quantified Mapping)
```bash
bd onboard       # Refresh your instructions
bd sync          # Get latest issues
bd devlog sync   # Ingest latest team knowledge
bd devlog verify --fix # Check graph integrity
```

## 2. 🔍 Map the Landscape (Mandatory)
Before using `ls`, `grep`, or `glob`, you MUST query the architectural graph:
- **Entities:** `bd devlog entities` (Identify key components)
- **Relationships:** `bd devlog graph "Subject"` (See dependencies)
- **History:** `bd devlog search "Keywords"` (Find past solutions)

## 3. 🎯 Select and Claim Task
- List ready work: `bd ready`
- Claim task: `bd update <id> --status in_progress`
- Resume context: `bd devlog resume --last 1`
