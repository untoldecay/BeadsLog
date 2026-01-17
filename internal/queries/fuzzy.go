package queries

import (
	"context"
	"database/sql"
)

// SuggestEntities finds potential entity matches for a term that yielded no results.
func SuggestEntities(ctx context.Context, db *sql.DB, term string) ([]string, error) {
	// Strategy 1: FTS Prefix Match on Entities
	// "auth" -> "auth*" matches "authentication-service"
	// FTS5 MATCH syntax: prefix query
	query := `
        SELECT name FROM entities_fts 
        WHERE entities_fts MATCH ? 
        LIMIT 5
    `
	matchTerm := term + "*"

	rows, err := db.QueryContext(ctx, query, matchTerm)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err == nil {
				// Avoid duplicates if we were to combine lists, 
				// but here we just return FTS results if found.
				// Actually, we should probably dedupe if we mix strategies.
				// For now, return early if FTS finds something.
				// Wait, rows.Next() loop needs to finish to capture all.
			}
		}
	}

	// Re-run properly capturing results
	var suggestions []string
	rows, err = db.QueryContext(ctx, query, matchTerm)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err == nil {
				suggestions = append(suggestions, name)
			}
		}
	}

	// Strategy 2: If FTS returned nothing (or few?), try fuzzy LIKE on base table
	if len(suggestions) == 0 {
		queryLike := `
            SELECT name FROM entities 
            WHERE name LIKE ? 
            ORDER BY length(name) ASC -- Shortest match first is a simple heuristic
            LIMIT 5
        `
		rowsLike, err := db.QueryContext(ctx, queryLike, "%"+term+"%")
		if err == nil {
			defer rowsLike.Close()
			for rowsLike.Next() {
				var name string
				if err := rowsLike.Scan(&name); err == nil {
					suggestions = append(suggestions, name)
				}
			}
		}
	}

	return suggestions, nil
}
