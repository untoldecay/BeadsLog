# Devlog List Command

## Overview

The `devlog list` command provides flexible listing and filtering of devlog entries. It can read from `index.md` files or query session events from the issues database.

## Features

- **Multiple Data Sources**:
  - Reads from `index.md` markdown files
  - Falls back to session events from `.beads/issues.jsonl`
  - Automatically detects and uses available data source

- **Type Filtering**:
  - Filter entries by type using `--type` flag
  - Searches in titles, descriptions, and entities
  - Case-insensitive matching

- **Flexible Output Formats**:
  - `table`: Human-readable format matching index.md structure
  - `json`: Machine-readable JSON output for scripting

- **Limit Results**:
  - Use `--limit` to restrict output to N entries
  - Useful for getting recent entries or summaries

## Installation

The devlog command is part of the beads project. Build it with:

```bash
cd cmd/devlog
go build -o devlog .
```

Or install it system-wide:

```bash
go install ./cmd/devlog
```

## Usage

### Basic Usage

List all devlog entries from the default `index.md`:

```bash
devlog list
```

### Filter by Type

List entries containing "authentication":

```bash
devlog list --type authentication
# or
devlog list -t authentication
```

List session events only:

```bash
devlog list --type session
```

### Limit Results

Show only the 5 most recent entries:

```bash
devlog list --limit 5
# or
devlog list -l 5
```

### Custom Index Path

Use a custom index.md file:

```bash
devlog list --index /path/to/index.md
# or
devlog list -i /path/to/index.md
```

### JSON Output

Get output in JSON format for scripting:

```bash
devlog list --format json
# or
devlog list -f json
```

Combined with filters:

```bash
devlog list --type bug --format json --limit 10
```

## Output Format

### Table Format (default)

The table format matches the original index.md structure:

```
# Devlog

## 2024-01-15 - Implemented user authentication
Added JWT-based authentication to the API.
Users can now login with email/password and receive tokens.

Entities: JWT, API, TODO

## 2024-01-16 - Fixed database connection bug
Fixed issue where connections were not being properly closed.
This was causing memory leaks in production.

ID: bd-123 | Status: closed | Created: 2024-01-16 10:30:00
```

### JSON Format

JSON output includes full entry details:

```json
[
  {
    "date": "2024-01-15",
    "title": "Implemented user authentication",
    "description": "Added JWT-based authentication...",
    "entities": ["JWT", "API", "TODO"],
    "timestamp": "2024-01-15T00:00:00Z",
    "line_number": 3
  }
]
```

## How It Works

### Reading from index.md

The command first tries to read from the specified `index.md` file (default: `./index.md`). It:

1. Parses the markdown file
2. Extracts entries in `## YYYY-MM-DD - Title` format
3. Detects entities (CamelCase, kebab-case, keywords, issue IDs)
4. Applies filters and limits
5. Formats output

### Reading from Sessions

If no `index.md` file is found, the command:

1. Reads `.beads/issues.jsonl`
2. Filters for `issue_type: "event"` (session events)
3. Applies type filtering based on title/description matching
4. Sorts by creation date (newest first)
5. Formats output with session metadata

## Type Filtering

The `--type` flag performs case-insensitive matching against:

- Entry titles
- Entry descriptions
- Detected entities

This makes it flexible for finding entries by:

- Feature name: `--type authentication`
- Bug type: `--type bug`
- Entity name: `--type MyFunction`
- Issue ID: `--type bd-123`
- Keywords: `--type TODO`

## Examples

### Find all bug-related entries

```bash
devlog list --type bug
```

### Get last 10 entries as JSON

```bash
devlog list --limit 10 --format json
```

### Find authentication-related work

```bash
devlog list --type authentication --format table
```

### List session events from issues database

```bash
# When no index.md exists, automatically queries sessions
cd /path/to/beads/project
devlog list
```

### Combine filters

```bash
# Find recent TODO items
devlog list --type TODO --limit 5

# Get session events in JSON
devlog list --type session --format json
```

## Integration with Beads

The devlog list command integrates with the beads issue tracker:

1. **Session Events**: Automatically detects and lists session events from issues.jsonl
2. **Issue References**: Recognizes issue IDs (e.g., `bd-123`) as entities
3. **Cross-Reference**: Links devlog entries to tracked issues

## Testing

Run the test suite:

```bash
cd cmd/devlog
go test -v
```

Run with coverage:

```bash
go test -cover
```

## Implementation Details

### File Structure

- `cmd/devlog/list.go`: Main list command implementation
- `cmd/devlog/list_test.go`: Test suite
- `cmd/devlog/import-md.go`: Index.md parser (shared with graph command)
- `cmd/devlog/main.go`: CLI entry point

### Key Functions

- `runList()`: Main command execution
- `parseIndexMD()`: Parse index.md files (from import-md.go)
- `filterRowsByType()`: Filter entries by type
- `outputTable()`: Format output as table
- `outputJSON()`: Format output as JSON
- `listFromSessions()`: Query sessions from issues.jsonl
- `readIssuesJSONL()`: Read issues database

### Data Flow

```
User Input (flags, args)
    ↓
runList()
    ↓
Try index.md → Parse → Filter → Limit → Format
    ↓ (if fails)
Try issues.jsonl → Query → Filter → Sort → Format
    ↓
Output (table or JSON)
```

## Future Enhancements

Potential improvements for future versions:

1. **Date Range Filtering**: `--after` and `--before` flags
2. **Entity Filtering**: `--entity` flag for specific entities
3. **Sort Options**: `--sort date|title|type`
4. **Tag Support**: Extract and filter by tags
5. **Export**: Write filtered results to new index.md
6. **Statistics**: Show summary stats (total entries, by type, etc.)
7. **Interactive**: Interactive filtering with prompts

## Troubleshooting

### No entries found

If you see "No entries found":

1. Check that `index.md` exists at the specified path
2. Verify the file has entries in `## YYYY-MM-DD - Title` format
3. Try without filters to see all entries
4. Check if issues.jsonl exists if relying on session data

### Filter not working

If type filtering isn't matching:

1. Try a more generic filter term
2. Remember filtering is case-insensitive
3. Check that the term appears in title, description, or entities
4. Use JSON output to see all available fields

### Cannot read issues.jsonl

If sessions aren't loading:

1. Verify you're in a beads project directory
2. Check that `.beads/issues.jsonl` exists
3. Ensure the file is readable
4. Try with `--index` flag to specify index.md directly

## Contributing

To add features or fix bugs:

1. Modify `list.go` for functionality
2. Add tests to `list_test.go`
3. Update this documentation
4. Run tests: `go test -v`
5. Submit pull request

## License

Part of the beads project. See main project LICENSE file.
