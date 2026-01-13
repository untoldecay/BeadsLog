package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateDevlogMissingFlag(db *sql.DB) error {
	var hasMissingColumn bool
	err := db.QueryRow(`
		SELECT COUNT(*) > 0 FROM pragma_table_info('sessions')
		WHERE name = 'is_missing'
	`).Scan(&hasMissingColumn)
	
	if err != nil {
		return fmt.Errorf("failed to check for is_missing column: %w", err)
	}
	
	if !hasMissingColumn {
		_, err = db.Exec(`ALTER TABLE sessions ADD COLUMN is_missing BOOLEAN DEFAULT 0`)
		if err != nil {
			return fmt.Errorf("failed to add is_missing column: %w", err)
		}
	}

	return nil
}
