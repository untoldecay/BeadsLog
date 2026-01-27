# Working Protocol: Task Loop

⚠️ **Load for every task during active work**

## 🔄 The Loop

### Step 1: Map It (BeadsLog First)
Before reading code or making a plan, you MUST use the graph:
```bash
bd devlog graph "ComponentName"  # Visualize dependencies
bd devlog impact "ComponentName" # Verify what depends on this
bd devlog search "error/feature" # Find how this was solved before
```

### Step 2: Verify It (Code Reading)
Read the actual code files identified in Step 1 to confirm architectural assumptions.

### Step 3: Implement & Crystallize
```bash
# Code change...
git add .
git commit -m "fix: message" # Auto-generates devlog
bd update <id> --status closed
```

## ✅ End Session
```bash
bd status          # Final health check
git push           # Share crystallized knowledge
```
