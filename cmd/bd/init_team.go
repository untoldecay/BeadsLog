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

// defaultSyncBranch is the dedicated branch beads commits go to by default, so
// beads data never lands on the user's work branch (develop/main). Isolated via
// an internal worktree; override with 'bd config set sync-branch <name>' or
// disable with 'bd init --inline'.
const defaultSyncBranch = "beads-metadata"

// runTeamWizard guides the user through team workflow setup
func runTeamWizard(ctx context.Context, store storage.Storage) error {
	fmt.Printf("\n%s %s\n\n", ui.RenderBold("bd"), ui.RenderBold("Team Workflow Setup Wizard"))
	fmt.Println("This wizard will configure beads for team collaboration.")
	fmt.Println()

	// Step 0: If this repo is currently in solo mode, reverse it — publish the
	// solo devlogs into the team dir, drop the excludes, restore config
	// (BeadsLog-9vd). Detected by the solo devlog dir being active.
	if cur, _ := store.GetConfig(ctx, "devlog_dir"); cur == soloDevlogDir {
		fmt.Printf("%s Solo mode detected — rejoining the team.\n", ui.RenderAccent("▶"))
		published, tErr := transitionToTeam(ctx, store)
		if tErr != nil {
			return fmt.Errorf("failed to reverse solo mode: %w", tErr)
		}
		fmt.Printf("%s Published %d solo devlog(s) to %s; solo excludes removed.\n", ui.RenderPass("✓"), published, teamDevlogDir)
		fmt.Printf("%s Note: rejoining un-excludes .beads/ — your local issue state now merges with the team (per-issue last-write-wins) on the next sync.\n\n", ui.RenderWarn("⚠"))
	}

	// Step 1: Check if we're in a git repository
	fmt.Printf("%s Detecting git repository setup...\n", ui.RenderAccent("▶"))

	if !isGitRepo() {
		fmt.Printf("%s Not in a git repository\n", ui.RenderWarn("⚠"))
		fmt.Println("\n  Initialize git first:")
		fmt.Println("  git init")
		fmt.Println()
		return fmt.Errorf("not in a git repository")
	}

	reader := bufio.NewReader(os.Stdin)

	// Step 2: Sync branch. Beads always commits to a dedicated branch, never your
	// work branch — a dedicated branch is safe whether or not main is protected
	// (that's all the old "protected main?" question decided), so we don't ask,
	// we just isolate. Honor a name already configured (e.g. from 'bd init').
	fmt.Printf("\n%s Sync branch\n", ui.RenderAccent("▶"))
	syncBranch := defaultSyncBranch
	if existing := syncbranch.GetFromYAML(); existing != "" {
		syncBranch = existing
	}
	// Set sync.branch config (GH#923: use syncbranch.Set for validation)
	if err := syncbranch.Set(ctx, store, syncBranch); err != nil {
		return fmt.Errorf("failed to set sync branch: %w", err)
	}
	fmt.Printf("%s Beads commits go to a dedicated %s branch via an internal worktree —\n", ui.RenderPass("✓"), ui.RenderAccent(syncBranch))
	fmt.Println("  never your work branches. Teammates pull/push it like normal git.")
	fmt.Printf("  Share to main when you want: %s (or 'bd sync --merge').\n", ui.RenderAccent("merge "+syncBranch+" → main via PR"))
	fmt.Printf("  Rename: %s   Disable: %s\n", ui.RenderAccent("bd config set sync-branch <name>"), ui.RenderAccent("bd init --inline"))

	// Step 3: Configure team settings
	fmt.Printf("\n%s Configuring team settings...\n", ui.RenderAccent("▶"))

	// Set team.enabled to true
	if err := store.SetConfig(ctx, "team.enabled", "true"); err != nil {
		return fmt.Errorf("failed to enable team mode: %w", err)
	}

	// Set team.sync_branch
	if err := store.SetConfig(ctx, "team.sync_branch", syncBranch); err != nil {
		return fmt.Errorf("failed to set team sync branch: %w", err)
	}

	fmt.Printf("%s Team mode enabled\n", ui.RenderPass("✓"))

	// Step 4: Configure auto-sync
	fmt.Println("\n  Enable automatic sync (daemon commits/pushes)?")
	fmt.Println("  • Auto-commit: Commits issue changes every 5 seconds")
	fmt.Println("  • Auto-push: Pushes commits to remote")
	fmt.Print("\nEnable auto-sync? [Y/n]: ")

	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	autoSync := !(response == "n" || response == "no")

	if autoSync {
		// GH#871: Write to config.yaml for team-wide settings (version controlled)
		// Use unified auto-sync config (replaces individual auto_commit/auto_push/auto_pull)
		if err := config.SetYamlConfig("daemon.auto-sync", "true"); err != nil {
			return fmt.Errorf("failed to enable auto-sync: %w", err)
		}

		fmt.Printf("%s Auto-sync enabled\n", ui.RenderPass("✓"))
	} else {
		if err := config.SetYamlConfig("daemon.auto-sync", "false"); err != nil {
			return fmt.Errorf("failed to disable auto-sync: %w", err)
		}
		fmt.Printf("%s Auto-sync disabled (manual sync with 'bd sync')\n", ui.RenderWarn("⚠"))
	}

	// Step 5: Summary
	fmt.Printf("\n%s %s\n\n", ui.RenderPass("✓"), ui.RenderBold("Team setup complete!"))

	fmt.Println("Configuration:")
	fmt.Printf("  Sync branch: %s\n", ui.RenderAccent(syncBranch))
	fmt.Printf("  Commits go to: %s (dedicated, never your work branches)\n", ui.RenderAccent(syncBranch))
	fmt.Printf("  Merge to main via: %s\n", ui.RenderAccent("Pull Request"))
	if autoSync {
		fmt.Printf("  Auto-sync: %s\n", ui.RenderAccent("enabled"))
	} else {
		fmt.Printf("  Auto-sync: %s\n", ui.RenderAccent("disabled"))
	}

	fmt.Println()
	fmt.Println("How it works:")
	fmt.Println("  • All team members work on the same repository")
	fmt.Println("  • Issues + devlogs are shared via the", syncBranch, "branch")
	fmt.Println("  • Use 'bd list' to see all team's issues")
	fmt.Println("  • Periodically merge", syncBranch, "to main via PR (optional)")

	if autoSync {
		fmt.Println("  • Daemon automatically commits and pushes changes")
	} else {
		fmt.Println("  • Run 'bd sync' manually to sync changes")
	}

	fmt.Println()
	fmt.Printf("Try it: %s\n", ui.RenderAccent("bd create \"Team planning issue\" -p 2"))
	fmt.Println()

	fmt.Println("Next steps:")
	fmt.Printf("  1. %s\n", "Teammates: git pull origin "+syncBranch+" (or just run 'bd sync')")
	fmt.Printf("  2. %s\n", "Periodically: merge "+syncBranch+" to main via PR")
	fmt.Println()

	return nil
}

// getGitBranch returns the current git branch name
// Uses symbolic-ref instead of rev-parse to work in fresh repos without commits (bd-flil)
func getGitBranch() (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// createSyncBranch creates a new branch for beads sync
func createSyncBranch(branchName string) error {
	// Check if branch already exists
	cmd := exec.Command("git", "rev-parse", "--verify", branchName)
	if err := cmd.Run(); err == nil {
		// Branch exists, nothing to do
		return nil
	}

	// Create new branch from current HEAD
	cmd = exec.Command("git", "checkout", "-b", branchName)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Switch back to original branch (checkout -b left us on the new branch)
	currentBranch, err := getGitBranch()
	if err == nil && currentBranch == branchName {
		cmd = exec.Command("git", "checkout", "-")
		_ = cmd.Run() // Ignore error, branch creation succeeded
	}

	return nil
}
