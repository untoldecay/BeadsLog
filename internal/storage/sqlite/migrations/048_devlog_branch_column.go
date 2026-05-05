package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateDevlogBranchColumn(db *sql.DB) error {
	// 1. Add branch column to sessions table
	_, err := db.Exec(`
		ALTER TABLE sessions ADD COLUMN branch TEXT;
	`)
	if err != nil {
		// Ignore if column already exists (idempotency)
	}

	// 2. Update FTS triggers if sessions_fts exists
	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='sessions_fts'").Scan(&name)
	if err == nil {
		// Drop existing triggers
		db.Exec("DROP TRIGGER IF EXISTS sessions_ai")
		db.Exec("DROP TRIGGER IF EXISTS sessions_au")
		db.Exec("DROP TRIGGER IF EXISTS sessions_ad")

		// Recreate FTS table with all desired columns including branch
		db.Exec("DROP TABLE IF EXISTS sessions_fts")
		_, err = db.Exec(`
			CREATE VIRTUAL TABLE sessions_fts USING fts5(
				id UNINDEXED,
				title,
				narrative,
				author,
				agent,
				branch,
				content='sessions',
				content_rowid='rowid'
			);
		`)
		if err != nil {
			return fmt.Errorf("failed to recreate sessions_fts: %w", err)
		}

		// Re-populate
		db.Exec("INSERT INTO sessions_fts(sessions_fts) VALUES('rebuild')")

		// Re-create triggers
		_, err = db.Exec(`
			CREATE TRIGGER sessions_ai AFTER INSERT ON sessions BEGIN
			  INSERT INTO sessions_fts(rowid, id, title, narrative, author, agent, branch)
			  VALUES (new.rowid, new.id, new.title, new.narrative, new.author, new.agent, new.branch);
			END;
			CREATE TRIGGER sessions_ad AFTER DELETE ON sessions BEGIN
			  INSERT INTO sessions_fts(sessions_fts, rowid, id, title, narrative, author, agent, branch)
			  VALUES('delete', old.rowid, old.id, old.title, old.narrative, old.author, old.agent, old.branch);
			END;
			CREATE TRIGGER sessions_au AFTER UPDATE ON sessions BEGIN
			  INSERT INTO sessions_fts(sessions_fts, rowid, id, title, narrative, author, agent, branch)
			  VALUES('delete', old.rowid, old.id, old.title, old.narrative, old.author, old.agent, old.branch);
			  INSERT INTO sessions_fts(rowid, id, title, narrative, author, agent, branch)
			  VALUES (new.rowid, new.id, new.title, new.narrative, new.author, new.agent, new.branch);
			END;
		`)
		if err != nil {
			return fmt.Errorf("failed to recreate FTS triggers: %w", err)
		}
	}

	return nil
}
