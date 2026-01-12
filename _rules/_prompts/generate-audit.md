# Generate Audit Analysis Prompt

## 🎯 **Purpose**

This prompt generates comprehensive audit analyses with a structured, step-by-step approach that enables clear separation of concerns and provides reusable results for users.

## 📋 **Usage**

```
CLI Command: /generate-audit <audit-theme> [options]
```

## 🏗️ **Audit Generation Workflow**

### **Step 1: Analysis Setup**
- Parse audit theme and context
- Create dedicated analysis folder: `_rules/_analysis/<audit-theme>-<YYYY-MM-DD>/`
- Generate README.md with audit scope and goals
- Define analysis methodology and success criteria

### **Step 2: Backend Analysis**
- Scan relevant backend files (server.js, routes, services, etc.)
- Extract all endpoints with methods, paths, authentication, middleware
- Categorize endpoints by functionality
- Document response formats, rate limiting, security measures
- Create `backend-endpoints-audit.md`

### **Step 3: Frontend Analysis**
- Scan frontend files (api clients, components, hooks, etc.)
- Extract all API calls with methods, paths, authentication
- Identify patterns (ApiClient vs direct fetch)
- Document error handling, parameter usage
- Create `frontend-endpoints-usage.md`

### **Step 4: Confrontation Analysis**
- Compare backend endpoints vs frontend usage
- Identify mismatches, gaps, and alignment issues
- Create detailed mapping with status indicators
- Categorize issues by severity (Critical, High, Medium, Low)
- Create `endpoint-confrontation-analysis.md`

### **Step 5: Findings & Recommendations**
- Generate executive summary with business impact
- Create prioritized implementation roadmap
- Define quality gates and success criteria
- Provide concrete code examples where applicable
- Create `endpoint-alignment-findings-and-recommendations.md`

## 📁 **Output Structure**

```
_rules/_analysis/<audit-theme>-<YYYY-MM-DD>/
├── README.md                                      # Audit overview and goals
├── backend-endpoints-audit.md                     # Backend endpoints inventory
├── frontend-endpoints-usage.md                    # Frontend API calls analysis
├── endpoint-confrontation-analysis.md             # Alignment comparison
└── endpoint-alignment-findings-and-recommendations.md # Executive summary
```

## 🎨 **File Generation Guidelines**

### **README.md** (Required)
- Audit date and scope
- Concise audit goal paragraph
- Tree view of analysis files
- Keep it minimal and focused

### **Backend Audit File**
- Complete endpoint inventory with categorization
- Include authentication, middleware, rate limiting details
- Format: Clean lists, clear descriptions
- Provide line numbers/code references

### **Frontend Usage File**
- All API calls from frontend codebase
- Patterns and usage analysis
- Authentication handling methods
- Error handling approaches

### **Confrontation Analysis**
- Detailed comparison tables
- Status indicators (✅ Aligned, ❌ Missing, ⚠️ Partial)
- Specific issue identification
- Severity classification

### **Findings Report**
- Executive summary
- Prioritized recommendations
- Implementation roadmap
- Quality gates and success criteria

## 🔧 **Flexibility Guidelines**

### **Customization Allowances**
- **Additional Analysis Files**: Can add specialized files based on audit theme
- **Modified Structure**: Can adapt structure for specific audit needs
- **Custom Categories**: Can create theme-specific endpoint categories
- **Enhanced Metrics**: Can add theme-specific success criteria

### **Core Requirements** (Must Keep)
- README.md with audit scope and goals
- Step-by-step workflow adherence
- Separation of concerns principle
- Tree view file organization
- Clear status indicators and categorization

### **Content Generation Rules**
- Use current context from conversation
- Adapt examples to match codebase specifics
- Include relevant file paths and line numbers
- Provide actionable, concrete recommendations
- Maintain consistent formatting across files

## 🎯 **Generation Principles**

1. **Context-Aware**: Use current conversation context and codebase state
2. **Structured**: Follow the step-by-step workflow methodically
3. **Actionable**: Generate practical, implementable recommendations
4. **Reusable**: Create results users can easily copy/paste for their needs
5. **Traceable**: Include specific references to files and code locations

## 📝 **Template Variables**

The prompt should dynamically substitute these variables:
- `{audit-theme}`: User-provided audit theme
- `{YYYY-MM-DD}`: Current date
- `{context-relevant-files}`: Files relevant to audit theme
- `{custom-requirements}`: Any user-specified custom needs

---

**Usage Example**: `/generate-audit authentication-security`
**Output**: Creates authentication-specific audit with security focus while maintaining core structure and principles.