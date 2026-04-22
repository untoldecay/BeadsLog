# Comprehensive Development Log: Fix Record Path and Protocol Wording

**Date:** 2026-04-23

### **Objective:**
To resolve a "double path" bug in the 'bd devlog record' command where absolute or prefixed paths were being stored literally, causing resolution failures during sync. Also, to update the 'FullBootloader' wording to promote the explicit use of 'record' instead of implying it is automatic.

---

### **Phase 1: Double Path Fix**

**Initial Problem:** Agents recording with '--file _rules/_devlog/file.md' resulted in '_rules/_devlog/_rules/_devlog/file.md' during sync.

*   **Action Taken:** Modified 'cmd/bd/devlog_cmds.go' in 'devlogRecordCmd' to use 'filepath.Base(file)' for the markdown link destination.
*   **Result:** The index now stores only the filename, which is correctly resolved relative to '_index.md'.
*   **Verification:** Confirmed in sandbox that sync now finds the files correctly even if the full path is provided to 'record'.

---

### **Phase 2: Protocol Wording Update**

**Initial Problem:** The 'FullBootloader' (AGENT.md) wording suggested 'git commit' auto-generates the devlog, which is misleading now that 'record' is the mandatory way to update the index.

*   **Action Taken:** Updated 'cmd/bd/init_templates.go' to replace '(auto-generates devlog)' with an explicit 'bd devlog record' step in the reasoning loop.
*   **Result:** Agents are now clearly instructed to run the record command.

---

### **Final Session Summary**

**Final Status:** Fixed and updated. The devlog system is more resilient to agent path errors, and the protocol is more precise.

---

### **Architectural Relationships**
- bd-record -> _index.md (updates)
- FullBootloader -> AGENT.md (injected_by)
