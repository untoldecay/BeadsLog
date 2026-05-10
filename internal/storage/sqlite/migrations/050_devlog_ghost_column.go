package migrations

import (
	"database/sql"
)

func MigrateDevlogGhostColumn(db *sql.DB) error {
	// Add is_ghost column to sessions table
	_, err := db.Exec(`
		ALTER TABLE sessions ADD COLUMN is_ghost INTEGER DEFAULT 0;
	`)
	if err != nil {
		// Ignore if column already exists
	}
	return nil
}
