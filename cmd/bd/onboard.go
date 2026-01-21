package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/untoldecay/BeadsLog/internal/storage"
	"github.com/untoldecay/BeadsLog/internal/ui"
)

// Candidates lists potential agent instruction files
var Candidates = []string{
	"AGENTS.md",
	".windsufrules",
	".cursorrules",
	"CLAUDE.md",
	".claude/rules",
	"GEMINI.md",
	".github/copilot-instructions.md",
	".github/COPILOT-INSTRUCTIONS.md",
}

// finalizeOnboarding unlocks the environment by installing the Full Bootloader
func finalizeOnboarding(ctx context.Context, store storage.Storage) {
	if store == nil {
		return
	}

	finalized, _ := store.GetConfig(ctx, "onboarding_finalized")
	if finalized == "true" {
		return
	}

	// Find and update all agent files
	found := false
	for _, file := range Candidates {
		if _, err := os.Stat(file); err == nil {
			content, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			// Check if it's our bootloader (has protocol tags)
			if strings.Contains(string(content), ProtocolStartTag) {
				finalBootloader := restoreCodeBlocks(FullBootloader)
				fullBlock := ProtocolStartTag + "\n" + finalBootloader + "\n" + ProtocolEndTag
				
				if err := os.WriteFile(file, []byte(fullBlock), 0644); err == nil {
					found = true
				}
			}
		}
	}

	if found {
		_ = store.SetConfig(ctx, "onboarding_finalized", "true")
		fmt.Printf("\n%s Session Initialized. Project context is now unlocked in your agent instructions.\n", ui.RenderPass("✅"))
	}
}

// executeOnboard generates the copy-paste protocol for the agent
func executeOnboard(ctx context.Context, store storage.Storage) error {
	// Ensure orchestration structure exists (for manually following the instructions)
	initializeOrchestration(false)

	// Prepare the Restricted Bootloader snippet
	finalBootloader := restoreCodeBlocks(RestrictedBootloader)
	fullSnippet := ProtocolStartTag + "\n" + finalBootloader + "\n" + ProtocolEndTag

	fmt.Printf("\n%s BeadsLog Onboarding - Copy-Paste Protocol\n", ui.RenderAccent("🏰"))
	fmt.Printf("------------------------------------------------------------\n")
	fmt.Printf("To integrate with this repository, you must update your instructions.\n\n")
	
	fmt.Printf("👉 **STEP 1: Copy this block to your instruction file**\n")
	fmt.Printf("Target files (choose yours): %s\n\n", strings.Join(Candidates, ", "))
	
	fmt.Printf("```markdown\n%s\n```\n\n", fullSnippet)

	fmt.Printf("👉 **STEP 2: Delete the 'run bd onboard' note from that file.**\n\n")

	fmt.Printf("👉 **STEP 3: Execute the Activation Sequence**\n")
	fmt.Printf("Read %s and run the commands inside to unlock full context:\n\n", ui.RenderAccent("_rules/_orchestration/PROTOCOL.md"))
	
	fmt.Printf("```bash\n")
	fmt.Printf("bd sync\n")
	fmt.Printf("bd devlog verify --fix\n")
	fmt.Printf("bd devlog sync\n")
	fmt.Printf("bd ready\n")
	fmt.Printf("```\n")

	// Set state to uninitialized
	if store != nil {
		_ = store.SetConfig(ctx, "onboarding_finalized", "false")
	}

	return nil
}

var onboardCmd = &cobra.Command{
	Use:     "onboard",
	GroupID: "setup",
	Short:   "Generate integration instructions for the agent",
	Long: `This command generates the necessary Markdown snippets to integrate
	Beads and Beads Devlog into your agent instruction files.
	
	Copy the generated snippet to your primary rule file (e.g. AGENTS.md)
	to begin the activation process.`,
	Run: func(cmd *cobra.Command, args []string) {
			if err := executeOnboard(rootCtx, store); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
			}
		},
}

func init() {
	rootCmd.AddCommand(onboardCmd)
}
