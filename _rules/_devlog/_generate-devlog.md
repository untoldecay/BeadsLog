# Prompt: Generate Chronological Debugging & Development Log (AI Enhanced)

## Objective:
Analyze the entire conversation history of the current session and generate a comprehensive, chronological development log. The primary purpose is to be a transparent record of the entire problem-solving process, detailing every assumption (especially flawed ones), every action taken, the resulting outcomes, and the evidence-based corrections that led to the final solution.

## ✨ Background AI Active:
Background AI Enrichment is ENABLED. You do NOT need to manually extract relationships or use arrows. Focus strictly on the technical narrative and your problem-solving journey. The system will automatically build the architectural graph from your prose in the background.

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
4.  **Record Session:** After creating/updating the devlog file, you MUST record the session using the `bd devlog record` command. This ensures the index is updated with proper metadata (Author, Date, etc.) automatically.
    *   **Command:** `bd devlog record --subject "[prefix] description" --problem "brief description" --file "_rules/_devlog/[filename].md"`
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

## End of Instructions
