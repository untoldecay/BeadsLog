# Devlog: The Architectural Memory

In BeadsLog, every work session is a **Bead**. The collection of these beads forms your project's **Devlog**—a living wiki of your system's evolution.

## What is a Devlog?
Unlike a Git commit (which focuses on *what* code changed), a Devlog focus on the **Scout's Narrative**:
- The initial problem or goal.
- The assumptions made (and which ones were wrong).
- The "Why" behind architectural decisions.
- The semantic relationships discovered during the build.

## Structured Example
A typical devlog file resides in `_rules/_devlog/` and looks like this:

```markdown
# Session: OAuth Integration Fix

**Date:** 2026-01-27

### Objective:
Resolve the redirect loop in the production environment.

### Phase 1: Debugging
Initial Assumption: The cookie domain was misconfigured.
Action: Checked nginx.conf and env vars.
Result: Domain was correct. The real issue was proxy_buffering truncating the header.

### Final Summary
Fixed by disabling proxy_buffering for the /auth endpoint.

### Architectural Relationships
- nginx -> AuthAPI (proxies_to)
- AuthAPI -> Redis (uses)
```

## The Workflow Stand
The Devlog is the **entry and exit point** of every task:
1. **Onboard:** The agent reads past devlogs to gain context.
2. **Execute:** The agent builds the feature.
3. **Crystallize:** At the end of the session, the agent generates a new Devlog.
4. **Sync:** The system ingests the devlog and updates the architectural graph.

## Crystallization
This is the "magic" of BeadsLog. If an agent forgets to add the `Architectural Relationships` block, the **Background AI worker** will read the narrative and append it for them. This turns temporary thoughts into permanent, version-controlled wiki data.
