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
const CurrentVersion = "0.51.2"


var entries = []Entry{
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
