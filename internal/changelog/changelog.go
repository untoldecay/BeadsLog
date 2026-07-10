package changelog

import (
	"fmt"
	"strconv"
	"strings"
)

type Entry struct {
	Version  string
	Date     string
	Features []string
	Protocol []string
}

// CurrentVersion is the latest version of the tool embedded in this binary
const CurrentVersion = "0.55.0"


var entries = []Entry{
	{
		Version: "0.55.0",
		Date:    "2026-07-10",
		Features: []string{
			"Trust Layer: ghost sessions excluded from resume/list/search/catchup; honest 'devlog status' health; sync reconciliation summary.",
			"Machine-First: '--json' on devlog list/show/resume/status/entities/impact and catchup.",
			"Graph Trust: extraction-time noise filtering, auto-merge of duplicate entity variants on sync, 'devlog prune --noise'.",
			"Entity Bonus: graph-linked sessions rank higher in search (alias-aware).",
			"Visual Graph: 'devlog graph <entity> --html <path>' exports an interactive force-directed graph.",
			"Genesis Devlog: fresh installs start with a real first session (0 ghosts, 0 incomplete).",
			"Catchup Digest: 'bd catchup --digest' groups activity by feature arc with narratives and lifecycle deltas.",
			"Ongoing State: off-branch work shows [🔄 ONGOING]; 'bd devlog ongoing' is the explicit un-pause.",
		},
		Protocol: []string{
			"Trust your reads: resume/search/list never return dead sessions; when sync reports ghosts or incompletes, run the suggested repair command.",
			"Prefer '--json' on devlog read commands for machine-readable output.",
			"PAUSED now always means deliberately parked with a reason; ONGOING is a healthy context switch — do not treat it as blocked.",
		},
	},
	{
		Version: "0.54.0",
		Date:    "2026-05-31",
		Features: []string{
			"Catchup Prompt: Automated creation/update of '_generate-catchup.md' during init and onboard.",
			"Improved UX: Reliable access to catchup summarization instructions for agents.",
		},
		Protocol: []string{
			"Use '_rules/_devlog/_generate-catchup.md' to summarize technical activity.",
		},
	},
	{
		Version: "0.53.1",
		Date:    "2026-05-31",
		Features: []string{
			"Time Utility: Added 'bd now' command to provide current system time in devlog format.",
			"Protocol Update: Injected protocol now instructs agents to use 'bd now' to prevent time hallucinations.",
		},
		Protocol: []string{
			"Run 'bd now' to get the current timestamp for devlog headers.",
		},
	},
	{
		Version: "0.53.0",

		Date:    "2026-05-30",
		Features: []string{
			"Write-First Workflow: Reverted 'Auto-Stub' in favor of a stricter 'Write -> Record' order.",
			"Auto-Metadata Extraction: 'bd devlog record' now automatically parses Subject and Problem from your file.",
			"Loud Error Reporting: High-signal directives when a devlog file is missing.",
		},
		Protocol: []string{
			"You MUST write the full Markdown devlog FIRST, then run 'bd devlog record --file <path>'.",
		},
	},
	{
		Version: "0.52.1",
		Date:    "2026-05-26",
		Features: []string{
			"Safe Aliasing: Added '--dry-run' flag to 'bd devlog alias' to preview graph impact.",
			"Impact Stats: Dry run displays the number of sessions and relationships to be merged.",
		},
		Protocol: []string{
			"Use '--dry-run' when unsure about a potential entity merge.",
		},
	},
	{
		Version: "0.52.0",
		Date:    "2026-05-26",
		Features: []string{
			"Graph Hardening (Minor): Major overhaul of entity extraction and relationship integrity.",
			"Shared Alias Registry: Aliases now persist in '.beads/aliases.jsonl' for team-wide consistency.",
			"Agent-Proofing: Atomic 'record' command with stub generation and 'verify' gatekeeping.",
			"Improved Discovery: Jaccard-based session overlap hints and robust pathfinding.",
			"Casing Preservation: UI now respects original entity capitalization (e.g. AuthService).",
			"History Resilience: Automatic session re-anchoring after squash-merges via patch-id.",
		},
		Protocol: []string{
			"Agents MUST verify '💡 OPPORTUNITY' hints and use 'bd devlog alias' for graph hygiene.",
			"Mandatory 'bd devlog verify --fix' before closing sessions to adopt orphans.",
		},
	},
	{
		Version: "0.51.9",
		Date:    "2026-05-26",
		Features: []string{
			"Catchup Fix: Default 'since' window increased to 7 days (from 24h) when no previous check exists.",
			"Improved UX: New message informing users when 'bd catchup' falls back to default window.",
		},
		Protocol: []string{
			"Run 'bd catchup --ack' regularly to maintain a sliding window of team activity.",
		},
	},
	{
		Version: "0.51.8",
		Date:    "2026-05-26",
		Features: []string{
			"Visibility: 'bd devlog status' now displays the count of incomplete or unfinalized sessions.",
			"Improved Grooming: Status now suggests 'bd devlog prune' for ghost sessions.",
		},
		Protocol: []string{
			"Monitor 'bd status' to ensure no 'Incomplete' sessions remain before closing.",
		},
	},
	{
		Version: "0.51.7",
		Date:    "2026-05-26",
		Features: []string{
			"Casing Preservation: RegexExtractor now preserves original casing for entities.",
			"Robust Discovery: Enhanced FindBeadsDir for cross-directory workspace stability.",
			"Unalias Fix: Case-sensitive unaliasing for consistent registry cleanup.",
		},
		Protocol: []string{
			"Entities now appear in the UI with their preferred capitalization.",
		},
	},
	{
		Version: "0.51.6",
		Date:    "2026-05-26",
		Features: []string{
			"Success Trap Prevention: 'bd devlog record' now outputs a loud AI Directive when a stub is created.",
			"Stub Detection: 'bd devlog verify' now detects and warns about unfinalized Markdown templates.",
		},
		Protocol: []string{
			"Never assume 'bd devlog record' has logged your work; you MUST fill in the 'Work Done' section.",
		},
	},
	{
		Version: "0.51.5",
		Date:    "2026-05-26",
		Features: []string{
			"Atomic 'record': Automatically creates a Markdown stub if the devlog file doesn't exist.",
			"Orphan Warnings: 'bd devlog sync' now warns about devlog files not present in the index.",
			"Non-interactive 'prune': New 'bd devlog prune' command to quickly purge ghost sessions.",
			"Auto-Flush: Maintenance commands (alias, unalias, prune, etc.) now automatically update JSONL files.",
			"Preferred Casing: Entities now preserve their original casing for better readability in UI results.",
		},
		Protocol: []string{
			"Agents MUST run 'bd devlog verify --fix' as part of the session close checklist.",
		},
	},
	{
		Version: "0.51.4",
		Date:    "2026-05-25",
		Features: []string{
			"Shared Alias Registry: Aliases are now stored in '.beads/aliases.jsonl' and shared via Git.",
			"Reproducible Graphs: Database reconstruction (bd init) now restores all aliases from the repo.",
		},
		Protocol: []string{
			"Aliases are now a shared source of truth in the repository.",
		},
	},
	{
		Version: "0.51.3",
		Date:    "2026-05-25",
		Features: []string{
			"Sticky Aliases: Entities merged via 'bd devlog alias' now stay merged even after a sync.",
			"Reversible Aliasing: New 'bd devlog unalias <name>' command to remove an alias and restore the original entity.",
		},
		Protocol: []string{
			"Use 'bd devlog unalias' to undo mistaken merges.",
		},
	},
	{
		Version: "0.51.2",
		Date:    "2026-05-25",
		Features: []string{
			"Database Migration: Added missing 'confidence' and 'source' columns to 'entity_deps' table for existing databases.",
		},
		Protocol: []string{
			"Agents SHOULD run 'bd devlog verify --fix' after upgrading to backfill relationship metadata.",
		},
	},
	{
		Version: "0.51.1",
		Date:    "2026-05-25",
		Features: []string{
			"Graph Aliasing: Collapse fragmented regex entities using 'bd devlog alias'.",
			"Opportunistic Discovery: Automatic hinting in search/graph outputs for potential entity merges.",
			"Extraction Hardening: More conservative Regex extractor with noise blacklist and confidence levels.",
			"Explicit Edge Promotion: Manual relationships in '### Architectural Relationships' are now gold-standard (1.0 confidence).",
			"Git-Based History Recovery: Automated re-anchoring of devlog sessions to new SHAs after squash-merges using patch-id.",
		},
		Protocol: []string{
			"Agents SHOULD verify '💡 OPPORTUNITY' hints in search results and run 'alias' command to maintain graph cleanliness.",
			"Prioritize manual architectural relationships (Markdown arrows) over AI-inferred ones.",
		},
	},
	{
		Version: "0.51.0",
		Date:    "2026-05-21",
		Features: []string{
			"Team-Aware Catchup: 'bd catchup' now displays Author/Actor for all activity items.",
			"Improved Summaries: Catchup prompt now groups activity by team member for better clarity.",
		},
		Protocol: []string{
			"Agents MUST group project activity by team member when generating catchup summaries.",
		},
	},
	{
		Version: "0.50.0",
		Date:    "2026-05-21",
		Features: []string{
			"Session Lifecycle States: Use 'bd devlog pause' or 'bd devlog abandon' to mark branch/task status.",
			"Proximity Warnings: Proactive warnings in 'bd ready' and 'bd resume' when working on abandoned scopes.",
			"Reachability Tracking: Native commit_sha tracking for reliable history even after squash-merges.",
			"Tool Changelog: Binary-embedded feature announcements on 'bd onboard'.",
			"Project Catchup: 'bd catchup' command to get a summary of what you missed (sessions, closed issues, states).",
		},
		Protocol: []string{
			"Agents MUST check lifecycle status ([⏸ PAUSED], [🚫 ABANDONED]) in search results before assuming code is baseline.",
			"Always read the 'Reason' for abandoned paths to avoid repeating failed historical experiments.",
			"Agents should use '_rules/_devlog/_generate-catchup.md' to summarize activity when requested.",
		},
	},
}

// GetLatest returns the latest changelog entry
func GetLatest() Entry {
	return entries[0]
}

// IsNewer returns true if v1 > v2 in semver terms (simple version)
func IsNewer(v1, v2 string) bool {
	if v2 == "" || v2 == "0.0.0" {
		return true
	}

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		n1, _ := strconv.Atoi(parts1[i])
		n2, _ := strconv.Atoi(parts2[i])

		if n1 > n2 {
			return true
		}
		if n1 < n2 {
			return false
		}
	}
	return len(parts1) > len(parts2)
}

// RenderLatest returns a formatted string of the latest changes for terminal output
func RenderLatest() string {
	latest := GetLatest()
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("✨ What's New in BeadsLog %s (%s)\n", latest.Version, latest.Date))
	sb.WriteString(strings.Repeat("-", 60) + "\n")
	
	if len(latest.Features) > 0 {
		sb.WriteString("🚀 New Features:\n")
		for _, f := range latest.Features {
			sb.WriteString(fmt.Sprintf("  • %s\n", f))
		}
	}
	
	if len(latest.Protocol) > 0 {
		sb.WriteString("\n📜 Protocol Updates (Mandatory for Agents):\n")
		for _, p := range latest.Protocol {
			sb.WriteString(fmt.Sprintf("  • %s\n", p))
		}
	}
	
	return sb.String()
}
