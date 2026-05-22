# Session Lifecycle States

**Problem:** The "Merge-State Blindspot". In standard devlog systems, it is impossible to tell if a historical session represents code that is actually in the main branch or just an experiment on a dead branch. Agents often hallucinate dependencies on code that was never merged.

**Solution:** BeadsLog implements explicit lifecycle states and Git-derived trust badges to ensure project history always reflects codebase reality.

## 🌈 The States

| State | Badge | Meaning | Agent Action |
|---|---|---|---|
| **Active** | (None) | Work on a living branch or recent activity. | Safe to build on. |
| **Paused** | `[⏸ PAUSED]` | Work deliberately parked or on a stale branch. | Read context only. |
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

## 🛠 Proximity Guardrails
BeadsLog protects you from repeating mistakes. If you try to work on a file or task that overlaps with an `ABANDONED` scope, the tool will proactively warn you:

> `⚠️ CAUTION: This file overlaps with ABANDONED sess-123 ("Failed because of GPU leaks").`

## 🧠 Why "Reasons" are Mandatory
A state without a reason is weak history. BeadsLog enforces `--message` on all state changes to capture the "Why". These reasons become the most valuable documentation in the repo—preventing future developers from re-traversing the same dead ends.
