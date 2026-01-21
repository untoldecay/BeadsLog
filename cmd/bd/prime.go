package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/untoldecay/BeadsLog"
	"github.com/untoldecay/BeadsLog/internal/config"
	"github.com/untoldecay/BeadsLog/internal/rpc"
	"github.com/untoldecay/BeadsLog/internal/syncbranch"
)

// isDaemonAutoSyncing checks if daemon is running with auto-commit and auto-push enabled.
// Returns false if daemon is not running or check fails (fail-safe to show full protocol).
// This is a variable to allow stubbing in tests.
var isDaemonAutoSyncing = func() bool {
	beadsDir := beads.FindBeadsDir()
	if beadsDir == "" {
		return false
	}

	socketPath := filepath.Join(beadsDir, "bd.sock")
	client, err := rpc.TryConnect(socketPath)
	if err != nil || client == nil {
		return false
	}
	defer func() { _ = client.Close() }()

	status, err := client.Status()
	if err != nil {
		return false
	}

	// Only check auto-commit and auto-push (auto-pull is separate)
	return status.AutoCommit && status.AutoPush
}

var (
	primeFullMode    bool
	primeMCPMode     bool
	primeStealthMode bool
	primeExportMode  bool
)

var primeCmd = &cobra.Command{
	Use:     "prime",
	GroupID: "setup",
	Short:   "Output AI-optimized workflow context",
	Long: `Output essential Beads workflow context in AI-optimized markdown format.

Automatically detects if MCP server is active and adapts output:
- MCP mode: Brief workflow reminders (~50 tokens)
- CLI mode: Full command reference (~1-2k tokens)

Designed for Claude Code hooks (SessionStart, PreCompact) to prevent
agents from forgetting bd workflow after context compaction.

Config options:
- no-git-ops: When true, outputs stealth mode (no git commands in session close protocol).
  Set via: bd config set no-git-ops true
  Useful when you want to control when commits happen manually.

Workflow customization:
- Place a .beads/PRIME.md file to override the default output entirely.
- Use --export to dump the default content for customization.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Find .beads/ directory (supports both database and JSONL-only mode)
		beadsDir := beads.FindBeadsDir()
		if beadsDir == "" {
			// Not in a beads project - silent exit with success
			// CRITICAL: No stderr output, exit 0
			// This enables cross-platform hook integration
			os.Exit(0)
		}

		// Detect MCP mode (unless overridden by flags)
		mcpMode := isMCPActive()
		if primeFullMode {
			mcpMode = false
		}
		if primeMCPMode {
			mcpMode = true
		}

		// Check for stealth mode: flag OR config (GH#593)
		// This allows users to disable git ops in session close protocol via config
		stealthMode := primeStealthMode || config.GetBool("no-git-ops")

		// Check for custom PRIME.md override (unless --export flag)
		// This allows users to fully customize workflow instructions
		// Check local .beads/ first (even if redirected), then redirected location
		if !primeExportMode {
			localPrimePath := filepath.Join(".beads", "PRIME.md")
			redirectedPrimePath := filepath.Join(beadsDir, "PRIME.md")

			// Try local first (user's clone-specific customization)
			// #nosec G304 -- path is relative to cwd
			if content, err := os.ReadFile(localPrimePath); err == nil {
				fmt.Print(string(content))
				return
			}
			// Fall back to redirected location (shared customization)
			// #nosec G304 -- path is constructed from beadsDir which we control
			if content, err := os.ReadFile(redirectedPrimePath); err == nil {
				fmt.Print(string(content))
				return
			}
		}

		// Output workflow context (adaptive based on MCP and stealth mode)
		if err := outputPrimeContext(os.Stdout, mcpMode, stealthMode); err != nil {
			// Suppress all errors - silent exit with success
			// Never write to stderr (breaks Windows compatibility)
			os.Exit(0)
		}
	},
}

func init() {
	primeCmd.Flags().BoolVar(&primeFullMode, "full", false, "Force full CLI output (ignore MCP detection)")
	primeCmd.Flags().BoolVar(&primeMCPMode, "mcp", false, "Force MCP mode (minimal output)")
	primeCmd.Flags().BoolVar(&primeStealthMode, "stealth", false, "Stealth mode (no git operations, flush only)")
	primeCmd.Flags().BoolVar(&primeExportMode, "export", false, "Output default content (ignores PRIME.md override)")
	rootCmd.AddCommand(primeCmd)
}

// isMCPActive detects if MCP server is currently active
func isMCPActive() bool {
	// Get home directory with fallback
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to HOME environment variable
		home = os.Getenv("HOME")
		if home == "" {
			// Can't determine home directory, assume no MCP
			return false
		}
	}

	settingsPath := filepath.Join(home, ".claude/settings.json")
	// #nosec G304 -- settings path derived from user home directory
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return false
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return false
	}

	// Check mcpServers section for beads
	mcpServers, ok := settings["mcpServers"].(map[string]interface{})
	if !ok {
		return false
	}

	// Look for beads server (any key containing "beads")
	for key := range mcpServers {
		if strings.Contains(strings.ToLower(key), "beads") {
			return true
		}
	}

	return false
}

// isEphemeralBranch detects if current branch has no upstream (ephemeral/local-only)
var isEphemeralBranch = func() bool {
	// git rev-parse --abbrev-ref --symbolic-full-name @{u}
	// Returns error code 128 if no upstream configured
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	err := cmd.Run()
	return err != nil
}

// primeHasGitRemote detects if any git remote is configured (stubbable for tests)
var primeHasGitRemote = func() bool {
	return syncbranch.HasGitRemote(context.Background())
}

// getRedirectNotice returns a notice string if beads is redirected
func getRedirectNotice(verbose bool) string {
	redirectInfo := beads.GetRedirectInfo()
	if !redirectInfo.IsRedirected {
		return ""
	}

	if verbose {
		return fmt.Sprintf(`> ⚠️ **Redirected**: Local .beads → %s
> You share issues with other clones using this redirect.

`, redirectInfo.TargetDir)
	}
	return fmt.Sprintf("**Note**: Beads redirected to %s (shared with other clones)\n\n", redirectInfo.TargetDir)
}

// outputPrimeContext outputs workflow context in markdown format
func outputPrimeContext(w io.Writer, mcpMode bool, stealthMode bool) error {
	if mcpMode {
		return outputMCPContext(w, stealthMode)
	}
	return outputCLIContext(w, stealthMode)
}

// getSessionCloseProtocol returns the dynamic checklist for session end
func getSessionCloseProtocol(stealthMode bool) (string, string, string) {
	ephemeral := isEphemeralBranch()
	noPush := config.GetBool("no-push")
	autoSync := isDaemonAutoSyncing()
	localOnly := !primeHasGitRemote()

	var protocol, note, workflow string

	if stealthMode || localOnly {
		protocol = `[ ] bd sync --flush-only    (export beads to JSONL only)`
		workflow = "```bash\nbd close <id1> <id2> ...    # Close all completed issues\nbd sync --flush-only        # Export to JSONL\n```"
		if localOnly && !stealthMode {
			note = "**Note:** No git remote configured. Issues are saved locally only."
		} else {
			note = "**Note:** Stealth mode active (no git operations)."
		}
	} else if autoSync && !ephemeral && !noPush {
		protocol = `[ ] 1. git status              (check what changed)
[ ] 2. git add <files>         (stage code changes)
[ ] 3. git commit -m "..."     (commit code)
[ ] 4. git push                (push to remote)`
		note = "**Note:** Daemon is auto-syncing beads changes. No manual `bd sync` needed."
		workflow = "```bash\nbd close <id1> <id2> ...    # Close all completed issues\ngit push                    # Push to remote (beads auto-synced)\n```"
	} else if ephemeral {
		protocol = `[ ] 1. git status              (check what changed)
[ ] 2. git add <files>         (stage code changes)
[ ] 3. bd sync --from-main     (pull beads updates from main)
[ ] 4. git commit -m "..."     (commit code changes)`
		note = "**Note:** Ephemeral branch (no upstream). Code is merged to main locally, not pushed."
		workflow = "```bash\nbd close <id1> <id2> ...    # Close all completed issues\nbd sync --from-main         # Pull latest beads from main\ngit add . && git commit -m \"...\"\n```"
	} else if noPush {
		protocol = `[ ] 1. git status              (check what changed)
[ ] 2. git add <files>         (stage code changes)
[ ] 3. bd sync                 (commit beads changes)
[ ] 4. git commit -m "..."     (commit code)
[ ] 5. bd sync                 (commit any new beads changes)`
		note = "**Note:** Push disabled via config. Run `git push` manually when ready."
		workflow = "```bash\nbd close <id1> <id2> ...    # Close all completed issues\nbd sync                     # Sync beads (push disabled)\n```"
	} else {
		protocol = `[ ] 1. git status              (check what changed)
[ ] 2. git add <files>         (stage code changes)
[ ] 3. bd sync                 (commit beads changes)
[ ] 4. git commit -m "..."     (commit code)
[ ] 5. bd sync                 (commit any new beads changes)
[ ] 6. git push                (push to remote)`
		note = "**NEVER skip this.** Work is not done until pushed."
		workflow = "```bash\nbd close <id1> <id2> ...    # Close all completed issues\nbd sync                     # Push to remote\n```"
	}

	return protocol, note, workflow
}

// outputMCPContext outputs minimal context for MCP users
func outputMCPContext(w io.Writer, stealthMode bool) error {
	protocol, _, _ := getSessionCloseProtocol(stealthMode)
	// Clean up protocol for single line if possible
	protocol = strings.ReplaceAll(protocol, "\n", " → ")
	protocol = strings.ReplaceAll(protocol, "[ ] ", "")

	redirectNotice := getRedirectNotice(false)

	context := `# Beads Issue Tracker Active

` + redirectNotice + `# 🚨 SESSION CLOSE PROTOCOL 🚨
Before saying "done": ` + protocol + `

## Core Rules
- Track strategic work in beads (multi-session, dependencies)
- Load WORKING_PROTOCOL.md for task loop
- When in doubt, prefer bd—persistence beats lost context

Start: Check ` + "`ready`" + ` tool for available work.
`
	_, _ = fmt.Fprint(w, context)
	return nil
}

// outputCLIContext outputs full CLI reference adapted for Progressive Disclosure
func outputCLIContext(w io.Writer, stealthMode bool) error {
	// Check if onboarding is finalized
	finalized := "false"
	if daemonClient != nil {
		resp, err := daemonClient.GetConfig(&rpc.GetConfigArgs{Key: "onboarding_finalized"})
		if err == nil && resp != nil {
			finalized = resp.Value
		}
	} else if store != nil {
		finalized, _ = store.GetConfig(rootCtx, "onboarding_finalized")
	}
	// fmt.Fprintf(os.Stderr, "DEBUG: finalized=[%s] store=%v\n", finalized, store != nil)

	bootloader := restoreCodeBlocks(FullBootloader)
	if strings.TrimSpace(finalized) != "true" {
		bootloader = restoreCodeBlocks(RestrictedBootloader)
	}

	// Try to load WORKING_PROTOCOL.md
	workingProtocol := ""
	if data, err := os.ReadFile("_rules/_orchestration/WORKING_PROTOCOL.md"); err == nil {
		workingProtocol = "\n---\n\n" + string(data)
	}

	protocol, note, workflow := getSessionCloseProtocol(stealthMode)
	redirectNotice := getRedirectNotice(true)

	context := bootloader + workingProtocol + `

` + redirectNotice + `
# 🚨 SESSION CLOSE PROTOCOL 🚨

**CRITICAL**: Before saying "done" or "complete", you MUST run this checklist:

` + "```" + `
` + protocol + `
` + "```" + `

` + note + `

## Completing Work
` + workflow + `
`
	_, _ = fmt.Fprint(w, context)
	return nil
}
