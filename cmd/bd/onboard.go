package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// candidates lists potential agent instruction files
var candidates = []string{
	"AGENTS.md",
	".windsufrules",
	".cursorrules",
	"CLAUDE.md",
	".claude/rules",
	"GEMINI.md",
	".github/copilot-instructions.md",
	".github/COPILOT-INSTRUCTIONS.md",
}

// injectProtocol reads a file, removes old bootstrap triggers, and injects the new protocol.
// It uses tags for safe replacement.
func injectProtocol(file, protocol string) {
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
	}

	modifiedTrigger := false
	for _, t := range triggers {
		if strings.Contains(strContent, t) {
			strContent = strings.ReplaceAll(strContent, t, "")
			modifiedTrigger = true
		}
	}

	// Prepare the full protocol block with tags
	fullProtocolBlock := ProtocolStartTag + "\n" + protocol + "\n" + ProtocolEndTag

	// Tag-based replacement logic
	startIndex := strings.Index(strContent, ProtocolStartTag)
	endIndex := strings.Index(strContent, ProtocolEndTag)

	var newContent string
	var action string

	if startIndex != -1 && endIndex != -1 && endIndex > startIndex {
		// Tags found: Replace content between them
		preBlock := strContent[:startIndex]
		postBlock := strContent[endIndex+len(ProtocolEndTag):]
		
		// Normalize whitespace around blocks
		preBlock = strings.TrimRight(preBlock, "\n")
		postBlock = strings.TrimLeft(postBlock, "\n")
		
		if preBlock != "" {
			newContent = preBlock + "\n\n" + fullProtocolBlock
		} else {
			newContent = fullProtocolBlock
		}
		
		if postBlock != "" {
			newContent += "\n\n" + postBlock
		}
		action = "refreshed existing protocol"

	} else {
		// Tags not found or broken: Prepend to top
		strContent = strings.TrimSpace(strContent)
		if strContent == "" {
			newContent = fullProtocolBlock
		} else {
			newContent = fullProtocolBlock + "\n\n" + strContent
		}
		action = "prepended new protocol"
	}

	// Idempotency check
	// Note: We compare trimmed versions to avoid noise from trailing newlines
	if strings.TrimSpace(newContent) == strings.TrimSpace(string(content)) && !modifiedTrigger {
		fmt.Printf("Skipping %s (protocol up to date)\n", file)
		return
	}

	if err := os.WriteFile(file, []byte(newContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to %s: %v\n", file, err)
		return
	}
	fmt.Printf("Updated %s (%s)\n", file, action)
}

// executeOnboard actively modifies agent instruction files.
func executeOnboard() error {
	// Use the embedded AgentProtocol directly
	unifiedProtocol := AgentProtocol

	found := false
	for _, file := range candidates {
		if _, err := os.Stat(file); err == nil {
			injectProtocol(file, unifiedProtocol)
			found = true
		}
	}

	if !found {
		// If no specific agent file, suggest creating AGENTS.md
		fmt.Println("No standard agent instruction file found. Creating AGENTS.md with the unified protocol...")
		fullBlock := ProtocolStartTag + "\n" + unifiedProtocol + "\n" + ProtocolEndTag
		if err := os.WriteFile("AGENTS.md", []byte(fullBlock), 0644); err != nil {
			return fmt.Errorf("error creating AGENTS.md: %w", err)
		}
		fmt.Println("✓ Created AGENTS.md")
	} else {
		fmt.Println("✓ Onboarding process completed for existing agent files.")
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
