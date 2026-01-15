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

// injectProtocol reads a file, removes old bootstrap triggers, and appends the new protocol.
// It includes idempotency checks.
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
	
	modified := false
	for _, t := range triggers {
		if strings.Contains(strContent, t) {
			strContent = strings.ReplaceAll(strContent, t, "")
			modified = true
		}
	}
	
	if modified {
		// Clean up potentially empty lines left behind, but preserve file structure generally
		strContent = strings.TrimSpace(strContent)
		if len(strContent) > 0 {
			strContent += "\n"
		}
	}

	// Idempotency: Don't inject if already present
	// For this issue, we will just check for the basic Devlog Protocol header.
	// More robust idempotency will be handled in later issues.
	if strings.Contains(strContent, "## Devlog Protocol (MANDATORY)") {
		fmt.Printf("Skipping %s (protocol already present)\n", file)
		// Still save to apply the bootstrap removal if happened
		if modified {
			if err := os.WriteFile(file, []byte(strContent), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", file, err)
			}
		}
		return
	}

	// Clean slate or Append
	var newContent string
	if strings.TrimSpace(strContent) == "" {
		newContent = protocol
	} else {
		newContent = strContent + "\n" + protocol
	}

	if err := os.WriteFile(file, []byte(newContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error appending to %s: %v\n", file, err)
		return
	}
	fmt.Printf("Updated %s\n", file)
}


// executeOnboard will contain the logic to actively modify agent instruction files.
func executeOnboard() error {
	protocolFilePath := "_rules/AGENTS.md.protocol" // Path to the external protocol file

	protocolContentBytes, err := os.ReadFile(protocolFilePath)
	if err != nil {
		return fmt.Errorf("failed to read unified agent protocol from %s: %w", protocolFilePath, err)
	}
	unifiedProtocol := string(protocolContentBytes)

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
		if err := os.WriteFile("AGENTS.md", []byte(unifiedProtocol), 0644); err != nil {
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