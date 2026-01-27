package main

const ProtocolMdTemplate = `# Protocol: Session Activation (First Message)

⚠️ **STOP: You MUST execute these steps in order before any manual file searching**

## 1. 🟢 Initialize Memory (Quantified Mapping)
[codeblock=bash]
bd onboard       # Refresh your instructions
bd sync          # Get latest issues
bd devlog sync   # Ingest latest team knowledge
bd devlog verify --fix # Check graph integrity
[/codeblock]

## 2. 🔍 Map the Landscape (Mandatory)
Before using ` + "`ls`" + `, ` + "`grep`" + `, or ` + "`glob`" + `, you MUST query the architectural graph:
- **Entities:** ` + "`bd devlog entities`" + ` (Identify key components)
- **Relationships:** ` + "`bd devlog graph \"Subject\"`" + ` (See dependencies)
- **History:** ` + "`bd devlog search \"Keywords\"`" + ` (Find past solutions)

## 3. 🎯 Select and Claim Task
- List ready work: ` + "`bd ready`" + `
- Claim task: ` + "`bd update <id> --status in_progress`" + `
- Resume context: ` + "`bd devlog resume --last 1`" + `

## ✅ Activation Complete
Load ` + "`WORKING_PROTOCOL.md`" + ` to begin the development loop.
`
const WorkingProtocolMdTemplate = `# Working Protocol: Task Loop

⚠️ **Load for every task during active work**

## 🔄 The Loop

### Step 1: Map It (BeadsLog First)
Before reading code or making a plan, you MUST use the graph:
[codeblock=bash]
bd devlog graph "ComponentName"  # Visualize dependencies
bd devlog impact "ComponentName" # Verify what depends on this
bd devlog search "error/feature" # Find how this was solved before
[/codeblock]

### Step 2: Verify It (Code Reading)
Read the actual code files identified in Step 1 to confirm architectural assumptions.

### Step 3: Implement & Crystallize
[codeblock=bash]
# Code change...
git add .
git commit -m "fix: message" # Auto-generates devlog
bd update <id> --status closed
[/codeblock]

## ✅ End Session
[codeblock=bash]
bd status          # Final health check
git push           # Share crystallized knowledge
[/codeblock]
`
const BeadsReferenceMdTemplate = `# Beads Commands

⚠️ **Load ONLY when bd --help insufficient**

## Issue Lifecycle
[codeblock=bash]
bd new "Title" --type bug --priority high
bd ready                           # P0 issues
bd update <id> --status in-progress
bd update <id> --assign @teammate
bd close <id>
bd split <id> "Sub-task"
bd block <current> <blocking>
[/codeblock]

## Query
[codeblock=bash]
bd list --status open
bd show <id>
bd search "keywords"
[/codeblock]

## Sync & Versioning
[codeblock=bash]
bd sync               # Sync local DB <-> JSONL <-> Git
bd deploy staging     # Deploy to staging environment
bd deploy production  # Deploy to production environment
[/codeblock]
`
const DevlogReferenceMdTemplate = `# Devlog Commands

⚠️ **Load ONLY when bd devlog --help insufficient**

## Search
[codeblock=bash]
bd devlog search "nginx timeout"
bd devlog search "modal" --type fix
bd devlog search "auth" --since 2026-01
bd devlog list --last 5
[/codeblock]

## Architecture
[codeblock=bash]
bd devlog graph "nginx"
bd devlog impact "AuthService"
bd devlog status
bd devlog verify --fix
[/codeblock]

## Maintenance
[codeblock=bash]
bd devlog sync        # Ingest new markdown files
bd devlog verify      # Check for missing metadata
bd devlog reset       # Clear local cache (rare)
[/codeblock]
`
const ProjectContextMdTemplate = `# Project Context

⚠️ **Load for tech stack, conventions, and architectural overview**

## Overview
This file contains project-specific knowledge.
It is populated by migrating legacy agent instructions or by the user.

## Tech Stack
- Frontend: [React/Vue/Svelte]
- Backend: [Go/Node/Python]
- Database: [SQLite/Postgres]

## Conventions
- Commits: Conventional Commits
- Branching: Feature Branch Workflow
`

// RestrictedBootloader is the initial trap that forces the agent to run the protocol
const RestrictedBootloader = `# BeadsLog Agent Protocol (SETUP IN PROGRESS)

## 🚧 Mandatory Initialization
You have successfully modularized the project rules.
However, your session is not yet initialized and project context is locked.

**You MUST execute the starting workflow to proceed.**

1. Read: _rules/_orchestration/PROTOCOL.md
2. Execute EVERY command in the "Starting Workflow" section.

## 📋 Required Sequence
[codeblock=bash]
bd sync
bd devlog verify --fix
bd devlog sync
bd ready  <-- RUN THIS TO UNLOCK CONTEXT
[/codeblock]

## 🔍 Discover More
bd --help                 # All bd commands
bd devlog --help          # All devlog commands
`

// FullBootloader is the final protocol injected once onboarding is complete
const FullBootloader = `# BeadsLog Agent Protocol

## 🚀 Session Start? (First Message Only)
IF this is your first message, you MUST execute these commands in order:
[codeblock=bash]
bd sync          # Get latest issues
bd devlog sync   # Ingest latest knowledge
bd devlog verify --fix # Ensure graph integrity
bd devlog resume --last 1 # Load context
[/codeblock]

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
`
