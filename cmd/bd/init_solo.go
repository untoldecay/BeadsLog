package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/untoldecay/BeadsLog/internal/config"
	"github.com/untoldecay/BeadsLog/internal/storage"
	"github.com/untoldecay/BeadsLog/internal/syncbranch"
	"github.com/untoldecay/BeadsLog/internal/ui"
)

// soloLocalBranch is the default local-only sync branch for solo mode.
const soloLocalBranch = "beads-local"

// applySoloConfig writes the local-only settings to .beads/config.yaml.
// sync.mode is the declarative marker (read by doctor); no-push and
// daemon.auto-sync are the enforcement.
func applySoloConfig() error {
	settings := [][2]string{
		{"sync.mode", "local-only"},
		{"no-push", "true"},
		{"daemon.auto-sync", "false"},
	}
	for _, kv := range settings {
		if err := config.SetYamlConfig(kv[0], kv[1]); err != nil {
			return fmt.Errorf("failed to set %s: %w", kv[0], err)
		}
	}
	return nil
}

// runSoloWizard configures beads for personal, local-only use in a shared repo:
// nothing beads-related is ever pushed to the remote.
//
// Returns excluded=true when the user chose to keep beads files out of git
// entirely (.git/info/exclude) — the caller must then skip hook and merge
// driver installation, since hooks would try to 'git add' excluded files.
func runSoloWizard(ctx context.Context, store storage.Storage, stealthAlreadySet bool) (excluded bool, err error) {
	fmt.Printf("\n%s %s\n\n", ui.RenderBold("bd"), ui.RenderBold("Solo Mode Setup (local-only)"))
	fmt.Println("Beads data stays on this machine: no push, no daemon auto-sync.")
	fmt.Println()

	if err := applySoloConfig(); err != nil {
		return false, err
	}
	fmt.Printf("%s sync.mode: local-only (no-push, daemon auto-sync off)\n", ui.RenderPass("✓"))

	if stealthAlreadySet {
		fmt.Printf("%s Stealth already active: beads files excluded from git\n", ui.RenderPass("✓"))
		return true, nil
	}
	if !isGitRepo() {
		return false, nil
	}

	// Without an exclude or a dedicated sync branch, the pre-commit hook stages
	// .beads/ changes into normal commits — beads data would leak to the remote
	// through regular feature-branch pushes.
	fmt.Println("\nHow should beads data relate to git?")
	fmt.Printf("  1. %s: exclude .beads/ via .git/info/exclude (recommended)\n", ui.RenderAccent("Invisible"))
	fmt.Printf("  2. %s: commit beads data to a local-only branch (%s), never pushed\n", ui.RenderAccent("Local branch"), soloLocalBranch)
	fmt.Print("\nChoice [1/2, default 1]: ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)

	if response == "2" {
		if err := syncbranch.Set(ctx, store, soloLocalBranch); err != nil {
			return false, fmt.Errorf("failed to set sync branch: %w", err)
		}
		if err := createSyncBranch(soloLocalBranch); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create sync branch: %v\n", err)
			fmt.Println("  You can create it manually: git checkout -b", soloLocalBranch)
		}
		fmt.Printf("\n%s Beads commits go to local branch %s (never pushed)\n", ui.RenderPass("✓"), ui.RenderAccent(soloLocalBranch))
		printSoloSummary(soloLocalBranch)
		return false, nil
	}

	if err := setupForkExclude(true); err != nil {
		return false, fmt.Errorf("failed to configure git exclude: %w", err)
	}
	fmt.Printf("\n%s .beads/ excluded from git — invisible to collaborators\n", ui.RenderPass("✓"))
	printSoloSummary("")
	return true, nil
}

func printSoloSummary(localBranch string) {
	fmt.Printf("\n%s %s\n\n", ui.RenderPass("✓"), ui.RenderBold("Solo setup complete!"))
	fmt.Println("How it works:")
	fmt.Println("  • Issues and devlogs work fully — locally")
	fmt.Println("  • 'bd sync' commits locally, never pushes (no-push: true)")
	fmt.Println("  • Daemon runs without auto-commit/push")
	fmt.Println("  • 'bd doctor' won't nag about unpushed beads data")
	if localBranch != "" {
		fmt.Printf("  • Beads data is versioned on local branch %s\n", ui.RenderAccent(localBranch))
	} else {
		fmt.Println("  • Teammates never see beads files (.git/info/exclude)")
	}
	fmt.Println()
	fmt.Printf("To go team mode later: %s\n", ui.RenderAccent("bd init --team --force"))
	fmt.Println()
}
