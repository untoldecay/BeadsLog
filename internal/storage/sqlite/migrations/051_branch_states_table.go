package migrations

import (
	"database/sql"
)

func MigrateBranchStatesTable(db *sql.DB) error {
	// Create branch_states table for human intent (paused, abandoned)
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS branch_states (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			state TEXT NOT NULL,           -- active, paused, abandoned
			scope_type TEXT NOT NULL,      -- branch, entity, file, task, session
			scope_ref TEXT NOT NULL,       -- the actual reference (e.g. branch name, entity ID)
			short_reason TEXT,
			full_reason_ref TEXT,          -- link to the special devlog entry ID
			actor TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			commit_sha TEXT,
			branch_ref TEXT,
			UNIQUE(scope_type, scope_ref)
		);
	`)
	if err != nil {
		return err
	}

	// Create branch_cache table for daemon's Git facts
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS branch_cache (
			branch_name TEXT PRIMARY KEY,
			last_validated_sha TEXT,
			is_merged INTEGER DEFAULT 0,
			is_deleted INTEGER DEFAULT 0,
			last_checked_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return err
	}

	return nil
}
