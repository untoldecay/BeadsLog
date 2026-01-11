# Devlog Commands Quick Reference

This document provides quick usage examples for all tested devlog commands.

## List Command

List all devlog entries:
```bash
./devlog list
```

Filter by type:
```bash
./devlog list --type fix
./devlog list --type feature
./devlog list --type bug
```

Limit output:
```bash
./devlog list --limit 5
```

JSON output:
```bash
./devlog list --format json
```

Custom index file:
```bash
./devlog list --index /path/to/index.md
```

## Graph Command

Show graph for specific entity:
```bash
./devlog graph manage-columns
./devlog graph MyFunction
./devlog graph bd-123
```

Control depth:
```bash
./devlog graph --depth 1 manage-columns
./devlog graph --depth 3 MyFunction
```

Show all entities:
```bash
./devlog graph
```

## Entities Command

List all entities:
```bash
./devlog entities
```

Filter by type:
```bash
./devlog entities --type CamelCase
./devlog entities --type kebab-case
./devlog entities --type keyword
./devlog entities --type issue-id
```

Minimum mentions:
```bash
./devlog entities --min 3
```

Limit results:
```bash
./devlog entities --limit 10
```

JSON output:
```bash
./devlog entities --format json
```

## Search Command

Basic search:
```bash
./devlog search migration
./devlog search authentication
./devlog search bug
```

Filter by type:
```bash
./devlog search "API" --type feature
```

Limit results:
```bash
./devlog search migration --limit 3
```

JSON output:
```bash
./devlog search "database" --format json
```

Include graph context:
```bash
./devlog search "session" --depth 2
```

## Show Command

Show by date:
```bash
./devlog show 2025-11-29
./devlog show 2024-01-15
```

Show by filename:
```bash
./devlog show 2024-01-15.md
./devlog show entries/my-feature.md
```

Custom index:
```bash
./devlog show 2025-11-29 --index /path/to/index.md
```

## Output Formats

### Table Format (default)
Human-readable markdown-style output

### JSON Format
Machine-readable JSON for scripting/integration

## Common Patterns

### Find all fixes this week
```bash
./devlog list --type fix --limit 10
```

### Trace entity relationships
```bash
./devlog graph MyFunction --depth 3
```

### Search with context
```bash
./devlog search "performance" --depth 2
```

### Get entity statistics
```bash
./devlog entities --min 5 --format json
```

### View specific entry
```bash
./devlog show 2025-11-29
```

## Entity Types

- **CamelCase**: ClassNames, FunctionNames, VariableNames
- **kebab-case**: file-names, url-paths, command-names
- **keyword**: TODO, FIXME, NOTE, HACK, XXX, BUG
- **issue-id**: bd-123, BD-456

## Tips

1. Use `--format json` for scripting and automation
2. Use `--limit` to reduce output for large datasets
3. Use `--depth` in graph to control relationship traversal
4. Combine filters for precise queries
5. Use quotes for multi-word search terms

## Examples from Testing

```bash
# Find all fix-type entries
./devlog list --type fix

# See what's related to manage-columns
./devlog graph manage-columns

# Get entity statistics
./devlog entities

# Search for migration-related work
./devlog search migration

# View specific day's entry
./devlog show 2025-11-29
```

## Getting Help

```bash
./devlog --help
./devlog [command] --help
```

Examples:
```bash
./devlog list --help
./devlog graph --help
./devlog entities --help
./devlog search --help
./devlog show --help
```
