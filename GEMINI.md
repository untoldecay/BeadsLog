BEFORE ANYTHING ELSE: run 'bd onboard'

<!-- BD_PROTOCOL_START -->
# BeadsLog Agent Protocol

## 🚀 Session Start? (First Message Only)
IF this is your first message:
1. Read: _rules/_orchestration/PROTOCOL.md  
2. Execute: Beads + Devlog starting workflows
3. Proceed to regular workflow

## 🔄 Regular Workflow (Every Task)
Read: _rules/_orchestration/WORKING_PROTOCOL.md

## 📋 Core Commands (Always Available)

### Beads (Issues)
bd ready                 # Find next task
bd update <id>           # Mark in progress  
bd close <id>            # Complete task
bd sync                  # Sync issues (auto via git hook)

### Devlog (Memory)
bd devlog resume --last 1    # Load last session
bd devlog search "query"     # Find past solutions
bd devlog graph "entity"     # See dependencies
bd devlog impact "entity"    # What depends on this?

### Commit (Auto-Devlog)
git commit -m "fix: message" # Generates devlog automatically

## 🔍 Discover More
bd --help                 # All bd commands
bd devlog --help          # All devlog commands

## 📚 On-Demand Files (Load Only When Needed)
| File | When to Load |
|------|-------------|
| **PROTOCOL.md** | First execution only |
| **WORKING_PROTOCOL.md** | Every task |
| **BEADS_REFERENCE.md** | bd --help insufficient |
| **DEVLOG_REFERENCE.md** | bd devlog --help insufficient |
| **PROJECT_CONTEXT.md** | Need project overview/architecture |

## ⚠️ Loading Rules
1. Always try --help first
2. Load PROTOCOL.md only once per session
3. Load WORKING_PROTOCOL.md at task start
4. Reference files only when commands fail
<!-- BD_PROTOCOL_END -->