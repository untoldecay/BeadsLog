package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/untoldecay/BeadsLog/internal/queries"
	"github.com/untoldecay/BeadsLog/internal/storage/sqlite"
	"github.com/untoldecay/BeadsLog/internal/ui"
)

var catchupCmd = &cobra.Command{
	Use:   "catchup",
	Short: "See what has happened in the project since your last check",
	Long: `Fetch all new sessions, closed issues, and state changes since the last catchup.
Use --ack to mark these changes as seen and update your catchup timestamp.`,
	Run: func(cmd *cobra.Command, args []string) {
		ack, _ := cmd.Flags().GetBool("ack")

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

		// 3. Render feed
		if isFirstCatchup {
			fmt.Printf("\n%s No previous catchup found. Showing last 7 days of activity.\n", ui.RenderAccent("💡"))
			fmt.Printf("   Run 'bd catchup --ack' to set your first checkpoint.\n\n")
		}
		renderCatchupFeed(delta)

		// 4. Update timestamp if ack
		if ack {
			now := time.Now().Format(time.RFC3339)
			_ = store.SetMetadata(rootCtx, "last_catchup_time", now)
			fmt.Printf("\n✓ Catchup acknowledged. Last check updated to: %s\n", now)
		} else if len(delta.Sessions) > 0 || len(delta.ClosedIssues) > 0 || len(delta.StateChanges) > 0 {
			fmt.Printf("\n%s Run 'bd catchup --ack' to mark these as seen.\n", ui.RenderAccent("💡"))
			fmt.Printf("%s Agents: Use '_rules/_devlog/_generate-catchup.md' to summarize this activity for the user.\n", ui.RenderAccent("🤖"))
		} else {
			fmt.Println("Nothing new since your last check.")
		}
	},
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
	rootCmd.AddCommand(catchupCmd)
}
