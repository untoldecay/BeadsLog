# Session Lifecycle States

**Problem:** The "Merge-State Blindspot". In standard devlog systems, it is impossible to tell if a historical session represents code that is actually in the main branch or just an experiment on a dead branch. Agents often hallucinate dependencies on code that was never merged.

**Solution:** BeadsLog implements explicit lifecycle states and Git-derived trust badges to ensure project history always reflects codebase reality.

## 🌈 The States

| State | Badge | Meaning | Agent Action |
|---|---|---|---|
| **Active** | (None) | Work on the current branch or recent activity. | Safe to build on. |
| **Ongoing** | `[🔄 ONGOING]` | Work alive on another branch — the owner context-switched away and will resume. Derived automatically for off-branch unmerged work. | Safe to read; coordinate before modifying. |
| **Paused** | `[⏸ PAUSED]` | Work **explicitly** parked with a reason (`bd devlog pause`). | Read context only. |
| **Abandoned**| `[🚫 ABANDONED]`| Work found to be flawed or rejected. | **Cautionary history only.** |

### 🟢 The "Validated" Badge
When you see `[🟢 VALIDATED]` in search results, it means the `bd-daemon` has verified that this session's code has actually landed in the project's mainline (via Git ancestry or patch-matching).

## 🚀 Management Commands

### 1. Pausing Work
When pivoting to a more urgent task:
```bash
bd devlog pause --scope branch:feature-xyz --message "Pivoting to landing page"
```

### 2. Killing a Path
When an experiment fails:
```bash
bd devlog abandon --scope entity:NewRenderer --message "Caused GPU memory leaks"
```

### 3. Reviving Work (the explicit un-pause)
When a paused scope is alive again — without needing to check out the branch:
```bash
bd devlog ongoing --scope branch:feature-xyz --message "Resuming after the landing page ships"
```
Note: you rarely need this for mere context switches — off-branch unmerged work
is shown as `[🔄 ONGOING]` automatically. `PAUSED` only ever means someone ran
`bd devlog pause` deliberately.

## 🛠 Proximity Guardrails
BeadsLog protects you from repeating mistakes. If you try to work on a file or task that overlaps with an `ABANDONED` scope, the tool will proactively warn you:

> `⚠️ CAUTION: This file overlaps with ABANDONED sess-123 ("Failed because of GPU leaks").`

## 🧠 Why "Reasons" are Mandatory
A state without a reason is weak history. BeadsLog enforces `--message` on all state changes to capture the "Why". These reasons become the most valuable documentation in the repo—preventing future developers from re-traversing the same dead ends.
