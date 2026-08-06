# Copy drift table — every "what is BeadsLog" surface (2026-08-06 audit)

Canonical target (from README proposal): *"BeadsLog is git-backed project memory for AI coding agents: an issue tracker (`bd`) + a devlog knowledge graph."*
Naming rule: **BeadsLog** = system · **bd** = CLI · **Beads** = upstream engine · **devlog** = memory layer.

| # | Surface | File | Current copy | Drift | Proposed fix |
|---|---|---|---|---|---|
| 1 | README | `README.md` | "Git-backed project memory that scales…" | ✅ Baseline — but missing JSONL/no-conflict story & naming rule | Replace with proposal draft |
| 2 | Root CLI help | `cmd/bd/main.go:217-220` | "bd - Dependency-aware issue tracker" / "Issues chained together like beads" | ❌ No memory/devlog at all — a `bd --help` user never learns the core value | Long: add one canonical sentence + "see `bd devlog --help` for the memory layer" |
| 3 | `bd devlog` help | `cmd/bd/devlog_cmds.go` | "Devlog management commands" | ❌ Zero purpose statement for the flagship feature | Long: "The devlog is BeadsLog's project memory: searchable session history + entity graph…" |
| 4 | BD_GUIDE.md template | generated (source in init templates) | "bd (beads) is a Git-backed issue tracker" | ❌ Wrong name, no memory framing — agents reading only this learn half the tool | Open with canonical sentence; add devlog section (resume/search/graph) |
| 5 | Website intro | `website/docs/intro.md` | "Beads (bd) is a git-backed issue tracker" | ❌ Upstream's identity, not ours; no devlog/memory anywhere | Rewrite intro from README proposal; keep feature tables below |
| 6 | CLAUDE.md | `CLAUDE.md` (dev section) | "beads … issue tracker. We dogfood our own tool." | ⚠️ Internal-facing; minor | Add canonical sentence at top, keep architecture content |
| 7 | Onboard wizard | `cmd/bd/onboard.go:159-168` | "Activation Guide… GOAL: Memory First" | ✅ Aligned | Keep; maybe echo canonical sentence in header |
| 8 | Agent protocol template | `cmd/bd/init_templates.go:158-219` | "Maintain project memory via git-backed devlog graph" | ✅ Aligned | Keep |
| 9 | AGENTS.md (root) | `AGENTS.md` | Same as #8 | ✅ Aligned | Keep |

**Pattern:** agent-facing surfaces (7-9) already say "memory"; human-facing surfaces (2-5) still say "issue tracker" — humans get the weaker pitch, which is exactly what happened in the Slack debate.

**Unification order (one PR each, no behavior change):** #2 + #3 (5 lines of Go strings, biggest reach) → #4 template → #5 website → #1 README swap → #6.
