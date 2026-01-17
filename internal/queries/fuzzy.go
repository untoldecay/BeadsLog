package queries

import (
	"context"
	"database/sql"
	"fmt"
)

type ResolvedEntity struct {
	ID   string
	Name string
}

// ResolveEntities finds entities matching the term using Hybrid FTS + LIKE logic.
// This ensures consistent behavior between "Search" and "Graph/Impact" commands.
func ResolveEntities(ctx context.Context, db *sql.DB, term string, limit int) ([]ResolvedEntity, error) {
	var results []ResolvedEntity
	seen := make(map[string]bool)

	// Strategy 1: FTS Prefix Match
	queryFTS := `
        SELECT name FROM entities_fts 
        WHERE entities_fts MATCH ? 
        LIMIT ?
    `
	matchTerm := term + "*"
	rows, err := db.QueryContext(ctx, queryFTS, matchTerm, limit)
	if err != nil {
		return nil, fmt.Errorf("fts query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("fts scan failed: %w", err)
		}
		if !seen[name] {
			// Fetch real ID using case-insensitive comparison
			var id string
			idQuery := "SELECT id FROM entities WHERE LOWER(name) = LOWER(?)"
			if err := db.QueryRowContext(ctx, idQuery, name).Scan(&id); err != nil {
				if err != sql.ErrNoRows { // Only return error if it's not simply "not found"
					return nil, fmt.Errorf("id resolution query failed for name '%s': %w", name, err)
				}
				// If ErrNoRows, it means FTS matched but base table didn't have it (corrupted FTS?), skip this specific name
				continue
			}
			results = append(results, ResolvedEntity{ID: id, Name: name})
			seen[name] = true
		}
	}

	// Strategy 2: Fallback to LIKE if we have few results or FTS failed to produce enough
	if len(results) < limit {
		remaining := limit - len(results)
		if remaining <= 0 { // Already reached limit with FTS results
			return results, nil
		}
		queryLike := `
            SELECT id, name FROM entities 
            WHERE LOWER(name) LIKE LOWER(?) 
            ORDER BY length(name) ASC
            LIMIT ?
        `
		rowsLike, err := db.QueryContext(ctx, queryLike, "%"+term+"%", remaining)
		if err != nil {
			return nil, fmt.Errorf("like query failed: %w", err)
		}
		defer rowsLike.Close()

		for rowsLike.Next() {
			var id, name string
			if err := rowsLike.Scan(&id, &name); err != nil {
				return nil, fmt.Errorf("like scan failed: %w", err)
			}
			if !seen[name] { // Add only if not already found by FTS
				results = append(results, ResolvedEntity{ID: id, Name: name})
				seen[name] = true
			}
		}
	}

	return results, nil
}

// SuggestEntities finds potential entity matches for a term that yielded no results.
// Wraps ResolveEntities but returns only names.
func SuggestEntities(ctx context.Context, db *sql.DB, term string) ([]string, error) {
	entities, err := ResolveEntities(ctx, db, term, 5) // Always limit suggestions to 5
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entities {
		names = append(names, e.Name)
	}
	return names, nil
}
