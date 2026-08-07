package migrations

import (
	"database/sql"
)

func MigrateAliasDismissals(db *sql.DB) error {
	// Reviewed-and-rejected alias suggestions. Pairs stored with name_a < name_b
	// (lexicographic) so each pair has one canonical row; suggestions matching a
	// dismissal are never surfaced again, letting agent review converge.
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS alias_dismissals (
			name_a TEXT NOT NULL,
			name_b TEXT NOT NULL,
			dismissed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			dismissed_by TEXT,
			PRIMARY KEY (name_a, name_b)
		);
	`)
	return err
}
