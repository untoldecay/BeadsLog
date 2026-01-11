# bd - Beads

**Distributed, git-backed graph issue tracker for AI agents.**

[![License](https://img.shields.io/github/license/untoldecay/BeadsLog)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/untoldecay/BeadsLog)](https://goreportcard.com/report/github.com/untoldecay/BeadsLog)
[![Release](https://img.shields.io/github/v/release/untoldecay/BeadsLog)](https://github.com/untoldecay/BeadsLog/releases)
[![npm version](https://img.shields.io/npm/v/@beads/bd)](https://www.npmjs.com/package/@beads/bd)
[![PyPI](https://img.shields.io/pypi/v/beads-mcp)](https://pypi.org/project/beads-mcp/)

## What is Beads?

Beads provides a **persistent, structured memory system** for AI coding agents. It replaces messy markdown plans with a **dependency-aware task graph**, enabling agents to handle long-horizon tasks without losing context or dropping work.

**Key capabilities:**
- **Graph-based task tracking** - Tasks with dependencies, priorities, and blockers
- **Git-backed storage** - All tasks versioned and synced via git (no external database)
- **Multi-platform CLI** - Native binary for macOS, Linux, FreeBSD, and Windows
- **MCP Server** - Model Context Protocol integration for Claude and other AI agents
- **Agent-optimized** - JSON output, auto-ready detection, and semantic memory compaction

**Use cases:**
- Coordinate multiple AI agents working on the same codebase
- Track long-running projects across many agent sessions
- Maintain task context in git repos without external services
- Replace markdown TODO lists with dependency-aware graphs

## 📚 Two CLI Tools

This project provides two complementary command-line tools:

| Tool | Purpose | Commands |
|------|---------|----------|
| **`bd`** | Issue tracker for task management | `bd init`, `bd create`, `bd ready`, `bd list`, `bd show`, `bd close` |
| **`devlog`** | Markdown devlog parser and analyzer | `devlog list`, `devlog graph`, `devlog search`, `devlog entities`, `devlog show` |

- **Use `bd`** to manage tasks, issues, and dependencies in a git-backed graph database
- **Use `devlog`** to parse and query markdown developer logs with entity extraction and graph visualization

Both tools work together: `devlog` can reference `bd` issue IDs (like `bd-123`) in your markdown logs, creating links between your devlog entries and your task tracker.

## ⚡ Quick Start

### For `bd` (Issue Tracker)

```bash
# Install (macOS/Linux/FreeBSD)
curl -fsSL https://raw.githubusercontent.com/untoldecay/BeadsLog/main/scripts/install.sh | bash

# Initialize (Humans run this once)
bd init

# Tell your agent
echo "Use 'bd' for task tracking" >> AGENTS.md
```

### For `devlog` (Markdown Parser)

```bash
# Install from Go
go install github.com/untoldecay/BeadsLog/cmd/devlog@latest

# Create a devlog
cat > index.md << 'EOF'
# Devlog

## 2025-01-15 - Started project
Initial implementation. MyFunction handles the core logic.
TODO: Add tests.
EOF

# Query your devlog
devlog list
devlog entities
```

## 📦 Installation

### One-Line Install (Recommended)

**macOS / Linux / FreeBSD:**
```bash
curl -fsSL https://raw.githubusercontent.com/untoldecay/BeadsLog/main/scripts/install.sh | bash
```
This downloads the latest binary for your platform and installs it to `~/.local/bin` (or adds to PATH).

### Package Managers

**npm (cross-platform):**
```bash
npm install -g @beads/bd
```
Wraps native binaries with automatic platform detection.

**Homebrew (macOS/Linux):**
```bash
brew install untoldecay/BeadsLog/bd
```
Installs the native binary with shell completion.

**Go (any platform):**
```bash
# Install bd (issue tracker)
go install github.com/untoldecay/BeadsLog/cmd/bd@latest

# Install devlog (markdown parser)
go install github.com/untoldecay/BeadsLog/cmd/devlog@latest
```
Installs to `$GOPATH/bin` (usually `~/go/bin`).

**Python MCP Server:**
```bash
pip install beads-mcp
```
Installs the Model Context Protocol server for Claude integration.

### From Source

```bash
# Clone repository
git clone https://github.com/untoldecay/BeadsLog.git
cd beads

# Build and install both tools
make install

# Or build individually
cd cmd/bd && go build -o bd
cd ../devlog && go build -o devlog
```
Builds from source and installs to `$GOPATH/bin`.

### System Requirements

- **Linux:** glibc 2.32+ (most distros from 2020+)
- **macOS:** 10.15 (Catalina) or later
- **FreeBSD:** 12.0 or later
- **Windows:** 10 or later (WSL recommended)
- **Go:** 1.24+ (building from source only)

### Verifying Installation

```bash
# Verify bd (issue tracker)
bd version

# Verify devlog (markdown parser)
devlog --help
```
Should output version information and help text.

### Shell Completion

Both tools include automatic shell completion for bash, zsh, fish, and powershell. Restart your shell after installation to enable.

## 🚀 Getting Started

### Initialize a Repository

```bash
# Navigate to your project
cd my-project

# Initialize beads (creates .beads/ directory)
bd init

# Optional: Use stealth mode (local-only, not committed to git)
bd init --stealth
```

**What `bd init` does:**
- Creates `.beads/` directory with SQLite database
- Adds `.beads/` to `.gitignore` (unless `--stealth` mode)
- Installs git hooks for auto-sync (pre-commit, post-merge, pre-push, post-checkout)
- Creates initial configuration in `.beads/config.yaml`

### Creating Tasks

```bash
# Create a simple task
bd create "Fix authentication bug"

# Create with priority (P0 = highest)
bd create "Add user login" -p 0

# Create with type
bd create "Refactor database" -t enhancement

# Create with description
bd create "Add tests" -d "Add unit tests for auth module"

# Create hierarchical tasks
bd create "Build auth system" -t epic
bd create "OAuth integration" -p 1    # Creates bd-a3f8.1
bd create "Token refresh" -p 2       # Creates bd-a3f8.1.1
```

### Task Dependencies

```bash
# Link tasks (child blocked by parent)
bd dep add bd-child bd-parent

# Mark tasks as related
bd dep add bd-task1 bd-task2 --type related

# View task graph
bd show bd-a3f8
```

### Working with Tasks

```bash
# List ready tasks (no open blockers)
bd ready

# List all tasks
bd list

# Show task details
bd show bd-a3f8

# Update task status
bd update bd-a3f8 --status in_progress

# Close a task
bd close bd-a3f8 --reason "Completed"
```

## 🛠 Features

### Git as Database

Beads stores issues as JSONL in `.beads/issues.jsonl`. Every task is:
- **Versioned** - Full history in git
- **Branched** - Task branches work like code branches
- **Merged** - No merge conflicts with hash-based IDs (`bd-a1b2`)

### Agent-Optimized

Designed specifically for AI agents:
- **JSON output** - All commands support `--json` flag
- **Auto-ready detection** - `bd ready` shows tasks with no open blockers
- **Dependency tracking** - Automatic blocker resolution
- **Semantic compaction** - Old closed tasks summarized to save context window

### Zero Conflicts

Hash-based IDs (`bd-a1b2`, `bd-a3f8.1`) prevent merge collisions in multi-agent/multi-branch workflows. No more "fix #123" conflicts.

### Invisible Infrastructure

- **SQLite cache** - Fast local queries with automatic sync
- **Background daemon** - Auto-sync every 5 seconds (optional)
- **Git hooks** - Automatic export/import on git operations

### Compaction

Semantic "memory decay" summarizes old closed tasks to save context window. Critical for long-running projects with thousands of tasks.

## 📖 Essential Commands

| Command | Action |
| --- | --- |
| `bd ready` | List tasks with no open blockers (agent starting point) |
| `bd create "Title" -p 0` | Create a P0 (highest priority) task |
| `bd dep add <child> <parent>` | Link tasks (child blocked by parent) |
| `bd show <id>` | View task details and audit trail |
| `bd close <id> --reason "Done"` | Close a task with reason |
| `bd sync` | Force immediate sync to git |
| `bd init --stealth` | Initialize without committing to git |

## 🔗 Hierarchy & Workflow

Beads supports hierarchical task IDs for organizing epics and large tasks:

```
bd-a3f8        # Epic: "Build authentication system"
bd-a3f8.1      # Task: "Add OAuth login"
bd-a3f8.1.1    # Sub-task: "Implement token refresh"
bd-a3f8.1.2    # Sub-task: "Add logout handler"
bd-a3f8.2      # Task: "Add email/password login"
```

**Stealth Mode:**
```bash
bd init --stealth
```
Use beads locally without committing `.beads/` files to the main repo. Perfect for personal use on shared projects.

## 🤖 Agent Integration

### For Claude (MCP Server)

Install the MCP server:
```bash
pip install beads-mcp
```

Add to your Claude Desktop config (`claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "beads": {
      "command": "uvx",
      "args": ["beads-mcp"]
    }
  }
}
```

Now Claude can directly query and manipulate tasks through tools like `beads_create`, `beads_list`, `beads_ready`, etc.

### For Other Agents

Add to your `AGENTS.md` file:
```markdown
# Task Tracking

Use `bd` for all task tracking and planning.

**Getting started:**
- Run `bd ready` to see available tasks
- Run `bd create "Title" -p 1` to create new tasks
- Run `bd show <id>` to view task details
- Always run `bd sync` after making changes

**Important:**
- Use JSON output: `bd ready --json`
- Update task status when starting work
- Close tasks when complete: `bd close <id> --reason "Done"`
- Never create tasks with "Test" prefix (use BEADS_DB for testing)
```

### Multi-Agent Coordination

Beads excels at coordinating multiple agents:
- **Hash-based IDs** prevent merge conflicts
- **Auto-sync** keeps all agents in sync
- **Dependency graph** prevents duplicate work
- **Audit trail** tracks who did what

## 🧪 Testing

### Manual Testing (Isolated Database)

**IMPORTANT:** Never pollute the production database with test issues!

```bash
# Use isolated database for testing
BEADS_DB=/tmp/test.db bd init --quiet --prefix test
BEADS_DB=/tmp/test.db bd create "Test issue" -p 1
```

### Automated Testing

```bash
# Run all tests (from repository root)
go test ./...

# Run short tests (skip integration tests)
go test -short ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## ⚙️ Configuration

Beads stores configuration in `.beads/config.yaml`:

```yaml
# Database settings
database:
  path: .beads/beads.db

# Git sync settings
git:
  auto_commit: false      # Auto-commit changes
  auto_push: false         # Auto-push to remote
  branch: main            # Branch for metadata

# Daemon settings
daemon:
  enabled: true           # Run background daemon
  sync_interval: 5        # Sync interval (seconds)

# Task settings
tasks:
  default_priority: 2     # Default priority (0-5)
  default_type: task      # Default task type
```

See [docs/CONFIG.md](docs/CONFIG.md) for all configuration options.

---

## 📝 Devlog CLI - Markdown Developer Log Parser

**`devlog`** is a companion CLI tool for parsing and analyzing markdown-based developer logs. It transforms your devlog into a queryable knowledge graph with entity extraction and relationship tracking.

### What is Devlog?

Devlog reads markdown files with entries in `## YYYY-MM-DD - Title` format and extracts:
- **CamelCase entities** - `MyFunction`, `ClassName`, `UserService`
- **kebab-case entities** - `my-function`, `url-path`, `api-endpoint`
- **Keywords** - `TODO`, `FIXME`, `NOTE`, `HACK`, `XXX`, `BUG`
- **Issue IDs** - References to `bd-123`, `BD-456`, etc.

**Use cases:**
- Maintain a searchable devlog alongside your code
- Track entity relationships across your codebase
- Find all work related to a specific component
- Generate reports from your daily development logs
- Resume work with full context from previous sessions

### Installing Devlog

```bash
# From Go (any platform)
go install github.com/untoldecay/BeadsLog/cmd/devlog@latest

# Or build locally
cd cmd/devlog
go build -o devlog

# Verify installation
devlog --help
```

### Quick Start with Devlog

**1. Create a devlog file:**

```bash
# Create index.md in your project
cat > index.md << 'EOF'
# Devlog

## 2025-01-15 - Implemented user authentication
Added JWT-based authentication to the API.
MyFunction handles token validation.
UserService manages user data.
TODO: Add refresh token support.

## 2025-01-16 - Fixed database connection bug
Fixed issue where connections were not being properly closed.
This was causing memory leaks in production.
Tracked by bd-123.

## 2025-01-17 - Added unit tests for UserService
Wrote comprehensive tests for user CRUD operations.
Coverage now at 85% for UserService.
EOF
```

**2. List your entries:**

```bash
devlog list
```

**3. Explore entity relationships:**

```bash
# Show graph for a specific entity
devlog graph UserService

# List all entities
devlog entities

# Search for specific topics
devlog search "database"
```

### Devlog Commands

| Command | Purpose | Example |
|---------|---------|---------|
| `devlog list` | List all devlog entries | `devlog list --limit 5 --type feature` |
| `devlog graph <entity>` | Show entity relationship graph | `devlog graph MyFunction --depth 3` |
| `devlog entities` | List all detected entities | `devlog entities --type CamelCase --min 3` |
| `devlog search <query>` | Full-text search with context | `devlog search "database" --depth 2` |
| `devlog show <date>` | Show full entry content | `devlog show 2025-01-15` |
| `devlog impact <entity>` | Show what depends on entity | `devlog impact UserService --depth 2` |
| `devlog resume` | Find matching sessions for context | `devlog resume --query auth --query JWT` |
| `devlog import-md` | Import markdown to beads database | `devlog import-md --file index.md` |

### Common Devlog Workflows

**Track feature development:**
```bash
# See all authentication work
devlog search "authentication"

# Explore related entities
devlog graph JWT
devlog graph UserService

# Check impact of changes
devlog impact UserService
```

**Investigate bugs:**
```bash
# Find bug-related entries
devlog list --type bug

# Search database issues
devlog search "database"

# Show full context
devlog show 2025-01-16
```

**Prepare for standup:**
```bash
# Show recent work
devlog list --limit 5

# Find all TODOs
devlog search "TODO"

# Check active entities
devlog entities --min 2
```

**Resume previous work:**
```bash
# Find sessions with context
devlog resume --query authentication --query JWT

# Explore entity graph
devlog graph MyFunction --depth 3
```

### Entity Types

| Type | Pattern | Examples |
|------|---------|----------|
| **CamelCase** | Capitalized words | `MyFunction`, `UserService`, `JWT`, `API` |
| **kebab-case** | Hyphenated lowercase | `my-function`, `url-path`, `api-endpoint` |
| **Keywords** | Special markers | `TODO`, `FIXME`, `NOTE`, `HACK`, `XXX`, `BUG` |
| **Issue IDs** | Beads issue references | `bd-123`, `BD-456` |

### Devlog Output Formats

**Table format (default, human-readable):**
```bash
devlog list
```

**JSON format (machine-readable):**
```bash
devlog list --format json | jq '.[] | {date, title, entities}'
```

### Integration with bd

Devlog and bd work together seamlessly:

```bash
# In your devlog, reference bd issues
## 2025-01-15 - Working on bd-a3f8
Implemented the token refresh logic.
bd-a3f8.1 covers the OAuth flow.

# Query devlog for issue references
devlog graph bd-a3f8
devlog search "bd-a3f8"
```

For complete devlog documentation, see [cmd/devlog/README.md](cmd/devlog/README.md).

---

## 🌐 Community Tools

See [docs/COMMUNITY_TOOLS.md](docs/COMMUNITY_TOOLS.md) for a curated list of community-built tools:
- Terminal interfaces (TUI)
- Web UIs and dashboards
- Editor extensions (VS Code, Neovim, etc.)
- Native desktop applications

## 📚 Documentation

### Getting Started
- [Quick Start Guide](docs/QUICKSTART.md) - 5-minute introduction
- [Installation Guide](docs/INSTALLING.md) - Detailed installation instructions
- [Setup Guide](docs/SETUP.md) - Initial configuration and setup

### Core Features
- [Agent Workflow Guide](AGENT_INSTRUCTIONS.md) - Using beads with AI agents
- [CLI Reference](docs/CLI_REFERENCE.md) - Complete command reference
- [Git Integration](docs/GIT_INTEGRATION.md) - Git sync and branching
- [Protected Branches](docs/PROTECTED_BRANCHES.md) - Separate branch for metadata

### Devlog CLI
- [Devlog Documentation](cmd/devlog/README.md) - Complete devlog CLI guide
- [Devlog Command Reference](cmd/devlog/COMMAND_QUICK_REFERENCE.md) - Quick command reference

### Advanced Topics
- [Architecture](docs/ARCHITECTURE.md) - System design and internals
- [Configuration](docs/CONFIG.md) - All configuration options
- [Daemon](docs/DAEMON.md) - Background sync daemon
- [Compaction](docs/ADVANCED.md#compaction) - Memory optimization
- [Multi-Repo Setup](docs/MULTI_REPO_AGENTS.md) - Managing multiple repositories

### Reference
- [FAQ](docs/FAQ.md) - Frequently asked questions
- [Troubleshooting](docs/TROUBLESHOOTING.md) - Common issues and solutions
- [Contributing](CONTRIBUTING.md) - Contribution guidelines
- [Security](SECURITY.md) - Security policy

### Community
- [![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/untoldecay/BeadsLog) - AI-powered Q&A
- [GitHub Discussions](https://github.com/untoldecay/BeadsLog/discussions) - Community discussions
- [GitHub Issues](https://github.com/untoldecay/BeadsLog/issues) - Bug reports and feature requests

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# Clone repository
git clone https://github.com/untoldecay/BeadsLog.git
cd beads

# Install dependencies
go mod download

# Run tests
go test ./...

# Run linter
golangci-lint run ./...

# Build
make build
```

### Code Standards
- **Go version:** 1.24+
- **Linting:** `golangci-lint run ./...`
- **Testing:** All features need tests
- **Documentation:** Update relevant .md files

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

Beads was created to solve the problem of persistent memory for AI coding agents. It draws inspiration from:
- Git's distributed architecture
- Graph databases for dependency tracking
- Issue trackers like Jira and GitHub Issues
- The AI agent community's need for structured memory
