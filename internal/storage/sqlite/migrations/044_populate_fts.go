package migrations

import (
	"database/sql"
	"fmt"
)

// MigratePopulateFTS ensures FTS tables are populated if the base tables are not empty.
// Since we used content='...', standard INSERTs don't auto-populate existing data.
// We must issue a 'rebuild' command.
func MigratePopulateFTS(db *sql.DB) error {
	// Check if sessions table has data
	var sessionCount int
	// Use pure SQL check, ignoring errors (table might theoretically be missing if schema init failed weirdly)
	err := db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&sessionCount)
	if err == nil && sessionCount > 0 {
		// Rebuild the FTS index from the content table
		_, err = db.Exec("INSERT INTO sessions_fts(sessions_fts) VALUES('rebuild')")
		if err != nil {
			return fmt.Errorf("failed to rebuild sessions_fts: %w", err)
		}
	}

	// Check if entities table has data
	var entityCount int
	err = db.QueryRow("SELECT COUNT(*) FROM entities").Scan(&entityCount)
	if err == nil && entityCount > 0 {
		// Rebuild the FTS index from the content table
		_, err = db.Exec("INSERT INTO entities_fts(entities_fts) VALUES('rebuild')")
		if err != nil {
			return fmt.Errorf("failed to rebuild entities_fts: %w", err)
		}
	}

	return nil
}
