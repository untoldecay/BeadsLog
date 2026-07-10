package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/untoldecay/BeadsLog/internal/queries"
	"github.com/untoldecay/BeadsLog/internal/storage/sqlite"
	"github.com/untoldecay/BeadsLog/internal/types"
	"github.com/untoldecay/BeadsLog/internal/ui"
)

var catchupCmd = &cobra.Command{
	Use:   "catchup",
	Short: "See what has happened in the project since your last check",
	Long: `Fetch all new sessions, closed issues, and state changes since the last catchup.
Use --ack to mark these changes as seen and update your catchup timestamp.`,
	Run: func(cmd *cobra.Command, args []string) {
		ack, _ := cmd.Flags().GetBool("ack")
		digest, _ := cmd.Flags().GetBool("digest")

		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()

		// 1. Get last catchup time
		lastCatchupStr, _ := store.GetMetadata(rootCtx, "last_catchup_time")

		isFirstCatchup := false
		since := time.Now().Add(-7 * 24 * time.Hour) // Default to 7 days ago (safer than 24h)
		if lastCatchupStr == "" {
			isFirstCatchup = true
		} else {
			if t, err := time.Parse(time.RFC3339, lastCatchupStr); err == nil {
				since = t
			} else {
				isFirstCatchup = true
			}
		}

		// 2. Fetch delta
		delta, err := queries.GetCatchupDelta(rootCtx, store.UnderlyingDB(), since)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching catchup data: %v\n", err)
			os.Exit(1)
		}

		// 3. Render
		if digest {
			groups, unattached := queries.GroupCatchupDelta(rootCtx, store.UnderlyingDB(), delta)
			if jsonOutput {
				outputJSON(map[string]interface{}{
					"since":                    delta.Since,
					"groups":                   groups,
					"unattached_state_changes": unattached,
					"closed_issues":            delta.ClosedIssues,
				})
			} else {
				renderCatchupDigest(delta, groups, unattached)
			}
		} else if jsonOutput {
			outputJSON(delta)
		} else {
			if isFirstCatchup {
				fmt.Printf("\n%s No previous catchup found. Showing last 7 days of activity.\n", ui.RenderAccent("💡"))
				fmt.Printf("   Run 'bd catchup --ack' to set your first checkpoint.\n\n")
			}
			renderCatchupFeed(delta)
		}

		// 4. Update timestamp if ack (nano precision: branch_states timestamps
		// carry sub-second precision and compare as strings)
		if ack {
			now := time.Now().Format(time.RFC3339Nano)
			_ = store.SetMetadata(rootCtx, "last_catchup_time", now)
			if !jsonOutput {
				fmt.Printf("\n✓ Catchup acknowledged. Last check updated to: %s\n", now)
			}
		} else if jsonOutput {
			// no prose in JSON mode
		} else if len(delta.Sessions) > 0 || len(delta.ClosedIssues) > 0 || len(delta.StateChanges) > 0 {
			fmt.Printf("\n%s Run 'bd catchup --ack' to mark these as seen.\n", ui.RenderAccent("💡"))
			fmt.Printf("%s Agents: Use '_rules/_devlog/_generate-catchup.md' to summarize this activity for the user.\n", ui.RenderAccent("🤖"))
		} else {
			fmt.Println("Nothing new since your last check.")
		}
	},
}

// renderCatchupDigest prints the delta grouped by feature arc, each with a
// short narrative, its sessions, entities, and attached lifecycle changes.
func renderCatchupDigest(delta *queries.CatchupDelta, groups []queries.FeatureGroup, unattached []types.BranchState) {
	days := int(time.Since(delta.Since).Hours()/24) + 1
	fmt.Printf("📰 Catchup since %s (last %d day(s))\n", delta.Since.Format("2006-01-02"), days)
	fmt.Println(strings.Repeat("=", 60))

	if len(groups) == 0 && len(unattached) == 0 && len(delta.ClosedIssues) == 0 {
		fmt.Println("\nNothing new since your last check.")
		return
	}

	for _, g := range groups {
		kind := ""
		if g.Kind == "entity" {
			kind = "  [entity arc]"
		}
		fmt.Printf("\n📦 %s — %d session(s) · %s%s\n", g.Key, len(g.Sessions), strings.Join(g.Authors, ", "), kind)
		if g.Narrative != "" {
			fmt.Printf("   %s\n", g.Narrative)
		}
		for _, s := range g.Sessions {
			date := s.Date
			if len(date) >= 10 {
				date = date[:10]
			}
			fmt.Printf("   • %s (%s)\n", s.Title, date)
		}
		if len(g.Entities) > 0 {
			fmt.Printf("   🔗 %s\n", strings.Join(g.Entities, ", "))
		}
		for _, sc := range g.StateChanges {
			icon := "⚠"
			if sc.State == types.StateAbandoned {
				icon = "🚫"
			} else if sc.State == types.StateOngoing {
				icon = "🔄"
			}
			fmt.Printf("   %s %s since you last looked — %q (%s)\n", icon, strings.ToUpper(string(sc.State)), sc.ShortReason, sc.Actor)
		}
	}

	if len(unattached) > 0 {
		fmt.Printf("\n⚠️  Other lifecycle changes since you last looked:\n")
		for _, sc := range unattached {
			fmt.Printf("   • [%s] %s:%s — %q (%s, %s)\n", strings.ToUpper(string(sc.State)), sc.ScopeType, sc.ScopeRef, sc.ShortReason, sc.Actor, sc.Timestamp.Format("2006-01-02"))
		}
	}

	if len(delta.ClosedIssues) > 0 {
		fmt.Printf("\n✔ %d issue(s) closed:\n", len(delta.ClosedIssues))
		for _, i := range delta.ClosedIssues {
			fmt.Printf("   • %s: %s (by %s)\n", i.ID, i.Title, i.Owner)
		}
	}
}

func renderCatchupFeed(delta *queries.CatchupDelta) {
	fmt.Printf("🔍 Activity since %s:\n", delta.Since.Format("2006-01-02 15:04"))
	fmt.Println(strings.Repeat("=", 60))

	if len(delta.StateChanges) > 0 {
		fmt.Printf("\n🚦 %d State Changes:\n", len(delta.StateChanges))
		for _, s := range delta.StateChanges {
			badge := string(s.State)
			fmt.Printf("  • [%s] %s:%s - %s (by %s)\n", badge, s.ScopeType, s.ScopeRef, s.ShortReason, s.Actor)
		}
	}

	if len(delta.ClosedIssues) > 0 {
		fmt.Printf("\n✅ %d Closed Issues:\n", len(delta.ClosedIssues))
		for _, i := range delta.ClosedIssues {
			fmt.Printf("  • %s: %s (%s) (by %s)\n", i.ID, i.Title, i.CloseReason, i.Owner)
		}
	}

	if len(delta.Sessions) > 0 {
		fmt.Printf("\n📄 %d New Sessions:\n", len(delta.Sessions))
		for _, s := range delta.Sessions {
			fmt.Printf("  • [%s] %s (Branch: %s) (by %s)\n", s.ID, s.Title, s.Branch, s.Author)
		}
	}
}

func init() {
	catchupCmd.Flags().Bool("ack", false, "Acknowledge the activity and update last catchup timestamp")
	catchupCmd.Flags().Bool("digest", false, "Group activity by feature arc with narratives and lifecycle deltas")
	rootCmd.AddCommand(catchupCmd)
}
