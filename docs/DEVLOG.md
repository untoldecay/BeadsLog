# Devlog: The Architectural Memory

In BeadsLog, every work session is a **Bead**. The collection of these beads forms your project's **Devlog**—a living wiki of your system's evolution.

## What is a Devlog?
Unlike a Git commit (which focuses on *what* code changed), a Devlog focus on the **Scout's Narrative**:
- The initial problem or goal.
- The assumptions made (and which ones were wrong).
- The "Why" behind architectural decisions.
- The semantic relationships discovered during the build.
- **Attribution:** Tracks both the **Human Author** (the owner) and the **AI Agent** (the actor).
- **Precision:** High-precision time logging (`HH:MM`) for exact sequencing of work.

## Structured Example
A typical devlog file resides in `_rules/_devlog/` and looks like this:

```markdown
# Session: OAuth Integration Fix

**Date:** 2026-01-27 14:30
**Author:** untoldecay
**Agent:** Gemini CLI

### Objective:
...
```

## The Workflow Stand
The Devlog is the **entry and exit point** of every task:
1. **Onboard:** The agent reads past devlogs to gain context.
2. **Execute:** The agent builds the feature.
3. **Record:** The agent uses `bd devlog record` to update the index automatically.
4. **Crystallize:** Discovered relationships are saved to the narrative.
5. **Sync:** The system ingests the devlog and updates the architectural graph.

## Crystallization
This is the "magic" of BeadsLog. If an agent forgets to add the `Architectural Relationships` block, the **Background AI worker** will read the narrative and append it for them. This turns temporary thoughts into permanent, version-controlled wiki data.

## Entity Naming Rules
To keep the graph high-signal, extraction filters out noise: names shorter than 3 characters, pure numbers, phrase fragments, bare common words, and trunk branch names (`master`, `main`, `develop`). This applies to **all** sources — including manually written arrows. If an arrow targets a filtered word, the edge is dropped with a stderr warning; use a qualified name instead:

```markdown
### Architectural Relationships
- AuthService -> master (deploys)            <- dropped with warning
- AuthService -> master-branch (deploys)     <- kept
```
