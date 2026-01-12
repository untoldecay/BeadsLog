# Testing Protocol Generation Prompt

**Purpose**: Generate comprehensive AI agent testing protocols for feature validation

---

## 📋 Prompt Template

Use this prompt to generate testing protocols for any feature:

```
Create a comprehensive testing protocol for [FEATURE_NAME] version [VERSION] following this structure:

## Context
- Feature: [FEATURE_NAME]
- Version: [VERSION]
- Environment: [STAGING_URL]
- Components: [LIST_COMPONENTS]
- API Endpoints: [LIST_ENDPOINTS]

## Requirements
Generate a testing protocol document that includes:

1. **Header Section**
   - Date, Environment URL, Status, Tester (AI Agent), Browser, Version, Overall Result
   - Testing Objective (what are we testing and why)
   - Test Environment details (URL, version, docker image, browser)
   - Pre-Test Checklist (cache clear, devtools, login, version check)

2. **Test Cases** (numbered sequentially)
   For each test case include:
   - Test number and descriptive title with ✅/❌ placeholder
   - **Purpose**: One-line explanation of what this test verifies
   - **Steps**: Numbered list of exact actions to perform
   - **Expected Results**: Bulleted list with ✅ checkmarks of what should happen
   - **Actual Results**: Code block with ⏳ PENDING placeholder
   - **Status**: ⏳ PENDING placeholder

3. **Test Categories to Cover**
   - Version verification
   - Navigation tests (menu items, routing)
   - View loading tests (UI components render)
   - API loading tests (data fetches correctly)
   - Dialog/Modal tests (forms open correctly)
   - Form submission tests (CRUD operations)
   - Form validation tests (required fields, button states)
   - Tab/Filter tests (data filtering works)
   - Error handling tests (network failures, API errors)
   - Console check (no unexpected errors)
   - Regression tests (existing features still work)

4. **Test Results Summary Section**
   - Total Tests count
   - Passed/Failed/Skipped counts
   - Pass Rate percentage
   - Critical Issues Found (list)
   - Blockers for Production (list)
   - Nice-to-Have Improvements (list)

5. **Test Completion Criteria**
   - Checklist of requirements for production readiness
   - Include: pass rate, no critical issues, no blockers, clean console, all CRUD works

6. **Notes for AI Agent**
   - Testing Instructions (how to run tests)
   - Common Issues to Watch For (known problems)
   - Success Indicators (what good looks like)
   - Documentation requirements (screenshots, logs, etc.)

## Formatting Rules
- Use Markdown with proper headers (###)
- Use ✅ for passed tests, ❌ for failed, ⏳ for pending
- Use code blocks for actual results
- Use bulleted lists with ✅ for expected results
- Number all test cases sequentially
- Keep test titles concise but descriptive
- Include specific API endpoints, URLs, and data in steps
- Make steps actionable (exact clicks, exact text to type)

## Output Format

### Step 1: Create Test Suite Folder
```bash
# Create numbered test suite folder in _tests/
mkdir _tests/XX-[feature-name]
cd _tests/XX-[feature-name]
```

### Step 2: Generate Test Protocol
Generate a complete `test-protocol.md` file following the template below.

Example: `_tests/05-multi-tenant-rls/test-protocol.md`

### Step 3: Create package.json
```json
{
  "name": "[kebab-case-feature-name]-tests",
  "version": "[version]",
  "description": "Test suite for [feature description]",
  "private": true,
  "scripts": {
    "test": "echo 'Manual testing protocol - see test-protocol.md for instructions'"
  }
}
```

### Step 4: Update _tests/README.md
Add new test suite to the "Available Test Suites" section with:
- Status, Version, Purpose, Environment, Test Cases count
- Quick Start instructions
- Test Protocol path reference
```

---

## 🎯 Usage Examples

### Example 1: Testing Multi-Tenant RLS Feature

**Input**:
```
Create a comprehensive testing protocol for Multi-Tenant RLS version 4.0.44 following this structure:

## Context
- Feature: Multi-Tenant RLS (Row Level Security)
- Version: 4.0.44
- Environment: https://postgreapi-staging.decaylab.com
- Components: TenantManagement view, RLSPolicyManagement view, API endpoints
- API Endpoints: 
  - GET /admin/tenants
  - POST /admin/tenants
  - PUT /admin/tenants/{id}
  - DELETE /admin/tenants/{id}
  - GET /admin/rls/policies?database={db}
  - POST /admin/rls/policies
  - DELETE /admin/rls/policies/{id}?database={db}
```

**Output**: Complete testing protocol with 23 test cases covering all CRUD operations

---

### Example 2: Testing Database Import/Export Feature

**Input**:
```
Create a comprehensive testing protocol for Database Import/Export version 4.0.41 following this structure:

## Context
- Feature: Database Import/Export
- Version: 4.0.41
- Environment: https://postgreapi-staging.decaylab.com
- Components: ImportExport view, File upload, Download functionality
- API Endpoints:
  - POST /admin/databases/{name}/export
  - POST /admin/databases/{name}/import
  - GET /admin/databases/{name}/export/status
```

---

### Example 3: Testing Vector Search Feature

**Input**:
```
Create a comprehensive testing protocol for Vector Search Advanced Features version 4.4.0 following this structure:

## Context
- Feature: Vector Search Advanced Features
- Version: 4.4.0
- Environment: https://postgreapi-staging.decaylab.com
- Components: VectorSearch view, Similarity search, Embedding visualization
- API Endpoints:
  - POST /api/vector/search
  - GET /api/vector/embeddings
  - POST /api/vector/visualize
```

---

## 📝 Template Variables

When generating a test protocol, replace these variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `[FEATURE_NAME]` | Name of feature being tested | Multi-Tenant RLS |
| `[VERSION]` | Version number | 4.0.44 |
| `[STAGING_URL]` | Staging environment URL | https://postgreapi-staging.decaylab.com |
| `[LIST_COMPONENTS]` | Frontend components involved | TenantManagement, RLSPolicyManagement |
| `[LIST_ENDPOINTS]` | Backend API endpoints | GET /admin/tenants, POST /admin/tenants |
| `[TEST_COUNT]` | Total number of tests | 23 |
| `[DOCKER_IMAGE]` | Expected docker image | untoldecay/pgvector-admin:staging-4.0.44-* |

---

## ✅ Quality Checklist

A good testing protocol should have:

- [ ] Clear testing objective stated upfront
- [ ] Pre-test checklist for setup
- [ ] At least 15-25 test cases (comprehensive coverage)
- [ ] Tests organized by category (navigation, API, CRUD, validation, etc.)
- [ ] Each test has: Purpose, Steps, Expected Results, Actual Results, Status
- [ ] Specific, actionable steps (exact clicks, exact text)
- [ ] Expected results with ✅ checkmarks
- [ ] Actual results as code blocks with ⏳ PENDING
- [ ] Test results summary section
- [ ] Test completion criteria checklist
- [ ] Notes for AI agent with common issues
- [ ] Proper Markdown formatting throughout
- [ ] Consistent numbering and structure

---

## 🔄 Updating Existing Protocols

When updating an existing test protocol:

1. **Keep the structure** - Don't change test numbers or order
2. **Update Actual Results** - Replace ⏳ PENDING with actual findings
3. **Update Status** - Change ⏳ PENDING to ✅ PASSED or ❌ FAILED
4. **Add screenshots/logs** - Include evidence in code blocks
5. **Update summary** - Recalculate pass rate, list issues
6. **Document blockers** - Clearly mark critical issues
7. **Preserve history** - Don't delete failed test results

---

## 🎨 Formatting Standards

### Test Case Format
```markdown
### Test [N]: [Category] - [Action] ✅/❌

**Purpose**: [One-line description]

**Steps**:
1. [Exact action]
2. [Exact action]
3. [Exact action]

**Expected Results**:
- ✅ [Expected outcome]
- ✅ [Expected outcome]
- ✅ [Expected outcome]

**Actual Results**:
\`\`\`
⏳ PENDING
\`\`\`

**Status**: ⏳ PENDING
```

### Status Icons
- ⏳ **PENDING** - Not yet tested
- ✅ **PASSED** - Test passed successfully
- ❌ **FAILED** - Test failed
- ⚠️ **SKIPPED** - Test skipped (dependency failed)

### Result Formatting
```markdown
**Actual Results**:
\`\`\`
✅ Feature works correctly
- API call: HTTP 200 - SUCCESS
- Response: Contains expected data
- UI: Updates correctly
- Console: No errors
\`\`\`

**Status**: ✅ PASSED
```

---

## 🚀 Best Practices

### For AI Agents Running Tests

1. **Run tests in order** - Dependencies may exist between tests
2. **Document everything** - Capture API requests, responses, console logs
3. **Take screenshots** - Visual evidence of UI state
4. **Check network tab** - Verify API calls, status codes, payloads
5. **Check console** - Look for errors, warnings, unhandled rejections
6. **Test edge cases** - Empty states, validation, error handling
7. **Test regressions** - Verify existing features still work
8. **Update in real-time** - Fill in Actual Results as you test
9. **Summarize at end** - Update statistics, list issues
10. **Be thorough** - Don't skip tests, document all findings

### For Humans Reviewing Test Results

1. **Check pass rate** - Should be 100% for production
2. **Review failed tests** - Understand why they failed
3. **Assess blockers** - Determine if production-ready
4. **Verify regressions** - Ensure no existing features broken
5. **Check console** - Should be clean (no red errors)
6. **Review API calls** - Correct status codes, proper data
7. **Test manually** - Spot-check critical flows
8. **Approve or reject** - Make go/no-go decision

---

## 📚 Related Documents

- **Example Protocol**: `_rules/_plans/done/[testing]frontend-phase3-ui-components-v4.0.40.md`
- **Current Protocol**: `_rules/_plans/started/[testing]multi-tenant-rls-v4.0.44.md`
- **Feature Plans**: `_rules/_plans/started/` and `_rules/_plans/todo/`
- **Mission Document**: `_rules/_prompts/mission.md`

---

## 🎯 Success Criteria

A testing protocol is successful when:

- ✅ All tests are documented with clear steps
- ✅ AI agent can run tests without human intervention
- ✅ Results are actionable (clear pass/fail)
- ✅ Issues are documented with evidence
- ✅ Production readiness is clear (go/no-go)
- ✅ Regressions are caught before production
- ✅ Protocol is reusable for future versions

---

**Remember**: The goal is to create testing protocols that AI agents can execute autonomously to verify feature quality before production deployment.
