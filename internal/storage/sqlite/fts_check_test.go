package sqlite

import (
	"context"
	"testing"
)

func TestFTS5Availability(t *testing.T) {
	// Use in-memory DB
	ctx := context.Background()
	store, err := New(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	db := store.UnderlyingDB()

	// Try to create an FTS5 table
	// schema.go already tries to create them, so if New() succeeded, 
	// chances are high, but let's test a manual explicit one to be sure.
	_, err = db.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS test_fts_check USING fts5(content)")
	if err != nil {
		t.Fatalf("FTS5 is NOT available: %v", err)
	}

	// Try to insert and match
	_, err = db.Exec("INSERT INTO test_fts_check(content) VALUES('hello world')")
	if err != nil {
		t.Fatalf("Failed to insert into FTS5: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_fts_check WHERE test_fts_check MATCH 'hello'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query FTS5: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 match, got %d", count)
	}
}
