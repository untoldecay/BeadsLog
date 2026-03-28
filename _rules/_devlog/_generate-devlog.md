# Prompt: Generate Chronological Debugging & Development Log (Manual Mode)

## Objective:
Analyze the session history and generate a comprehensive development log. Document assumptions, actions, outcomes, and corrections.

## ⚠️ MANDATORY: Architectural Relationships
Since Background AI Enrichment is DISABLED, manually append relationships at the bottom:
```markdown
### Architectural Relationships
- EntityA -> EntityB (uses)
```

## persona:
Meticulous technical writer documenting the learning journey.

## File Handling:
1.  **Check for Existing Log:** Find today's file in `_rules/_devlog/`.
2.  **Update or Create:** Append to today's log or create `_rules/_devlog/[YYYY-MM-DD]_[title].md`.
3.  **Record Session:** After saving the file, you MUST run:
    *   `bd devlog record --subject "[prefix] description" --problem "description" --file "_rules/_devlog/[filename].md"`

## Output Structure:
---
# Comprehensive Development Log: [Goal]
**Date:** [YYYY-MM-DD]
...
