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
const CurrentVersion = "0.50.0"

var entries = []Entry{
	{
		Version: "0.50.0",
		Date:    "2026-05-19",
		Features: []string{
			"Session Lifecycle States: Use 'bd devlog pause' or 'bd devlog abandon' to mark branch/task status.",
			"Proximity Warnings: Proactive warnings in 'bd ready' and 'bd resume' when working on abandoned scopes.",
			"Reachability Tracking: Native commit_sha tracking for reliable history even after squash-merges.",
		},
		Protocol: []string{
			"Agents MUST check lifecycle status ([⏸ PAUSED], [🚫 ABANDONED]) in search results before assuming code is baseline.",
			"Always read the 'Reason' for abandoned paths to avoid repeating failed historical experiments.",
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
