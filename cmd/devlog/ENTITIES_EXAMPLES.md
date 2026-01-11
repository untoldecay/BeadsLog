# Entities Command - Examples and Usage

## Overview
The `entities` command lists all entities from your devlog index.md file, sorted by mention count with statistics.

## Command Syntax
```bash
devlog entities [path/to/index.md] [flags]
```

## Features
- **Entity Type Detection**: Automatically categorizes entities as:
  - CamelCase (e.g., MyFunction, UserService)
  - kebab-case (e.g., my-function, user-service)
  - Keywords (e.g., TODO, FIXME, NOTE)
  - Issue IDs (e.g., bd-123)

- **Statistics Provided**:
  - Mention count (frequency)
  - First seen date
  - Last seen date
  - Contexts (entries where entity appears)

- **Sorting**: Entities are sorted by mention count (descending) by default

## Examples

### 1. Show all entities
```bash
devlog entities
```
Output:
```
📊 Entity Statistics Report

Total Entities: 8
Total Mentions: 12

Breakdown by Type:
  CamelCase: 4
  kebab-case: 2
  keyword: 1
  issue-id: 1

Top Entities (by mention count):

  Entity           Type         Mentions  First Seen   Last Seen    Contexts
  ---------------  -----------  -------  -----------  -----------  --------------------------------------------------
  UserService      CamelCase           3  2024-01-15   2024-01-17   [3] 2024-01-15: Implemented user authentication (+2 more)
  MyFunction       CamelCase           2  2024-01-17   2024-01-18   [2] 2024-01-17: Added unit tests for UserService
  JWT              CamelCase           1  2024-01-15   2024-01-15   [1] 2024-01-15: Implemented user authentication
  ...
```

### 2. Show only CamelCase entities
```bash
devlog entities --type CamelCase
```

### 3. Show entities mentioned at least 3 times
```bash
devlog entities --min 3
```

### 4. Show top 10 entities
```bash
devlog entities --limit 10
```

### 5. Output in JSON format
```bash
devlog entities --format json
```
Output:
```json
{
  "total_entities": 8,
  "total_mentions": 12,
  "by_type": {
    "CamelCase": 4,
    "kebab-case": 2,
    "keyword": 1,
    "issue-id": 1
  },
  "sorted_by": "mention_count",
  "entities": [
    {
      "name": "UserService",
      "type": "CamelCase",
      "mention_count": 3,
      "first_seen": "2024-01-15",
      "last_seen": "2024-01-17",
      "contexts": [
        "2024-01-15: Implemented user authentication",
        "2024-01-16: Fixed database bug",
        "2024-01-17: Added unit tests"
      ]
    },
    ...
  ]
}
```

### 6. Combine multiple filters
```bash
devlog entities --type CamelCase --min 2 --limit 5
```

### 7. Use a custom index file
```bash
devlog entities /path/to/custom/index.md
```

### 8. Show help
```bash
devlog entities --help
```

## Flags
- `-f, --format <type>`: Output format (table or json) [default: table]
- `-t, --type <type>`: Filter by entity type (CamelCase, kebab-case, keyword, issue-id)
- `-l, --limit <n>`: Limit number of entities shown (0 = unlimited)
- `-m, --min <n>`: Minimum mention count to include [default: 1]

## Use Cases

### Track Most Referenced Components
```bash
devlog entities --type CamelCase --limit 10
```
Shows your top 10 most mentioned CamelCase entities, helping you identify core components.

### Find TODOs and FIXMEs
```bash
devlog entities --type keyword
```
Lists all keyword mentions like TODO, FIXME, NOTE.

### Identify Active Issues
```bash
devlog entities --type issue-id --min 2
```
Shows issue IDs mentioned multiple times, indicating active work items.

### Export for Analysis
```bash
devlog entities --format json > entities-report.json
```
Export entity statistics for further analysis or visualization.

## Implementation Details

### Entity Type Detection
- **CamelCase**: Starts with uppercase, contains mixed case (e.g., MyFunction)
- **kebab-case**: Lowercase words with hyphens (e.g., my-function)
- **Keywords**: Common markers (TODO, FIXME, NOTE, HACK, XXX, BUG, OPTIMIZE, REFACTOR)
- **Issue IDs**: Matches bd-XXX pattern (case insensitive)

### Statistics Calculation
- **Mention Count**: Number of entries where the entity appears
- **First Seen**: Earliest date of entity mention
- **Last Seen**: Most recent date of entity mention
- **Contexts**: List of entry titles/dates where entity appears

### Sorting
Primary sort: Mention count (descending)
Secondary sort: Entity name (ascending)

## Testing
Run the test suite:
```bash
cd /projects/devlog/cmd/devlog
go test -v -run TestEntities
```

## Related Commands
- `devlog list`: List all devlog entries
- `devlog graph`: Show entity relationship graph
- `devlog search`: Search across entries
