# File Modularization with Facade Pattern

This prompt guides the systematic restructuring of monolithic files into logical, maintainable modules using the Facade pattern while maintaining backward compatibility.
Please create a feature .md file plan in '_rules/_features' for the modularization process in which you will update the modularization ETA.

> **Modularization Threshold**: Files exceeding 1000 lines should be considered candidates for modularization to improve maintainability and reduce the risk of errors during modifications.

## Input
- **Target File/Folder**: File or folder to modularize
- **Current Issues**: Problems with the current structure
- **Line Count**: Total lines in the file(s) being modularized

## ⚠️ IMPORTANT ALERT

> **DO NOT WRITE NEW CODE DURING MODULARIZATION UNLESS ABSOLUTELY NECESSARY**
>
> When modularizing, only:
>
> 1. Move existing code to appropriate module files
> 2. Add necessary import/export statements
> 3. Update references between functions
>
> Adding new functionality during modularization significantly increases the risk of introducing bugs and makes it harder to identify the source of any issues that arise. Keep the refactoring phase separate from feature development.

## Step 1: Analysis Process

1. **Identify Functional Categories**
   * Review the provided file/folder and categorize functions by purpose
   * Note line ranges for each category 
   * Document particularly complex or large functions (> 100 lines)

2. **Create Dependency Map**
   * Document cross-references between functions
   * Identify shared utilities and common functionality
   * Detect potential circular dependencies
   * Note shared state or variables

## Step 2: Module Design

1. **Define Module Structure**
   * Create logical groupings based on functionality
   * Plan new file structure with clear separation of concerns
   * Keep each module focused on a single responsibility
   * Target 300-500 lines maximum per file

2. **Proposed File Structure**
   ```
   src/
   ├── components/
   │   ├── ComponentName/
   │   │   ├── index.js        # Main entry point (facade)
   │   │   ├── moduleA.js      # Category A functions
   │   │   ├── moduleB.js      # Category B functions
   │   │   └── ...
   ```

3. **Facade Pattern Implementation**
   * Original file becomes a facade that re-exports all functions
   * Group exports by category for readability
   * Maintain backward compatibility for all public APIs
   * Document the purpose of each module

4. **Module Structure Pattern**

   ```javascript
   // Import dependencies
   import { dependencyA, dependencyB } from './other-module.js';
   
   /**
    * Function documentation
    * @param {Type} param - Description
    */
   export function functionName(param) {
     try {
       // Implementation with debug logs
       console.log('Function executing with:', param); // Debug log
       
       // Logic
       
       // Return value if needed
     } catch (error) {
       console.error('Error in functionName:', error);
     }
   }
   ```

## Step 3: Implementation Plan

1. **Create Module Files**
   * Define file names based on functional categories
   * Add descriptive header comments to each file
   * Create consistent directory structure

2. **Extract Functions**
   * Move related functions to appropriate module files
   * Maintain original function signatures
   * Add proper imports for dependencies
   * Add debug logs at function entry/exit points
   * Ensure comprehensive error handling with try/catch

3. **Update Facade**
   * Import all module functions in the original file
   * Re-export everything with explicit named exports
   * Add clear section comments for each category
   * Document any interface changes

4. **Main File Refactoring Example**
   ```javascript
   // Example: index.js (facade)
   import { functionA1, functionA2 } from './moduleA.js';
   import { functionB1, functionB2 } from './moduleB.js';
   
   // Category A exports
   export { functionA1, functionA2 };
   
   // Category B exports
   export { functionB1, functionB2 };
   ```

## Step 4: Technical Considerations

1. **Regular Expression Handling**
   * Be cautious with escape sequences in regex when moving code
   * Verify all regex patterns work the same way

2. **Circular Dependencies**
   * Avoid circular imports by using utility modules for shared code
   * Consider dependency injection for complex circular references

3. **Error Handling**
   * Every function should have try/catch blocks
   * Add descriptive error messages with context
   * Use consistent logging approach

4. **State Management**
   * Carefully handle shared state across modules
   * Document where state is managed and accessed
   * Consider options for reducing shared state

## Step 5: Testing Strategy

1. **Verification Approach**
   * Compile time testing (resolve all import/export errors)
   * Runtime testing of core functionality
   * Compare behavior before and after modularization
   * Look for edge cases specific to the modularized code

2. **Test Categories**
   * Unit tests for each module
   * Integration tests for module interactions
   * Performance tests if applicable
   * Edge case testing for complex functionality
   * Verify identical behavior before and after modularization

## Guidelines
- One responsibility per file
- Clear imports/exports
- Minimize circular dependencies
- Add debug logs at module boundaries
- Document each file's purpose
- Use consistent naming conventions
- Target 300-500 lines maximum per file
- Keep functions under 100 lines when possible

## Implementation Process
1. Create new directory structure
2. Extract one module at a time
3. Update imports/exports
4. Test after each extraction
5. Refactor main file last

## Implementation Plan Template

```
## Implementation Steps

1. **Create Module Files**
   - [moduleA].js - [Purpose description]
   - [moduleB].js - [Purpose description]
   - [moduleC].js - [Purpose description]

2. **Extract Functions by Category**
   - Category A → [moduleA].js
     - functionA1, functionA2, functionA3
   - Category B → [moduleB].js
     - functionB1, functionB2
   - Category C → [moduleC].js
     - functionC1, functionC2, functionC3

3. **Update Imports/Exports**
   - Import shared dependencies in each module
   - Export all functions from their respective modules
   - Import and re-export all functions in the facade

4. **Testing Prioritization**
   - Test core functionality: [list key functions]
   - Test error handling: [list error scenarios]
   - Test edge cases: [list edge cases]
```

## Time Estimation Template

- Analysis: [X] hours
- Module creation: [X] hours
- Testing: [X] hours
- Debugging: [X] hours
- Total estimated time: [X] hours
