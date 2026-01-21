package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
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

// migrateAndInjectProtocol handles the Progressive Disclosure migration.
// 1. Extracts legacy content (anything outside BD_PROTOCOL tags).
// 2. Moves legacy content to _rules/_orchestration/PROJECT_CONTEXT.md.
// 3. Replaces agent file content with the Bootloader Protocol.
func migrateAndInjectProtocol(file string) {
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
			// If initializeOrchestration created the default template, we append to it.
			header := fmt.Sprintf("\n\n## Legacy Content from %s\n", file)
			
			// Use restoreCodeBlocks on the template just in case
			baseContent := restoreCodeBlocks(ProjectContextMdTemplate)
			
			newContext := baseContent + header + legacyContent
			if err := os.WriteFile(contextPath, []byte(newContext), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing context file: %v\n", err)
			} else {
				fmt.Printf("  %s Migrated legacy content from %s to %s\n", ui.RenderPass("✓"), file, contextPath)
			}
		}
	}

	// Inject Bootloader (Overwrite agent file)
	// The bootloader is purely the protocol block.
	// Restore blocks just in case AgentProtocol uses them in future
	finalBootloader := restoreCodeBlocks(AgentProtocol)
	fullBlock := ProtocolStartTag + "\n" + finalBootloader + "\n" + ProtocolEndTag
	
	// Idempotency check
	if strings.TrimSpace(string(content)) == strings.TrimSpace(fullBlock) {
		fmt.Printf("  %s %s (Already up to date)\n", ui.RenderPass("✓"), file)
		return
	}

	if err := os.WriteFile(file, []byte(fullBlock), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to %s: %v\n", file, err)
		return
	}
	fmt.Printf("  %s Updated %s (Bootloader installed)\n", ui.RenderPass("✓"), file)
}

// executeOnboard actively modifies agent instruction files.
func executeOnboard() error {
	// Ensure orchestration structure exists
	initializeOrchestration(false)

	found := false
	for _, file := range Candidates {
		if _, err := os.Stat(file); err == nil {
			migrateAndInjectProtocol(file)
			found = true
		}
	}

	if !found {
		// If no specific agent file, suggest creating AGENTS.md
		fmt.Println("No standard agent instruction file found. Creating AGENTS.md with the unified protocol...")
		
		finalBootloader := restoreCodeBlocks(AgentProtocol)
		fullBlock := ProtocolStartTag + "\n" + finalBootloader + "\n" + ProtocolEndTag
		
		if err := os.WriteFile("AGENTS.md", []byte(fullBlock), 0644); err != nil {
			return fmt.Errorf("error creating AGENTS.md: %w", err)
		}
		fmt.Printf("  %s Created AGENTS.md\n", ui.RenderPass("✓"))
	} else {
		fmt.Printf("\n%s Onboarding complete. Agents are now using Progressive Disclosure.\n", ui.RenderPass("✓"))
	}

	return nil
}

var onboardCmd = &cobra.Command{
	Use:     "onboard",
	GroupID: "setup",
	Short:   "Set up agent instruction files for Beads and Devlog integration",
	Long: `This command actively modifies agent instruction files (e.g., AGENTS.md)
	to integrate Beads and Beads Devlog workflows. It injects a unified
	protocol that guides agents on issue tracking, session memory, and proper
	workflow.

	This approach replaces the old method of printing instructions for manual
	copy-pasting, ensuring consistency and correctness across agent setups.`,
	Run: func(cmd *cobra.Command, args []string) {
			if err := executeOnboard(); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
			}
		},
}

func init() {
	rootCmd.AddCommand(onboardCmd)
}
