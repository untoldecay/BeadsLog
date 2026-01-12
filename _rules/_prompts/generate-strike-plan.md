# Generate Strike Plan (Quick Action Plan)

## Purpose
Strike plans are **lean, actionable plans** for rapid execution. Unlike comprehensive plans, strike plans focus on clarity and speed with minimal text. Use ✅ for completed items (not [x]) for better visual scanning. Use for features that are well-understood and need quick tracking.

**When to use**:
- Clear, scoped features with known requirements
- Quick iterations or bug fixes
- Features already discussed/documented elsewhere
- Work that needs immediate action tracking

**When NOT to use**:
- Complex features needing detailed breakdown
- New architecture requiring extensive documentation
- Features with many unknowns or dependencies
- Long-term projects (>1 month)

---

## Strike Plan Structure

### 1. Header (Single Line)
```markdown
# [Feature Name]

**Version**: [version] | **Updated**: [YYYY-MM-DD] | **Branch**: [branch-name]
```

### 2. Context (1-2 Lines MAX)
Brief summary of what this is and current state. Reference roadmap or other docs if needed.

**Example**:
```markdown
## Context
Complete multi-tenant RLS system deployed to staging (v4.0.41). Backend API, frontend UI, and database schema all implemented. Currently awaiting browser cache-clear retest before production deployment to v4.1.0.
```

### 3. Dependencies (Bullet List)
Non-exhaustive list of key dependencies. One item per line, concise.

**Format**:
```markdown
## Dependencies
- [Branch/file/concept]
- [Component/library]
- [Database/table]
- [External system]
- Related: [link to roadmap/plan]
```

**Example**:
```markdown
## Dependencies
- `api-security` branch (backend + frontend changes)
- shadcn/ui components (TenantManagement, RLSPolicyManagement)
- Database: tenants table with 4 sample tenants
- Cloudron deployment platform
- Related: Phase 3 completion milestone from roadmap
```

### 4. Feature Status (Checkboxes)
Simple checked/unchecked list. Group by logical category. NO percentages, NO time estimates, NO extra text.

**Use ✅ for completed items, [ ] for pending items** - better visual clarity.

**Format**:
```markdown
## 📊 Feature Status

### [Category 1]
- ✅ [Completed item]
- ✅ [Completed item]
- [ ] [Pending item]

### [Category 2]
- ✅ [Completed item]
- [ ] [Pending item]
- [ ] [Pending item]
```

**Example**:
```markdown
## 📊 Feature Status

### Backend
- ✅ Type definitions & API integration
- ✅ Authentication context with tenant support
- ✅ Tenant management endpoints
- [ ] Production deployment

### Frontend
- ✅ TenantManagement view
- ✅ RLSPolicyManagement view
- [ ] Final smoke tests

### Testing
- [ ] Retest on v4.0.41
- [ ] Production smoke tests
```

### 5. Next Actions (Detailed Steps)
This is the **working section** - rewrite as work progresses. Break into numbered steps with clear instructions.

**Format**:
```markdown
## 🎯 Next Actions

### Step [N]: [Step Name]
**[One-line description of what this step accomplishes]**

1. **[Action 1]**
   - [Sub-instruction or detail]
   - [Command or specific method]
   
2. **[Action 2]**
   - [ ] [Testable item]
   - [ ] [Testable item]

3. **[Action 3]**
   ```bash
   # Commands if needed
   ```

4. **If [Condition]**
   - [What to do]
   - [Where to document]

5. **If [Other Condition]**
   - [Next step]

---

### Step [N+1]: [Next Step Name]
**[Brief description]**

[Repeat format above]
```

**Example**:
```markdown
## 🎯 Next Actions

### Step 1: Retest Staging (v4.0.41)
**Clear browser cache and verify all features work**

1. **Clear Cache**
   - Open Chrome DevTools → Right-click refresh → "Empty Cache and Hard Reload"
   - OR use Incognito Mode (Cmd+Shift+N)
   - Navigate to: https://api-security.decaylab.com

2. **Verify Version Loaded**
   ```bash
   curl https://api-security.decaylab.com/health | jq
   # Expected: "version": "4.0.41"
   ```

3. **Run Manual Tests**
   - [ ] Login works
   - [ ] 4 sample tenants visible
   - [ ] Create tenant works
   - [ ] No 500 errors in console

4. **If Tests Fail**
   - Document errors in this plan
   - Create bug fix tasks
   - Retest after fixes

5. **If Tests Pass**
   - Move to Step 2 (Production Prep)

---

### Step 2: Deploy to Production
**Deploy v4.1.0 to production environment**

1. **Pre-Deployment Checklist**
   - [ ] All staging tests passed
   - [ ] Release notes created
   - [ ] Backup plan ready

[Continue with deployment steps...]
```

---

## Generation Instructions

### STEP 0: Read Roadmap ⚠️ MANDATORY
```markdown
1. Read `_rules/_plans/roadmap.md`
2. Verify no duplicate work exists
3. Determine version target and priority
4. Understand dependencies on other plans
```

### STEP 1: Gather Minimal Info
**Only collect what's essential**:
- Feature name
- Current version/state
- What's done vs. what's pending
- Immediate next actions (next 1-3 steps)
- Key dependencies (files, components, systems)

**Skip**:
- Detailed architecture (unless critical to actions)
- Long explanations
- Historical context
- Nice-to-have features
- Far-future plans

### STEP 2: Create Strike Plan
**File location**: `_rules/_plans/[status]/[prefix]feature-name.md`

**Structure (in order)**:
1. Header (1 line: title + version + date + branch)
2. Context (1-2 lines max)
3. Dependencies (5-10 bullet points)
4. Feature Status (grouped checkboxes, no fluff)
5. Next Actions (detailed steps, rewritable)

### STEP 3: Keep It Actionable
**Focus on ACTIONS, not descriptions**:
- ✅ "Clear browser cache and retest"
- ❌ "The browser cache issue is a known problem that affects..."

**Use commands and specific instructions**:
- ✅ `curl https://api.com/health | jq`
- ❌ "Check the health endpoint"

**Make it testable**:
- ✅ Checkbox items with clear pass/fail
- ❌ Vague items like "verify it works"

### STEP 4: Update Roadmap ⚠️ MANDATORY
```markdown
1. Open `_rules/_plans/roadmap.md`
2. Update "Last Updated" date
3. Add new plan to appropriate section
4. Use format: | Priority | **Name** (filename) | Version | Description | ETA |
5. Update Quick Stats
```

---

## Strike Plan Patterns

### Pattern: Testing & Deployment
```markdown
## 🎯 Next Actions

### Step 1: Test on Staging
**Verify all features work on staging environment**

1. **Access Environment**
   - URL: https://staging.example.com
   - Clear browser cache first

2. **Run Test Suite**
   - [ ] Feature A works
   - [ ] Feature B works
   - [ ] No console errors

3. **If Tests Pass** → Move to Step 2
4. **If Tests Fail** → Document errors, fix, retest

---

### Step 2: Deploy to Production
**Push tested version to production**

1. **Backup Production**
   ```bash
   ./backup.sh production
   ```

2. **Deploy**
   ```bash
   ./deploy.sh production v1.2.3
   ```

3. **Verify**
   - [ ] Version endpoint shows v1.2.3
   - [ ] Smoke tests pass
   - [ ] Monitor for 30 minutes

4. **Update Roadmap** → Mark complete
```

### Pattern: Bug Fix
```markdown
## 📊 Feature Status

### Issue
- ✅ Bug identified and reproduced
- ✅ Root cause found (auth token expiry)
- [ ] Fix implemented
- [ ] Tested on staging
- [ ] Deployed to production

---

## 🎯 Next Actions

### Step 1: Implement Fix
**Update token refresh logic in auth middleware**

1. **Files to Modify**
   - `server/middleware/auth.js` (line 45-60)
   - `server/utils/token.js` (add refresh function)

2. **Changes**
   - Add token refresh before expiry (5 min buffer)
   - Handle refresh errors gracefully
   - Log refresh events

3. **Test Locally**
   - [ ] Token refreshes automatically
   - [ ] Expired tokens rejected
   - [ ] New tokens issued correctly

4. **If Tests Pass** → Deploy to staging
```

### Pattern: New Feature (Simple)
```markdown
## 📊 Feature Status

### Backend
- [ ] API endpoint created
- [ ] Database schema updated
- [ ] Tests written

### Frontend
- [ ] UI component created
- [ ] API integration
- [ ] Validation added

### Deployment
- [ ] Staging deployment
- [ ] Production deployment

---

## 🎯 Next Actions

### Step 1: Create Backend Endpoint
**Add POST /api/widgets endpoint**

1. **Create Endpoint**
   - File: `server/routes/widgets.js`
   - Method: POST
   - Body: `{ name: string, type: string }`
   - Response: `{ id: uuid, name: string, type: string }`

2. **Update Database**
   ```sql
   CREATE TABLE widgets (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     name VARCHAR(255) NOT NULL,
     type VARCHAR(50) NOT NULL,
     created_at TIMESTAMP DEFAULT NOW()
   );
   ```

3. **Write Tests**
   - [ ] Create widget succeeds
   - [ ] Validation rejects invalid data
   - [ ] Returns correct response format

4. **If Tests Pass** → Move to Step 2 (Frontend)

---

### Step 2: Create Frontend Component
**Build WidgetForm component**

1. **Create Component**
   - File: `components/widgets/WidgetForm.tsx`
   - Use shadcn/ui Dialog + Form components
   - Fields: name (text), type (select)

2. **API Integration**
   - POST to /api/widgets
   - Show success toast
   - Refresh widget list

3. **Test UI**
   - [ ] Form validates inputs
   - [ ] Submission works
   - [ ] Success message shows
   - [ ] Widget appears in list

4. **If Tests Pass** → Deploy to staging
```

---

## File Naming & Location

### Choose Status Folder
- `todo/` - Not started yet
- `started/` - Currently working on
- `validation/` - Testing/validating
- `archived/` - Completed (moved after done)

### Choose Prefix
- `[next]` - Highest priority in that folder
- `[current]` - Main active work (started/ only)
- `[started]` - Secondary active work (started/ only)
- `[todo]` - Standard priority (todo/ only)
- `[testing]` - Being tested (validation/ only)
- No prefix - Lower priority/backlog

### Examples
- `started/[current]multi-tenant-rls-complete-and-deploy.md`
- `todo/[next]database-token-system-v4.2.0.md`
- `validation/[testing]frontend-components-v4.0.40.md`

---

## Best Practices

### ✅ DO:
- **Read roadmap.md before creating**
- **Keep context to 1-2 lines**
- **Use ✅ for completed items, [ ] for pending** (not [x])
- **Make next actions copy-pasteable**
- **Include actual commands and URLs**
- **Update the plan as work progresses**
- **Rewrite "Next Actions" section frequently**
- **Move completed items to ✅ status**
- **Update roadmap.md when done**

### ❌ DON'T:
- Write long explanations
- Add time estimates (no "2 hours", "3 days")
- Include percentages (no "95% complete")
- Add "nice to know" information
- Duplicate info from other docs
- Keep outdated "Next Actions" steps
- Leave vague instructions
- Forget to update roadmap

---

## Strike Plan vs. Full Plan

| Aspect | Strike Plan | Full Plan |
|--------|-------------|-----------|
| **Length** | 1-2 pages | 5-15 pages |
| **Context** | 1-2 lines | Multiple sections |
| **Status** | Checkboxes only | Detailed progress tables |
| **Actions** | Next 1-3 steps | Full implementation roadmap |
| **Timeline** | No estimates | Detailed timelines |
| **History** | None | Deployment history |
| **Technical** | Only if needed for actions | Comprehensive technical notes |
| **Use Case** | Quick, clear features | Complex, long-term projects |

**Example when to use each**:
- **Strike Plan**: "Add export button to table view"
- **Full Plan**: "Migrate entire codebase to Next.js 14"

---

## Maintenance

### When to Update
- **Daily**: Check off completed items in status
- **After each step**: Rewrite "Next Actions" section
- **When blocked**: Add blocker info to current step
- **When priorities change**: Adjust step order
- **When complete**: Update roadmap, move to archived/

### When to Convert to Full Plan
If a strike plan grows beyond 2 pages or needs:
- Extensive technical documentation
- Complex architecture decisions
- Multiple phases spanning weeks
- Historical tracking of many deployments
- Detailed API documentation

**Action**: Create full plan using `generate-plan.md`, reference strike plan for completed work

---

## Example Strike Plan (Complete)

```markdown
# Database Export Feature

**Version**: v4.0.42 | **Updated**: 2025-10-15 | **Branch**: `feature/export`

## Context
Add CSV export functionality to table view. Backend endpoint exists, need to wire up frontend button and download handler.

## Dependencies
- Backend: `GET /api/tables/:table/export` endpoint (already exists)
- Frontend: TableView component (`components/views/TableView.tsx`)
- shadcn/ui Button and DropdownMenu components
- File download utility (create new)

---

## 📊 Feature Status

### Backend
- ✅ Export endpoint exists
- ✅ CSV generation works
- [ ] Add progress tracking

### Frontend
- [ ] Add export button to table header
- [ ] Implement download handler
- [ ] Show progress indicator
- [ ] Add error handling

### Testing
- [ ] Test with small table (<100 rows)
- [ ] Test with large table (>10k rows)
- [ ] Test error cases

---

## 🎯 Next Actions

### Step 1: Add Export Button
**Add download button to TableView header**

1. **Update Component**
   - File: `components/views/TableView.tsx`
   - Add Button with Download icon in table header
   - Place next to existing filter/search buttons

2. **Wire Up Handler**
   ```typescript
   const handleExport = async () => {
     try {
       const response = await apiClient.get(
         `/api/tables/${tableName}/export`,
         { responseType: 'blob' }
       );
       downloadFile(response.data, `${tableName}.csv`);
     } catch (error) {
       toast.error('Export failed');
     }
   };
   ```

3. **Create Download Utility**
   - File: `lib/downloadFile.ts`
   - Function: `downloadFile(blob: Blob, filename: string)`
   - Use blob URL + temporary anchor element

4. **Test**
   - [ ] Button appears in UI
   - [ ] Clicking triggers download
   - [ ] File downloads as CSV
   - [ ] Filename is correct

5. **If Tests Pass** → Move to Step 2

---

### Step 2: Add Progress Indicator
**Show progress for large table exports**

1. **Update Backend** (if needed)
   - Add progress events to export endpoint
   - Stream progress via Server-Sent Events

2. **Update Frontend**
   - Show loading spinner during export
   - Display "Exporting... please wait" toast
   - Handle long-running exports (>30s)

3. **Test**
   - [ ] Progress shows during export
   - [ ] Large exports don't timeout
   - [ ] User can cancel export

4. **If Tests Pass** → Deploy to staging

---

### Step 3: Deploy & Verify
**Push to staging and test**

1. **Deploy**
   ```bash
   git push origin feature/export
   ./deploy.sh staging
   ```

2. **Verify**
   - [ ] Export button visible
   - [ ] Small table export works
   - [ ] Large table export works
   - [ ] Error handling works

3. **Update Roadmap**
   - Mark "Database Export Feature" as complete
   - Move plan to archived/
```

---

**Rule Application**: Plan and Break Down, Readability

**Roadmap Maintenance**: Read before creating, update after creating/completing
