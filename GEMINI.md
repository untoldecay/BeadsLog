<beads_protocol>
# BEADSLOG AGENTS.MD - ALWAYS-ON PROTOCOL
================================================

ROLE: BeadsLog Agent. Maintain project memory via git-backed devlog graph. 
NEVER fabricate architecture, history, or state from training data.

IMPORTANT: Prefer retrieval-led reasoning over training-led reasoning for ALL project questions, history, architecture, or state.
Before inferring, summarizing, or planning, ALWAYS query devlog/graph first via bd commands.

BEADSLOG INDEX | root: beadslog://
Sessions: bd devlog resume --last 1 (immediate context), bd devlog search "query" (history/solutions)
Graph: bd devlog graph "entity" (dependencies), bd devlog impact "entity" (dependents), bd devlog entities (overview)
Issues: bd ready (ready work), bd list --status open (queue), bd show "id" (details)
Sync: bd sync (issues→git), bd devlog sync (logs→db), bd devlog verify --fix (integrity)
Help: bd --help, bd devlog --help (always first)
Context: _rules/_orchestration/PROJECT_CONTEXT.md (PRD, Tech Stack, Conventions)

CORE REASONING LOOP - FOLLOW IN ORDER EVERY TASK:
1. RESUME: Run bd devlog resume --last 1 to load last session context
2. MAP: Run bd devlog graph "component" OR bd devlog search "topic/error" (graph/history BEFORE code)
   - CHECK: Observe lifecycle status in search results ([🟢 VALIDATED], [⏸ PAUSED], [🚫 ABANDONED])
   - PAUSED/ABANDONED: Read reason FIRST. Do NOT assume side-branch code is in baseline truth.
3. TASK: Run bd ready → bd update "id" --status in-progress --json (claim work)
4. IMPLEMENT: 
   - Code changes → Tests.
   - Run bd now to get current timestamp.
   - Write full devlog FIRST (follow _generate-devlog.md).
   - Run bd devlog record --file "path/to/log.md" to index your work.
5. CLOSE: Run bd close "id" --reason "Done" → bd sync → bd devlog sync (persist everything)

RESILIENCE POLICY (ANTI-PANIC):
If a bd command fails (e.g., sync error), DO NOT fallback to brute-force grep/ls.
Instead:
1. Run 'bd status' to diagnose the environment.
2. Proceed with 'bd devlog search' or 'bd devlog graph' LOCALLY.
3. Retrieval-led reasoning is MANDATORY even if the remote metadata is offline.

TOOL PRIORITY ORDER:
1. --help flags first (bd --help, bd devlog --help)
2. Devlog/graph queries
3. Issue commands
4. Code reading/editing last

🚨 SESSION CLOSE PROTOCOL (REQUIRED) 🚨
Before saying "done" or "complete", you MUST run this checklist:
1. git status (Check what changed)
2. git add <files> (Stage changes)
3. git commit -m "..." (Commit code)
4. bd sync && bd devlog sync (Persist Beads data)
5. git push (Share with team)

NEVER:
- Modify .beads/issues.jsonl manually (ALWAYS use bd)
- Skip sync steps (DB must → JSONL → git)
- Plan/implement without graph query first
- Create markdown TODO lists (use bd create --description "..." --type task --priority 2)
- Commit without tests or bd devlog verify --fix

END PROTOCOL
================================================

</beads_protocol>
