# Comprehensive Development Log: Debug bd init Hook Prompt

**Date:** 2026-01-19

### **Objective:**
To investigate and fix issue `bd-52o`, where the `bd init` interactive prompt for enabling devlog enforcement was being skipped. Additionally, implemented a minor UX enhancement for `bd config --help` and improved the flow of the `bd init` prompts.

---

### **Phase 1: UX Enhancement for bd config**

**Initial Problem:** Users might not know how to list available configurations from the help text.

*   **Action Taken:** Modified `cmd/bd/config.go` to add "Run bd config list to list all configuration available." to the command's Long description.
*   **Result:** Verified with `bd config --help`.

---

### **Phase 2: Investigation of Missing Init Prompt**

**Initial Problem:** `bd init` was supposed to ask "Do you want to ENFORCE devlog updates...?" but it was skipping this prompt.

*   **My Assumption/Plan #1:** The logic in `cmd/bd/init.go` checks if the value is already configured. I assumed it was checking for an empty string, but the recent addition of default values (in `bd-k7o`) caused `GetYamlConfig` to return "false" instead of empty string, making the system think it was already configured.
    *   **Action Taken:** Created a reproduction sandbox `_sandbox/debug_hooks` and confirmed the prompt was missing. Analyzed `cmd/bd/init.go`.
    *   **Result:** Confirmed the code used `config.GetYamlConfig(...) != ""`. Since a default of `false` was added in `internal/config/config.go`, this condition was now always true.
    *   **Analysis/Correction:** I needed to check the *source* of the configuration value, using `config.GetValueSource`. If the source is `SourceDefault`, we should prompt.

*   **My Assumption/Plan #2:** Switch to `config.GetValueSource` to detect if the user explicitly set the config.
    *   **Action Taken:** Modified `cmd/bd/init.go` to use `config.GetValueSource`. Added debug prints.
    *   **Result:** The debug prints showed `source=config_file configured=true`. This was unexpected because I was in a fresh directory.
    *   **Analysis/Correction:** `config.Initialize()` walks up the directory tree looking for `config.yaml`. Since I was running the test inside a subdirectory of the Beads repository itself (`_sandbox/debug_hooks`), Viper found the *repository's own* `.beads/config.yaml` and loaded it. This meant `bd init` believed the setting was already configured (by the parent repo).

*   **My Assumption/Plan #3:** Ensure `bd init` only respects the configuration file in the *target* directory.
    *   **Action Taken:**
        *   Added `GetConfigFileUsed()` helper to `internal/config/config.go`.
        *   Modified `cmd/bd/init.go` to compare the path of the config file used by Viper against the expected path in the new `.beads` directory.
        *   If they don't match (i.e., we loaded a parent config), we treat the setting as unconfigured for the purpose of the prompt.
    *   **Result:** Verified in the sandbox. The prompt "[Devlog Policy] Do you want to ENFORCE..." now appears correctly.

---

### **Phase 3: Relocating the Prompt**

**Initial Problem:** The "Devlog Enforcement" question was appearing before the "Git Hooks" section, which felt disjointed since enforcement relies on hooks.

*   **Action Taken:**
    *   Moved the enforcement prompt logic from `cmd/bd/init.go` to `cmd/bd/devlog_cmds.go` inside the `initializeDevlog` function.
    *   Placed it immediately after the "Install auto-sync hooks?" prompt.
    *   Adapted the logic to work within `devlog_cmds.go` (calculating expected config path from `dbPath`).
*   **Result:** The `bd init` flow is now smoother, with all git-hook related questions grouped under the `[Log Memory]` -> `Git hooks` section.

---

### **Phase 4: UX Polish**

**Initial Problem:** The prompts were technically accurate but mixed mechanisms (hooks) with value (automation). The flow was also a bit dense.

*   **Action Taken:**
    *   Redesigned the "Git Hooks Integration" section in `cmd/bd/devlog_cmds.go` to be an "[Automation Setup]" block.
    *   Separated "Auto-Sync" and "Enforcement" into numbered steps 1 and 2.
    *   Added clear value-proposition descriptions to each prompt.
    *   Updated the enforcement prompt to explicitly mention the `git commit --no-verify` bypass.
*   **Result:** `bd init` now presents a clear, value-driven automation menu:
    ```text
    [Automation Setup]
      1. Enable Auto-Sync? [Y/n]
         (Keeps your issue tracker up-to-date in the background)
      2. Enforce Devlogs? [y/N]
         (Prevents commits until you update the devlog...)
    ```

---

### **Phase 5: Bootstrap Trigger Verification Fix**

**Initial Problem:** When running `bd init` on a project that already had agent files with full "Devlog Protocol" (e.g., from a previous installation), only CLAUDE.md was getting the bootstrap trigger, while GEMINI.md was being skipped. This prevented agents from running `bd onboard` to get protocol updates.

*   **My Assumption/Plan #1:** The `injectBootstrapTrigger` function was checking if a file contained "Devlog Protocol" text and skipping it.
    *   **Action Taken:** Analyzed `cmd/bd/devlog_cmds.go` line 249. Confirmed the logic:
        ```go
        if strings.Contains(sContent, trigger) || strings.Contains(sContent, "Devlog Protocol") {
            return false // Already configured
        }
        ```
    *   **Result:** This prevented any file with an existing protocol from getting the bootstrap trigger, breaking the agent trap cycle.

*   **My Assumption/Plan #2:** Instead of skipping files with protocols, replace them with the bootstrap trigger so agents will run `bd onboard` to get the latest version.
    *   **Action Taken:** Refactored `injectBootstrapTrigger` function to:
        1. Check if file has bootstrap trigger → skip (idempotent)
        2. Check if file has full protocol (with start/end tags) → replace protocol block with bootstrap trigger
        3. Check if file has broken/incomplete protocol → prepend bootstrap trigger
        4. Check if file has neither → prepend bootstrap trigger
    *   **Result:** All edge cases now handled correctly. Verified with test suite:
        - ✅ Full protocol → replaced with bootstrap trigger
        - ✅ Bootstrap trigger present → skipped (idempotent)
        - ✅ Broken protocol → prepends bootstrap trigger
        - ✅ Empty file → adds bootstrap trigger

*   **Analysis/Correction:** This creates a clean reset cycle:
    1. User runs `bd init` → replaces any existing protocol with bootstrap trigger
    2. Agent starts session → sees "BEFORE ANYTHING ELSE: run 'bd onboard'"
    3. Agent runs `bd onboard` → gets latest protocol from embedded binary
    This ensures every `bd init` forces agents to refresh to the current protocol version.

*   **My Assumption/Plan #3:** Handle edge case where file has both bootstrap trigger AND full protocol (leftover cruft from previous updates).
    *   **Action Taken:** Added check at the beginning of `injectBootstrapTrigger` to detect if file contains both trigger and protocol tags. If so:
        1. Remove old bootstrap trigger from beforeProtocol content
        2. Remove entire protocol block between tags
        3. Rebuild with just bootstrap trigger and user content outside protocol
    *   **Result:** Files with leftover cruft (both trigger and protocol) are now cleaned up correctly.
    *   **Analysis/Correction:** The "both trigger and protocol" case should never happen in normal usage, but handles situations where previous updates left breadcrumbs. By removing the old trigger from beforeProtocol, we prevent duplicate triggers in the final output.

---

### **Final Session Summary**

**Final Status:** Issue `bd-52o` is fixed and significantly improved. `bd init` logic for prompting is robust against inherited config, the user experience for setting up automation is now structured and clear, and the bootstrap trigger mechanism now correctly enforces protocol updates across all agent files.
**Key Learnings:**
*   **Viper Configuration Inheritance:** Viper's config loading strategy (walking up directories) can interfere with initialization logic when running nested instances (e.g., tests or sub-repos). Explicitly checking the config file path is necessary to distinguish between "inherited" and "local" configuration.
*   **UX Grouping:** Grouping related configuration options (like git hooks and enforcement policies) under a single header ("Automation Setup") reduces cognitive load compared to scattering them across different phases of initialization.

---

### **Architectural Relationships**
- cmd/bd/init.go -> internal/config/config.go (uses GetValueSource, GetConfigFileUsed)
- cmd/bd/devlog_cmds.go -> internal/config/config.go (uses GetValueSource, SetYamlConfig)
- cmd/bd/devlog_cmds.go -> injectBootstrapTrigger (modifies agent files)
- injectBootstrapTrigger -> Agent instruction files (replaces protocol with trigger)
- bd onboard -> injectProtocol (replaces trigger with full protocol)
