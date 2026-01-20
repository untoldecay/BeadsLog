# Prompt: Generate Go Install Link

## Objective:
Generate a `go install` command that installs the `bd` binary from a specific git branch or commit hash.

## Persona:
CLI installation guide generator

## Input:
- Current git branch name (or use specific branch/commit)
- Short commit hash (8-9 characters) - optional
- Installation preference: branch, commit hash, release tag, or latest

## Output Format:

Generate a code block with the appropriate `go install` command based on the installation preference.

## Installation Options:

### Option 1: Install from Current Branch
Use when you want to install from the branch you're currently working on.

**Template:**
```bash
go install github.com/untoldecay/BeadsLog/cmd/bd@<branch-name>
```

**Example:**
```bash
# Get current branch name
git symbolic-ref --short HEAD
# Output: dev/beads-01-trap-verification

# Install from that branch
go install github.com/untoldecay/BeadsLog/cmd/bd@dev/beads-01-trap-verification
```

---

### Option 2: Install from Specific Branch
Use when you want to install from a different branch than your current one.

**Template:**
```bash
go install github.com/untoldecay/BeadsLog/cmd/bd@<branch-name>
```

**Examples:**
```bash
# Install from main branch
go install github.com/untoldecay/BeadsLog/cmd/bd@main

# Install from specific feature branch
go install github.com/untoldecay/BeadsLog/cmd/bd@dev/beads-01-trap-verification
```

---

### Option 3: Install from Commit Hash
Use when you need to install from a specific commit (more precise than branch name).

**Template:**
```bash
go install github.com/untoldecay/BeadsLog/cmd/bd@<short-hash>
```

**Example:**
```bash
# Get short commit hash (8 characters)
git rev-parse --short HEAD
# Output: 8d5260f5

# Install from that specific commit
go install github.com/untoldecay/BeadsLog/cmd/bd@8d5260f5
```

---

### Option 4: Install from Release Tag
Use when installing an official release.

**Template:**
```bash
go install github.com/untoldecay/BeadsLog/cmd/bd@v<version>
```

**Example:**
```bash
go install github.com/untoldecay/BeadsLog/cmd/bd@v0.47.0
```

---

### Option 5: Install from Latest
Use when you want the most recent release from the default branch.

**Template:**
```bash
go install github.com/untoldecay/BeadsLog/cmd/bd@latest
```

**Example:**
```bash
go install github.com/untoldecay/BeadsLog/cmd/bd@latest
```

## Important Notes:

1. **Why `cmd/bd` path:** This is the Go module path containing the `main` package
2. **Branch vs Hash:**
   - **Branch names** are more readable for feature branches and better for sharing
   - **Commit hashes** are more precise for testing specific states
   - **Both are supported** by Go's install mechanism
3. **Installation Location:** `go install` places the binary in `$GOPATH/bin` or `$HOME/go/bin`
4. **Path Considerations:** Add `$GOPATH/bin` or `$HOME/go/bin` to your `$PATH` if needed

## Recommended Usage:

### For Development/Testing:
Use branch name (current or specific) for easier identification:

```bash
# Install from current branch
go install github.com/untoldecay/BeadsLog/cmd/bd@$(git symbolic-ref --short HEAD)

# Verify installation
bd --version

# Run bd init in your project
cd /path/to/your/project
bd init
```

### For Production:
Use specific version tag or `@latest`:

```bash
# Install latest release
go install github.com/untoldecay/BeadsLog/cmd/bd@latest
```
