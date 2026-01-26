package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateExtractionEvolSchema(db *sql.DB) error {
	// Add confidence and source columns to entities table
	// SQLite doesn't support adding multiple columns in one ALTER TABLE
	
	// Add confidence column
	var hasConfidence bool
	err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('entities') WHERE name='confidence'").Scan(&hasConfidence)
	if err != nil {
		return fmt.Errorf("failed to check for confidence column: %w", err)
	}
	if !hasConfidence {
		_, err = db.Exec("ALTER TABLE entities ADD COLUMN confidence REAL DEFAULT 1.0")
		if err != nil {
			return fmt.Errorf("failed to add confidence column: %w", err)
		}
	}

	// Add source column
	var hasSource bool
	err = db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('entities') WHERE name='source'").Scan(&hasSource)
	if err != nil {
		return fmt.Errorf("failed to check for source column: %w", err)
	}
	if !hasSource {
		_, err = db.Exec("ALTER TABLE entities ADD COLUMN source TEXT DEFAULT 'regex'")
		if err != nil {
			return fmt.Errorf("failed to add source column: %w", err)
		}
	}

	// Create extraction_log table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS extraction_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			extractor TEXT,  -- regex|ollama
			input_length INTEGER,
			entities_found INTEGER,
			duration_ms INTEGER,
			FOREIGN KEY(session_id) REFERENCES sessions(id)
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create extraction_log table: %w", err)
	}

	return nil
}
