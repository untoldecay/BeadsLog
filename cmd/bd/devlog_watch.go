package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/untoldecay/BeadsLog/internal/storage/sqlite"
)

// devlogWatchCmd live-tails the devlog index and re-indexes new/changed sessions
// incrementally, reusing the daemon's FileWatcher + Debouncer. It runs as a
// standalone foreground command (like `tail -f`); it does not touch git and
// never commits, so it is safe alongside a running daemon (SQLite WAL).
var devlogWatchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Live-tail the devlog and re-index new sessions incrementally",
	Long: `Watch the devlog directory and re-index sessions as they are recorded.

  bd devlog watch      # blocks until Ctrl-C, indexing changes as they arrive

Reuses the daemon's file-watcher/debounce; runs standalone (no daemon needed).`,
	Run: func(cmd *cobra.Command, args []string) {
		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			FatalError("failed to open database: %v", err)
		}
		defer store.Close()

		devlogDir, _ := store.GetConfig(rootCtx, "devlog_dir")
		if devlogDir == "" {
			devlogDir = "_rules/_devlog"
		}
		indexPath := filepath.Join(devlogDir, "_index.md")
		log := SetupStderrLogger(false, slog.LevelInfo)

		// syncNow re-reads the index and upserts every row; SyncSession is a
		// no-op when a session's content hash is unchanged, so this is naturally
		// incremental — only new/edited sessions do real work.
		syncNow := func() {
			rows, err := parseIndexMD(indexPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "watch: index parse error: %v\n", err)
				return
			}
			for _, row := range rows {
				updated, err := SyncSession(store, row)
				if err != nil {
					fmt.Fprintf(os.Stderr, "watch: sync error (%s): %v\n", row.Subject, err)
					continue
				}
				if updated {
					fmt.Printf("  indexed: %s\n", row.Subject)
				}
			}
		}

		fmt.Printf("Watching %s (Ctrl-C to stop)\n", indexPath)
		syncNow() // full pass on startup so the DB is current

		ctx, cancel := signal.NotifyContext(rootCtx, os.Interrupt, syscall.SIGTERM)
		defer cancel()

		debounce := NewDebouncer(500*time.Millisecond, syncNow)
		defer debounce.Cancel()

		watcher, err := NewFileWatcher(indexPath, func() { debounce.Trigger() })
		if err != nil {
			FatalError("failed to start watcher: %v", err)
		}
		defer func() { _ = watcher.Close() }()
		watcher.Start(ctx, log)

		<-ctx.Done()
		fmt.Println("\nwatch stopped")
	},
}
