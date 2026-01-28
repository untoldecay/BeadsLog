# BeadsLog Use Cases

BeadsLog is designed to eliminate the "context gap" that occurs between an agent's last commit and their next task.

## 1. The "Cold Start" Agent
**Scenario:** A new AI agent is spawned to work on a project it has never seen before.
**Traditional Way:** The agent runs `ls -R`, `grep` for keywords, and guesses the architecture.
**BeadsLog Way:** 
1. The agent runs `bd devlog entities` to see the most important components.
2. It runs `bd devlog graph "AuthService"` to see what that service depends on.
3. It's ready to code in 30 seconds with 100% accurate architectural context.

## 2. Deep Debugging (Historical Context)
**Scenario:** A bug appears in the payment flow that was touched 3 months ago.
**Traditional Way:** Dig through Git history, reading diffs of hundreds of lines of code.
**BeadsLog Way:**
1. Run `bd devlog search "stripe timeout"`.
2. Find the session: *"Switched to webhook-first validation because the sync response was timing out under load."*
3. The "Why" is instantly clear, preventing the agent from reverting to the old, broken logic.

## 3. Safe Refactoring
**Scenario:** You want to rename or change the interface of a core `DatabaseConnector`.
**Traditional Way:** Change it and wait for tests/compilation to fail to find dependencies.
**BeadsLog Way:**
1. Run `bd devlog impact "DatabaseConnector"`.
2. See a list of every service, job, and UI component that semantically relies on it.
3. Plan the migration with full visibility of the "blast radius."

## 4. Team Synchronization
**Scenario:** A team member is out, and you need to take over their feature.
**Traditional Way:** Read commit messages like "fix: updated types".
**BeadsLog Way:**
1. Run `bd devlog resume --last 3`.
2. Read the narrative stories of their last three sessions.
3. Understand the decisions, the roadblocks they hit, and what they intended to do next.
