package fix

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/untoldecay/BeadsLog/internal/beads"
	"github.com/untoldecay/BeadsLog/internal/configfile"
	"github.com/untoldecay/BeadsLog/internal/devlog"
	"github.com/untoldecay/BeadsLog/internal/storage/sqlite"
)

// DevlogGhosts prunes ghost sessions from the index and database
func DevlogGhosts(path string) error {
	beadsDir := resolveBeadsDir(filepath.Join(path, ".beads"))

	// 1. Get database path
	var dbPath string
	if cfg, err := configfile.Load(beadsDir); err == nil && cfg != nil && cfg.Database != "" {
		dbPath = cfg.DatabasePath(beadsDir)
	} else {
		dbPath = filepath.Join(beadsDir, beads.CanonicalDatabaseName)
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil
	}

	ctx := context.Background()
	store, err := sqlite.New(ctx, dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { _ = store.Close() }()

	// 2. Prune, with the same safety rules as 'bd devlog prune'
	devlogDir, _ := store.GetConfig(ctx, "devlog_dir")
	if devlogDir == "" {
		devlogDir = "_rules/_devlog"
	}
	if !filepath.IsAbs(devlogDir) {
		devlogDir = filepath.Join(path, devlogDir)
	}
	res, err := devlog.PruneGhosts(ctx, store.UnderlyingDB(), devlog.NewPresence(devlogDir))
	if err != nil {
		return err
	}
	if res.Restored > 0 {
		fmt.Printf("  Kept %d session(s) whose file exists on another branch or worktree\n", res.Restored)
	}
	if res.RowsRemoved > 0 {
		fmt.Printf("  Removed %d ghost entries from _index.md\n", res.RowsRemoved)
	}
	fmt.Printf("  Deleted %d ghost sessions from database\n", res.Deleted)
	return nil
}
