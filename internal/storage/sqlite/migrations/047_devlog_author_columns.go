package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateAuthorColumns(db *sql.DB) error {
	// 1. Add author columns to sessions table
	_, err := db.Exec(`
		ALTER TABLE sessions ADD COLUMN author TEXT;
	`)
	if err != nil {
		// Ignore if column already exists (idempotency)
	}

	_, err = db.Exec(`
		ALTER TABLE sessions ADD COLUMN author_email TEXT;
	`)
	if err != nil {
		// Ignore if column already exists
	}

	// 2. Update FTS triggers if sessions_fts exists
	// We check if the table exists first
	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='sessions_fts'").Scan(&name)
	if err == nil {
		// It exists, we need to drop and recreate it to include the new columns if we want them searchable
		// For now, let's at least ensure the triggers don't break or are updated.
		// Actually, the best way to handle FTS column additions is to drop and recreate.
		
		// Drop existing triggers
		db.Exec("DROP TRIGGER IF EXISTS sessions_ai")
		db.Exec("DROP TRIGGER IF EXISTS sessions_au")
		db.Exec("DROP TRIGGER IF EXISTS sessions_ad")

		// Recreate FTS table with author
		db.Exec("DROP TABLE IF EXISTS sessions_fts")
		_, err = db.Exec(`
			CREATE VIRTUAL TABLE sessions_fts USING fts5(
				id UNINDEXED,
				title,
				narrative,
				author,
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
			  INSERT INTO sessions_fts(rowid, id, title, narrative, author)
			  VALUES (new.rowid, new.id, new.title, new.narrative, new.author);
			END;
			CREATE TRIGGER sessions_ad AFTER DELETE ON sessions BEGIN
			  INSERT INTO sessions_fts(sessions_fts, rowid, id, title, narrative, author)
			  VALUES('delete', old.rowid, old.id, old.title, old.narrative, old.author);
			END;
			CREATE TRIGGER sessions_au AFTER UPDATE ON sessions BEGIN
			  INSERT INTO sessions_fts(sessions_fts, rowid, id, title, narrative, author)
			  VALUES('delete', old.rowid, old.id, old.title, old.narrative, old.author);
			  INSERT INTO sessions_fts(rowid, id, title, narrative, author)
			  VALUES (new.rowid, new.id, new.title, new.narrative, new.author);
			END;
		`)
		if err != nil {
			return fmt.Errorf("failed to recreate FTS triggers: %w", err)
		}
	}

	return nil
}
