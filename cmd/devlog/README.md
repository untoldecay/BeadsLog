# devlog - Markdown Developer Log CLI

**A powerful CLI tool for parsing and analyzing devlog markdown files with intelligent entity extraction and graph-based relationship tracking.**

devlog transforms your markdown-based developer logs into a queryable knowledge graph, making it easy to track entities, relationships, and patterns across your development work.

## ⚡ Quick Start

```bash
# Install from source
go install github.com/steveyeggie/beads/cmd/devlog@latest

# Or build locally
cd cmd/devlog
go build -o devlog

# Run commands
./devlog list
./devlog graph MyFunction
./devlog search "authentication"
```

## 📋 Table of Contents

- [Features](#-features)
- [Installation](#-installation)
- [Quick Start Guide](#-quick-start-guide)
- [Commands Reference](#-commands-reference)
- [Workflow Examples](#-workflow-examples)
- [Graph Traversal](#-graph-traversal)
- [Entity Types](#-entity-types)
- [Output Formats](#-output-formats)
- [Advanced Usage](#-advanced-usage)
- [Examples](#examples)

## 🛠 Features

- **Markdown Parsing**: Parse `index.md` files with entries in `## YYYY-MM-DD - Title` format
- **Entity Extraction**: Automatically detect and track:
  - CamelCase identifiers (MyFunction, ClassName)
  - kebab-case identifiers (my-function, url-path)
  - Keywords (TODO, FIXME, NOTE, HACK, XXX, BUG)
  - Issue IDs (bd-123, BD-456)
- **Graph Visualization**: Display entity relationship graphs with hierarchical connections
- **Full-Text Search**: Search across entries with graph context and entity relationships
- **Multiple Output Formats**: Table (human-readable) and JSON (machine-readable)
- **Session Tracking**: Import and query session events from beads issues database
- **Impact Analysis**: Show what entities depend on a given entity (reverse graph traversal)
- **Resume Context**: Find matching sessions with full entity graph context for continuity

## 📦 Installation

### From Source

```bash
go install github.com/steveyeggie/beads/cmd/devlog@latest
```

### Local Build

```bash
git clone https://github.com/steveyeggie/beads.git
cd beads/cmd/devlog
go build -o devlog
```

### Requirements

- Go 1.24 or later
- Linux, macOS, or Windows

## 🚀 Quick Start Guide

### 1. Create Your Devlog

Create an `index.md` file in your project directory:

```markdown
# Devlog

## 2024-01-15 - Implemented user authentication
Added JWT-based authentication to the API.
Users can now login with email/password and receive tokens.
MyFunction handles the token validation.
TODO: Add refresh token support.

## 2024-01-16 - Fixed database connection bug
Fixed issue where connections were not being properly closed.
This was causing memory leaks in production.
bd-123 was the tracking issue.

## 2024-01-17 - Added unit tests for UserService
Wrote comprehensive tests for user CRUD operations.
Coverage now at 85% for UserService.
```

### 2. List Your Entries

```bash
./devlog list
```

**Output:**
```
# Devlog

## 2024-01-17 - Added unit tests for UserService
Wrote comprehensive tests for user CRUD operations.
Coverage now at 85% for UserService.

Entities: UserService, CRUD

## 2024-01-16 - Fixed database connection bug
Fixed issue where connections were not being properly closed.
This was causing memory leaks in production.

Entities: bd-123, database, memory-leak

## 2024-01-15 - Implemented user authentication
Added JWT-based authentication to the API.
Users can now login with email/password and receive tokens.
MyFunction handles the token validation.

Entities: JWT, API, MyFunction, TODO
```

### 3. Explore Entity Relationships

```bash
./devlog graph MyFunction
```

**Output:**
```
Entity: MyFunction (CamelCase)
Mentions: 1
Related entries:
  - 2024-01-15: Implemented user authentication
```

### 4. Search Your Devlog

```bash
./devlog search "database"
```

**Output:**
```
Found 1 matching entry:

## 2024-01-16 - Fixed database connection bug
Fixed issue where connections were not being properly closed.
This was causing memory leaks in production.

Entities: bd-123, database, memory-leak
```

## 📖 Commands Reference

### `list` - List Devlog Entries

List all devlog entries with optional filtering by type and limits.

```bash
devlog list [flags]
```

**Flags:**
- `-f, --format string`: Output format: `table` or `json` (default "table")
- `-h, --help`: help for list
- `-i, --index string`: Path to index.md file (default "./index.md")
- `-l, --limit int`: Maximum number of entries to show (0 = unlimited)
- `-t, --type string`: Filter by type (e.g., event, feature, bug)

**Examples:**

```bash
# List all entries
./devlog list

# Filter by type
./devlog list --type fix
./devlog list --type feature
./devlog list --type bug

# Limit output
./devlog list --limit 5

# JSON output
./devlog list --format json

# Custom index file
./devlog list --index /path/to/index.md
```

### `graph` - Display Entity Relationship Graph

Show the relationship graph for a specific entity or list all entities.

```bash
devlog graph [entity] [flags]
```

**Flags:**
- `-d, --depth int`: Graph traversal depth (default 2)
- `-h, --help`: help for graph

**Examples:**

```bash
# Show graph for specific entity
./devlog graph manage-columns
./devlog graph MyFunction
./devlog graph bd-123

# Control traversal depth
./devlog graph --depth 1 manage-columns
./devlog graph --depth 3 MyFunction

# Show all entities with graphs
./devlog graph
```

### `entities` - List All Entities

List all detected entities sorted by mention count.

```bash
devlog entities [flags]
```

**Flags:**
- `-f, --format string`: Output format: `table` or `json` (default "table")
- `-h, --help`: help for entities
- `-l, --limit int`: Maximum number of entities to show (0 = unlimited)
- `-m, --min int`: Minimum mention count (default 1)
- `-t, --type string`: Filter by entity type (CamelCase, kebab-case, keyword, issue-id)

**Examples:**

```bash
# List all entities
./devlog entities

# Filter by type
./devlog entities --type CamelCase
./devlog entities --type kebab-case
./devlog entities --type keyword
./devlog entities --type issue-id

# Minimum mentions
./devlog entities --min 3

# Limit results
./devlog entities --limit 10

# JSON output
./devlog entities --format json
```

### `search` - Full-Text Search

Search across devlog entries with full-text search and optional graph context.

```bash
devlog search [query] [flags]
```

**Flags:**
- `-d, --depth int`: Graph context depth (default 1)
- `-f, --format string`: Output format: `table` or `json` (default "table")
- `-h, --help`: help for search
- `-i, --index string`: Path to index.md file (default "./index.md")
- `-l, --limit int`: Maximum number of results to show (0 = unlimited)
- `-t, --type string`: Filter by type (e.g., event, feature, bug)

**Examples:**

```bash
# Basic search
./devlog search migration
./devlog search authentication
./devlog search bug

# Filter by type
./devlog search "API" --type feature

# Limit results
./devlog search migration --limit 3

# JSON output
./devlog search "database" --format json

# Include graph context
./devlog search "session" --depth 2
```

### `show` - Show Full Entry Content

Display the full content of a specific devlog entry.

```bash
devlog show [date-or-filename] [flags]
```

**Flags:**
- `-h, --help`: help for show
- `-i, --index string`: Path to index.md file (default "./index.md")

**Examples:**

```bash
# Show by date
./devlog show 2024-01-15

# Show by filename
./devlog show 2024-01-15.md
./devlog show entries/my-feature.md

# Custom index
./devlog show 2024-01-15 --index /path/to/index.md
```

### `impact` - Show Entity Dependencies

Show what entities depend on a given entity (reverse graph traversal).

```bash
devlog impact [entity] [flags]
```

**Flags:**
- `-d, --depth int`: Impact traversal depth (default 2)
- `-f, --format string`: Output format: `table` or `json` (default "table")
- `-h, --help`: help for impact

**Examples:**

```bash
# Show impact for entity
./devlog impact MyFunction
./devlog impact database

# Control depth
./devlog impact --depth 3 MyFunction
```

### `resume` - Resume Work with Context

Find matching sessions with full entity graph context for continuity.

```bash
devlog resume [flags]
```

**Flags:**
- `-d, --depth int`: Graph context depth (default 2)
- `-f, --format string`: Output format: `table` or `json` (default "table")
- `-h, --help`: help for resume
- `-q, --query strings`: Search queries to find matching sessions

**Examples:**

```bash
# Resume with recent context
./devlog resume

# Resume with specific queries
./devlog resume --query authentication --query JWT

# Control graph depth
./devlog resume --depth 3
```

### `import-md` - Import Markdown Entries

Import devlog entries from a markdown file into the beads issues database.

```bash
devlog import-md [flags]
```

**Flags:**
- `-f, --file string`: Path to markdown file to import
- `-h, --help`: help for import-md

**Examples:**

```bash
# Import from index.md
./devlog import-md --file index.md

# Import from custom file
./devlog import-md --file /path/to/devlog.md
```

## 🔄 Workflow Examples

### Example 1: Transition from Markdown to Devlog CLI

**Before (Manual Markdown Review):**

You have a devlog in markdown format and want to find all authentication-related work.

```bash
# Old way: manually grep through files
grep -i "auth" index.md
grep -i "JWT" **/*.md
```

**After (With Devlog CLI):**

```bash
# New way: intelligent search with entity tracking
./devlog search "authentication"

# See related entities
./devlog entities --type CamelCase | grep -i auth

# Explore full context
./devlog graph JWT --depth 3
```

### Example 2: Track Feature Development

**Scenario:** Track the development of a user authentication feature.

```bash
# 1. Start with initial implementation
## 2024-01-15 - Implemented user authentication
Added JWT-based authentication. MyFunction handles validation.

# 2. List all authentication work
./devlog list --type authentication

# 3. See entity relationships
./devlog graph JWT

# 4. Check what depends on JWT
./devlog impact JWT

# 5. Search for related work
./devlog search "token" --depth 2
```

### Example 3: Bug Investigation

**Scenario:** Investigate a database connection bug.

```bash
# 1. Find bug-related entries
./devlog search "database"

# 2. Show full entry details
./devlog show 2024-01-16

# 3. Trace related entities
./devlog graph database

# 4. Check impact on other components
./devlog impact database --depth 2

# 5. Find all bug entries
./devlog list --type bug
```

### Example 4: Daily Standup Preparation

**Scenario:** Prepare for daily standup meeting.

```bash
# 1. Show recent work (last 5 entries)
./devlog list --limit 5

# 2. Find all TODOs
./devlog search "TODO"

# 3. Check entities in active development
./devlog entities --min 2

# 4. Get specific entry details
./devlog show 2024-01-17
```

### Example 5: Code Review Context

**Scenario:** Review recent changes before a PR.

```bash
# 1. List recent feature work
./devlog list --type feature --limit 10

# 2. Search for specific components
./devlog search "UserService"

# 3. See entity relationships
./devlog graph UserService --depth 2

# 4. Check impact of changes
./devlog impact UserService
```

## 🔍 Graph Traversal

### Understanding Entity Graphs

The devlog CLI builds a knowledge graph from your markdown entries:

```
Entry 1 (2024-01-15)
├── Entities: [JWT, API, MyFunction]
└── Related: Entry 2 (shares JWT)

Entry 2 (2024-01-16)
├── Entities: [database, MyFunction]
└── Related: Entry 1 (shares MyFunction)
```

### Traversal Examples

**Depth 1: Direct Mentions**
```bash
./devlog graph MyFunction --depth 1
# Shows entries directly mentioning MyFunction
```

**Depth 2: Extended Context**
```bash
./devlog graph MyFunction --depth 2
# Shows entries mentioning MyFunction
# Plus entries related to those entries
```

**Depth 3: Full Context**
```bash
./devlog graph MyFunction --depth 3
# Shows extended relationship network
```

### Reverse Traversal (Impact Analysis)

Find what depends on an entity:

```bash
./devlog impact MyFunction
# Shows: "If I change MyFunction, what else is affected?"
```

**Practical Example:**

```bash
# Before refactoring
./devlog impact UserService

# Output shows:
# - MyFunction (uses UserService)
# - UserController (depends on UserService)
# - TestSuite (tests UserService)

# Now you know what to test after refactoring
```

## 🏷 Entity Types

### CamelCase
Uppercase-first identifiers commonly used in code.

**Examples:** `MyFunction`, `ClassName`, `HTTPServer`, `UserService`, `JWT`

**Detected by:** Regex pattern for capitalized words

**Use case:** Track class names, function names, type names

### kebab-case
Lowercase identifiers with hyphens.

**Examples:** `my-function`, `parse-index-md`, `user-auth`, `url-path`

**Detected by:** Regex pattern for hyphenated lowercase words

**Use case:** Track file names, URL paths, command names

### Keywords
Special markers indicating work status or type.

**Examples:** `TODO`, `FIXME`, `NOTE`, `HACK`, `XXX`, `BUG`, `OPTIMIZE`, `REFACTOR`

**Detected by:** Predefined keyword list

**Use case:** Track work items, technical debt, notes

### Issue IDs
References to external issue tracking systems.

**Examples:** `bd-123`, `BD-456`, `beads-789`

**Detected by:** Regex pattern for issue IDs

**Use case:** Link devlog entries to issue tracker

## 📊 Output Formats

### Table Format (Default)

Human-readable markdown-style output.

```bash
./devlog list
```

**Output:**
```
# Devlog

## 2024-01-17 - Added unit tests for UserService
Wrote comprehensive tests for user CRUD operations.
Coverage now at 85% for UserService.

Entities: UserService, CRUD
```

### JSON Format

Machine-readable JSON for scripting and integration.

```bash
./devlog list --format json
```

**Output:**
```json
[
  {
    "date": "2024-01-17",
    "title": "Added unit tests for UserService",
    "description": "Wrote comprehensive tests for user CRUD operations.",
    "entities": ["UserService", "CRUD"],
    "timestamp": "2024-01-17T00:00:00Z",
    "line_number": 3
  }
]
```

**Use JSON with jq:**

```bash
# Extract just titles
./devlog list --format json | jq -r '.[].title'

# Count entries by type
./devlog list --format json | jq 'group_by(.type) | map({type: .[0].type, count: length})'

# Filter by date range
./devlog list --format json | jq 'select(.date >= "2024-01-15")'
```

## 🚀 Advanced Usage

### Combining Commands

```bash
# Find all TODOs related to authentication
./devlog search "TODO" | grep -i auth

# Get entity stats for recent work
./devlog list --limit 10 --format json | jq '[.[] | .entities] | add | group_by(.) | map({entity: .[0], count: length})'

# Find cross-references between entities
./devlog graph EntityA
./devlog graph EntityB
# Compare outputs to find shared entries
```

### Aliases for Common Tasks

```bash
# Add to your .bashrc or .zshrc
alias devlog-todo='devlog search "TODO"'
alias devlog-recent='devlog list --limit 5'
alias devlog-bugs='devlog list --type bug'
alias devlog-features='devlog list --type feature'
alias devlog-sessions='devlog list --type session'
alias devlog-entities='devlog entities --min 2'
```

### Integration with Git

```bash
# Show devlog entries for recent commits
git log --since="1 week ago" --format="%h %s"
devlog list --limit 7

# Create a release summary
git tag v1.0.0
devlog list --format json | jq '[.[] | select(.date >= "2024-01-01")]'
```

### Integration with Other Tools

```bash
# With fzf for interactive selection
devlog entities | fzf | xargs devlog graph

# With bat for pretty printing
devlog show 2024-01-15 | bat -l markdown

# With jq for data processing
devlog list --format json | jq '[.[] | {date, title, entities}]'
```

## 📚 Examples

### Example 1: Project Onboarding

```bash
# New developer joins the team
# They want to understand recent work

# 1. See recent activity
./devlog list --limit 20

# 2. Find major features
./devlog list --type feature

# 3. Explore key entities
./devlog entities --min 3

# 4. Understand architecture
./devlog graph UserService --depth 3
./devlog graph APIController --depth 3

# 5. Check technical debt
./devlog search "FIXME"
./devlog search "TODO"
```

### Example 2: Release Preparation

```bash
# Prepare release notes for v1.0.0

# 1. Find all features in this release
./devlog list --type feature --format json | \
  jq '[.[] | select(.date >= "2024-01-01" and .date <= "2024-01-31")]'

# 2. Find all bug fixes
./devlog list --type bug --format json | \
  jq '[.[] | select(.date >= "2024-01-01")]'

# 3. Generate release summary
./devlog list --limit 50 --format json | \
  jq 'group_by(.type) | map({type: .[0].type, entries: length})'
```

### Example 3: Refactoring Planning

```bash
# Plan refactoring of UserService

# 1. Understand current usage
./devlog graph UserService --depth 3

# 2. Check impact of changes
./devlog impact UserService --depth 3

# 3. Find related test coverage
./devlog search "UserService" --depth 2

# 4. Identify technical debt
./devlog search "UserService" | grep -i "FIXME\|REFACTOR\|TODO"
```

### Example 4: Debugging Production Issues

```bash
# Production incident: database slowdown

# 1. Find recent database changes
./devlog search "database" --limit 10

# 2. Check related components
./devlog impact database --depth 2

# 3. Find recent changes to related code
./devlog list --limit 20 | grep -i "database\|connection\|pool"

# 4. View full context of suspicious changes
./devlog show 2024-01-16
```

### Example 5: Knowledge Transfer

```bash
# Transfer knowledge to team members

# 1. Export entity documentation
./devlog entities --format json > entities.json

# 2. Generate architecture overview
./devlog entities --min 2 | \
  while read entity; do
    echo "## $entity"
    ./devlog graph "$entity" --depth 1
    echo ""
  done > architecture.md

# 3. Create workflow documentation
./devlog list --type feature > features.md
./devlog list --type bug > bugs.md
```

## 🔧 Tips and Best Practices

### Writing Devlog Entries

1. **Use Descriptive Titles**
   ```markdown
   ## 2024-01-15 - Implemented user authentication
   ```

2. **Include Entity References**
   ```markdown
   MyFunction handles token validation.
   UserService calls the database.
   Fixed bd-123.
   ```

3. **Add Keywords for Tracking**
   ```markdown
   TODO: Add refresh token support
   FIXME: Handle edge case in error handling
   NOTE: Performance optimization needed
   ```

4. **Link Related Work**
   ```markdown
   Related to bd-456.
   See also: MyFunction, UserService.
   ```

### Query Best Practices

1. **Start Broad, Then Narrow**
   ```bash
   ./devlog search "auth"           # Broad
   ./devlog search "authentication" # Narrower
   ./devlog graph JWT              # Specific
   ```

2. **Use Graph Context**
   ```bash
   # Without context
   ./devlog search "database"

   # With context
   ./devlog search "database" --depth 2
   ```

3. **Leverage JSON Output**
   ```bash
   # For scripts and automation
   ./devlog list --format json | jq '.'
   ```

4. **Combine with Other Tools**
   ```bash
   # Use with grep, jq, fzf, etc.
   ./devlog list | grep -i "urgent"
   ./devlog entities | fzf
   ```

## 🤝 Contributing

Contributions are welcome! Please see the main [CONTRIBUTING.md](../../CONTRIBUTING.md) file for guidelines.

## 📄 License

This project is part of the beads project. See the [LICENSE](../../LICENSE) file for details.

## 🐛 Troubleshooting

### Common Issues

**Problem:** "no index.md found at ./index.md"
```bash
# Solution: Specify the correct path
./devlog list --index /path/to/index.md
```

**Problem:** "No entities found"
```bash
# Solution: Your entries might not have entity patterns
# Check your markdown format
cat index.md | grep -E "([A-Z][a-z]+){2,}|[a-z]+-[a-z]+"
```

**Problem:** JSON output is empty
```bash
# Solution: Check if entries match your filter
./devlog list --format json | jq 'length'
```

## 📞 Getting Help

```bash
# General help
./devlog --help

# Command-specific help
./devlog list --help
./devlog graph --help
./devlog search --help
```

## 🎯 Summary

devlog transforms your markdown developer logs into a queryable knowledge graph, making it easy to:

- **Track entities** across your codebase
- **Explore relationships** between components
- **Search intelligently** with graph context
- **Understand impact** of changes
- **Resume work** with full context

Start using devlog today to unlock the knowledge hidden in your development logs!
