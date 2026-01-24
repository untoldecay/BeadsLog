# Working Protocol

⚠️ **Load for every task during active work**

## 🔄 The Loop (Repeat)

### Before Coding
```bash
bd devlog graph "ComponentName"  # Dependencies
bd devlog impact "ComponentName" # What breaks if changed?
bd devlog search "error/feature" # Past solutions?
```

### Code + Commit (Auto-Devlog)
```bash
git add .
git commit -m "fix: descriptive message"
```
*Pre-commit automatically generates devlog*

### Update Issue
```bash
bd update <id> --status "in-progress" | closed
```

## 🆘 Common Scenarios
**Split work?** `bd split <id> "sub-task"`
**Blocked?** `bd block <current> <blocking>`
**New bug?** `bd new "Bug title" --priority high`

## ✅ End Session
```bash
bd status          # Verify sync
git push           # Share with team
```

## 🔍 Still Need Help?
bd --help | bd devlog --help → Load *_REFERENCE.md
