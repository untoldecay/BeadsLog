package migrations

import (
	"database/sql"
)

func MigrateDevlogCommitSHA(db *sql.DB) error {
	// Add commit_sha column to sessions table
	_, err := db.Exec(`
		ALTER TABLE sessions ADD COLUMN commit_sha TEXT;
	`)
	if err != nil {
		// Ignore if column already exists
	}
	return nil
}
