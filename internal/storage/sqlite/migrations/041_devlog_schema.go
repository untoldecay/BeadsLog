package migrations

import (
	"database/sql"
	"fmt"
)

func MigrateDevlogSchema(db *sql.DB) error {
	// Create sessions table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
		  id TEXT PRIMARY KEY,
		  title TEXT NOT NULL,
		  timestamp DATETIME NOT NULL,
		  status TEXT DEFAULT 'closed',
		  type TEXT, -- fix, feature, enhance, etc.
		  filename TEXT,
		  narrative TEXT,
		  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}

	// Create entities table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS entities (
		  id TEXT PRIMARY KEY,
		  name TEXT UNIQUE NOT NULL,
		  type TEXT DEFAULT 'component',
		  first_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
		  mention_count INTEGER DEFAULT 1
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create entities table: %w", err)
	}

	// Create session_entities table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS session_entities (
		  session_id TEXT,
		  entity_id TEXT,
		  relevance TEXT DEFAULT 'mentioned',
		  PRIMARY KEY(session_id, entity_id),
		  FOREIGN KEY(session_id) REFERENCES sessions(id),
		  FOREIGN KEY(entity_id) REFERENCES entities(id)
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create session_entities table: %w", err)
	}

	// Create entity_deps table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS entity_deps (
		  from_entity TEXT,
		  to_entity TEXT,
		  relationship TEXT,
		  discovered_in TEXT,
		  PRIMARY KEY(from_entity, to_entity, relationship),
		  FOREIGN KEY(from_entity) REFERENCES entities(id),
		  FOREIGN KEY(to_entity) REFERENCES entities(id),
		  FOREIGN KEY(discovered_in) REFERENCES sessions(id)
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create entity_deps table: %w", err)
	}

	return nil
}
