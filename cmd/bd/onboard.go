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

// migrateAndInjectRestrictedProtocol handles the Progressive Disclosure migration.
// 1. Extracts legacy content (anything outside BD_PROTOCOL tags).
// 2. Moves legacy content to _rules/_orchestration/PROJECT_CONTEXT.md.
// 3. Replaces agent file content with the RESTRICTED Bootloader Protocol.
// 4. Sets onboarding_finalized = false in the database.
func migrateAndInjectRestrictedProtocol(ctx context.Context, store storage.Storage, file string) {
	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", file, err)
		return
	}

	strContent := string(content)

	// Self-healing: Remove bootstrap triggers
	triggers := []string{
		"BEFORE ANYTHING ELSE: run 'bd devlog onboard'",
		"BEFORE ANYTHING ELSE: run 'bd onboard'",
		"> **IMPORTANT:** Run `bd onboard` to configure this agent.",
	}

	for _, t := range triggers {
		if strings.Contains(strContent, t) {
			strContent = strings.ReplaceAll(strContent, t, "")
		}
	}

	// Identify Legacy Content
	startIndex := strings.Index(strContent, ProtocolStartTag)
	endIndex := strings.Index(strContent, ProtocolEndTag)

	var legacyContent string
	if startIndex != -1 && endIndex != -1 && endIndex > startIndex {
		// Content BEFORE the tag + Content AFTER the tag
		preBlock := strContent[:startIndex]
		postBlock := strContent[endIndex+len(ProtocolEndTag):]
		legacyContent = strings.TrimSpace(preBlock + "\n" + postBlock)
	} else {
		// No tags? The whole file is legacy content (unless it's empty)
		legacyContent = strings.TrimSpace(strContent)
	}

	// Move Legacy Content to PROJECT_CONTEXT.md
	if legacyContent != "" {
		contextPath := "_rules/_orchestration/PROJECT_CONTEXT.md"
		
		// Ensure directory exists (idempotent)
		initializeOrchestration(false)

		// Read existing context
		existingContext, err := os.ReadFile(contextPath)
		if err == nil {
			// Check if we already migrated this content (simple string check)
			if !strings.Contains(string(existingContext), legacyContent) {
				// Append with a header
				header := fmt.Sprintf("\n\n## Legacy Content from %s\n", file)
				newContext := string(existingContext) + header + legacyContent
				if err := os.WriteFile(contextPath, []byte(newContext), 0644); err != nil {
					fmt.Fprintf(os.Stderr, "Error updating context file: %v\n", err)
				} else {
					fmt.Printf("  %s Migrated legacy content from %s to %s\n", ui.RenderPass("✓"), file, contextPath)
				}
			}
		} else {
			// Create new (should have been created by initializeOrchestration, but safety first)
			header := fmt.Sprintf("\n\n## Legacy Content from %s\n", file)
			baseContent := restoreCodeBlocks(ProjectContextMdTemplate)
			newContext := baseContent + header + legacyContent
			if err := os.WriteFile(contextPath, []byte(newContext), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing context file: %v\n", err)
			} else {
				fmt.Printf("  %s Migrated legacy content from %s to %s\n", ui.RenderPass("✓"), file, contextPath)
			}
		}
	}

	// Inject Restricted Bootloader (Overwrite agent file)
	// This forces the agent to run the protocol to unlock the context.
	finalBootloader := restoreCodeBlocks(RestrictedBootloader)
	fullBlock := ProtocolStartTag + "\n" + finalBootloader + "\n" + ProtocolEndTag
	
	if err := os.WriteFile(file, []byte(fullBlock), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to %s: %v\n", file, err)
		return
	}
	fmt.Printf("  %s Updated %s (Restricted Bootloader installed)\n", ui.RenderPass("✓"), file)

	// Set flag in database
	if store != nil {
		_ = store.SetConfig(ctx, "onboarding_finalized", "false")
	}
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

// executeOnboard actively modifies agent instruction files.
func executeOnboard(ctx context.Context, store storage.Storage) error {
	// Ensure orchestration structure exists
	initializeOrchestration(false)

	found := false
	for _, file := range Candidates {
		if _, err := os.Stat(file); err == nil {
			migrateAndInjectRestrictedProtocol(ctx, store, file)
			found = true
		}
	}

	if !found {
		// If no specific agent file, suggest creating AGENTS.md
		fmt.Println("No standard agent instruction file found. Creating AGENTS.md with restricted protocol...")
		
		finalBootloader := restoreCodeBlocks(RestrictedBootloader)
		fullBlock := ProtocolStartTag + "\n" + finalBootloader + "\n" + ProtocolEndTag
		
		if err := os.WriteFile("AGENTS.md", []byte(fullBlock), 0644); err != nil {
			return fmt.Errorf("error creating AGENTS.md: %w", err)
		}
		if store != nil {
			_ = store.SetConfig(ctx, "onboarding_finalized", "false")
		}
		fmt.Printf("  %s Created AGENTS.md\n", ui.RenderPass("✓"))
	}

	fmt.Printf("\n%s Phase 1 complete. Your environment is partially configured.\n", ui.RenderPass("✓"))
	fmt.Printf("👉 **MANDATORY NEXT STEP:** Follow the sequence in %s to unlock project context.\n", ui.RenderAccent("_rules/_orchestration/PROTOCOL.md"))

	return nil
}

var onboardCmd = &cobra.Command{
	Use:     "onboard",
	GroupID: "setup",
	Short:   "Set up agent instruction files for Beads and Devlog integration",
	Long: `This command actively modifies agent instruction files (e.g., AGENTS.md)
	to integrate Beads and Beads Devlog workflows. It injects a restricted
	protocol that guides agents to initialize their session.

	Once the initialization protocol is complete, the full project context will be unlocked.`,
	Run: func(cmd *cobra.Command, args []string) {
			if err := executeOnboard(rootCtx, store); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
			}
		},
}

func init() {
	rootCmd.AddCommand(onboardCmd)
}