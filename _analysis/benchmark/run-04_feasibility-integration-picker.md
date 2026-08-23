# Run 04 — Feasibility/impact/dependency assessment (bd's decisive win)

- **Date:** 2026-08-23
- **Corpus:** full BeadsLog codebase + devlogs.
- **Model:** sonnet (both arms + judge). n=1 (directional).
- **Task type:** a real engineering assessment, not recall.

## The question
> Assess the FEASIBILITY, IMPACT, and DEPENDENCIES of: "During `bd init`, ask the
> user their preferred issue-tracking integration (Jira/Linear/GitHub) via a
> charmbracelet/huh interactive form, and store the choice." Write a comprehensive
> report (feasibility, impact, dependencies & prior art, risks, implementation sketch).

Arm A = bd (`bd devlog search/graph/impact` `--json` + normal tools). Arm B = grep only.

## Result — bd won on efficiency AND quality

| | Arm A — bd | Arm B — grep |
|---|---|---|
| Tokens | **68,083** | 81,844 (bd −17%) |
| Tool-calls | **59** | 88 (bd −33%) |
| Judge total | **17/20** | 11/20 |
| — grounding | **5** | 2 |
| — impact | 4 | 3 |
| — prior-art | 4 | 3 |
| — actionability | 4 | 3 |

## Blind judge verdict (Set 1 = grep, Set 2 = bd; verbatim)
> **STORAGE VERDICT:** YAML is correct — every value the `bd init` wizard persists
> is written via `config.SetYamlConfig` (init.go 817–826: devlog.author,
> devlog.branch-tracking, entity_extraction.*, ollama.model), not SQLite;
> `jira.*`/`linear.*` in SQLite are runtime credentials/sync state, not a
> project-level preference declared at init time.
>
> **STRONGER: Set 2 (bd)** — it correctly identifies that the wizard persists via
> `SetYamlConfig`, while Set 1's SQLite recommendation contradicts how every other
> wizard-captured preference is actually stored.
>
> **KEY DIFFERENCE:** Set 1 (grep) infers storage from where integration credentials
> live (SQLite) rather than from where init-wizard choices land (YAML) — the wrong
> analogy for a project-level preference.

## Why bd won
The task required understanding *how the project actually works*, not finding a
string. bd recalled the right architecture/history (init-wizard persistence
pattern, "Linear is the canonical template" from the GitHub research session,
`ExternalRef`/`SourceSystem` already exist) and reasoned by the correct analogy
(→ YAML). grep reconstructed the codebase from scratch (88 tool-calls), reasoned
from the integration-credential code it happened to read (→ SQLite), and got the
central design call **wrong**.

---

## Arm A (bd) — full report

# Feasibility / Impact / Dependency Report: Integration Picker in `bd init`

## 1. FEASIBILITY
charmbracelet/huh v0.8.0 is pinned in `go.mod` and already used — `cmd/bd/init.go`
(lines 606–789) is a multi-step huh form (NewSelect/NewInput/NewConfirm/
NewMultiSelect); also `create_form.go`, `eval_task_opencode.go`. All forms gated by
`!quiet && ui.IsTerminal()`. Init prompts today: Auto-Sync, Enforce Devlogs, Author,
Branch Tracking, AI Enrichment (+conditional Ollama picker), Agent Instructions.
Difficulty: very easy (~30 lines). Open decisions: storage layer, options,
team-mode gating.

## 2. IMPACT
- `cmd/bd/init.go`: one `huh.NewSelect[string]` step; persist via
  `config.SetYamlConfig("integration.preferred", …)` (same pattern as devlog.author).
- `internal/ui/init_render.go`: optional `Integration` field on InitResult.
- config.yaml template: optional commented key.
- Tests use `--quiet`, no change. No changes to jira.go/linear.go/internal/linear/.

## 3. DEPENDENCIES & PRIOR ART
Three integrations: Linear (full, `internal/linear/`), Jira (full scaffolding),
GitHub (research only, devlog sess-aa65e3, no Go code, go-github not in go.mod).
Research session: "Linear is a complete template for the brick." Dual-layer config:
YAML (`config.SetYamlConfig` — devlog.author, sync-branch, issue-prefix) vs SQLite
(`store.SetConfig` — linear.api_key, jira.url, runtime/secrets). **Recommends YAML**
— it's a project-level declaration like issue-prefix, shareable, survives `.beads`
wipe. `types.go:49-50` ExternalRef/SourceSystem already exist. Sources: init.go,
linear.go, jira.go, types.go, config.yaml, go.mod, devlog sess-aa65e3, sess-da993b
(init-wizard origin).

## 4. RISKS
Wizard-length fatigue (mitigate: last step, "None" default, or team-mode only);
"None" must be a safe default; `integration.preferred` has no readers yet (advisory);
GitHub option without backend; interactive path untested (only `--quiet` path is).

## 5. SKETCH
init.go: `var preferredIntegration string` + huh.NewSelect (None/Linear/Jira/
GitHub-coming-soon) inside the TTY block; `config.SetYamlConfig(...)` after; optional
InitResult field + render; config.yaml commented key. ~40 lines, 2 files.

---

## Arm B (grep) — full report

# Feasibility / Impact / Dependency Report: Issue-Tracker Integration Picker in `bd init`

## 1. FEASIBILITY
huh v0.8.0 direct dep, used in 4 files (init.go 6 forms, create_form.go,
devlog_cmds.go, eval_task_opencode.go). Wizard init.go 606-789, guard
`!quiet && ui.IsTerminal()`. Steps: Auto-Sync, Enforce Devlogs, Author, Branch
Tracking, AI Enrichment (+Ollama), Agent Instructions. Difficulty low.

## 2. IMPACT
Two files: init.go (add huh.NewSelect, `selectedTracker:="none"`, persist via
**`store.SetConfig(ctx,"integration.preferred_tracker",…)`** — SQLite) and
init_render.go (PreferredTracker field, cosmetic).

## 3. DEPENDENCIES & PRIOR ART
Linear (full, internal/linear/, SQLite config), Jira (full, SQLite config). GitHub
planned not built — plan `_rules/_plans/github-notion-issue-sync.md` (paused,
BeadsLog-3c2); github.* namespace documented `config.go:29` but no getter/setter.
**Argues SQLite** because jira.*/linear.* are NOT in YamlOnlyKeys
(yaml_config.go:20) nor the prefix list (:91). [Judge: wrong analogy — the wizard
writes prefs via YAML.] Prior decisions: GitHub paused not abandoned (include as
signal); wizard degrades non-TTY; wizard already long (UX doc
2026-01-15_init-ux-test-summary.md); external_ref is per-issue, orthogonal.

## 4. RISKS
Wizard-length fatigue (primary); GitHub option without backend; store open at form
time (opened line 298, closed 537); `--quiet` skips form; no YamlOnlyKey needed.

## 5. SKETCH
init.go: `selectedTracker:="none"` + huh.NewSelect (None/Linear/Jira/GitHub-coming-
soon), persist `store.SetConfig(...)` if not none; init_render PreferredTracker;
wire InitResult; next-step hint appending `bd config set linear.api_key ...`.
~40 lines, 2 files.

---

## Note
n=1. Both reports are thorough and mostly accurate; the decisive gap was the
**storage-design call** (bd right / grep wrong), which the blind judge verified
against `init.go:817–826`. Confirm with 2–3 more feasibility-style runs before
citing the numbers publicly.
