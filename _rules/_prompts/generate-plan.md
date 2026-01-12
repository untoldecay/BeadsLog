# Generate Feature Implementation Plan

## Purpose
This prompt template guides AI assistants in creating comprehensive, structured implementation plans for features or migrations. The generated plan should serve as a living document that tracks progress, documents decisions, and provides technical context.

## Quick Reference
**Output Location**: `_rules/_plans/{status}/`  
**Filename Format**: `[prefix]descriptive-name.md`  
**Status Folders**: `todo/`, `started/`, `validation/`, `done/`  
**Prefix Purpose**: Indicates priority/urgency within each folder  
**Example**: `_rules/_plans/started/[current]user-authentication-plan.md`

---

## Input Requirements

Before generating a plan, gather the following information:

### 1. **Feature/Project Context**
- Feature name or migration title
- Current version/state
- Target completion date or milestone
- Overall objective and scope

### 2. **Technical Stack**
- Frontend framework and libraries
- Backend framework and dependencies
- Database and extensions
- Deployment platform
- Key architectural patterns

### 3. **Feature Breakdown**
- List of main features or components
- Sub-features or capabilities for each main feature
- Dependencies between features
- Priority levels (P1, P2, P3, etc.)

### 4. **Existing Infrastructure**
- Available backend endpoints (with HTTP methods and paths)
- Existing UI components
- Shared utilities or services
- Authentication/authorization mechanisms

### 5. **Constraints & Requirements**
- Technical limitations
- Performance requirements
- Compatibility requirements
- Security considerations

---

## Plan Structure Template

Generate a plan following this exact structure:

### **Header Section**
```markdown
# [Feature/Project Name]

**Current Version**: [version number]
**Last Updated**: [YYYY-MM-DD]
**Status**: [percentage]% [status description]

---
```

### **Section 1: Feature Status Overview**

#### Format:
```markdown
## 📊 Feature Status Overview

### ✅ **COMPLETE Features** ([percentage]%)

#### [Category Number]. [Category Name]
- ✅ [Feature item 1]
- ✅ [Feature item 2]
- ⏳ [In-progress item] (if applicable)
- ❌ [Pending item] (if applicable)

[Repeat for each category]

---

### ⏳ **IN PROGRESS Features** ([percentage]%)
[List features currently being worked on]

---

### ❌ **PENDING Features** ([percentage]%)
[List features not yet started]

---
```

#### Requirements:
- Use emoji indicators: ✅ (complete), ⏳ (in progress), ❌ (pending)
- Group features by logical categories (Auth, Database Ops, UI/UX, etc.)
- Include percentage completion for each section
- Be specific and actionable in feature descriptions

---

### **Section 2: Implementation Priorities**

#### Format:
```markdown
## 🎯 Implementation Priorities

### **Priority 1: [Priority Name]** [Status Emoji] **[Status Text]**
[Brief description of priority group]

[Number]. **[Feature Name]** [Status Emoji] **[Status Text]**
   - [Status Emoji] **UI**: [UI implementation details]
   - [Status Emoji] **Modal/Component**: [Component details]
   - [Status Emoji] **Backend**: [Backend endpoint/logic details]
   - [Status Emoji] **Update**: [Side effects or updates needed]
   - [Status Emoji] **Deployed**: [Version deployed]

[Repeat for each priority level]

---
```

#### Requirements:
- Organize features by implementation priority (P1, P2, P3)
- Break down each feature into implementation layers (UI, Backend, Integration)
- Include deployment version tracking
- Mark optional features clearly
- Provide technical implementation hints

---

### **Section 3: Available Backend Endpoints**

#### Format:
```markdown
## 📋 Available Backend Endpoints

### Fully Implemented & Working ([version]):
- `[HTTP METHOD] [endpoint path]` - [Description] [Status Emoji if new]
- `[HTTP METHOD] [endpoint path]` - [Description]

### Future Enhancements (Optional):
- `[HTTP METHOD] [endpoint path]` - [Description]

---
```

#### Requirements:
- List ALL available endpoints with HTTP methods
- Group by implementation status (working vs. planned)
- Mark newly added endpoints with ✅ NEW
- Include brief descriptions
- Maintain alphabetical or logical ordering

---

### **Section 4: Progress Tracking**

#### Format:
```markdown
## 📈 Progress Tracking

**Overall Completion**: [percentage]% [Status Emoji]

| Category | Complete | Remaining | Progress |
|----------|----------|-----------|----------|
| [Category 1] | X/Y | Z | [percentage]% [Emoji] |
| [Category 2] | X/Y | Z | [percentage]% [Emoji] |
| **[Priority Feature]** | **X/Y** | **Z** | **[percentage]%** [Emoji] |

---
```

#### Requirements:
- Create a table with completion metrics
- Bold priority or newly completed categories
- Include emoji indicators for 100% completion
- Calculate accurate percentages
- Update regularly as work progresses

---

### **Section 5: Technical Notes**

#### Format:
```markdown
## 🎓 Technical Notes

### Recent Achievements ([Month Year])
- **[Feature Name] ([version])**: [Brief description]
- **[Feature Name] ([version])**: [Brief description]

### Key Learnings
- [Technical insight or gotcha]
- [Best practice discovered]
- [Problem and solution]

### Architecture
- **Frontend**: [Framework + key libraries]
- **Backend**: [Framework + key middleware]
- **Database**: [Database + extensions]
- **Deployment**: [Platform + approach]
- **Authentication**: [Auth strategy]

---
```

#### Requirements:
- Document recent achievements with versions
- Capture key learnings and gotchas
- Summarize architecture decisions
- Include lessons learned for future reference
- Keep entries concise and actionable

---

### **Section 6: Deployment History**

#### Format:
```markdown
## 🚀 Deployment History

### v[version] ([YYYY-MM-DD]) [Status Emoji] [CURRENT/ARCHIVED]
- **Docker Image**: `[registry]/[image]:[tag]`
- **[Platform] App ID**: `[app-id]` (if applicable)
- **Status**: [Status Emoji] [Status text]
- **Features**:
  - [Feature or fix 1]
  - [Feature or fix 2]
- **Files Modified**: [count] files ([file names with line counts])
- **Dependencies Added**: [list of new dependencies]
- **Bug Fixes**: (if applicable)
  - [Bug description and fix]
- **Deployment Notes**: (if applicable)
  - [Special deployment considerations]

[Repeat for each version, newest first]

---
```

#### Requirements:
- List deployments in reverse chronological order (newest first)
- Include Docker image tags or build identifiers
- List all features/fixes in each deployment
- Document files modified with line counts
- Note any new dependencies
- Include deployment-specific notes or gotchas
- Mark current version clearly

---

### **Footer Section**

#### Format:
```markdown
**Created**: [YYYY-MM-DD]
**Last Updated**: [YYYY-MM-DD HH:MM]
**Status**: [percentage]% [status description] [Status Emoji]
**Git Commit**: [commit hash or branch name]

---
```

---

## Output Location & Naming Convention

### **File Location**
All generated plans MUST be saved in the appropriate status folder:
- `_rules/_plans/todo/` - Plans not yet started
- `_rules/_plans/started/` - Plans currently being worked on
- `_rules/_plans/validation/` - Plans being tested/validated
- `_rules/_plans/done/` - Completed plans

### **File Naming Convention**
Use the following prefix format: `[prefix]plan-name.md`

**Prefix System by Folder:**

Prefixes indicate **priority/urgency within each state**, allowing multiple items to coexist with different attention levels.

#### `todo/` Folder:
- `[next]` - Next priority item(s) to start (highest urgency)
- `[todo]` - Standard todo items (medium priority)
- No prefix - Backlog/lower priority items

#### `started/` Folder:
- `[current]` - Main active work (primary focus)
- `[started]` - Other ongoing work items (secondary focus)

#### `validation/` Folder:
- `[testing]` - Items being tested/validated
- `[validation]` - Items awaiting validation

#### `done/` Folder:
- `[done]` - Completed items
- `[cancelled]` - Abandoned plans (for historical reference)

**Examples:**
- `_rules/_plans/todo/[next]user-profile-management.md` (next to start)
- `_rules/_plans/todo/[todo]api-optimization.md` (standard todo)
- `_rules/_plans/todo/database-backup-feature.md` (backlog)
- `_rules/_plans/started/[current]api-migration-plan.md` (main work)
- `_rules/_plans/started/[started]frontend-refactor.md` (secondary work)
- `_rules/_plans/validation/[testing]frontend-components-v4.0.40.md`
- `_rules/_plans/done/[done]authentication-implementation.md`

**Naming Rules:**
- Use lowercase with hyphens for readability
- Be descriptive but concise
- Include feature/project name clearly
- **Choose appropriate prefix** based on priority/urgency within the folder
- **Move file to new folder AND update prefix** when status changes
- **Multiple items can share the same prefix** in a folder (e.g., multiple `[next]` items)

---

## Generation Instructions

When generating a plan using this template:

### 0. **Read and Prepare Roadmap** ⚠️ CRITICAL
**BEFORE creating any plan**, read the project roadmap:
- **File**: `_rules/_plans/roadmap.md`
- **Purpose**: Understand current priorities, active work, and version targets
- **Check**: Verify new plan doesn't duplicate existing work
- **Context**: Use roadmap to set appropriate priority and version number
- **Plan**: Determine where new plan fits in the overall roadmap

### 1. **Analyze Input Context**
- Parse all provided feature requirements
- Identify logical groupings and categories
- Determine dependencies and priorities
- Extract technical stack details
- **Cross-reference with roadmap** to ensure alignment

### 2. **Structure the Plan**
- Follow the template structure exactly
- Use consistent emoji indicators throughout
- Maintain clear hierarchy (H2 → H3 → H4)
- Use bold for emphasis on key items

### 3. **Be Specific and Actionable**
- Break down features into concrete tasks
- Include technical implementation details
- Specify endpoints, components, and files
- Provide version tracking

### 4. **Track Progress Accurately**
- Calculate percentages based on completed items
- Update status indicators consistently
- Maintain the progress tracking table
- Document deployment history

### 5. **Capture Technical Context**
- Document architecture decisions
- Record key learnings and gotchas
- Note deployment considerations
- Include troubleshooting hints

### 6. **Maintain as Living Document**
- Update timestamps on every change
- Move completed items to appropriate sections
- Add new learnings as discovered
- Track version history
- **Move file to new folder** when status changes (e.g., `todo/` → `started/` → `validation/` → `done/`)
- **Update roadmap.md** when:
  - Plan moves between folders (update tables)
  - Plan priority changes (update priority indicators)
  - Plan is completed (move to "Recently Completed")
  - Plan is archived (add to "Recently Completed" with link to archived/)
  - Version target changes (update version column)

### 7. **Create the Plan File**
- Save to appropriate `_rules/_plans/{status}/` directory
- Use appropriate prefix based on priority/urgency (typically `[todo]` for new plans, or `[next]` if high priority)
- Follow naming convention: `[prefix]descriptive-feature-name.md`
- Confirm file creation with full path including folder

### 8. **Update the Roadmap** ⚠️ CRITICAL
**AFTER creating the plan**, update the roadmap:
- **File**: `_rules/_plans/roadmap.md`
- **Update "Last Updated" date** at the top
- **Add new plan** to appropriate section:
  - If `[next]` prefix → Add to "Next Priorities (Todo)" table
  - If `[todo]` prefix → Add to "Next Priorities (Todo)" table
  - If `[current]` → Add to "Current Focus" AND "Active Work" table
  - If `[started]` → Add to "Active Work" table
  - If `[testing]` → Add to "Validation & Testing" table
- **Update Quick Stats** (active plan counts, focus areas)
- **Update Long-Term Roadmap** if this changes version targets
- **Verify no duplicates** in roadmap entries

**Roadmap Entry Format**:
- Include priority indicator (🔴/🟡/🟢)
- Link to plan filename (relative path from _plans/)
- Add version target (e.g., v4.2.0)
- Brief 1-line description
- ETA estimate (days/weeks/months)

**Example Update**:
```markdown
| 🔴 P1 | **Database Token System** ([next]database-token-system-v4.2.0.md) | v4.2.0 | Per-database API tokens for N8N security | 1-2 weeks |
```

---

## Example Usage

**User provides:**
- Feature: "Add user profile management"
- Stack: React + Express + MongoDB
- Requirements: Avatar upload, bio editing, privacy settings

**AI generates:**
1. **Creates file**: `_rules/_plans/todo/[todo]user-profile-management.md`
2. **Populates with complete plan** following the template:
   - Feature breakdown (Profile Display, Avatar Upload, Bio Editor, Privacy Controls)
   - Implementation priorities (P1: Basic Profile, P2: Avatar, P3: Privacy)
   - Backend endpoints needed
   - Progress tracking table
   - Technical notes on file upload strategy
   - Deployment history section (empty, ready for updates)
3. **Confirms creation** with full file path including folder

---

## Best Practices

### ✅ DO:
- **Read `roadmap.md` BEFORE creating any new plan**
- **Update `roadmap.md` AFTER creating, moving, or completing plans**
- **Save plans to `_rules/_plans/{status}/` with appropriate `[prefix]`**
- **Move file to new folder AND update prefix when status/priority changes**
- **Use prefixes to indicate priority/urgency within each folder**
- Keep feature descriptions concise but complete
- Use consistent formatting throughout
- Update the plan immediately after changes
- Document both successes and failures
- Include version numbers for all deployments
- Cross-reference related features
- Maintain chronological deployment history
- Keep roadmap.md synchronized with actual plan files

### ❌ DON'T:
- **Forget to read/update `roadmap.md` when working with plans**
- **Create plans outside `_rules/_plans/{status}/` directories**
- **Forget the `[prefix]` in filename** (except for backlog items in `todo/`)
- **Leave files in wrong folder** (e.g., `[done]` file in `todo/` folder)
- **Use mismatched prefix and folder** (e.g., `[current]` in `todo/` folder)
- **Create duplicate plans** - always check roadmap.md first
- Leave status indicators inconsistent
- Skip technical details in implementation notes
- Forget to update percentages and timestamps
- Remove historical information (archive instead)
- Use vague descriptions ("improve UI")
- Mix completed and pending items in same section

---

## Token Efficiency Notes

This plan format is optimized for:
- **Quick scanning**: Emoji indicators and bold headers
- **Context retrieval**: Structured sections with clear hierarchy
- **Progress tracking**: Tables and percentages at a glance
- **Technical reference**: Endpoints and architecture in dedicated sections
- **Historical context**: Deployment history with versions

The format balances human readability with AI parsing efficiency, making it ideal for long-running projects with multiple contributors (human and AI).

---

**Rule Application**: Plan and Break Down, Extended Documentation, Learning Capture

---

## 📋 Roadmap Maintenance Checklist

When working with plans, always:
- [ ] Read `_rules/_plans/roadmap.md` BEFORE creating new plan
- [ ] Check for duplicate or overlapping work
- [ ] Determine appropriate priority and version target
- [ ] Create the plan file in correct folder with correct prefix
- [ ] Update roadmap.md with new plan entry
- [ ] Update "Last Updated" date in roadmap
- [ ] Update Quick Stats if counts changed
- [ ] Verify roadmap tables are accurate and current

When moving or completing plans:
- [ ] Move plan file to new folder
- [ ] Update plan prefix if needed
- [ ] Update roadmap.md tables (remove from old section, add to new)
- [ ] Update "Recently Completed" if archiving
- [ ] Update "Last Updated" date in roadmap
- [ ] Update Quick Stats

**The roadmap is the single source of truth for project status - keep it synchronized!**
