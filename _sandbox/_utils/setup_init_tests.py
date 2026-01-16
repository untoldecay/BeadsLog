import os
import shutil
import subprocess

SANDBOX_DIR = "_sandbox"

def clean_sandbox():
    if os.path.exists(SANDBOX_DIR):
        shutil.rmtree(SANDBOX_DIR)
    os.makedirs(SANDBOX_DIR)

def create_git_repo(path):
    subprocess.run(["git", "init"], cwd=path, capture_output=True)
    # Configure user for commits
    subprocess.run(["git", "config", "user.email", "you@example.com"], cwd=path, capture_output=True)
    subprocess.run(["git", "config", "user.name", "Your Name"], cwd=path, capture_output=True)

def setup_fresh(path):
    os.makedirs(path)
    create_git_repo(path)

def setup_existing_clean(path):
    os.makedirs(path)
    create_git_repo(path)
    
    # Create .beads structure
    os.makedirs(os.path.join(path, ".beads"))
    with open(os.path.join(path, ".beads", "issues.jsonl"), "w") as f:
        f.write("")
    
    # Create devlog structure
    devlog_dir = os.path.join(path, "_rules", "_devlog")
    os.makedirs(devlog_dir)
    
    index_content = """# Development Log Index

> [!IMPORTANT]
> **AI AGENT INSTRUCTIONS:**
> 1. **APPEND ONLY:** Always add new session rows to the **existing table** at the bottom of this file.
> 2. **NO DUPLICATES:** Never create a new "Work Index" header or a second table.
> 3. **STAY AT BOTTOM:** Ensure the table remains the very last element in this file.

This index provides a concise record of all development work for easy scanning and pattern recognition across sessions.

## Nomenclature Rules:
- **[fix]** - Bug fixes and error resolution

## Work Index

| Subject | Problems | Date | Devlog |
|---------|----------|------|---------|
| [init] Setup | Initial devlog structure setup | 2024-01-01 | [2024-01-01_setup.md](2024-01-01_setup.md) |
"""
    with open(os.path.join(devlog_dir, "_index.md"), "w") as f:
        f.write(index_content)
        
    with open(os.path.join(devlog_dir, "_generate-devlog.md"), "w") as f:
        f.write("# Prompt content")

def setup_existing_corrupt(path):
    os.makedirs(path)
    create_git_repo(path)
    
    devlog_dir = os.path.join(path, "_rules", "_devlog")
    os.makedirs(devlog_dir)
    
    # Content AFTER the table - should trigger strict parsing error
    index_content = """# Development Log Index

## Work Index

| Subject | Problems | Date | Devlog |
|---------|----------|------|---------|
| [init] Setup | Initial devlog structure setup | 2024-01-01 | [2024-01-01_setup.md](2024-01-01_setup.md) |

This footer should not be here and should cause an error.
"""
    with open(os.path.join(devlog_dir, "_index.md"), "w") as f:
        f.write(index_content)

def setup_missing_prompt(path):
    os.makedirs(path)
    create_git_repo(path)
    
    devlog_dir = os.path.join(path, "_rules", "_devlog")
    os.makedirs(devlog_dir)
    
    index_content = """# Development Log Index

## Work Index

| Subject | Problems | Date | Devlog |
|---------|----------|------|---------|
"""
    with open(os.path.join(devlog_dir, "_index.md"), "w") as f:
        f.write(index_content)
    # _generate-devlog.md is intentionally NOT created

def setup_agent_configured(path):
    os.makedirs(path)
    create_git_repo(path)
    
    content = """
# Agent Instructions

Some existing instructions.

BEFORE ANYTHING ELSE: run 'bd devlog onboard'
"""
    with open(os.path.join(path, "AGENTS.md"), "w") as f:
        f.write(content)

def setup_hooks_installed(path):
    os.makedirs(path)
    create_git_repo(path)
    
    hooks_dir = os.path.join(path, ".git", "hooks")
    os.makedirs(hooks_dir, exist_ok=True)
    
    hook_content = """#!/bin/sh
# Auto-sync devlogs to beads database
if [ -f "./bd" ]; then
    ./bd devlog sync >/dev/null 2>&1 &
elif command -v bd >/dev/null 2>&1;
    bd devlog sync >/dev/null 2>&1 &
fi
"""
    with open(os.path.join(hooks_dir, "post-commit"), "w") as f:
        f.write(hook_content)
    os.chmod(os.path.join(hooks_dir, "post-commit"), 0o755)
    
    with open(os.path.join(hooks_dir, "post-merge"), "w") as f:
        f.write(hook_content)
    os.chmod(os.path.join(hooks_dir, "post-merge"), 0o755)

def setup_git_not_initialized(path):
    os.makedirs(path)
    # Intentionally do NOT run git init here

def setup_hook_conflict(path):
    os.makedirs(path)
    create_git_repo(path)
    hooks_dir = os.path.join(path, ".git", "hooks")
    os.makedirs(hooks_dir, exist_ok=True)
    
    # Existing non-beads hook
    hook_content = """#!/bin/sh
echo "Running existing hook..."
# Some linting command
exit 0
"""
    with open(os.path.join(hooks_dir, "post-commit"), "w") as f:
        f.write(hook_content)
    os.chmod(os.path.join(hooks_dir, "post-commit"), 0o755)

def setup_agent_file_variation(path):
    os.makedirs(path)
    create_git_repo(path)
    # Create .cursorrules instead of AGENTS.md
    with open(os.path.join(path, ".cursorrules"), "w") as f:
        f.write("# Cursor Rules\n\nSome existing rules.")

def setup_partially_initialized(path):
    os.makedirs(path)
    create_git_repo(path)
    # Create dir but no files
    os.makedirs(os.path.join(path, "_rules", "_devlog"))

def setup_agent_no_tags(path):
    os.makedirs(path)
    create_git_repo(path)
    with open(os.path.join(path, "AGENTS.md"), "w") as f:
        f.write("# My Agent\n\nExisting rules.")

def setup_agent_outdated_tags(path):
    os.makedirs(path)
    create_git_repo(path)
    content = """<!-- BD_PROTOCOL_START -->
# Old Protocol
Do this.
<!-- BD_PROTOCOL_END -->

# My Agent
Existing rules.
"""
    with open(os.path.join(path, "AGENTS.md"), "w") as f:
        f.write(content)

def setup_agent_garbage_tags(path):
    os.makedirs(path)
    create_git_repo(path)
    content = """<!-- BD_PROTOCOL_START -->
# Broken Protocol
<!-- BD_PROTOCOL_END missing -->

# My Agent
Existing rules.
"""
    with open(os.path.join(path, "AGENTS.md"), "w") as f:
        f.write(content)

def main():
    print("Cleaning sandbox...")
    clean_sandbox()
    
    scenarios = [
        ("Test-01-Fresh", setup_fresh),
        ("Test-02-Existing-Clean", setup_existing_clean),
        ("Test-03-Existing-Corrupt", setup_existing_corrupt),
        ("Test-04-Missing-Prompt", setup_missing_prompt),
        ("Test-05-Agent-Configured", setup_agent_configured),
        ("Test-06-Hooks-Installed", setup_hooks_installed),
        ("Test-07-Git-Not-Initialized", setup_git_not_initialized),
        ("Test-08-Hook-Conflict", setup_hook_conflict),
        ("Test-09-Agent-File-Variation", setup_agent_file_variation),
        ("Test-10-Partially-Initialized", setup_partially_initialized),
        ("Test-11-Onboard-Fresh", setup_fresh),
        ("Test-12-Onboard-Existing-No-Tags", setup_agent_no_tags),
        ("Test-13-Onboard-Existing-With-Tags-Outdated", setup_agent_outdated_tags),
        ("Test-15-Onboard-Existing-Garbage-Tags", setup_agent_garbage_tags),
    ]

    for name, setup_func in scenarios:
        print(f"Setting up {name}...")
        setup_func(os.path.join(SANDBOX_DIR, name))
    
    print("\nTest scenarios generation complete.")

if __name__ == "__main__":
    main()