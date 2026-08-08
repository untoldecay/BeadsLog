package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/untoldecay/BeadsLog/internal/changelog"
	"github.com/untoldecay/BeadsLog/internal/ui"
)

// upgradeCmd is a single command (BeadsLog-q9o.2): show the current version,
// check GitHub for a newer one, surface its changelog, and offer to install.
// The old status/review/ack subcommands are gone — version-seen tracking is now
// automatic and per-machine (see .local_changelog_seen), and "what's new in the
// binary I already have" is handled by `bd refresh` / `bd onboard`.
var upgradeCmd = &cobra.Command{
	Use:     "upgrade",
	GroupID: "maint",
	Short:   "Check for a newer bd version and install it",
	Long: `Check GitHub for a newer bd version, show what's new, and offer to install.

  bd upgrade            check, show changelog, prompt to install
  bd upgrade --install  install immediately if a newer version exists
  bd upgrade --json     machine-readable {current, latest, upgrade_available}`,
	Run: func(cmd *cobra.Command, _ []string) {
		force, _ := cmd.Flags().GetBool("force")
		install, _ := cmd.Flags().GetBool("install")

		if !jsonOutput {
			fmt.Printf("Current version: v%s\n", Version)
			fmt.Println("Checking GitHub for a newer version...")
		}

		latest, features, err := fetchRemoteChangelog()
		if err != nil {
			if jsonOutput {
				outputJSON(map[string]interface{}{"current_version": Version, "error": err.Error()})
			} else {
				fmt.Printf("Error checking remote version: %v\n", err)
			}
			return
		}

		available := changelog.IsNewer(latest, Version)

		if jsonOutput {
			outputJSON(map[string]interface{}{
				"current_version":   Version,
				"latest_version":    latest,
				"upgrade_available": available,
			})
			return
		}

		if !available && !force {
			fmt.Printf("✓ You are up to date (v%s)\n", Version)
			return
		}

		fmt.Printf("\n✨ New version available: v%s\n", latest)
		if len(features) > 0 {
			fmt.Printf("\nWhat's new in v%s:\n", latest)
			for _, f := range features {
				fmt.Printf("  • %s\n", f)
			}
		}
		fmt.Printf("\nUpdate: %s\n", ui.RenderAccent(UpdateCommand))

		if install || ui.PromptYesNo("Install the update now?", false) {
			runUpdate()
		}
	},
}

func init() {
	upgradeCmd.Flags().Bool("force", false, "Show upgrade details even if already on the latest version")
	upgradeCmd.Flags().Bool("install", false, "Install immediately if a newer version exists")
	rootCmd.AddCommand(upgradeCmd)
}

// pluralize returns "s" for counts != 1, "" otherwise.
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
