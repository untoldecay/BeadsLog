# Context-Aware Operation of Beads (`bd`)

The Beads (`bd`) command-line tool operates in a highly context-aware manner, meaning its behavior and the data it accesses are specific to the directory from which you invoke it. While `bd` is a single executable installed globally on your system, it intelligently adapts to your current project.

## How it Works: The `.beads/` Directory

When you run any `bd` command, the tool performs the following steps to determine its operational context:

1.  **Local Search:** It first checks the current directory for a subdirectory named `.beads/`.
2.  **Ancestral Search:** If `.beads/` is not found in the current directory, `bd` recursively searches up the directory tree (to parent directories) until it finds a `.beads/` folder.
3.  **Context Establishment:** Once a `.beads/` directory is located, `bd` establishes that directory as the root of its current operational context. All subsequent commands (e.g., `bd list`, `bd create`, `bd devlog sync`) will interact with the database (`.beads/beads.db`), configuration (`.beads/config.yaml`), and issue data (`.beads/issues.jsonl`) found within that specific `.beads/` context.

## Implications for Multi-Project Workflows

This design enables a seamless multi-project workflow:

*   **Isolated Data:** Each project has its own `.beads/` directory, ensuring that issues, devlogs, and configurations are completely isolated. Running `bd list` in `~/my-project-a` will only show issues relevant to `my-project-a`.
*   **Consistent Tooling:** You use the exact same `bd` commands regardless of which project you are working on. The tool automatically adapts its context.
*   **Easy Collaboration:** Project-specific data lives within the project's repository, simplifying version control and collaboration for issues and devlogs.

## Example:

Consider the following directory structure:

```
~/
├── my-project-a/
│   ├── .beads/
│   └── src/
└── my-project-b/
    ├── .beads/
    └── docs/
```

*   If you run `cd ~/my-project-a` and then `bd list`, Beads will load data from `~/my-project-a/.beads/`.
*   If you then run `cd ~/my-project-b/docs` and execute `bd list`, Beads will search upwards, find `~/my-project-b/.beads/`, and load data relevant to `my-project-b`.

This context-aware behavior simplifies development by providing a single, powerful tool that intelligently manages data across all your projects.
