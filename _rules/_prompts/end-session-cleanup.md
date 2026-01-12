# Prompt: End of Session Cleanup & Documentation

## Objective:
Automatically update all project "_rules/_documentation/", "_rules/_plans/", and "_rules/_devlog/" at the end of a development session. This ensures consistency across all tracking files and prevents documentation drift.

## Persona:
Act as a meticulous project manager and technical writer, ensuring all documentation is up-to-date, consistent, and properly organized.

## When to Use:
Run this prompt at the end of any significant development session, especially when:
- New features have been implemented
- Bugs have been fixed
- Architecture decisions have been made
- Plans or roadmaps need updating
- Documentation needs to be consolidated

## Execution Steps:

### 1. Create or Update Session Devlog

**Action:** Generate a comprehensive development log.

**File Handling:**
- Gather key insights from "_rules/_documentation/" to add them to the session devlog (see below).
- Check "_rules/_devlog/" for existing log with today's date (YYYY-MM-DD format)
- If exists: Append new session phases to existing file
- If not, create a new one following the dedictated prompt "_rules/_prompts/generate-devlog.md"

### 2. Clean Duplicate Documentation

**Action:** Review "_rules/_documentation/" and remove or consolidate duplicate files

**Keep These Files:**
- Files with unique, valuable information
- Files referenced in plans or roadmap
- Files with mock API updates (e.g., LOGIN_FLOW_ANALYSIS.md)
- Latest version of any document series

**Remove These Files:**
- Older versions of the same document
- Superseded implementation guides
- Redundant reports (keep only latest)
- Temporary analysis files that have been incorporated elsewhere

**Consolidation Rules:**
- If multiple files cover the same topic, merge into one comprehensive document

**Exception:** Never delete files that contain:
- Mock API update instructions
- Frontend modification tracking
- Unique architectural decisions
- Historical context that might be needed

### 3. Update Backend ETA Log

**Action:** Update "_rules/_documentation/backend-ETA.md"

**Add New Entries:**
- Follow document guidelines for formatting
- Only add entries for actually implemented features, do not add planned features


### 4. Update Current Plan

**Action:** Update the active plan in "_rules/_plans/started/"

**Updates Required:**
- Add "Latest Updates" section at top with today's date
- List session accomplishments
- List documentation created
- Update deployment status
- Mark completed items with ✅
- Update "Current Implementation Status" date
- Add any new blockers or dependencies

**Move Plan If Needed:**
- If feature is 100% complete: Move to `_rules/_plans/done/`
- Rename file: `[done]feature-name-vX.X.X.md`
- Update roadmap.md with new location

### 5. Update Roadmap

**Action:** Update `_rules/_plans/roadmap.md`

**Updates Required:**
- Update "Last Updated" date at top
- Update current version status
- Add today's completed items with ✅
- Update timeline estimates
- Add devlog reference link
- Update remaining tasks
- Adjust priorities if needed

**Format for Active Feature:**
```markdown
## 🚀 vX.X.X: Feature Name (STATUS)
**Plan**: [started/[next]feature-name.md](started/[next]feature-name.md)
**Devlog**: [_devlog/YYYY-MM-DD_description.md](../_devlog/YYYY-MM-DD_description.md)

**✅ Completed (YYYY-MM-DD):**
- ✅ Item 1
- ✅ Item 2

**▢ Remaining:**
- ▢ Item 3
- ▢ Item 4

```

### 6. Summary Report

**Action:** Generate a brief summary of all changes made

**Include:**
- Files created
- Files updated
- Files deleted
- Plans moved
- Key documentation links
- Next session recommendations

**Format:**
```markdown
## Session Cleanup Summary

**Date:** YYYY-MM-DD

### Files Created:
- `_rules/_devlog/YYYY-MM-DD_description.md`
- `_rules/_documentation/NEW_DOC.md`

### Files Updated:
- `_rules/_documentation/backend-ETA.md`
- `_rules/_plans/roadmap.md`
- `_rules/_plans/started/[next]feature.md`

### Files Deleted:
- `_rules/_documentation/OLD_DOC.md` (superseded by NEW_DOC.md)

### Plans Status:
- ✅ Feature X: Backend complete, frontend pending
- 🚀 Feature Y: In progress

### Next Session:
- [ ] Fix frontend logout issue
- [ ] Update mock API
- [ ] Implement Phase 3 UI components
```

## Checklist for Execution:

- [ ] Create/update devlog with session details
- [ ] Review and clean duplicate documentation
- [ ] Update backend-ETA.md with new features
- [ ] Update current plan in _plans/started/
- [ ] Move plan to done/ if feature complete
- [ ] Update roadmap.md with progress
- [ ] Generate summary report
- [ ] Commit all changes with descriptive message

## Example Commit Message:

```
docs: end of session cleanup and documentation update

Session: [Brief description]
Date: YYYY-MM-DD

Updates:
- Created session devlog
- Updated backend ETA with new features
- Updated roadmap and current plan
- Cleaned duplicate documentation
- Consolidated [X] files into [Y]

Files modified: [count]
Documentation: Up to date
Plans: Current
```

## Notes:

- This prompt should be run BEFORE committing final session changes
- All documentation updates should be committed together
- Keep documentation concise but comprehensive
- Use consistent formatting across all files
- Link related documents together
- Maintain chronological order in logs
- Update dates consistently (YYYY-MM-DD format)

## Automation Potential:

This prompt could be enhanced with:
- Automated duplicate detection
- Link validation across documents
- Consistency checking (dates, formatting)
- Auto-generation of summary reports
- Git integration for commit messages
