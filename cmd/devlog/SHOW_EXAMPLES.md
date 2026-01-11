# Devlog Show Command - Examples

This document demonstrates the usage of the `devlog show` command.

## Overview

The `devlog show` command displays full devlog entry content by date or filename.

## Usage

```bash
devlog show [date|filename]
```

## Examples

### Example 1: Show Entry by Date

```bash
$ devlog show 2024-01-15
```

**Output:**
```
## 2024-01-15 - Implemented user authentication

Added JWT-based authentication to the API.
Users can now login with email/password and receive tokens.
TODO: Add refresh token support.

**Entities:** JWT, API, TODO

---
Date: 2024-01-15
Line: 3
```

### Example 2: Show Entry by Date (with index flag)

```bash
$ devlog show 2024-01-16 --index ./test-index.md
```

**Output:**
```
## 2024-01-16 - Fixed database connection bug

Fixed issue where connections were not being properly closed.
This was causing memory leaks in production.

**Entities:** bd-123, database

---
Date: 2024-01-16
Line: 8
```

### Example 3: Show Entry from Filename

```bash
$ devlog show 2024-01-15.md
```

**Output:**
```
# 2024-01-15.md

# My Feature Implementation

Detailed notes about the feature implementation...

## Code Changes

- Modified file1.go
- Added file2.go

## Testing

Created unit tests for all new functionality.
```

### Example 4: Show Entry from Relative Path

```bash
$ devlog show entries/my-feature.md
```

**Output:**
```
# entries/my-feature.md

# My Feature

Feature implementation details...
```

### Example 5: Error - No Argument Provided

```bash
$ devlog show
```

**Output:**
```
Error: requires a date or filename argument

Usage: devlog show [date|filename]
```

### Example 6: Error - Date Not Found

```bash
$ devlog show 2024-12-31
```

**Output:**
```
Error: no entry found for date: 2024-12-31
```

### Example 7: Error - File Not Found

```bash
$ devlog show nonexistent.md
```

**Output:**
```
Error: failed to read file nonexistent.md: open nonexistent.md: no such file or directory
```

## Features

### 1. Date Format Detection

The command automatically detects if the input is a date (YYYY-MM-DD format) or a filename:

- `2024-01-15` → Treated as date
- `2024-01-15.md` → Treated as filename
- `my-feature.md` → Treated as filename

### 2. Automatic .md Extension

If you provide a filename without the `.md` extension, it will be added automatically:

```bash
devlog show my-feature  # Equivalent to: devlog show my-feature.md
```

### 3. Path Resolution

The command looks for files in:
1. Current directory (relative path)
2. Directory containing the index.md file

```bash
# If index.md is at ./devlog/index.md
devlog show entries/feature.md  # Looks for ./devlog/entries/feature.md
```

### 4. Entity Display

When showing entries by date, detected entities are displayed:

```bash
$ devlog show 2024-01-17
```

**Output:**
```
## 2024-01-17 - Added unit tests for UserService

Wrote comprehensive tests for user CRUD operations.
Coverage now at 85% for UserService.

**Entities:** UserService, CRUD, MyFunction

---
Date: 2024-01-17
Line: 13
```

## Use Cases

### 1. Review Daily Work

```bash
# Show what you worked on yesterday
devlog show $(date -d "yesterday" +%Y-%m-%d)
```

### 2. View Specific Entry

```bash
# View a specific feature's notes
devlog show feature-authentication.md
```

### 3. Check Session Details

```bash
# View session notes for a specific date
devlog show 2024-01-19
```

### 4. Read External Files

```bash
# Read linked markdown files
devlog show entries/2024-01-15-bugfix.md
```

## Integration with Other Commands

### With `devlog list`

```bash
# List entries, then show details
devlog list --limit 5
devlog show 2024-01-15
```

### With `devlog graph`

```bash
# Show graph, then read specific entry
devlog graph
devlog show authentication-flow.md
```

## Implementation Details

### Date Recognition

Uses regex pattern: `^\d{4}-\d{2}-\d{2}$`

Matches:
- ✓ `2024-01-15`
- ✗ `2024-1-15`
- ✗ `2024/01/15`
- ✗ `01-15-2024`

### File Reading

- Reads entire file content
- Displays raw markdown
- Preserves formatting

### Index Parsing

- Uses existing `parseIndexMD()` function
- Matches date field exactly
- Returns all matching entries (if multiple exist for same date)

## Comparison: devlog list vs devlog show

| Feature | devlog list | devlog show |
|---------|-------------|-------------|
| Purpose | List multiple entries | Show single entry |
| Input | Filters (--type, --limit) | Date or filename |
| Output | Summary view | Full content |
| File content | No | Yes (for filename mode) |
| Best for | Browsing | Detailed review |

## Tips

1. **Use tab completion** for dates if your shell supports it
2. **Create aliases** for frequently viewed entries:
   ```bash
   alias devlog-today='devlog show $(date +%Y-%m-%d)'
   alias devlog-yesterday='devlog show $(date -d "yesterday" +%Y-%m-%d)'
   ```
3. **Combine with less** for long entries:
   ```bash
   devlog show long-entry.md | less
   ```
