package fix

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/untoldecay/BeadsLog/internal/beads"
	"github.com/untoldecay/BeadsLog/internal/configfile"
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

	db := store.UnderlyingDB()

	// 2. Identify ghost sessions (ID and Filename)
	rows, err := db.Query("SELECT id, filename, title FROM sessions WHERE is_ghost = 1")
	if err != nil {
		return fmt.Errorf("failed to query ghosts: %w", err)
	}
	defer rows.Close()

	ghostIDs := make(map[string]bool)
	ghostFiles := make(map[string]bool)
	ghostTitles := make(map[string]bool)
	
	for rows.Next() {
		var id, file, title string
		if err := rows.Scan(&id, &file, &title); err == nil {
			ghostIDs[id] = true
			ghostFiles[file] = true
			ghostTitles[title] = true
		}
	}

	if len(ghostIDs) == 0 {
		return nil
	}

	// 3. Prune from _index.md
	devlogDir := ""
	db.QueryRow("SELECT value FROM config WHERE key = 'devlog_dir'").Scan(&devlogDir)
	if devlogDir == "" {
		devlogDir = "_rules/_devlog"
	}
	indexPath := filepath.Join(path, devlogDir, "_index.md")

	if _, err := os.Stat(indexPath); err == nil {
		content, err := os.ReadFile(indexPath)
		if err == nil {
			lines := strings.Split(string(content), "\n")
			newLines := make([]string, 0, len(lines))
			prunedFromIndex := 0

			inTable := false
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.Contains(trimmed, "| Subject | Problems |") {
					inTable = true
					newLines = append(newLines, line)
					continue
				}
				
				if inTable && strings.HasPrefix(trimmed, "|") && !strings.HasPrefix(trimmed, "|---") {
					parts := strings.Split(trimmed, "|")
					if len(parts) >= 6 {
						subj := strings.TrimSpace(parts[1])
						// Check if this row matches a ghost session
						if ghostTitles[subj] {
							prunedFromIndex++
							continue
						}
					}
				}
				newLines = append(newLines, line)
			}

			if prunedFromIndex > 0 {
				err = os.WriteFile(indexPath, []byte(strings.Join(newLines, "\n")), 0644)
				if err != nil {
					return fmt.Errorf("failed to update index: %w", err)
				}
				fmt.Printf("  Removed %d ghost entries from %s\n", prunedFromIndex, indexPath)
			}
		}
	}

	// 4. Prune from database
	// 4a. Prune session links
	_, _ = db.Exec("DELETE FROM session_entities WHERE session_id IN (SELECT id FROM sessions WHERE is_ghost = 1)")
	
	// 4b. Prune relationship links
	_, _ = db.Exec("DELETE FROM entity_deps WHERE discovered_in IN (SELECT id FROM sessions WHERE is_ghost = 1)")

	// 4c. Prune extraction log
	_, _ = db.Exec("DELETE FROM extraction_log WHERE session_id IN (SELECT id FROM sessions WHERE is_ghost = 1)")

	// 4d. Delete the sessions
	res, err := db.Exec("DELETE FROM sessions WHERE is_ghost = 1")
	if err != nil {
		return fmt.Errorf("failed to delete ghost sessions: %w", err)
	}
	
	deleted, _ := res.RowsAffected()
	fmt.Printf("  Deleted %d ghost sessions from database\n", deleted)

	return nil
}
