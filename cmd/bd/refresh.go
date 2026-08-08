package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/untoldecay/BeadsLog/cmd/bd/doctor"
	"github.com/untoldecay/BeadsLog/internal/beads"
	"github.com/untoldecay/BeadsLog/internal/changelog"
	"github.com/untoldecay/BeadsLog/internal/config"
	"github.com/untoldecay/BeadsLog/internal/ui"
)

// refreshCmd is the one command to run after a binary update (BeadsLog-q9o.1).
// It orchestrates the mandatory-but-scattered post-update steps and prints a
// devlog-status-style summary. Everything is local except the namespace step,
// which is gated to team mode and only ever FETCHES — never pushes.
var refreshCmd = &cobra.Command{
	Use:     "refresh",
	GroupID: "maint",
	Short:   "Bring your repo up to date after a bd update (one command)",
	Long: `Run after updating the bd binary. In one step it:
  • reports the version change and what's new
  • confirms schema migrations (they apply automatically on DB open)
  • restarts the daemon onto the new binary
  • runs update-relevant doctor checks
  • (team mode only) probes the remote namespace and adopts a signed prefix
    migration — fetch only, never a push

Flags:
  --fix      also apply doctor fixes (git hooks, etc.)
  --devlog   also re-index the devlog and repair the graph (slower)`,
	Run: func(cmd *cobra.Command, _ []string) {
		fix, _ := cmd.Flags().GetBool("fix")
		withDevlog, _ := cmd.Flags().GetBool("devlog")

		beadsDir := beads.FindBeadsDir()
		if beadsDir == "" {
			fmt.Fprintln(os.Stderr, "Not in a bd repository (no .beads/ found). Run 'bd init' first.")
			os.Exit(1)
		}
		path := filepath.Dir(beadsDir) // repo root (parent of .beads)

		fmt.Println("bd refresh")
		fmt.Println("==========")

		changed := false

		// 1. Binary version + what's new (per-machine detection).
		if versionUpgradeDetected && previousVersion != "" {
			fmt.Printf("Binary:     v%s → v%s\n", previousVersion, Version)
			changed = true
		} else {
			fmt.Printf("Binary:     v%s (current)\n", Version)
		}

		// 2. Migrations already ran when the DB opened (idempotent). Report state.
		fmt.Println("Migrations: applied automatically on open (schema current)")

		// 3. Daemon: restart onto the new binary when a version bump was seen.
		if versionUpgradeDetected {
			if restartDaemonForVersionMismatch() {
				fmt.Printf("Daemon:     restarted on v%s\n", Version)
			} else {
				fmt.Printf("Daemon:     no running daemon to restart\n")
			}
			changed = true
		} else {
			fmt.Println("Daemon:     already on current binary")
		}

		// 4. Doctor — curated, update-relevant checks.
		fmt.Printf("Doctor:     %s\n", refreshDoctorLine(path))
		if fix {
			runDoctorFix(path)
		}

		// 5. Protocol — re-inject the agent protocol block only if it drifted
		// from this binary's embedded version (acknowledge a new protocol
		// without the heavyweight full onboard).
		if drifted := refreshProtocol(true); len(drifted) > 0 {
			fmt.Printf("Protocol:   updated agent instructions (%s)\n", strings.Join(drifted, ", "))
			changed = true
		} else {
			fmt.Println("Protocol:   agent instructions current")
		}

		// 6. Namespace — team mode only, fetch-only.
		fmt.Printf("Namespace:  %s\n", refreshNamespace(beadsDir))

		// Optional: devlog re-index + graph repair (slow).
		if withDevlog {
			fmt.Println("Devlog:     re-indexing + verifying graph…")
			_ = runSelfCommand("devlog", "sync")
			_ = runSelfCommand("devlog", "verify", "--fix")
		}

		fmt.Println()
		if changed {
			fmt.Printf("%s Up to date on v%s. No further action needed.\n", ui.RenderAccent("✨"), Version)
		} else {
			fmt.Printf("%s Already current (v%s) — nothing to do.\n", ui.RenderAccent("✨"), Version)
		}

		// Record this machine as having seen the current changelog (per-machine).
		if changelog.IsNewer(changelog.CurrentVersion, readLocalChangelogSeen(beadsDir)) {
			writeLocalChangelogSeen(beadsDir, changelog.CurrentVersion)
		}
	},
}

// refreshDoctorLine runs the update-relevant checks and returns a compact
// "✓ N passed  ○ M warning(s) (first fix)" summary.
func refreshDoctorLine(path string) string {
	checks := []doctor.DoctorCheck{
		doctor.CheckInstallation(path),
		doctor.CheckGitHooks(),
		doctor.CheckSchemaCompatibility(path),
		doctor.CheckDatabaseVersion(path, Version),
	}
	passed, warnings, errors := 0, 0, 0
	var firstIssue *doctor.DoctorCheck
	for i := range checks {
		switch checks[i].Status {
		case doctor.StatusOK:
			passed++
		case doctor.StatusWarning:
			warnings++
			if firstIssue == nil {
				firstIssue = &checks[i]
			}
		case doctor.StatusError:
			errors++
			firstIssue = &checks[i] // errors outrank warnings
		}
	}
	line := fmt.Sprintf("✓ %d passed", passed)
	if warnings > 0 {
		line += fmt.Sprintf("   ○ %d warning(s)", warnings)
	}
	if errors > 0 {
		line += fmt.Sprintf("   ✖ %d error(s)", errors)
	}
	if firstIssue != nil && firstIssue.Fix != "" {
		line += fmt.Sprintf(" — %s (%s)", firstIssue.Message, firstIssue.Fix)
	}
	return line
}

// refreshNamespace probes the remote for a signed prefix migration and adopts
// it, but ONLY in team mode — a solo user may keep a remote from before going
// solo, so we gate on mode, not remote presence. Fetch only; never pushes.
func refreshNamespace(beadsDir string) string {
	if config.GetString("sync-mode") == "local-only" {
		return "skipped (solo / local-only mode)"
	}
	if !hasGitRemote(rootCtx) {
		return "skipped (no git remote)"
	}

	// Reuse the battle-tested pull+signature-gated-adopt path with push disabled.
	out, err := runSelfCommandOutput("sync", "--no-push")
	if err != nil {
		return fmt.Sprintf("check failed (%v) — run 'bd sync' manually", err)
	}
	// Report ONLY a real adoption. Do not surface sync's internal
	// "Ignoring prefix mismatches (all are tombstones)" line — that's benign
	// old-prefix tombstone cleanup, not a namespace change.
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(strings.ToLower(line), "adopted") {
			return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "✓ "))
		}
	}
	return "in sync (no namespace change)"
}

// refreshProtocol re-injects the agent protocol block ONLY when the version in
// the agent files drifts from the binary's embedded FullBootloader. It touches
// just the block between the protocol tags — never PROJECT_CONTEXT or user
// prose — so an update acknowledges a new protocol without the full onboard.
// Returns the files that were (or would be) updated.
func refreshProtocol(reinject bool) []string {
	desired := ProtocolStartTag + "\n" + restoreCodeBlocks(FullBootloader) + "\n" + ProtocolEndTag
	var drifted []string
	for _, file := range Candidates {
		content, err := os.ReadFile(file) // #nosec G304 - Candidates is a fixed allowlist
		if err != nil {
			continue
		}
		s := string(content)
		start := strings.Index(s, ProtocolStartTag)
		end := strings.Index(s, ProtocolEndTag)
		if start == -1 || end == -1 || end < start {
			continue // no protocol block here — that's onboard's job, not refresh's
		}
		current := s[start : end+len(ProtocolEndTag)]
		if current == desired {
			continue
		}
		drifted = append(drifted, file)
		if reinject {
			newContent := s[:start] + desired + s[end+len(ProtocolEndTag):]
			_ = os.WriteFile(file, []byte(newContent), 0644) // #nosec G306 - agent instruction file
		}
	}
	return drifted
}

// runSelfCommand execs this same bd binary for a subcommand, streaming output.
func runSelfCommand(args ...string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	c := exec.CommandContext(rootCtx, exe, append([]string{"--no-daemon"}, args...)...) // #nosec G204 - trusted self
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// runSelfCommandOutput execs this bd binary and captures combined output.
func runSelfCommandOutput(args ...string) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	c := exec.CommandContext(rootCtx, exe, append([]string{"--no-daemon"}, args...)...) // #nosec G204 - trusted self
	out, err := c.CombinedOutput()
	return string(out), err
}

func runDoctorFix(path string) {
	fmt.Println("            (applying doctor --fix…)")
	_ = runSelfCommand("doctor", "--fix")
}

func init() {
	refreshCmd.Flags().Bool("fix", false, "Also apply doctor fixes (git hooks, etc.)")
	refreshCmd.Flags().Bool("devlog", false, "Also re-index the devlog and repair the graph (slower)")
	rootCmd.AddCommand(refreshCmd)
}
