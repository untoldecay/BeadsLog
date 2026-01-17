package queries

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

type SearchResult struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Narrative string  `json:"narrative"` // Snippet
	Score     float64 `json:"score"`
	Reason    string  `json:"reason"`    // e.g. "text match", "entity:Modal"
}

type SearchOptions struct {
	Query    string
	Limit    int
	Strict   bool // If true, only BM25 on sessions, no expansion
	TextOnly bool // If true, only BM25, no entity lookup
}

func HybridSearch(ctx context.Context, db *sql.DB, opts SearchOptions) ([]SearchResult, error) {
	results := make(map[string]SearchResult)

	// Prepare the match query
	// If not strict, we might want to make it friendlier
	// For now, we pass it through, assuming the user might use FTS syntax
	matchQuery := opts.Query
	if !opts.Strict {
		// UX Enhancement: If it's a simple word, append * for prefix match
		// e.g. "mod" -> "mod*"
		if !strings.ContainsAny(matchQuery, " \"*:()") {
			matchQuery = matchQuery + "*"
		}
	}

	// 1. BM25 Text Search
	textQuery := "" +
	             `
        SELECT s.id, s.title, snippet(sessions_fts, 1, '<b>', '</b>', '...', 64), bm25(sessions_fts) 
        FROM sessions_fts 
        JOIN sessions s ON sessions_fts.rowid = s.rowid
        WHERE sessions_fts MATCH ? 
        ORDER BY bm25(sessions_fts) 
        LIMIT ?
    `

	rows, err := db.QueryContext(ctx, textQuery, matchQuery, opts.Limit)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var r SearchResult
			if err := rows.Scan(&r.ID, &r.Title, &r.Narrative, &r.Score); err == nil {
				r.Reason = "text match"
				results[r.ID] = r
			}
		}
	}
	// Note: We ignore errors here (e.g. syntax error in query) to allow falling back
	// or returning what we have. But a syntax error likely fails everything.

	if opts.Strict || opts.TextOnly {
		return mapToSlice(results), nil
	}

	// 2. Entity Search & Expansion
	entityQuery := `
        SELECT rowid, name FROM entities_fts WHERE entities_fts MATCH ? LIMIT 5
    `
	eRows, err := db.QueryContext(ctx, entityQuery, matchQuery)
	var matchedEntities []string
	if err == nil {
		defer eRows.Close()
		for eRows.Next() {
			var id int64
			var name string
			if err := eRows.Scan(&id, &name); err == nil {
				matchedEntities = append(matchedEntities, name)
			}
		}
	}

	// For each matched entity, find related sessions
	for _, entityName := range matchedEntities {
		relatedQuery := `
            SELECT s.id, s.title, s.narrative
            FROM sessions s
            JOIN session_entities se ON s.id = se.session_id
            JOIN entities e ON se.entity_id = e.id
            WHERE e.name = ?
            LIMIT ?
        `
		rRows, err := db.QueryContext(ctx, relatedQuery, entityName, opts.Limit)
		if err == nil {
			defer rRows.Close()
			for rRows.Next() {
				var r SearchResult
				// Placeholder score for entity matches (since we don't have BM25 for them in this query)
				// We start with a base score.
				var baseScore float64 = -5.0 // Stronger than a weak text match?
				
				if err := rRows.Scan(&r.ID, &r.Title, &r.Narrative); err == nil {
					if existing, ok := results[r.ID]; ok {
						// Boost existing result
						// Lower score is better in FTS5 BM25
						existing.Score -= 2.0 
						if !strings.Contains(existing.Reason, "entity:") {
							existing.Reason += fmt.Sprintf(", entity:%s", entityName)
						}
						results[r.ID] = existing
					} else {
						// Add new result from entity relation
						r.Score = baseScore
						r.Reason = fmt.Sprintf("entity:%s", entityName)
						results[r.ID] = r
					}
				}
			}
		}
	}

	return mapToSlice(results), nil
}

func mapToSlice(m map[string]SearchResult) []SearchResult {
	s := make([]SearchResult, 0, len(m))
	for _, v := range m {
		s = append(s, v)
	}
	// Sort by Score (ascending = better relevance)
	sort.Slice(s, func(i, j int) bool {
		return s[i].Score < s[j].Score
	})
	return s
}
