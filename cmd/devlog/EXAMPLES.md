# Devlog List Command - Examples

This document shows example outputs from the `devlog list` command.

## Example 1: Basic List Output (Table Format)

```bash
$ devlog list
```

**Output:**
```
# Devlog

## 2024-01-19 - Session: Feature implementation sprint
Completed sprint for new feature implementation.
Implemented 5 new features including user dashboard.
Closed 3 related bugs.

ID: bd-abc123 | Status: closed | Created: 2024-01-19 17:30:00
Created by: beads/crew/grip

## 2024-01-17 - Added unit tests for UserService
Wrote comprehensive tests for user CRUD operations.
Coverage now at 85% for UserService.

Entities: UserService, CRUD, MyFunction
```

## Example 2: Filter by Type

```bash
$ devlog list --type authentication
```

**Output:**
```
# Devlog

## 2024-01-15 - Implemented user authentication
Added JWT-based authentication to the API.
Users can now login with email/password and receive tokens.

Entities: JWT, API, TODO
```

## Example 3: Limit Results

```bash
$ devlog list --limit 2
```

**Output:**
```
# Devlog

## 2024-01-19 - Session: Feature implementation sprint
Completed sprint for new feature implementation.

ID: bd-abc123 | Status: closed | Created: 2024-01-19 17:30:00

## 2024-01-17 - Added unit tests for UserService
Wrote comprehensive tests for user CRUD operations.

Entities: UserService, CRUD
```

## Example 4: JSON Format

```bash
$ devlog list --format json --limit 1
```

**Output:**
```json
[
  {
    "date": "2024-01-19",
    "title": "Session: Feature implementation sprint",
    "description": "Completed sprint for new feature implementation.\nImplemented 5 new features including user dashboard.",
    "entities": ["sprint", "feature", "dashboard"],
    "timestamp": "2024-01-19T00:00:00Z",
    "line_number": 3
  }
]
```

## Example 5: Session Events from Issues Database

```bash
# When no index.md exists, queries from .beads/issues.jsonl
$ devlog list --type session
```

**Output:**
```
# Devlog Sessions

## 2026-01-08 - Session ended: gt-beads-crew-wolf

ID: bd-0t8ak | Status: closed | Created: 2026-01-08 14:37:26
Created by: beads/crew/wolf
Reason: auto-closed session event

## 2026-01-07 - Session ended: gt-beads-crew-grip

ID: bd-03ze8 | Status: closed | Created: 2026-01-07 19:20:04
Created by: beads/crew/grip
Reason: auto-closed session event
```

## Example 6: Combined Filters

```bash
$ devlog list --type bug --limit 3 --format table
```

**Output:**
```
# Devlog

## 2024-01-16 - Fixed database connection bug
Fixed issue where connections were not being properly closed.
This was causing memory leaks in production.

Entities: bd-123, database, memory-leak
```

## Example 7: Custom Index Path

```bash
$ devlog list --index /path/to/project/index.md --type feature
```

**Output:**
```
# Devlog

## 2024-01-15 - Implemented user authentication
Added JWT-based authentication to the API.

Entities: JWT, authentication, feature
```

## Example 8: Help Command

```bash
$ devlog list --help
```

**Output:**
```
List devlog entries from index.md or from session events.

Supports filtering by type and various output formats.
The display matches the original index.md structure with dates and titles.

Usage:
  devlog list [flags]

Flags:
  -f, --format string   Output format: table or json (default "table")
  -h, --help            help for list
  -i, --index string    Path to index.md file (default "./index.md")
  -l, --limit int       Maximum number of entries to show (0 = unlimited)
  -t, --type string     Filter by type (e.g., event, feature, bug)
```

## Example 9: Session Event Querying

When working with a beads project that has session events:

```bash
$ cd /path/to/beads/project
$ devlog list
# Automatically queries .beads/issues.jsonl for session events
```

**Output:**
```
# Devlog Sessions

## 2026-01-08 - Session ended: gt-beads-crew-wolf

ID: bd-0t8ak | Status: closed | Created: 2026-01-08 14:37:26
Created by: beads/crew/wolf
Reason: auto-closed session event
```

## Example 10: Entity Detection

The command automatically detects entities in entries:

```bash
$ devlog list
```

**Output shows detected entities:**
```
## 2024-01-15 - Implemented user authentication
Added JWT-based authentication to the API.
MyFunction handles the token validation.

Entities: JWT, API, MyFunction, authentication
```

Detected entities include:
- **CamelCase**: MyFunction, UserService, JWT
- **kebab-case**: my-function, user-auth
- **Keywords**: TODO, FIXME, NOTE
- **Issue IDs**: bd-123, BD-456

## Use Cases

### 1. Daily Standup
```bash
# Show yesterday's work
devlog list --limit 1
```

### 2. Bug Triage
```bash
# Find all bug-related entries
devlog list --type bug
```

### 3. Feature Tracking
```bash
# Track progress on specific feature
devlog list --type authentication --format json | jq .
```

### 4. Code Review
```bash
# See what was worked on recently
devlog list --limit 10
```

### 5. Session Audit
```bash
# Review all agent sessions
devlog list --type session
```

### 6. TODO Tracking
```bash
# Find all TODO items
devlog list --type TODO
```

## Performance Notes

- **index.md parsing**: Fast, reads entire file into memory
- **issues.jsonl querying**: Reads entire file, filters in-memory
- **Filtering**: Case-insensitive substring matching
- **Limiting**: Applied after filtering, for efficiency

## Integration Examples

### With Git

```bash
# Log devlog entries with git commits
git log --since="1 week ago" --format="%h %s" > recent-commits.txt
devlog list --limit 7 > recent-devlog.txt
```

### With jq (JSON processing)

```bash
# Extract just titles from JSON output
devlog list --format json | jq -r '.[].title'
```

### With grep

```bash
# Further filter output
devlog list | grep -i "urgent"
```

## Error Handling

### File not found
```bash
$ devlog list --index nonexistent.md
Error: no index.md found at nonexistent.md and no issues.jsonl at .beads/issues.jsonl
```

### Invalid format
```bash
$ devlog list --format xml
Error: invalid format: xml (must be 'table' or 'json')
```

### No matching entries
```bash
$ devlog list --type nonexistent-filter
No entries found.
```

## Tips and Tricks

1. **Use aliases** for common queries:
   ```bash
   alias devlog-todo='devlog list --type TODO'
   alias devlog-recent='devlog list --limit 5'
   alias devlog-sessions='devlog list --type session'
   ```

2. **Combine with other tools**:
   ```bash
   devlog list --format json | jq '[.[] | {date, title}]'
   ```

3. **Export to files**:
   ```bash
   devlog list --type feature > features.md
   devlog list --format json > devlog.json
   ```

4. **Use in scripts**:
   ```bash
   #!/bin/bash
   count=$(devlog list --format json | jq 'length')
   echo "Found $count devlog entries"
   ```
