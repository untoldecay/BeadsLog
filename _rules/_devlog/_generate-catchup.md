# Prompt: High-Signal Activity Catchup Summary

## Objective
Analyze the raw activity feed provided by `bd catchup` and generate a concise, human-readable technical summary that aligns the agent and the user on the project's current state.

## Persona
A meticulous Lead Software Engineer ensuring that architectural shifts and blocked paths are clearly communicated to prevent duplicate work or context loss.

---

## Context Analysis
The input for this prompt is the raw output from `bd catchup`. 
Focus your summary on:
1.  **Activity by Team Member**: Group technical progress by the person/agent who performed it.
2.  **Landed Work**: What features or fixes were actually merged and validated.
3.  **Directional Shifts**: Any new epics, major task splits, or modularization attempts.
4.  **The Grave (Tombstones)**: Abandoned paths and the reasons WHY they were dropped.
5.  **Paused Context**: Parked branches and what is blocking them.

## Summary Requirements
The summary MUST follow this structured format:

### 🔄 Project Evolution Since [Date]

[One paragraph explaining the major themes of the recent work and any significant architectural drift.]

#### 👥 Activity by Team Member
- **[Author Name]**:
  - [High-level summary of their primary focus]
  - [Bullet points of specific landed or paused work]

#### ✨ Landed Features & Enhancements
- **[Component/Feature]**: [Concise technical summary of what was completed]
- **[Landed via [SessionID]]**: [Brief impact on the codebase]

#### ⚠️ Blocked & Abandoned Paths (Critical)
- **[Abandoned Scope]**: [The 'Why' - explain the specific failure or reason for rejection]
- **[Paused Scope]**: [What is blocking this work? Link to reasoning session if available]

#### 🎯 Suggested Next Steps
- [Based on `bd ready`, what are the highest priority items to address now?]

## Success Criteria
- A human reading this summary should know exactly what they missed without opening individual devlog files.
- The summary must have a high signal-to-noise ratio: prioritize "Why" over "What".
- Blocked paths must be highly visible to prevent the agent from re-proposing failed approaches.
