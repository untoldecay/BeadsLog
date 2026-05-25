package migrations

import (
	"database/sql"
)

func MigrateEntityDepsHardening(db *sql.DB) error {
	// Add confidence column to entity_deps
	_, _ = db.Exec(`ALTER TABLE entity_deps ADD COLUMN confidence REAL DEFAULT 1.0;`)
	
	// Add source column to entity_deps
	_, _ = db.Exec(`ALTER TABLE entity_deps ADD COLUMN source TEXT DEFAULT 'manual';`)

	return nil
}
