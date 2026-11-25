package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/untoldecay/BeadsLog/internal/changelog"
	"github.com/untoldecay/BeadsLog/internal/channel"
	"github.com/untoldecay/BeadsLog/internal/ui"
)

// upgradeCmd is a single command (BeadsLog-q9o.2): show the current version,
// check for a newer one on this machine's channel, surface its changelog, and
// offer to install. Version-seen tracking is automatic and per-machine (see
// .local_changelog_seen), and "what's new in the binary I already have" is
// handled by `bd refresh` / `bd onboard`.
var upgradeCmd = &cobra.Command{
	Use:     "upgrade",
	GroupID: "maint",
	Short:   "Check for a newer bd version and install it",
	Long: `Check for a newer bd version, show what's new, and offer to install.

  bd upgrade                   check, show changelog, prompt to install
  bd upgrade --install         install immediately if a newer version exists
  bd upgrade --channel dev     follow pre-releases (X.Y.Z-dev.N) on this machine
  bd upgrade --channel stable  go back to releases only
  bd upgrade --json            machine-readable {current, latest, channel, upgrade_available}

The channel is remembered per machine. Dev builds can be unstable and replace
the bd used by every project on this machine.`,
	Run: func(cmd *cobra.Command, _ []string) {
		force, _ := cmd.Flags().GetBool("force")
		install, _ := cmd.Flags().GetBool("install")

		ch := channel.Current(Version)
		switching := cmd.Flags().Changed("channel")
		if switching {
			name, _ := cmd.Flags().GetString("channel")
			parsed, err := channel.Parse(name)
			if err != nil {
				FatalError("%v", err)
			}
			if err := channel.Save(parsed); err != nil {
				FatalError("saving channel: %v", err)
			}
			ch = parsed
		}

		if !jsonOutput {
			fmt.Printf("Current version: v%s (channel: %s)\n", Version, ch)
			fmt.Println("Checking for a newer version...")
		}

		latest, err := channel.FetchLatest(ch)
		if err != nil {
			if jsonOutput {
				outputJSON(map[string]interface{}{"current_version": Version, "channel": ch, "error": err.Error()})
			} else {
				fmt.Printf("Error checking remote version: %v\n", err)
			}
			return
		}

		available := changelog.IsNewer(latest, Version)
		// Asking for stable on a dev build means going back, even though
		// that is an older version.
		goingBack := switching && ch == channel.Stable && channel.IsDev(Version) && !available

		if jsonOutput {
			outputJSON(map[string]interface{}{
				"current_version":   Version,
				"latest_version":    latest,
				"channel":           ch,
				"upgrade_available": available,
			})
			return
		}

		switch {
		case available:
			fmt.Printf("\n✨ New version available: v%s\n", latest)
			if features := fetchReleaseFeatures(latest); len(features) > 0 {
				fmt.Printf("\nWhat's new in v%s:\n", latest)
				for _, f := range features {
					fmt.Printf("  • %s\n", f)
				}
			} else {
				fmt.Printf("Notes: %s\n", fmt.Sprintf(releaseURL, latest))
			}
		case goingBack:
			fmt.Printf("\n↩ Back to stable: v%s (older than v%s)\n", latest, Version)
			fmt.Println("If a dev build migrated a database, the older bd may not understand it; keep a backup of .beads/.")
		case force:
			fmt.Printf("\nLatest %s version: v%s\n", ch, latest)
		default:
			fmt.Printf("✓ You are up to date (v%s)\n", Version)
			if ch == channel.Stable && channel.IsDev(Version) {
				fmt.Printf("  This is a dev build; run %s to go back to v%s.\n", ui.RenderAccent("bd upgrade --channel stable"), latest)
			}
			return
		}
		fmt.Printf("\nUpdate: %s\n", ui.RenderAccent(installCommand(latest)))

		if install || ui.PromptYesNo(fmt.Sprintf("Install v%s now?", latest), false) {
			runUpdate(latest)
		}
	},
}

func init() {
	upgradeCmd.Flags().Bool("force", false, "Show upgrade details even if already on the latest version")
	upgradeCmd.Flags().Bool("install", false, "Install immediately if a newer version exists")
	upgradeCmd.Flags().String("channel", "", "Switch this machine to a release channel: stable or dev")
	rootCmd.AddCommand(upgradeCmd)
}

// pluralize returns "s" for counts != 1, "" otherwise.
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
