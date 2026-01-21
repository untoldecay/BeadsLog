package main

const ProtocolMdTemplate = `# Protocol: First Execution

⚠️ **Load ONLY on first message per session**

## 1. Beads Starting Workflow
[codeblock=bash]
bd sync          # Get latest issues
bd status        # Health check
bd ready         # Find prioritized work
[/codeblock]

## 2. Devlog Starting Workflow
[codeblock=bash]
bd devlog verify --fix # Health check (Fix if needed)
bd devlog sync         # Get latest team knowledge
bd devlog resume --last 1  # Load your last session
bd devlog status       # Verify database state
[/codeblock]

## 3. Pick Task
- Choose from ` + "`bd ready`" + `
- ` + "`bd update <id>`" + ` to claim
- Check: ` + "`bd devlog search \"<issue keywords>\"`" + `

## ✅ Now Ready
Load WORKING_PROTOCOL.md for task loop.
`
const WorkingProtocolMdTemplate = `# Working Protocol

⚠️ **Load for every task during active work**

## 🔄 The Loop (Repeat)

### Before Coding
[codeblock=bash]
bd devlog graph "ComponentName"  # Dependencies
bd devlog impact "ComponentName" # What breaks if changed?
bd devlog search "error/feature" # Past solutions?
[/codeblock]

### Code + Commit (Auto-Devlog)
[codeblock=bash]
git add .
git commit -m "fix: descriptive message"
[/codeblock]
*Pre-commit automatically generates devlog*

### Update Issue
[codeblock=bash]
bd update <id> --status "in-progress" | closed
[/codeblock]

## 🆘 Common Scenarios
**Split work?** ` + "`bd split <id> \"sub-task\"`" + `
**Blocked?** ` + "`bd block <current> <blocking>`" + `
**New bug?** ` + "`bd new \"Bug title\" --priority high`" + `

## ✅ End Session
[codeblock=bash]
bd status          # Verify sync
git push           # Share with team
[/codeblock]

## 🔍 Still Need Help?
bd --help | bd devlog --help → Load *_REFERENCE.md
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

