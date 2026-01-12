# Test Protocol Creation Prompt

**Purpose:** Guide AI agents to create comprehensive test protocols following project architecture

---

## 📋 Overview

When creating test protocols for this project, follow the established architecture and format. Test protocols are detailed documents that guide manual or automated testing and provide a structured way to document test results.

---

## 🎯 Test Protocol Requirements

### 1. File Location
- **Directory:** `_tests/XX-test-suite-name/`
- **Filename:** `test-protocol.md`
- **Numbering:** Sequential (01-, 02-, 03-, etc.)
- **Example:** `_tests/02-phase3-ui-components/test-protocol.md`

### 2. File Structure

Every test protocol **MUST** include these sections:

```markdown
# [Feature Name] Testing - [Version]

**Date**: YYYY-MM-DD
**Environment**: [Staging/Production URL]
**Status**: ⏳ PENDING / ✅ PASSED / ❌ FAILED
**Tester**: [Agent Name or Human Name]
**Browser**: [Chrome/Firefox/Safari/Automated]
**Version Tested**: [Version number]
**Overall Result**: [X/Y tests passed]

## 🔍 Key Findings Summary

**Testing Date**: YYYY-MM-DD
**Testing Duration**: ~XX minutes
**Testing Environment**: [URL or environment details]

**Critical Issues Identified**:
1. ❌/✅ [Issue description]
2. ❌/✅ [Issue description]

**Test Results by Category**:
- ✅/❌ [Category Name]: [Status and brief description]

---

## 🎯 Testing Objective

[Clear description of what is being tested and why]

This includes:
- [Feature 1]
- [Feature 2]
- [Feature 3]

---

## 📋 Test Environment

**URL**: [Test environment URL]
**Backend Version**: [Version]
**Frontend Version**: [Version]
**Docker Image**: [Image tag if applicable]
**Browser**: [Browser details]

### What's New in [Version]

#### Frontend Components:
1. **[Component Name]** ✅
   - [Feature 1]
   - [Feature 2]

#### Backend Changes:
- [Change 1]
- [Change 2]

---

## 🧪 Test Cases

### Test [Number]: [Test Name] ✅/❌

**Purpose**: [What this test validates]

**Steps**:
1. [Step 1]
2. [Step 2]
3. [Step 3]

**Expected Results**:
- ✅ [Expected result 1]
- ✅ [Expected result 2]
- ✅ [Expected result 3]

**Actual Results**:
```
[Detailed description of what actually happened]
- [Result 1]: [Status and details]
- [Result 2]: [Status and details]
```

**Status**: ✅ PASSED / ❌ FAILED / ⚠️ SKIPPED

**Agent Reviewer Notes**:
```
[SPACE FOR AGENT TO ADD FINDINGS]
- [Finding 1]
- [Finding 2]
- [Screenshots or evidence if applicable]
```

---

[Repeat for each test case]

---

## 📊 Test Results Summary

### Overall Statistics
- **Total Tests**: XX
- **Passed**: XX (XX%)
- **Failed**: XX (XX%)
- **Skipped**: XX (XX%)

### Tests by Category
| Category | Passed | Failed | Skipped | Total |
|----------|--------|--------|---------|-------|
| [Category 1] | X | X | X | X |
| [Category 2] | X | X | X | X |

### Critical Issues
1. **[Issue Title]** - [Severity: Critical/High/Medium/Low]
   - **Impact**: [Description]
   - **Workaround**: [If available]
   - **Fix Required**: [Yes/No]

### Known Limitations
- [Limitation 1]
- [Limitation 2]

---

## 🔧 Troubleshooting

### Common Issues

**Issue 1: [Problem Description]**
- **Symptom**: [What you see]
- **Cause**: [Why it happens]
- **Solution**: [How to fix]

**Issue 2: [Problem Description]**
- **Symptom**: [What you see]
- **Cause**: [Why it happens]
- **Solution**: [How to fix]

---

## 📝 Test Execution Log

### Run 1: [Date]
- **Tester**: [Name]
- **Duration**: XX minutes
- **Result**: XX/XX passed
- **Notes**: [Any observations]

### Run 2: [Date]
- **Tester**: [Name]
- **Duration**: XX minutes
- **Result**: XX/XX passed
- **Notes**: [Any observations]

---

## ✅ Sign-Off

**Test Completion Date**: YYYY-MM-DD
**Tested By**: [Name]
**Approved By**: [Name]
**Status**: ✅ Ready for Production / ❌ Needs Fixes / ⚠️ Conditional Approval

**Reviewer Comments**:
[Final comments from reviewer]

---

**Last Updated**: YYYY-MM-DD
**Protocol Version**: 1.0
```

---

## 🎨 Formatting Guidelines

### Status Indicators
- ✅ **PASSED** - Test completed successfully
- ❌ **FAILED** - Test failed, issue found
- ⚠️ **SKIPPED** - Test not run (blocked or not applicable)
- ⏳ **PENDING** - Test not yet executed

### Severity Levels
- 🔴 **Critical** - Blocks release, must fix
- 🟠 **High** - Significant impact, should fix
- 🟡 **Medium** - Moderate impact, nice to fix
- 🟢 **Low** - Minor issue, can defer

### Section Emojis
- 🔍 Key Findings
- 🎯 Testing Objective
- 📋 Test Environment
- 🧪 Test Cases
- 📊 Results Summary
- 🔧 Troubleshooting
- 📝 Execution Log
- ✅ Sign-Off

---

## 📐 Agent Reviewer Section

**IMPORTANT:** Every test case MUST include an "Agent Reviewer Notes" section where the testing agent can add:

1. **Detailed Findings**
   - Specific observations
   - Edge cases discovered
   - Performance notes

2. **Evidence**
   - Console logs
   - Network requests
   - Error messages
   - Screenshots (if applicable)

3. **Recommendations**
   - Suggested fixes
   - Improvements
   - Follow-up tests needed

**Example:**
```markdown
**Agent Reviewer Notes**:
```
✅ Test passed successfully
- Login form validation works correctly
- Password strength indicator shows appropriate colors
- Email validation provides real-time feedback
- Console: No errors logged
- Network: POST /auth/login returned 200 OK
- Response time: 245ms

⚠️ Minor observation:
- Password field could benefit from show/hide toggle
- Consider adding "Remember me" persistence across sessions

📸 Evidence:
- Console clean, no warnings
- Network tab shows proper Authorization header
```
```

---

## 🎯 Test Protocol Checklist

Before finalizing a test protocol, verify:

- [ ] File saved in `_tests/XX-test-suite-name/test-protocol.md`
- [ ] All required sections present
- [ ] Test cases numbered sequentially
- [ ] Each test has Purpose, Steps, Expected, Actual, Status
- [ ] Agent Reviewer Notes section in every test
- [ ] Status indicators used consistently (✅ ❌ ⚠️)
- [ ] Summary statistics calculated correctly
- [ ] Critical issues clearly documented
- [ ] Troubleshooting section included
- [ ] Test execution log started
- [ ] Sign-off section ready for approval

---

## 📚 Reference Example

See `_tests/_plans/done/[testing]frontend-phase3-ui-components-v4.0.40.md` for a complete example of a well-structured test protocol.

---

## 🚀 Quick Start for Agents

When asked to create a test protocol:

1. **Determine test suite number** (next available: 02-, 03-, etc.)
2. **Create folder structure**: `_tests/XX-test-suite-name/`
3. **Copy this template** and fill in all sections
4. **Write test cases** based on feature requirements
5. **Include Agent Reviewer Notes** in each test case
6. **Save as** `test-protocol.md` in the test suite folder
7. **Update** `_tests/README.md` with new test suite info

---

## 💡 Tips for Writing Good Test Cases

### Do's ✅
- Write clear, actionable steps
- Include specific expected values
- Document actual results in detail
- Use consistent formatting
- Add context in Agent Reviewer Notes
- Include edge cases
- Test error scenarios

### Don'ts ❌
- Don't skip Agent Reviewer Notes sections
- Don't use vague descriptions
- Don't assume prior knowledge
- Don't skip documentation of failures
- Don't forget to update status indicators
- Don't omit troubleshooting steps

---

**Created**: 2025-10-18
**Purpose**: Standardize test protocol creation across all test suites
**Maintained By**: Development Team
