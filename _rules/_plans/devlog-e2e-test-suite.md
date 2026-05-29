# BeadsLog Feature Plan: Devlog E2E Test Suite
**Date:** 2026-05-26
**Status:** Active Testing

## 1. Context & Goal
To ensure the stability of the Devlog system and its recent enhancements (Extraction Hardening, Aliasing, Search UI, and Lifecycle), we are defining a comprehensive End-to-End (E2E) test suite. This suite validates core components in an isolated Git repository sandbox.

## 2. Test Scenarios

### Test 1: Username Edit & Record Command
- **Action:** Update `git config user.name`, then run `bd devlog record` with CLI flags (non-interactive).
- **Validation:** 
  1. The generated Markdown file contains the updated username.
  2. `bd devlog search` displays the new username, proving the database indexed it correctly.

### Test 2: Commit Hook Verification
- **Action:** Perform a `git commit` inside the initialized repository.
- **Validation:** Ensure that `.beads-hooks/` or `.git/hooks/` are correctly configured by `bd init` and actively tracking the commit.

### Test 3: Extraction Hardening (Noise Reduction & Explicit Edges)
- **Action:** Create a devlog containing noise words (`using`, `component`, `module`, `state`) and strong keywords (`AuthService`, `UserModal`). Add an explicit markdown edge (`- FragmentA -> TargetB (uses)`).
- **Validation:** 
  1. Noise words are completely absent from `bd devlog entities`.
  2. Strong keywords are successfully extracted.
  3. The explicit edge is extracted and linked with 1.0 confidence.

### Test 4: Aliasing & Registry Sync
- **Action:** Run `bd devlog alias TargetB FragmentA` and flush the DB.
- **Validation:** 
  1. `TargetB` and `FragmentA` are merged.
  2. `.beads/aliases.jsonl` contains the persistent mapping.

### Test 5: Search UI & Metadata
- **Action:** Run `bd devlog search TargetB --preview`.
- **Validation:** The output correctly displays the `Author` (with 👤), `Agent` (with 🤖), and `Branch` metadata, along with 3-line snippets.

### Test 6: Graph & Impact
- **Action:** Run `bd devlog graph` and `bd devlog impact`.
- **Validation:** Ensure the dependency tree renders without throwing errors after the alias merge.

### Test 7: Unalias & Verification
- **Action:** Run `bd devlog unalias FragmentA`, followed by `bd devlog verify --fix`.
- **Validation:** 
  1. The alias is removed from `.beads/aliases.jsonl`.
  2. The original distinct entities are restored via markdown re-parsing.

### Test 8: Session Lifecycle
- **Action:** Run `bd devlog pause` and `bd devlog abandon`.
- **Validation:** `bd devlog status` correctly surfaces the states and short reasons.

### Test 9: Atomic Record & Success Trap Prevention
- **Action:** Run `bd devlog record` for a file that does NOT exist.
- **Validation:** 
  1. The command creates the file automatically.
  2. The command output contains the `🚀 **AI ACTION REQUIRED:**` directive.

### Test 10: Incomplete Stub Detection
- **Action:** Run `bd devlog verify` on the newly created stub (which contains placeholders).
- **Validation:** `verify` must report the session as an "unfinalized stub" and issue an AI Directive.

### Test 11: Orphan Detection
- **Action:** Create a new `.md` file in the devlog directory but do NOT record it.
- **Validation:** `bd devlog sync` must print a warning about the orphaned file.

### Test 12: Non-Interactive Prune
- **Action:** Delete a recorded file and run `bd devlog sync` to mark it as a ghost. Then run `bd devlog prune`.
- **Validation:** The ghost session is removed from the database without interactive prompts.

### Test 13: Preferred Casing
- **Action:** Extract an entity with specific casing (e.g. `UserAuthenticationService`).
- **Validation:** `bd devlog entities` must display the name with its original casing, not just lowercase.

### Test 14: Auto-Flush Metadata
- **Action:** Run `bd devlog alias`.
- **Validation:** `.beads/aliases.jsonl` must be updated immediately without an explicit manual sync.
