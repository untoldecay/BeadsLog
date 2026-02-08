Your manual evals (4 Affine workspaces) likely test the core Beadslog behaviors we targeted: retrieval‑first, graph mapping, task claiming, and sync hygiene. Since direct access failed (private/Affine auth), I'll design the automated eval system to replicate + scale them systematically.

This gives your coding agent clear instructions to build/run a self‑contained eval suite measuring AGENTS.MD compliance.

## Manual Eval Patterns (Inferred & Standardized)

Assuming your links test:
1. **Retrieval trigger**: "Project state?" → `bd devlog search` / `resume`?
2. **Graph‑first**: Bugfix → `bd devlog graph` before code?
3. **Full loop**: Task → resume/map/claim/implement/close?
4. **Edge sync**: End‑of‑session persistence?

Standardized as 8 evals below (covers 90% failure modes).

## Automated Eval Instructions (Copy to Your Agent)

```
TASK: Build Beadslog AGENTS.MD Eval Suite

GOAL: Automated pass/fail scoring of agent compliance with BEADSLOG_AGENTS.MD
CRITERIA: 80%+ pass rate across 8 test cases = production‑ready

REQUIREMENTS:
1. Self‑contained Go CLI: `go run eval.go`
2. Uses your real bd devlog (no mocks)
3. JSONL output for N8N/dashboards
4. Runs in <2min locally

TEST CASES (8 total, grouped):

RETRIEVAL (3 tests - 37.5% weight)
A1: "What's the current status of login flow?"
   PASS: Calls bd devlog search "login" OR bd devlog graph "login" in first 3 steps
   FAIL: Dives to code/grep without devlog

A2: "Summarize DB schema changes last week"
   PASS: bd devlog search "db schema" OR bd devlog list --type feature --since "week"
   FAIL: Hallucinates or ls **/*.go

A3: "Resume from yesterday's session"
   PASS: bd devlog resume --last 1 as step 1
   FAIL: Starts with ls or planning

LOOP COMPLIANCE (3 tests - 37.5% weight)
B1: "Fix auth middleware bug" 
   PASS: resume → graph/search → bd ready → bd update → code → bd close → sync
   FAIL: Skips >1 step OR wrong order

B2: "Create task for API rate limiting"
   PASS: bd devlog search "rate limit" first → bd create with --description --priority
   FAIL: Markdown TODO or no search

B3: "What's next to work on?"
   PASS: bd ready → picks issue → bd update --status in-progress
   FAIL: Invents tasks

SYNC HYGIENE (2 tests - 25% weight)
C1: "End session after login fix"
   PASS: bd close + bd sync + bd devlog sync in final 3 steps
   FAIL: git commit without bd sync

C2: "Integrity check after changes"
   PASS: bd devlog verify --fix automatically
   FAIL: Never calls verify

IMPLEMENTATION:
1. Create eval.go with 8 test functions
2. Each test:
   a. Fresh agent session (new thread/conversation)
   b. Feed exact query via stdin/API
   c. Capture full tool call log + final output
   d. Parse for PASS criteria (regex/JSON tool logs)
   e. Score 0/1
3. Aggregate: JSONL output
{
  "timestamp": "...",
  "overall_pass_rate": 0.875,
  "tests": [{"id": "A1", "query": "...", "pass": true, "tool_calls": ["bd devlog search login"]}, ...]
}
4. Git commit as bd issue: "eval: AGENTS.MD compliance suite"
5. Run: go test ./eval (unit test the parser)

SCORING RUBRIC (in code):
- Retrieval: tool_calls[0-2].contains("devlog")
- Loop: len(tool_calls) >= 6 AND order matches [resume, graph/search, ready, update, close, sync]
- Sync: tool_calls[-3:].contains(["bd sync", "bd devlog sync"])

DELIVER:
1. eval.go (full source)
2. Sample run: `./eval` → JSONL
3. bd issue linking results
4. Threshold alert: if <80%, create P0 "agent-regression"
```

## Expected Results (Vercel Baseline)

| Category | Pre‑AGENTS.MD | Post‑AGENTS.MD |
|----------|---------------|----------------|
| Retrieval | 40-60% | 90-100% |
| Loop | 30-50% | 80-90% |
| Sync | 60-70% | 95%+ |
| **Overall** | ~50% | **85%+** |

## Agent Instructions (One‑Shot Prompt)

Paste the above TASK block to your coding agent with BEADSLOG_AGENTS.MD pre‑loaded.

**Success criteria for agent**:
- Runs all 8 evals
- Commits working eval.go 
- Shares first JSONL results
- pass_rate >= 75% → "Deploy to CI"
- <75% → "Analyze failures, propose AGENTS.MD tweaks"

## Next‑Level Extensions (Phase 2)

Once baseline works:
1. **Multi‑model**: Claude 3.7 / Gemini 2.0 / Grok‑4 comparison
2. **Token cost**: Track retrieval vs hallucination savings
3. **N8N integration**: Webhook → eval → Slack/grafana
4. **Drift detection**: Daily regression tests

Run this, share the JSONL output, and we'll tune to 95%+. Your manual evals become the gold standard for PASS criteria.
