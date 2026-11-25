package devlog

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PruneResult reports what PruneGhosts did.
type PruneResult struct {
	Deleted     int // ghost sessions removed from the database
	RowsRemoved int // dead rows removed from _index.md
	Restored    int // sessions wrongly marked as ghosts, un-marked
}

// PruneGhosts deletes ghost sessions whose file is in no checkout and on no
// branch tip, and drops their _index.md rows so the next sync does not
// resurrect them. It is the only place that deletes ghosts.
func PruneGhosts(ctx context.Context, db *sql.DB, p *Presence) (PruneResult, error) {
	var res PruneResult
	if reason := p.Restriction(); reason != "" {
		return res, fmt.Errorf("refusing to prune: %s, and the database is shared. Prune from the main checkout after merging", reason)
	}
	if err := p.Err(); err != nil {
		return res, fmt.Errorf("refusing to prune: cannot check other branches: %w", err)
	}
	restored, err := RestoreFound(ctx, db, p)
	if err != nil {
		return res, err
	}
	res.Restored = restored

	rows, err := db.QueryContext(ctx, "SELECT id, filename FROM sessions WHERE is_ghost = 1")
	if err != nil {
		return res, err
	}
	var ids []string
	dead := map[string]bool{}
	for rows.Next() {
		var id, fn string
		if err := rows.Scan(&id, &fn); err != nil {
			_ = rows.Close()
			return res, err
		}
		ids = append(ids, id)
		dead[filepath.Base(fn)] = true
	}
	if err := rows.Close(); err != nil {
		return res, err
	}
	if len(ids) == 0 {
		return res, nil
	}

	res.RowsRemoved, err = removeIndexRows(filepath.Join(p.dir, "_index.md"), dead)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not update index: %v\n", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return res, err
	}
	defer func() { _ = tx.Rollback() }()
	for _, id := range ids {
		// Explicit cleanup: old databases have no cascades (and may lack a table).
		for _, q := range []string{
			"DELETE FROM session_entities WHERE session_id = ?",
			"DELETE FROM extraction_log WHERE session_id = ?",
			"DELETE FROM entity_deps WHERE discovered_in = ?",
		} {
			_, _ = tx.ExecContext(ctx, q, id)
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM sessions WHERE id = ?", id); err != nil {
			return res, err
		}
	}
	if err := tx.Commit(); err != nil {
		return res, err
	}
	res.Deleted = len(ids)
	return res, nil
}

// RestoreFound clears the ghost flag of sessions whose file exists after all
// (here, on a branch tip or in another worktree).
func RestoreFound(ctx context.Context, db *sql.DB, p *Presence) (int, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, filename FROM sessions WHERE is_ghost = 1")
	if err != nil {
		return 0, err
	}
	var ids []string
	for rows.Next() {
		var id, fn string
		if err := rows.Scan(&id, &fn); err != nil {
			_ = rows.Close()
			return 0, err
		}
		if s := p.Lookup(fn); s == Here || s == Elsewhere {
			ids = append(ids, id)
		}
	}
	if err := rows.Close(); err != nil {
		return 0, err
	}
	for _, id := range ids {
		if _, err := db.ExecContext(ctx, "UPDATE sessions SET is_ghost = 0, is_missing = 0 WHERE id = ?", id); err != nil {
			return 0, err
		}
	}
	return len(ids), nil
}

func removeIndexRows(indexPath string, dead map[string]bool) (int, error) {
	content, err := os.ReadFile(indexPath) // #nosec G304 - configured devlog index
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	var kept []string
	removed := 0
	for _, line := range strings.Split(string(content), "\n") {
		if rowIsDead(line, dead) {
			removed++
			continue
		}
		kept = append(kept, line)
	}
	if removed == 0 {
		return 0, nil
	}
	if err := os.WriteFile(indexPath, []byte(strings.Join(kept, "\n")), 0644); err != nil { // #nosec G306 - tracked markdown
		return 0, err
	}
	return removed, nil
}

func rowIsDead(line string, dead map[string]bool) bool {
	for name := range dead {
		if strings.Contains(line, "]("+name) {
			return true
		}
	}
	return false
}
