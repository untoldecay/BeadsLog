package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateDevlogFileHash(db *sql.DB) error {
	var hasHashColumn bool
	err := db.QueryRow(`
		SELECT COUNT(*) > 0 FROM pragma_table_info('sessions')
		WHERE name = 'file_hash'
	`).Scan(&hasHashColumn)
	
	if err != nil {
		return fmt.Errorf("failed to check for file_hash column: %w", err)
	}
	
	if !hasHashColumn {
		_, err = db.Exec(`ALTER TABLE sessions ADD COLUMN file_hash TEXT DEFAULT ''`)
		if err != nil {
			return fmt.Errorf("failed to add file_hash column: %w", err)
		}
	}

	return nil
}
