# Prompt: Generate Chronological Debugging & Development Log

## Objective:
Analyze the entire conversation history of the current session and generate a comprehensive, chronological development log. The primary purpose is to be a transparent record of the entire problem-solving process, detailing every assumption (especially flawed ones), every action taken, the resulting outcomes, and the evidence-based corrections that led to the final solution.

## Persona:
Act as a meticulous technical writer and project manager, documenting the development journey with a focus on learning from mistakes.

## Input:
The full conversation history of the current development session.

## File Handling Logic:
1.  **Check for Existing Log:** Before generating, list the files in the `_rules/_devlog/` directory.
2.  **Identify Today's Log:** Find the most recent file. Check if its filename matches today's date (e.g., `2025-07-04_session_summary.md` or `2025-07-04_specific-title.md`).
3.  **Update or Create:**
    *   **If a log for today exists:** Read that file and append the new phases from the current session to it. Do not create a new file.
    *   **If no log for today exists:** Create a new file named `_rules/_devlog/[YYYY-MM-DD]_[concise-title-separated-by-dashes].md`.
        *   **Naming Convention:** The title **MUST NOT** be generic like `session_summary`. It must be descriptive of the main task (e.g., `2025-07-04_csv-import-fix.md`, `2025-10-12_auth-refactor-and-docs.md`).
4.  **Maintain Index:** Always update the `_rules/_devlog/_index.md` file with work subjects from the current session.
    *   **If index doesn't exist:** Create the index file with the current session's work subjects.
    *   **If index exists:** Append new work subjects to the existing table.
    *   **Nomenclature Rules:** Use prefix format `[prefix]description` for subjects (e.g., `[fix]user-authentication`, `[feature]csv-import`, `[deploy]v4.1.0`).

## Output Structure (Embedded Template):
Generate or update a single markdown file with the following structure.

---

# Comprehensive Development Log: [Briefly Describe Main Goal of the Session]

**Date:** [Current Date: YYYY-MM-DD]

### **Objective:**
To provide a complete, transparent, and chronological log of the entire development and troubleshooting process for the features worked on during this session. This document details every assumption, every action taken, the resulting errors, and the evidence-based corrections, serving as a definitive record to prevent repeating these mistakes.

---

### **Phase [X]: [Name of the First Major Task or Problem]**

**Initial Problem:** [Describe the starting problem or goal for this phase.]

*   **My Assumption/Plan #1:** [Describe the initial plan or assumption.]
    *   **Action Taken:** [Detail the specific steps taken, e.g., "Modified file X to do Y", "Ran command Z".]
    *   **Result:** [Describe the outcome. Was it a success, failure, or partial success? Include any errors or unexpected behavior.]
    *   **Analysis/Correction:** [Explain why the initial assumption was right or wrong. If wrong, what was the evidence (e.g., error message, user feedback, file inspection) that led to the correction? What was the fix?]

*(Repeat for all assumptions and plans within the phase)*

---

### **Phase [Y]: [Name of the Second Major Task or Problem]**

[Repeat the structure from the previous phase for each major part of the session.]

---

### **Final Session Summary**

**Final Status:** [Briefly describe the state of the feature(s) at the end of the session.]
**Key Learnings:**
*   [A key technical takeaway, e.g., "Electron-builder's `asarUnpack` is required for native addons to preserve their directory structure."]
*   [Another key learning, e.g., "Backspace handling in contenteditable requires differentiating between empty and non-empty states to provide intuitive merging vs. de-escalation."]

---

### **Architectural Relationships**
<!-- Format: [From Entity] -> [To Entity] (relationship type) -->
- EntityA -> EntityB (uses)
- EntityC -> EntityA (depends on)

---

## Guidelines for Generation:
1.  **Chronological Order:** The phases must follow the order in which they occurred in the conversation.
2.  **Focus on the "Why":** Don't just list actions. Explain the *reasoning* behind each action (the assumption) and the *analysis* of the result. The goal is to capture the thought process.
3.  **Be Honest About Mistakes:** The most valuable parts of the log are the "Flawed Assumptions" or incorrect plans. Document them clearly.
4.  **Use Evidence:** When a correction is made, mention the evidence that prompted it (e.g., "The user provided 'before' and 'after' HTML that showed...", "The error message `net::ERR_FILE_NOT_FOUND` indicated...").
5.  **First-Person Narrative:** Write from the perspective of the AI assistant who performed the work (e.g., "My flawed assumption was...", "I modified the file...").

---

## Index Maintenance Instructions

**Index Reference:** All work subjects from this session must be referenced in the `_rules/_devlog/_index.md` file.

### **CRITICAL AI UPDATE RULES:**
1. **APPEND ONLY:** Add new rows to the **existing Markdown table** at the very bottom of the index file.
2. **NO NEW HEADERS:** Do not create a new "## Work Index" header. Use the one already there.
3. **ONE ROW PER SUBJECT:** Each distinct work subject gets its own line.

### Index Structure:
```markdown
| [prefix] subject-description | Brief problem description | YYYY-MM-DD | [filename.md](filename.md) |
```

### Subject Nomenclature:
- **[fix]** - Bug fixes and error resolution
- **[feature]** - New feature implementation
- **[enhance]** - Improvements to existing functionality
- **[rationalize]** - Code cleanup and consolidation
- **[deploy]** - Deployment activities and version releases
- **[security]** - Security fixes and vulnerability patches
- **[debug]** - Troubleshooting and investigation
- **[test]** - Testing and validation activities

### Example Subjects:
- `[rationalize] Export endpoint system` - Consolidated 5 redundant export endpoints to 2 unified endpoints
- `[fix] Vector export format detection` - Added intelligent format selection for vector vs regular tables
- `[enhance] API client export support` - Updated frontend to use rationalized export endpoints
- `[deploy] Export rationalization v4.1.141` - Successfully deployed unified export system to staging

**Important:** Each distinct work subject in a session should be listed on its own line in the index, even if multiple subjects reference the same devlog file.

**Note:** Add a reference to this index maintenance in the devlog's "Final Session Summary" section to remind users that subjects must be referenced in the `_index.md` file in case the AI assistant doesn't follow this prompt directly.