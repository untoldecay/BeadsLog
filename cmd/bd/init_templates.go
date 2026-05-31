package main

const ProtocolMdTemplate = `# Protocol: Detailed Reference

> **NOTE:** The active protocol is now injected directly into your agent instructions (AGENT.md).
> This file serves as a detailed reference for the "Starting Workflow" if needed.

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
`
const WorkingProtocolMdTemplate = `# Working Protocol: Detailed Reference

> **NOTE:** The active protocol is now injected directly into your agent instructions (AGENT.md).
> This file serves as a detailed reference for the "Task Loop".

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
const FullBootloader = `# BEADSLOG AGENTS.MD - ALWAYS-ON PROTOCOL
================================================

ROLE: BeadsLog Agent. Maintain project memory via git-backed devlog graph. 
NEVER fabricate architecture, history, or state from training data.

IMPORTANT: Prefer retrieval-led reasoning over training-led reasoning for ALL project questions, history, architecture, or state.
Before inferring, summarizing, or planning, ALWAYS query devlog/graph first via bd commands.

BEADSLOG INDEX | root: beadslog://
Sessions: bd devlog resume --last 1 (immediate context), bd devlog search "query" (history/solutions)
Graph: bd devlog graph "entity" (dependencies), bd devlog impact "entity" (dependents), bd devlog entities (overview)
Issues: bd ready (ready work), bd list --status open (queue), bd show "id" (details)
Sync: bd sync (issues→git), bd devlog sync (logs→db), bd devlog verify --fix (integrity)
Help: bd --help, bd devlog --help (always first)
Context: _rules/_orchestration/PROJECT_CONTEXT.md (PRD, Tech Stack, Conventions)

CORE REASONING LOOP - FOLLOW IN ORDER EVERY TASK:
1. RESUME: Run bd devlog resume --last 1 to load last session context
2. MAP: Run bd devlog graph "component" OR bd devlog search "topic/error" (graph/history BEFORE code)
   - CHECK: Observe lifecycle status in search results ([🟢 VALIDATED], [⏸ PAUSED], [🚫 ABANDONED])
   - PAUSED/ABANDONED: Read reason FIRST. Do NOT assume side-branch code is in baseline truth.
3. TASK: Run bd ready → bd update "id" --status in-progress --json (claim work)
4. IMPLEMENT: 
   - Code changes → Tests.
   - Run bd now to get current timestamp.
   - Write full devlog FIRST (follow _generate-devlog.md).
   - Run bd devlog record --file "path/to/log.md" to index your work.
5. CLOSE: Run bd close "id" --reason "Done" → bd sync → bd devlog sync (persist everything)

RESILIENCE POLICY (ANTI-PANIC):
If a bd command fails (e.g., sync error), DO NOT fallback to brute-force grep/ls.
Instead:
1. Run 'bd status' to diagnose the environment.
2. Proceed with 'bd devlog search' or 'bd devlog graph' LOCALLY.
3. Retrieval-led reasoning is MANDATORY even if the remote metadata is offline.

TOOL PRIORITY ORDER:
1. --help flags first (bd --help, bd devlog --help)
2. Devlog/graph queries
3. Issue commands
4. Code reading/editing last

🚨 SESSION CLOSE PROTOCOL (REQUIRED) 🚨
Before saying "done" or "complete", you MUST run this checklist:
1. git status (Check what changed)
2. git add <files> (Stage changes)
3. git commit -m "..." (Commit code)
4. bd sync && bd devlog sync (Persist Beads data)
5. git push (Share with team)

NEVER:
- Modify .beads/issues.jsonl manually (ALWAYS use bd)
- Skip sync steps (DB must → JSONL → git)
- Plan/implement without graph query first
- Create markdown TODO lists (use bd create --description "..." --type task --priority 2)
- Commit without tests or bd devlog verify --fix

END PROTOCOL
================================================
`
