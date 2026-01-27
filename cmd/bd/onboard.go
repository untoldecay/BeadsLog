package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/untoldecay/BeadsLog/internal/storage"
	"github.com/untoldecay/BeadsLog/internal/storage/sqlite"
	"github.com/untoldecay/BeadsLog/internal/ui"
)

// Candidates lists potential agent instruction files
var Candidates = []string{
	"AGENTS.md",
	"AGENT.md",
	".windsufrules",
	".cursorrules",
	"CLAUDE.md",
	".claude/rules",
	"GEMINI.md",
	"CODEBASE.md",
	".github/copilot-instructions.md",
	".github/COPILOT-INSTRUCTIONS.md",
}

// AgentTool represents a logical AI tool that may look for multiple file paths
type AgentTool struct {
	Name  string
	Files []string
}

// AgentToolCandidates groups configuration files by their respective tools
var AgentToolCandidates = []AgentTool{
	{Name: "Standard Agent", Files: []string{"AGENTS.md", "AGENT.md"}},
	{Name: "Windsurf", Files: []string{".windsufrules"}},
	{Name: "Cursor", Files: []string{".cursorrules"}},
	{Name: "Claude", Files: []string{"CLAUDE.md", ".claude/rules"}},
	{Name: "Gemini", Files: []string{"GEMINI.md"}},
	{Name: "Codebase", Files: []string{"CODEBASE.md"}},
	{Name: "GitHub Copilot", Files: []string{".github/copilot-instructions.md", ".github/COPILOT-INSTRUCTIONS.md"}},
}

// finalizeOnboarding unlocks the environment by installing the Full Bootloader.
// It searches for either the Protocol tags OR the initial bootstrap trap.
func finalizeOnboarding(ctx context.Context, store storage.Storage) {
	if store == nil {
		return
	}

	finalized, _ := store.GetConfig(ctx, "onboarding_finalized")
	if finalized == "true" {
		return
	}

	bootstrapTrigger := "BEFORE ANYTHING ELSE: run 'bd onboard' and follow ALL instructions to initialize your session."
	fullBootloader := restoreCodeBlocks(FullBootloader)
	fullBlock := ProtocolStartTag + "\n" + fullBootloader + "\n" + ProtocolEndTag

	found := false
	for _, file := range Candidates {
		if _, err := os.Stat(file); err == nil {
			content, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			sContent := string(content)
			var newContent string
			modified := false

			if strings.Contains(sContent, ProtocolStartTag) {
				// Case 1: Already has tags, just update between them
				startIndex := strings.Index(sContent, ProtocolStartTag)
				endIndex := strings.Index(sContent, ProtocolEndTag)
				if startIndex != -1 && endIndex != -1 && endIndex > startIndex {
					pre := sContent[:startIndex]
					post := sContent[endIndex+len(ProtocolEndTag):]
					newContent = pre + fullBlock + post
					modified = true
				}
			} else if strings.Contains(sContent, bootstrapTrigger) {
				// Case 2: Fresh init, replace the trap line
				newContent = strings.Replace(sContent, bootstrapTrigger, fullBlock, 1)
				modified = true
			} else {
				// Case 3: File exists but missing trigger/protocol - PREPEND full protocol
				// This handles cases where onboarding was somehow missed but we are finalizing
				newContent = fullBlock + "\n" + sContent
				modified = true
			}

			if modified {
				if err := os.WriteFile(file, []byte(newContent), 0644); err == nil {
					found = true
				}
			}
		}
	}

	if found {
		_ = store.SetConfig(ctx, "onboarding_finalized", "true")
		fmt.Printf("\n%s Session Initialized. Agent instructions updated with full project context.\n", ui.RenderPass("✅"))
	}
}

// executeOnboard acts as a guide for the agent
func executeOnboard(ctx context.Context, store storage.Storage) error {
	// Ensure orchestration structure exists
	initializeOrchestration(false)

	// Establish store for config checks
	if daemonClient != nil && store == nil && dbPath != "" {
		var err error
		store, err = sqlite.New(ctx, dbPath)
		if err == nil {
			defer store.Close()
		}
	}

	// Ensure agent rules are configured (inject bootstrap trigger if missing)
	configureAgentRules(false, true, Candidates)

	// Refresh devlog instructions if config changed
	if sqliteStore, ok := store.(*sqlite.SQLiteStorage); ok {
		refreshDevlogPrompt(sqliteStore)
	}

	fmt.Printf("\n%s BeadsLog Activation Guide\n", ui.RenderAccent("🚀"))

	if store != nil {
		_ = store.SetConfig(ctx, "onboarding_finalized", "false")
	}

	return nil
}

var onboardCmd = &cobra.Command{
	Use:     "onboard",
	GroupID: "setup",
	Short:   "Guide the agent through session initialization",
	Long: `This command provides a step-by-step guide for agents to activate
	their session. Following the guide will lead to the 'bd ready' command,
	which unlocks the full project context in the agent's instructions.`,
	Run: func(cmd *cobra.Command, args []string) {
			if err := executeOnboard(rootCtx, store); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
			}
		},
}

func init() {
	rootCmd.AddCommand(onboardCmd)
}