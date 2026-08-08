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

	"github.com/untoldecay/BeadsLog/internal/changelog"
	"github.com/untoldecay/BeadsLog/internal/config"
	"github.com/untoldecay/BeadsLog/internal/ui"
)

const (
	// Raw URL to the changelog file in the main repository
	RemoteChangelogURL = "https://raw.githubusercontent.com/untoldecay/BeadsLog/main/internal/changelog/changelog.go"
	UpdateCommand      = "go install github.com/untoldecay/BeadsLog/cmd/bd@main"
)

// fetchRemoteChangelog fetches the remote changelog.go once and returns the
// latest version plus a best-effort list of that version's Features bullets,
// so `bd upgrade` can surface what's new without the new binary installed.
func fetchRemoteChangelog() (version string, features []string, err error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(RemoteChangelogURL)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("failed to fetch remote changelog: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}
	text := string(body)

	re := regexp.MustCompile(`const\s+CurrentVersion\s+=\s+"([^"]+)"`)
	matches := re.FindStringSubmatch(text)
	if len(matches) < 2 {
		return "", nil, fmt.Errorf("could not find version string in remote changelog")
	}
	return matches[1], extractLatestFeatures(text), nil
}

// extractLatestFeatures pulls the quoted strings from the FIRST Features block
// in the changelog source (entries are newest-first, so that's the latest
// version). Best-effort — returns nil if the shape doesn't match.
func extractLatestFeatures(text string) []string {
	start := strings.Index(text, "Features: []string{")
	if start == -1 {
		return nil
	}
	rest := text[start:]
	end := strings.Index(rest, "}")
	if end == -1 {
		return nil
	}
	quoted := regexp.MustCompile(`"((?:[^"\\]|\\.)*)"`).FindAllStringSubmatch(rest[:end], -1)
	var features []string
	for _, m := range quoted {
		features = append(features, strings.ReplaceAll(m[1], `\"`, `"`))
	}
	return features
}

// fetchRemoteVersion is the thin version-only wrapper used by the background
// update check.
func fetchRemoteVersion() (string, error) {
	v, _, err := fetchRemoteChangelog()
	return v, err
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

	fmt.Printf("\n%s Update complete! Then run %s in your repo to sync it to the new version.\n", ui.RenderPass("✅"), ui.RenderAccent("bd refresh"))
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
			fmt.Printf("   Run %s to review and install.\n", ui.RenderAccent("bd upgrade"))
		}
	}()
}
