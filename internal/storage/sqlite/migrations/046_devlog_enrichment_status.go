package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateDevlogEnrichmentStatus(db *sql.DB) error {
	// Add enrichment_status to sessions table
	// 0: pending, 1: regex_done (waiting for AI), 2: ai_crystallized (done), 3: failed
	var hasStatus bool
	err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('sessions') WHERE name='enrichment_status'").Scan(&hasStatus)
	if err != nil {
		return fmt.Errorf("failed to check for enrichment_status column: %w", err)
	}
	if !hasStatus {
		_, err = db.Exec("ALTER TABLE sessions ADD COLUMN enrichment_status INTEGER DEFAULT 0")
		if err != nil {
			return fmt.Errorf("failed to add enrichment_status column: %w", err)
		}
	}

	return nil
}