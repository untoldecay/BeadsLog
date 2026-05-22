package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/untoldecay/BeadsLog/internal/changelog"
	"github.com/untoldecay/BeadsLog/internal/config"
	"github.com/untoldecay/BeadsLog/internal/ui"
)

const (
	// Raw URL to the changelog file in the main repository
	RemoteChangelogURL = "https://raw.githubusercontent.com/untoldecay/BeadsLog/main/internal/changelog/changelog.go"
	UpdateCommand      = "go install github.com/untoldecay/BeadsLog/cmd/bd@main"
)

var upgradeCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for new tool versions on GitHub",
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")
		install, _ := cmd.Flags().GetBool("install")

		fmt.Printf("Checking for updates at %s...\n", RemoteChangelogURL)
		
		latestVersion, err := fetchRemoteVersion()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking remote version: %v\n", err)
			os.Exit(1)
		}

		if !changelog.IsNewer(latestVersion, Version) && !force {
			fmt.Printf("✓ You are up to date (v%s)\n", Version)
			return
		}

		fmt.Printf("✨ New version available: %s (Current: %s)\n", latestVersion, Version)
		fmt.Println("\nTo update, run:")
		fmt.Printf("  %s\n", ui.RenderAccent(UpdateCommand))

		if install || ui.PromptYesNo("Do you want to install the update now?", false) {
			runUpdate()
		}
	},
}

func fetchRemoteVersion() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(RemoteChangelogURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch remote changelog: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Regex to extract const CurrentVersion = "X.Y.Z"
	re := regexp.MustCompile(`const\s+CurrentVersion\s+=\s+"([^"]+)"`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return "", fmt.Errorf("could not find version string in remote changelog")
	}

	return matches[1], nil
}

func runUpdate() {
	fmt.Printf("\n🚀 Running update: %s\n", UpdateCommand)
	
	// Split command for exec
	parts := strings.Fields(UpdateCommand)
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Update failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n%s Update complete! Please restart your agent session to use the new version.\n", ui.RenderPass("✅"))
}

// maybeCheckForUpdates is a non-blocking check for the onboarding flow
func maybeCheckForUpdates() {
	// Throttle checks to once every 24 hours
	lastCheckStr := config.GetString("last-remote-version-check")
	if lastCheckStr != "" {
		if t, err := time.Parse(time.RFC3339, lastCheckStr); err == nil {
			if time.Since(t) < 24*time.Hour {
				return
			}
		}
	}

	// Perform check in background-ish (fast)
	go func() {
		latest, err := fetchRemoteVersion()
		if err != nil {
			return
		}

		// Update last check time
		_ = config.SetYamlConfig("last-remote-version-check", time.Now().Format(time.RFC3339))

		if changelog.IsNewer(latest, Version) {
			fmt.Printf("\n%s A new version of BeadsLog is available: %s (Current: %s)\n", ui.RenderAccent("✨"), latest, Version)
			fmt.Printf("   Run %s to update.\n", ui.RenderAccent("bd upgrade check --install"))
		}
	}()
}

func init() {
	upgradeCheckCmd.Flags().Bool("force", false, "Force check even if already on latest version")
	upgradeCheckCmd.Flags().Bool("install", false, "Automatically install if an update is found")
	upgradeCmd.AddCommand(upgradeCheckCmd)
}
