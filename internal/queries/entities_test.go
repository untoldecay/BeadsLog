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
		source TEXT DEFAULT 'manual',
		PRIMARY KEY(from_entity, to_entity, relationship),
		FOREIGN KEY(from_entity) REFERENCES entities(id),
		FOREIGN KEY(to_entity) REFERENCES entities(id),
		FOREIGN KEY(discovered_in) REFERENCES sessions(id)
	);
	CREATE TABLE alias_dismissals (
		name_a TEXT NOT NULL,
		name_b TEXT NOT NULL,
		dismissed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		dismissed_by TEXT,
		PRIMARY KEY (name_a, name_b)
	);
	CREATE TABLE link_dismissals (
		name_a TEXT NOT NULL,
		name_b TEXT NOT NULL,
		dismissed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		dismissed_by TEXT,
		PRIMARY KEY (name_a, name_b)
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

	// e1/e2: similar names (containment variant), co-occur in 2 sessions → suggested
	_, _ = db.Exec("INSERT INTO entities (id, name) VALUES ('e1', 'ollama-extractor')")
	_, _ = db.Exec("INSERT INTO entities (id, name) VALUES ('e2', 'ollamaextract')")
	// e4/e5: dissimilar names, perfect co-occurrence → filtered (related ≠ identical)
	_, _ = db.Exec("INSERT INTO entities (id, name) VALUES ('e4', 'cut-release')")
	_, _ = db.Exec("INSERT INTO entities (id, name) VALUES ('e5', 'bump-version')")
	// e6/e7: similar names but only ONE shared session → filtered (session floor)
	_, _ = db.Exec("INSERT INTO entities (id, name) VALUES ('e6', 'auth-service')")
	_, _ = db.Exec("INSERT INTO entities (id, name) VALUES ('e7', 'auth-services')")

	for _, pair := range [][2]string{{"s1", "e1"}, {"s1", "e2"}, {"s2", "e1"}, {"s2", "e2"},
		{"s1", "e4"}, {"s1", "e5"}, {"s2", "e4"}, {"s2", "e5"},
		{"s3", "e6"}, {"s3", "e7"}} {
		_, _ = db.Exec("INSERT INTO session_entities (session_id, entity_id) VALUES (?, ?)", pair[0], pair[1])
	}

	suggestions, err := GetAliasSuggestions(ctx, db, 0.8, 0)
	if err != nil {
		t.Fatalf("GetAliasSuggestions failed: %v", err)
	}

	if len(suggestions) != 1 {
		t.Fatalf("Expected 1 suggestion (similar names, 2+ sessions), got %d: %+v", len(suggestions), suggestions)
	}
	if suggestions[0].EntityA != "ollama-extractor" && suggestions[0].EntityB != "ollama-extractor" {
		t.Errorf("Expected ollama pair, got %+v", suggestions[0])
	}
	if suggestions[0].Similarity != 1.0 {
		t.Errorf("Expected session overlap 1.0, got %f", suggestions[0].Similarity)
	}
	if suggestions[0].NameSimilarity < 0.55 {
		t.Errorf("Expected name similarity >= 0.55, got %f", suggestions[0].NameSimilarity)
	}

	// Dismiss the pair — never suggested again, regardless of argument order
	if err := DismissAliasPair(ctx, db, "ollamaextract", "ollama-extractor", "test"); err != nil {
		t.Fatalf("DismissAliasPair failed: %v", err)
	}
	suggestions, err = GetAliasSuggestions(ctx, db, 0.8, 0)
	if err != nil {
		t.Fatalf("GetAliasSuggestions after dismissal failed: %v", err)
	}
	if len(suggestions) != 0 {
		t.Fatalf("Expected 0 suggestions after dismissal, got %d: %+v", len(suggestions), suggestions)
	}
}

func TestAutoAliasDuplicates(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	_, _ = db.Exec("INSERT INTO sessions (id, title, timestamp) VALUES ('s1', 'Session 1', '2026-05-01T10:00:00Z')")
	_, _ = db.Exec("INSERT INTO entities (id, name, mention_count) VALUES ('e1', 'ollamaextractor', 10)")
	_, _ = db.Exec("INSERT INTO entities (id, name, mention_count) VALUES ('e2', 'ollama-extractor', 5)")
	_, _ = db.Exec("INSERT INTO entities (id, name, mention_count) VALUES ('e3', 'ollama', 7)")
	_, _ = db.Exec("INSERT INTO session_entities (session_id, entity_id) VALUES ('s1', 'e2')")

	merged, err := AutoAliasDuplicates(ctx, db)
	if err != nil {
		t.Fatalf("AutoAliasDuplicates failed: %v", err)
	}
	if merged != 1 {
		t.Errorf("expected 1 merge, got %d", merged)
	}

	// Canonical (highest mention_count) absorbed the variant's mentions
	var count int
	if err := db.QueryRow("SELECT mention_count FROM entities WHERE id = 'e1'").Scan(&count); err != nil {
		t.Fatalf("canonical entity missing: %v", err)
	}
	if count != 15 {
		t.Errorf("expected mention_count 15, got %d", count)
	}

	// Variant entity gone, its name preserved in the registry
	var n int
	_ = db.QueryRow("SELECT COUNT(*) FROM entities WHERE id = 'e2'").Scan(&n)
	if n != 0 {
		t.Error("variant entity row should be deleted")
	}
	var canonical string
	if err := db.QueryRow("SELECT canonical_id FROM entity_aliases WHERE alias_name = 'ollama-extractor'").Scan(&canonical); err != nil || canonical != "e1" {
		t.Errorf("registry mapping missing or wrong: %q, %v", canonical, err)
	}

	// Session link moved to canonical
	_ = db.QueryRow("SELECT COUNT(*) FROM session_entities WHERE entity_id = 'e1'").Scan(&n)
	if n != 1 {
		t.Errorf("session link not moved to canonical, got %d", n)
	}

	// Near-match "ollama" must NOT be merged
	_ = db.QueryRow("SELECT COUNT(*) FROM entities WHERE id = 'e3'").Scan(&n)
	if n != 1 {
		t.Error("near-match 'ollama' was wrongly merged")
	}
}

func TestLinkSuggestionsAndManualLink(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	for _, s := range []string{"s1", "s2", "s3", "s4"} {
		_, _ = db.Exec("INSERT INTO sessions (id, title, timestamp) VALUES (?, ?, '2026-01-01')", s, s)
	}
	_, _ = db.Exec("INSERT INTO entities (id, name) VALUES ('e-ed','editor'),('e-sl','slashmenu'),('e-x','lonely')")
	// editor & slashmenu co-occur in 4 sessions, no explicit edge.
	for _, p := range [][2]string{{"s1", "e-ed"}, {"s1", "e-sl"}, {"s2", "e-ed"}, {"s2", "e-sl"},
		{"s3", "e-ed"}, {"s3", "e-sl"}, {"s4", "e-ed"}, {"s4", "e-sl"}, {"s1", "e-x"}} {
		_, _ = db.Exec("INSERT INTO session_entities (session_id, entity_id) VALUES (?, ?)", p[0], p[1])
	}

	sug, err := GetLinkSuggestions(ctx, db, 4, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(sug) != 1 || sug[0].CoSessions != 4 {
		t.Fatalf("expected 1 suggestion with 4 shared sessions, got %+v", sug)
	}

	// Creating the edge removes the suggestion.
	if err := AddManualLink(ctx, db, "editor", "slashmenu", "uses"); err != nil {
		t.Fatalf("AddManualLink: %v", err)
	}
	sug, _ = GetLinkSuggestions(ctx, db, 4, 0)
	if len(sug) != 0 {
		t.Fatalf("expected 0 suggestions after linking, got %+v", sug)
	}

	// The manual link exports under source='user-link' (not extracted edges).
	links, _ := GetManualLinks(ctx, db)
	if len(links) != 1 || links[0].FromName != "editor" || links[0].ToName != "slashmenu" {
		t.Fatalf("expected 1 manual link editor->slashmenu, got %+v", links)
	}

	// Unknown entity errors.
	if err := AddManualLink(ctx, db, "editor", "ghost", "uses"); err == nil {
		t.Error("expected error linking to unknown entity")
	}

	// Dismissal removes a pair from suggestions permanently.
	_, _ = db.Exec("DELETE FROM entity_deps") // reset so editor/slashmenu resurfaces
	if err := DismissLinkPair(ctx, db, "slashmenu", "editor", "test"); err != nil {
		t.Fatal(err)
	}
	sug, _ = GetLinkSuggestions(ctx, db, 4, 0)
	if len(sug) != 0 {
		t.Fatalf("expected 0 suggestions after dismissal, got %+v", sug)
	}
}
