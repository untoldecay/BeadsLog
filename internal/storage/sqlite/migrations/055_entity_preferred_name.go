package migrations

import (
	"database/sql"
)

func MigrateEntityPreferredName(db *sql.DB) error {
	// Add preferred_name column to entities
	_, _ = db.Exec(`ALTER TABLE entities ADD COLUMN preferred_name TEXT;`)
	
	// Backfill preferred_name with existing lowercase name if null
	_, _ = db.Exec(`UPDATE entities SET preferred_name = name WHERE preferred_name IS NULL;`)

	return nil
}
