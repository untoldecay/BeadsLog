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
	"github.com/untoldecay/BeadsLog/internal/channel"
	"github.com/untoldecay/BeadsLog/internal/config"
	"github.com/untoldecay/BeadsLog/internal/ui"
)

const (
	modulePath = "github.com/untoldecay/BeadsLog/cmd/bd"
	releaseURL = "https://github.com/untoldecay/BeadsLog/releases/tag/v%s"

	// The changelog as of a given tag
	remoteChangelogURL = "https://raw.githubusercontent.com/untoldecay/BeadsLog/v%s/internal/changelog/changelog.go"
)

// installCommand installs exactly the given published version.
func installCommand(version string) string {
	return fmt.Sprintf("go install %s@v%s", modulePath, version)
}

// fetchReleaseFeatures returns the Features bullets of a published version, or
// nil. Dev builds are cut without a changelog entry of their own; the previous
// version's bullets must not be shown as theirs.
func fetchReleaseFeatures(version string) []string {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fmt.Sprintf(remoteChangelogURL, version))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	return releaseFeatures(string(body), version)
}

func releaseFeatures(changelogSource, version string) []string {
	re := regexp.MustCompile(`const\s+CurrentVersion\s+=\s+"([^"]+)"`)
	m := re.FindStringSubmatch(changelogSource)
	if len(m) < 2 || m[1] != version {
		return nil
	}
	return extractLatestFeatures(changelogSource)
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

func runUpdate(version string) {
	fmt.Printf("\n🚀 Running update: %s\n", installCommand(version))

	install := func(extraEnv ...string) error {
		parts := strings.Fields(installCommand(version))
		cmd := exec.Command(parts[0], parts[1:]...) // #nosec G204 - version comes from a validated tag
		cmd.Env = append(os.Environ(), extraEnv...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	err := install()
	if err != nil {
		// The module proxy may not know a tag pushed minutes ago.
		fmt.Println("Retrying without the module proxy...")
		err = install("GOPROXY=direct")
	}
	if err != nil {
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
		latest, err := channel.FetchLatest(channel.Current(Version))
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
