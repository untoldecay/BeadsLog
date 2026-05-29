package queries

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}

	schema := `
	CREATE TABLE sessions (
		id TEXT PRIMARY KEY,
		title TEXT,
		timestamp DATETIME
	);
	CREATE TABLE entities (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		preferred_name TEXT,
		mention_count INTEGER DEFAULT 1,
		first_seen DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE entity_aliases (
		alias_name TEXT PRIMARY KEY,
		canonical_id TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(canonical_id) REFERENCES entities(id) ON DELETE CASCADE
	);
	CREATE TABLE session_entities (
		session_id TEXT,
		entity_id TEXT,
		relevance TEXT DEFAULT 'mentioned',
		PRIMARY KEY(session_id, entity_id),
		FOREIGN KEY(session_id) REFERENCES sessions(id),
		FOREIGN KEY(entity_id) REFERENCES entities(id)
	);
	CREATE TABLE entity_deps (
		from_entity TEXT,
		to_entity TEXT,
		relationship TEXT,
		discovered_in TEXT,
		PRIMARY KEY(from_entity, to_entity, relationship),
		FOREIGN KEY(from_entity) REFERENCES entities(id),
		FOREIGN KEY(to_entity) REFERENCES entities(id),
		FOREIGN KEY(discovered_in) REFERENCES sessions(id)
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestAliasEntities(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	// Setup data
	_, _ = db.Exec("INSERT INTO sessions (id, title, timestamp) VALUES ('s1', 'Session 1', '2026-05-01T10:00:00Z')")
	_, _ = db.Exec("INSERT INTO sessions (id, title, timestamp) VALUES ('s2', 'Session 2', '2026-05-02T10:00:00Z')")
	
	_, _ = db.Exec("INSERT INTO entities (id, name, mention_count, first_seen) VALUES ('e1', 'Target', 10, '2026-05-01T10:00:00Z')")
	_, _ = db.Exec("INSERT INTO entities (id, name, mention_count, first_seen) VALUES ('e2', 'Alias1', 5, '2026-05-02T10:00:00Z')")
	_, _ = db.Exec("INSERT INTO entities (id, name, mention_count, first_seen) VALUES ('e3', 'Alias2', 3, '2026-04-30T10:00:00Z')")

	// Session mentions
	_, _ = db.Exec("INSERT INTO session_entities (session_id, entity_id) VALUES ('s1', 'e1')")
	_, _ = db.Exec("INSERT INTO session_entities (session_id, entity_id) VALUES ('s1', 'e2')") // Conflict potential
	_, _ = db.Exec("INSERT INTO session_entities (session_id, entity_id) VALUES ('s2', 'e3')")

	// Dependencies
	_, _ = db.Exec("INSERT INTO entity_deps (from_entity, to_entity, relationship, discovered_in) VALUES ('e2', 'some-other', 'uses', 's1')")
	_, _ = db.Exec("INSERT INTO entity_deps (from_entity, to_entity, relationship, discovered_in) VALUES ('another', 'e3', 'calls', 's2')")

	err := AliasEntities(ctx, db, "e1", []ResolvedEntity{
		{ID: "e2", Name: "Alias1"},
		{ID: "e3", Name: "Alias2"},
	})
	if err != nil {
		t.Fatalf("AliasEntities failed: %v", err)
	}

	// Verify Target Entity
	var count int
	var firstSeen string
	err = db.QueryRow("SELECT mention_count, first_seen FROM entities WHERE id = 'e1'").Scan(&count, &firstSeen)
	if err != nil {
		t.Fatalf("Target entity not found: %v", err)
	}
	if count != 18 { // 10 + 5 + 3
		t.Errorf("Expected 18 mentions, got %d", count)
	}
	if firstSeen != "2026-04-30T10:00:00Z" { // earliest of all
		t.Errorf("Expected 2026-04-30T10:00:00Z, got %s", firstSeen)
	}

	// Verify Aliases Deleted
	var exists bool
	_ = db.QueryRow("SELECT 1 FROM entities WHERE id IN ('e2', 'e3')").Scan(&exists)
	if exists {
		t.Error("Alias entities still exist")
	}

	// Verify Session Entities
	var sCount int
	db.QueryRow("SELECT COUNT(*) FROM session_entities WHERE entity_id = 'e1'").Scan(&sCount)
	if sCount != 2 { // 's1' and 's2'
		t.Errorf("Expected 2 sessions for target, got %d", sCount)
	}

	// Verify Deps
	var fromCount, toCount int
	db.QueryRow("SELECT COUNT(*) FROM entity_deps WHERE from_entity = 'e1'").Scan(&fromCount)
	db.QueryRow("SELECT COUNT(*) FROM entity_deps WHERE to_entity = 'e1'").Scan(&toCount)
	if fromCount != 1 {
		t.Errorf("Expected 1 outgoing dep for target, got %d", fromCount)
	}
	if toCount != 1 {
		t.Errorf("Expected 1 incoming dep for target, got %d", toCount)
	}
}

func TestGetAliasSuggestions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	// Setup data
	_, _ = db.Exec("INSERT INTO sessions (id, title, timestamp) VALUES ('s1', 'Session 1', '2026-05-01T10:00:00Z')")
	_, _ = db.Exec("INSERT INTO sessions (id, title, timestamp) VALUES ('s2', 'Session 2', '2026-05-02T10:00:00Z')")
	_, _ = db.Exec("INSERT INTO sessions (id, title, timestamp) VALUES ('s3', 'Session 3', '2026-05-03T10:00:00Z')")
	
	_, _ = db.Exec("INSERT INTO entities (id, name) VALUES ('e1', 'A')")
	_, _ = db.Exec("INSERT INTO entities (id, name) VALUES ('e2', 'B')")
	_, _ = db.Exec("INSERT INTO entities (id, name) VALUES ('e3', 'C')")

	// e1 and e2 appear together in all sessions (Similarity = 1.0)
	_, _ = db.Exec("INSERT INTO session_entities (session_id, entity_id) VALUES ('s1', 'e1')")
	_, _ = db.Exec("INSERT INTO session_entities (session_id, entity_id) VALUES ('s1', 'e2')")
	_, _ = db.Exec("INSERT INTO session_entities (session_id, entity_id) VALUES ('s2', 'e1')")
	_, _ = db.Exec("INSERT INTO session_entities (session_id, entity_id) VALUES ('s2', 'e2')")

	// e1 and e3 overlap only in s1
	_, _ = db.Exec("INSERT INTO session_entities (session_id, entity_id) VALUES ('s1', 'e3')")

	suggestions, err := GetAliasSuggestions(ctx, db, 0.8)
	if err != nil {
		t.Fatalf("GetAliasSuggestions failed: %v", err)
	}

	if len(suggestions) != 1 {
		t.Fatalf("Expected 1 suggestion, got %d", len(suggestions))
	}

	if suggestions[0].Similarity != 1.0 {
		t.Errorf("Expected similarity 1.0, got %f", suggestions[0].Similarity)
	}
}
