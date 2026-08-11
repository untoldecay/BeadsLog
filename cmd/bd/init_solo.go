package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/untoldecay/BeadsLog/internal/config"
	"github.com/untoldecay/BeadsLog/internal/storage"
	"github.com/untoldecay/BeadsLog/internal/syncbranch"
	"github.com/untoldecay/BeadsLog/internal/ui"
)

// soloLocalBranch is the default local-only sync branch for solo mode.
const soloLocalBranch = "beads-local"

// applySoloConfig writes the local-only settings. sync.mode is the declarative
// marker (read by doctor); no-push and daemon.auto-sync are the enforcement.
// When local is true they go to the git-excluded config.local.yaml so they
// never leak into the committed config.yaml (invisible/stealth modes); when
// false they go to config.yaml (local-branch mode commits them to a
// never-pushed branch; non-git repos have nothing to leak to). (BeadsLog-9vd)
func applySoloConfig(local bool) error {
	settings := [][2]string{
		{"sync.mode", "local-only"},
		{"no-push", "true"},
		{"daemon.auto-sync", "false"},
	}
	set := config.SetYamlConfig
	if local {
		set = config.SetLocalYamlConfig
	}
	for _, kv := range settings {
		if err := set(kv[0], kv[1]); err != nil {
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

	if stealthAlreadySet {
		// Beads files are already git-excluded — keep the solo config local too.
		if err := applySoloConfig(true); err != nil {
			return false, err
		}
		fmt.Printf("%s sync.mode: local-only (no-push, daemon auto-sync off)\n", ui.RenderPass("✓"))
		fmt.Printf("%s Stealth already active: beads files excluded from git\n", ui.RenderPass("✓"))
		return true, nil
	}
	if !isGitRepo() {
		// No git → nothing to leak into; write to config.yaml directly.
		if err := applySoloConfig(false); err != nil {
			return false, err
		}
		fmt.Printf("%s sync.mode: local-only (no-push, daemon auto-sync off)\n", ui.RenderPass("✓"))
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
		// Local-branch mode commits beads (incl. config) to a never-pushed
		// branch, so the solo config belongs in the committed config.yaml.
		if err := applySoloConfig(false); err != nil {
			return false, err
		}
		fmt.Printf("%s sync.mode: local-only (no-push, daemon auto-sync off)\n", ui.RenderPass("✓"))
		branch := soloLocalBranch
		// Continue from an existing sync branch (config or a well-known branch
		// like beads-metadata) instead of orphaning its history (BeadsLog-a2l).
		if existing := detectExistingSyncBranch(); existing != "" {
			fmt.Printf("\n%s Existing sync branch detected: %s\n", ui.RenderAccent("▶"), ui.RenderAccent(existing))
			fmt.Print("  Continue from it (keeps beads history)? [Y/n]: ")
			resp, _ := reader.ReadString('\n')
			resp = strings.TrimSpace(strings.ToLower(resp))
			if resp != "n" && resp != "no" {
				branch = existing
			}
		}
		if err := syncbranch.Set(ctx, store, branch); err != nil {
			return false, fmt.Errorf("failed to set sync branch: %w", err)
		}
		if err := createSyncBranch(branch); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create sync branch: %v\n", err)
			fmt.Println("  You can create it manually: git checkout -b", branch)
		}
		fmt.Printf("\n%s Beads commits go to local branch %s (never pushed)\n", ui.RenderPass("✓"), ui.RenderAccent(branch))
		printSoloSummary(branch)
		return false, nil
	}

	// Devlog graph choice: fresh vs carry the team history.
	fmt.Println("\nDevlog graph when going solo:")
	fmt.Printf("  1. %s: start clean (your solo notes only)\n", ui.RenderAccent("Fresh"))
	fmt.Printf("  2. %s: carry the team's devlog history into your private graph\n", ui.RenderAccent("Continuity"))
	fmt.Print("\nChoice [1/2, default 2]: ")
	gResp, _ := reader.ReadString('\n')
	continuity := strings.TrimSpace(gResp) != "1"

	// Invisible mode: solo config goes to the git-excluded config.local.yaml so
	// nothing leaks into the committed config.yaml (BeadsLog-9vd).
	if err := applySoloConfig(true); err != nil {
		return false, err
	}
	fmt.Printf("%s sync.mode: local-only (no-push, daemon auto-sync off)\n", ui.RenderPass("✓"))

	carried, err := transitionToSolo(ctx, store, continuity)
	if err != nil {
		return false, fmt.Errorf("failed to set up solo devlog space: %w", err)
	}
	fmt.Printf("\n%s .beads/ + %s/ + beads scaffolding excluded from git — invisible to collaborators\n", ui.RenderPass("✓"), soloDevlogDir)
	printSoloDevlogNote(carried, continuity)
	// skip-worktree can hide an UNcommitted protocol block, but not one already
	// in history. If a teammate would pull the block, tell the user to strip it.
	if committed := protocolCommittedFiles(); len(committed) > 0 {
		fmt.Printf("%s Heads up: the beads protocol block is already committed in %s — it's in git history, so hiding it now only stops future changes. To fully remove it from the team's view, delete the <beads_protocol>…</beads_protocol> block and commit that.\n",
			ui.RenderWarn("⚠"), strings.Join(committed, ", "))
	}
	printSoloSummary("")
	return true, nil
}

// detectExistingSyncBranch returns a pre-existing sync branch to continue
// from: the configured sync-branch if set, else a local or remote
// "beads-metadata" branch (the conventional default). Empty if none.
func detectExistingSyncBranch() string {
	if configured := syncbranch.GetFromYAML(); configured != "" {
		return configured
	}
	for _, ref := range []string{"beads-metadata", "origin/beads-metadata"} {
		if exec.Command("git", "rev-parse", "--verify", ref).Run() == nil {
			return "beads-metadata"
		}
	}
	return ""
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
