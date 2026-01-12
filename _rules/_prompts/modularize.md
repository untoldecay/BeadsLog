# File Modularization Guide

Create a plan to break down a large file into smaller, function-based modules to improve maintainability. Save output to `_rules/_features/{filename}_modularization.md`.

> **Modularization Threshold**: Files exceeding 1000 lines should be considered candidates for modularization to improve maintainability and reduce the risk of errors during modifications.

## Input
- **Target File**: File to modularize
- **Current Issues**: Problems with the large file
- **Line Count**: Total lines in the file (files > 1000 lines are prime candidates)

## Output Structure

### 1. File Analysis
- List all functions/methods
- Group related functions
- Identify dependencies between functions
- Note shared state/variables
- Highlight particularly large functions (> 100 lines)

### 2. Proposed File Structure
```
src/
├── components/
│   ├── ComponentName/
│   │   ├── index.js        # Main entry point
│   │   ├── functionGroup1.js
│   │   ├── functionGroup2.js
│   │   └── ...
```

### 3. Extraction Steps
For each new file:
- Functions to include
- Required imports
- Exports
- Shared state handling

### 4. Implementation Guide
```js
// Example: functionGroup1.js
import { dependency } from './otherModule.js';

export function function1() {
  // Implementation
}

export function function2() {
  // Implementation
}
```

### 5. Main File Refactoring
```js
// Example: index.js
import { function1, function2 } from './functionGroup1.js';
import { function3, function4 } from './functionGroup2.js';

export default {
  function1,
  function2,
  function3,
  function4
};
```

### 6. Testing Strategy
- Test each module independently
- Test integration points
- Verify identical behavior

## Guidelines
- One responsibility per file
- Clear imports/exports
- Minimize circular dependencies
- Add debug logs at module boundaries
- Document each file's purpose
- Use consistent naming
- Target 300-500 lines maximum per file
- Keep functions under 100 lines when possible

## Implementation Process
1. Create new directory structure
2. Extract one module at a time
3. Update imports/exports
4. Test after each extraction
5. Refactor main file last
