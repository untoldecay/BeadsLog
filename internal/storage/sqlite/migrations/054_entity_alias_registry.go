package migrations

import (
	"database/sql"
)

func MigrateEntityAliasRegistry(db *sql.DB) error {
	// Create entity_aliases table to make aliases "sticky" and "reversible"
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS entity_aliases (
			alias_name TEXT PRIMARY KEY,
			canonical_id TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(canonical_id) REFERENCES entities(id) ON DELETE CASCADE
		);
	`)
	return err
}
