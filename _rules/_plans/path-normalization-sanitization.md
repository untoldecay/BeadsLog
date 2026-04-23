# Plan: Devlog Path Normalization & Database Sanitization

**Status:** Completed ✅
**Objective:** Resolve the "double path" bug where `_rules/_devlog/` prefixes are incorrectly persisted in the index and database, leading to resolution failures during synchronization.

---

## **Current Problem**
1.  **Index Pollution:** Some entries in `_rules/_devlog/_index.md` contain full relative paths (e.g., `[file.md](_rules/_devlog/file.md)`) instead of base filenames.
2.  **Database Inconsistency:** The `sessions` table stores these literal paths. During `bd devlog sync`, the system joins the `devlog_dir` prefix again, resulting in `_rules/_devlog/_rules/_devlog/file.md`.
3.  **Manual Toil:** Currently, fixing this requires manual database edits or a complete reset/re-ingest, which destroys historical metadata.

---

## **Proposed Solution**

### **1. Defensive Parsing (Prevention)**
Update the parser to be "prefix-aware". Even if an index file contains the prefix, the system should normalize it to the base filename during ingestion.
- **Target:** `parseIndexMD` in `cmd/bd/devlog_core.go`.
- **Action:** If the extracted filename starts with the `Dir` (the devlog directory), strip the prefix.

### **2. Self-Healing Repair (Cleanup)**
Integrate path sanitization into the existing `bd devlog verify --fix` workflow.
- **Target:** `devlogVerifyCmd` in `cmd/bd/devlog_cmds.go`.
- **Action:** 
    - **Scan Index:** Detect rows with prefixed paths.
    - **Rewrite Index:** Update the file on disk to use normalized base filenames.
    - **Migrate Database:** Execute `UPDATE sessions SET filename = ? WHERE filename = ?` for all corrected paths to prevent duplication and ensure ID stability.

---

## **Implementation Steps**

### **Phase 1: Parser Normalization**
- [ ] Modify `parseIndexMD` to detect and strip redundant directory prefixes from filenames.
- [ ] Add unit test case for "double path" index rows.

### **Phase 2: Sanitization Logic in Verify**
- [ ] Add `normalizePathsInternal` function to `cmd/bd/devlog_cmds.go`.
- [ ] Integrate this function into the `verify --fix` command.
- [ ] Ensure it performs both the `_index.md` rewrite and the SQL `UPDATE` in a single transaction-like flow.

### **Phase 3: Validation**
- [ ] Create a sandbox with "corrupted" double-path entries.
- [ ] Run `bd devlog verify --fix`.
- [ ] Confirm `_index.md` is clean.
- [ ] Confirm `bd devlog sync` runs without warnings and finds all files.

---

## **Key Learnings & Risks**
- **ID Stability:** The `sessions` table uses `filename + title` for lookups. Updating the filename is safer than letting `sync` create new sessions with new IDs.
- **Relative Path Resolution:** Always ensure normalization respects `filepath.Join` logic used by the rest of the system.
