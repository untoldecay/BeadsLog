package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/untoldecay/BeadsLog/internal/config"
	"github.com/untoldecay/BeadsLog/internal/debug"
)

var checkCmd = &cobra.Command{
	Use:    "check",
	Hidden: true, // Internal command for hooks/CI
	Short:  "Perform compliance checks",
	Run: func(cmd *cobra.Command, args []string) {
		// Hook selection logic
		hook, _ := cmd.Flags().GetString("hook")
		if hook == "pre-commit" {
			runPreCommitCheck()
		} else {
			fmt.Fprintf(os.Stderr, "Error: --hook flag required (supported: pre-commit)\n")
			os.Exit(1)
		}
	},
}

func init() {
	checkCmd.Flags().String("hook", "", "The hook context to run checks for (e.g., pre-commit)")
	rootCmd.AddCommand(checkCmd)
}

// runPreCommitCheck implements the logic for verifying devlog updates before commit
func runPreCommitCheck() {
	// 1. Check Configuration
	if !config.GetDevlogEnforceOnCommit() {
		debug.Logf("Devlog enforcement disabled, skipping check.")
		os.Exit(0)
	}

	debug.Logf("Running pre-commit devlog compliance check...")

	// 2. Identify Devlog Directory
	devlogDir := config.GetDevlogDir()
	if devlogDir == "" {
		devlogDir = "_rules/_devlog" // Default fallback
	}

	// 3. Check for Staged Changes in Devlog Directory
	// Use --name-only to get list of staged files
	// Use --cached to look at index (staged files)
	cmd := exec.Command("git", "diff", "--name-only", "--cached")
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking git status: %v\n", err)
		os.Exit(1) // Fail open or closed? Here we fail closed to be safe.
	}

	stagedFiles := strings.Split(string(output), "\n")
	hasDevlogChange := false
	hasIndexChange := false
	hasCodeChange := false

	// Define what constitutes "code" or "work" that requires a log
	// We skip .gitignore, README.md, and purely automated files if needed
	// For now, strict: ANY staged file counts as work.

	for _, file := range stagedFiles {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}

		// Normalize paths
		cleanFile := filepath.Clean(file)
		cleanDevlogDir := filepath.Clean(devlogDir)

		// Check if file is inside devlog directory
		if strings.HasPrefix(cleanFile, cleanDevlogDir) {
			hasDevlogChange = true
			if strings.HasSuffix(cleanFile, "_index.md") {
				hasIndexChange = true
			}
		} else {
			// Ignore .beads directory changes (metadata updates)
			if strings.HasPrefix(cleanFile, ".beads"+string(filepath.Separator)) {
				continue
			}
			// Ignore .gitignore and other common non-code files if desired?
			// For now, let's keep it strict but allow metadata syncs.
			hasCodeChange = true
		}
	}

	// 4. Logic Decision
	if !hasCodeChange {
		// No code/docs changed? Maybe just a devlog update or something else.
		// If they are ONLY updating the devlog, we let it pass.
		// If they are updating nothing (empty commit), git usually blocks it anyway.
		debug.Logf("No code changes detected, skipping enforcement.")
		os.Exit(0)
	}

	if hasDevlogChange {
		// We have code changes AND devlog changes.
		// Strict check: Did they update the index?
		if !hasIndexChange {
			// This is a warning sign. Protocol requires index update.
			// But maybe they just added a log file?
			// Let's enforce index update too, as per "The Agent Trap".
			printBlockerMessage("The '_index.md' file must be updated to reference your new session log.")
			os.Exit(1)
		}

		// Verify that _index.md actually has *content* changes (not just touched)
		// git diff --cached --exit-code returns 1 if differences exist, 0 if no diff.
		// BUT we want to check specifically for _index.md.
		// If we are here, we know _index.md is in the list of staged files.
		// Git normally only lists files with actual content changes in diff --name-only.
		// So we are likely good.
		debug.Logf("Devlog update verified. Commit allowed.")
		os.Exit(0)
	}

	// 5. Failure Case: Code changed, but no Devlog change.
	printBlockerMessage("No meaningful devlog update detected.")
	os.Exit(1)
}

func printBlockerMessage(reason string) {
	devlogDir := config.GetDevlogDir()
	if devlogDir == "" {
		devlogDir = "_rules/_devlog"
	}

	msg := fmt.Sprintf(`
❌ BLOCKER: Devlog Update Missing

Reason: %s

This project requires a devlog update for every commit that modifies code.

👉 AGENT ACTION REQUIRED:
1. Run the devlog generation prompt:
   cat %s/_generate-devlog.md
2. Generate the session log (YYYY-MM-DD_*.md).
3. Update the index (%s/_index.md).
4. Stage these files:
   git add %s/
5. Retry the commit.

(To bypass: use 'git commit --no-verify' if you are sure)
`, reason, devlogDir, devlogDir, devlogDir)

	fmt.Fprintln(os.Stderr, msg)
}
