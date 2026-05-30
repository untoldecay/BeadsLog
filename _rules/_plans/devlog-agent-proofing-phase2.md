# Plan: Devlog Agent-Proofing & UX Evolution (Phase 2)

**Status:** Proposed
**Objective:** Eliminate the remaining friction points in the `bd devlog record` workflow, specifically addressing the "Success Trap" associated with a two-step (create stub -> edit file) process.

---

## **Current Problem (The 2-Step Trap)**

In v0.52.x, the `bd devlog record` command creates a Markdown stub and registers it in `_index.md`. It outputs a loud directive (`🚀 **AI ACTION REQUIRED:**`) telling the agent to open the file and fill in the blanks.

**The Risk:** 
Despite the loud warning and the `bd devlog verify --fix` guardrail catching incomplete stubs at the end of the session, this still requires the agent to manage state across multiple steps:
1. Agent runs `bd devlog record`
2. Agent reads directive
3. Agent opens the file
4. Agent writes the content

If the context window gets crowded or the agent hallucinates, they might push the file as-is, triggering the guardrail failure and causing a loop, or abandoning the sync.

---

## **Proposed Alternatives (Single-Step Flow)**

We want to reduce logging to a single action, where the creation of the content and the registration in the index happen simultaneously or sequentially in a way that guarantees completeness.

### **Alternative A: "Write First, Record Later"**
Instead of `record` creating the stub, the agent creates the file first using the standard `_generate-devlog.md` prompt.
- **Workflow:**
  1. Agent writes `_rules/_devlog/2026-05-26_my-feature.md` completely.
  2. Agent runs `bd devlog record --file _rules/_devlog/2026-05-26_my-feature.md`.
- **Logic:** `record` reads the file, parses the title (subject) and problem, and simply appends it to `_index.md`. If the file doesn't exist, it errors out immediately instead of creating a stub.
- **Pros:** 100% guarantees the work is written *before* it is indexed. No stubs are ever created by the tool.

### **Alternative B: "Pass Content via Stdin / Flag"**
Similar to `git commit -m`, but allowing for full markdown payloads.
- **Workflow:**
  1. Agent generates the markdown content internally.
  2. Agent pipes it into the command: `echo "$MARKDOWN_CONTENT" | bd devlog record --file _rules/_devlog/2026-05-26_my-feature.md`
- **Logic:** `record` receives the content, creates the file with that exact content, and appends the entry to `_index.md` in one atomic operation.
- **Pros:** Atomic. The file and the index are created simultaneously with the final content.

---

## **Conclusion**
Moving to a model where the agent writes the content *first* (either to disk or via stdin) is conceptually safer than giving the agent an empty template and trusting them to fill it in afterward.
