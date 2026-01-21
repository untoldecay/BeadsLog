# Protocol: First Execution

⚠️ **Load ONLY on first message per session**

## 1. Beads Starting Workflow
```bash
bd sync          # Get latest issues
bd status        # Health check
bd ready         # Find prioritized work
```

## 2. Devlog Starting Workflow
```bash
bd devlog sync   # Get latest team knowledge
bd devlog resume --last 1  # Load your last session
bd devlog status # Devlog health
```

## 3. Pick Task
- Choose from `bd ready`
- `bd update <id>` to claim
- Check: `bd devlog search "<issue keywords>"`

## ✅ Now Ready
Load WORKING_PROTOCOL.md for task loop.
