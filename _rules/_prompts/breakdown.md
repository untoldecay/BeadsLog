Break down the following feature into separate logical prompts that can be supplied to an AI coding assistant for implementation in logical chunks. Do not perform any code generation at this stage.

Guidelines for Feature Breakdown:

1. Identify core components of the feature
2. List dependencies and prerequisites
3. Break down into small, manageable tasks
4. Consider error handling and edge cases
5. Note UI/UX requirements
6. Identify potential integration points
7. Consider testing requirements
8. Emphasize important aspects from previous discussions
9. Ensure plan aligns with user's approach and philosophy

Template for Response:

1. Feature Overview
   * Brief description of the feature
   * Main objectives and expected outcome
   * User's approach and key priorities
2. Component Breakdown
   * List major components
   * Dependencies between components
   * Technical requirements
3. Implementation Steps
   * Ordered list of tasks with estimated complexity
   * Dependencies between tasks
   * Request commits after significant steps
4. Considerations
   * Security, performance, and error handling needs
   * Testing requirements
5. Validation Criteria
   * Success criteria and test scenarios
   * Edge cases to consider
6. Learning Capture Plan
   * Key technical issues likely to be encountered
   * Concise solutions for common problems (token-efficient format)
   * Focus on AI-relevant knowledge, not human documentation
7. ETA
   * Estimated completion time for entire feature
   * Breakdown of time for each implementation step

After user validation of your proposal:

1. **CRITICAL: CREATE FEATURE FOLDER AND PLAN FILE IMMEDIATELY**
   * Create a dedicated folder in "_rules/_features" named after the feature (e.g., "_rules/_features/[feature-name]")
   * Inside this folder, create a plan .md file named after the feature (e.g., "_rules/_features/[feature-name]/[feature-name]-plan.md")
   * Create a 'devlog' subfolder for implementation logs (e.g., "_rules/_features/[feature-name]/devlog/")
   * DO NOT implement any code until folder structure is created and confirmed

2. **Plan File Format**
   * Use checklist format with sections:
     - [ ] Feature requirements
     - [ ] Component implementation
     - [ ] Testing
     - [ ] Learning capture
     - [ ] Review
   * Include ETA section to update after each step
   * If based on analysis from "_rules/_analysis", include reference link

⚠️ **IMPORTANT WARNING**: 
- DO NOT GENERATE ANY OTHER DOCUMENT THAN REQUESTED. Only create files specified in these instructions.
- NEVER skip creating the feature folder and plan file in "_rules/_features".
- ALWAYS create the plan file BEFORE implementing any code.
- DO NOT OUTPUT THE ENTIRE SUMMARY OF WHAT YOU DONE IN THE CHAT CREATING THE FEATURE FOLDER AND PLAN FILE, JUST A COMPACT OVERVIEW (too overwhelming otherwise)
- ALWAYS ITERATE ON EXISTING FILES unless specifically instructed otherwise by the user. Do not create draft files or temporary versions - modify the actual target files directly to minimize file proliferation and cognitive load.
- MAINTAIN PLAN CONSISTENCY: When marking tasks as completed, ALWAYS update BOTH the component checkboxes at the top of the plan AND the implementation steps at the bottom. Keep all progress indicators in sync.

Implementation Process and User Flow:

1. **Initial Implementation**
   * Implement tasks in first step of breakdown plan
   * Thoroughly test implementation
   * Document challenges/deviations in "01-[feature-title]-devlog-[title]-[date].md" inside the 'devlog' folder (use incremental numbers: 01, 02, etc. for each new devlog file)

2. **Validation Request:**
   * Present completed step to user for validation
   * Include what has been implemented, test results, and next steps
   * Ask directly for validation or change requests without artificial UI elements
   * Example: "I've completed [step]. Here's what was implemented. Does this look good or would you like changes?"

3. **After User Validation:**
   * Update ETA in feature plan
   * Mark step as done in BOTH component sections AND implementation steps: `- [x] Completed step`
   * Document challenges in "XX-[feature-name]-devlog-[title]-[date].md" in the 'devlog' folder (continuing the incremental numbering)
   * Request commit with suggested message
   * Example: "Should I commit these changes with message: 'feat: implement [feature] - [functionality]'?"

4. **Before Next Step:**
   * Confirm updated ETA
   * Verify documentation is current
   * Validate plan consistency to ensure both component checkboxes and implementation steps are in sync
   * Ask about proceeding to next step
   * Example: "Updated ETA is [estimate]. Plan consistency has been validated. Ready to proceed with [next step]?"

5. **Repeat Process** for each implementation step

6. **Final Validation:**
   * Present entire feature implementation to user for final validation
   * Compile token-efficiently the most important learnings from 'devlog' folder into '_rules/_learnings/ai/' for AI assistant reference
   * Focus ONLY on problems encountered and solutions found to prevent repeating mistakes in future
   * Use concise, structured format with minimal narrative and maximum information density
   * Include implementation summary, test results, and documentation status
   * Ask directly about documentation needs

7. **Documentation Completion:**
   * If requested, create documentation in "_rules/_documentation" folder following '_rules/_prompts/documentation.md'
   * Update feature tracking file
   * Suggest final review
