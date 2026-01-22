package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/untoldecay/BeadsLog/internal/ui"
)

// restoreCodeBlocks replaces placeholders with actual markdown code blocks
func restoreCodeBlocks(content string) string {
	content = strings.ReplaceAll(content, "[codeblock=bash]", "```bash")
	content = strings.ReplaceAll(content, "[/codeblock]", "```")
	return content
}

// initializeOrchestration scaffolds the Progressive Disclosure directory structure.

// It is safe to run multiple times (idempotent).

// Returns a list of files that were actually created or already existed.

func initializeOrchestration(verbose bool) []string {

	orchestrationDir := "_rules/_orchestration"

	var results []string

	

	// Create directory

	if err := os.MkdirAll(orchestrationDir, 0755); err != nil {

		if verbose {

			fmt.Fprintf(os.Stderr, "Error creating orchestration dir: %v\n", err)

		}

		return nil

	}

	

	if verbose {

		fmt.Printf("  %s Orchestration space: %s\n", ui.RenderPass("✓"), orchestrationDir)

	}



	// Helper to write file if not exists

	writeFile := func(name, content string) {

		path := filepath.Join(orchestrationDir, name)

		

		// Restore real markdown code blocks

		finalContent := restoreCodeBlocks(content)



		if _, err := os.Stat(path); os.IsNotExist(err) {

			if err := os.WriteFile(path, []byte(finalContent), 0644); err != nil {

				if verbose {

					fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", path, err)

				}

			} else {

				if verbose {

					fmt.Printf("    %s Created %s\n", ui.RenderPass("✓"), name)

				}

				results = append(results, fmt.Sprintf("%s (Created)", name))

			}

		} else {

			if verbose {

				fmt.Printf("    %s %s (Already exists)\n", ui.RenderPass("✓"), name)

			}

			results = append(results, fmt.Sprintf("%s (Exists)", name))

		}

	}



	writeFile("PROTOCOL.md", ProtocolMdTemplate)

	writeFile("WORKING_PROTOCOL.md", WorkingProtocolMdTemplate)

	writeFile("BEADS_REFERENCE.md", BeadsReferenceMdTemplate)

	writeFile("DEVLOG_REFERENCE.md", DevlogReferenceMdTemplate)

	

	// We create PROJECT_CONTEXT.md if it doesn't exist, but we don't overwrite it

	// because it might contain migrated content.

	writeFile("PROJECT_CONTEXT.md", ProjectContextMdTemplate)



	return results

}
