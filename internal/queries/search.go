package queries

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type SearchResult struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Narrative string  `json:"narrative"` // Snippet or Preview
	Score     float64 `json:"score"`
	Reason    string  `json:"reason"`    // e.g. "text match", "entity:Modal"
	
	// Explanation fields (for --explain)
	BM25        float64 `json:"bm25"`
	PhraseBonus float64 `json:"phrase_bonus"`
	NearBonus   float64 `json:"near_bonus"`
	EntityBonus float64 `json:"entity_bonus"`
	RecencyBonus float64 `json:"recency_bonus"`
	
	IsLowConfidence bool `json:"is_low_confidence"`
}

type SearchOptions struct {
	Query    string
	Author   string
	Limit    int
	Strict   bool // If true, only BM25 on sessions, no expansion
	TextOnly bool // If true, only BM25, no entity lookup
	Preview  bool // If true, show first 3 lines instead of snippet
	Explain  bool // If true, show score breakdown
}

type SearchResponse struct {
	Results         []SearchResult
	RelatedEntities []string
	Strategy        string // e.g. "exact", "hybrid", "fallback"
}

// decomposeQuery transforms raw input into various FTS5 query forms
func decomposeQuery(query string) (phrase, near, and, or string, tokens []string) {
	// Clean query and extract tokens
	// Match alpha-numeric and dashes, treat others as separators
	re := regexp.MustCompile(`[^\w\s*]+`)
	clean := re.ReplaceAllString(query, " ")
	rawTokens := strings.Fields(clean)
	
	for _, t := range rawTokens {
		t = strings.TrimSpace(t)
		if t != "" {
			tokens = append(tokens, t)
		}
	}

	if len(tokens) == 0 {
		return "", "", "", "", nil
	}

	// Phrase: "token1 token2"
	phrase = "\"" + strings.Join(tokens, " ") + "\""
	
	// NEAR: NEAR(token1 token2, 5)
	if len(tokens) > 1 {
		near = "NEAR(" + strings.Join(tokens, " ") + ", 5)"
	} else {
		near = tokens[0] + "*"
	}
	
	// AND: token1* AND token2*
	var andParts []string
	for _, t := range tokens {
		if !strings.HasSuffix(t, "*") {
			andParts = append(andParts, t+"*")
		} else {
			andParts = append(andParts, t)
		}
	}
	and = strings.Join(andParts, " AND ")
	
	// OR: token1* OR token2*
	or = strings.Join(andParts, " OR ")
	
	return phrase, near, and, or, tokens
}

func HybridSearch(ctx context.Context, db *sql.DB, opts SearchOptions) (SearchResponse, error) {
	var response SearchResponse
	
	phrase, near, and, or, tokens := decomposeQuery(opts.Query)
	if len(tokens) == 0 {
		return response, nil
	}

	// 1. Initial attempt: Phrase + NEAR + AND (Hybrid)
	response.Strategy = "hybrid"
	results, err := executeRankedSearch(ctx, db, phrase, near, and, tokens, opts)
	
	// 2. Fallback: If no results, try OR (Broad)
	if (err != nil || len(results) == 0) && !opts.Strict {
		response.Strategy = "fallback"
		results, err = executeRankedSearch(ctx, db, phrase, near, or, tokens, opts)
		for i := range results {
			results[i].IsLowConfidence = true
		}
	}

	if err != nil {
		return response, err
	}

	// 3. Entity Search & Expansion (only if not Strict/TextOnly)
	if !opts.Strict && !opts.TextOnly {
		entityQuery := `SELECT name FROM entities_fts WHERE entities_fts MATCH ? LIMIT 5`
		// Use OR query for entity matching to be broader
		eRows, err := db.QueryContext(ctx, entityQuery, or)
		if err == nil {
			defer eRows.Close()
			for eRows.Next() {
				var name string
				if err := eRows.Scan(&name); err == nil {
					response.RelatedEntities = append(response.RelatedEntities, name)
				}
			}
		}
	}

	response.Results = results
	return response, nil
}

func executeRankedSearch(ctx context.Context, db *sql.DB, phrase, near, mainQuery string, tokens []string, opts SearchOptions) ([]SearchResult, error) {
	// snippet column is index 2 (narrative)
	snippetFunc := "snippet(sessions_fts, 2, '<b>', '</b>', '...', 64)"

	// Hybrid Scoring Formula Implementation
	// weights: id(0.0), title(10.0), narrative(1.0), author(0.5), agent(0.5), branch(0.5)
	
	sqlQuery := fmt.Sprintf(`
		WITH hits AS (
			SELECT 
				rowid, 
				bm25(sessions_fts, 0.0, 10.0, 1.0, 0.5, 0.5, 0.5) as bm25_score,
				%s as snippet
			FROM sessions_fts
			WHERE sessions_fts MATCH ?
		)
		SELECT 
			s.id, 
			s.title, 
			h.snippet,
			s.narrative,
			h.bm25_score,
			(julianday('now') - julianday(s.timestamp)) as age_days
		FROM hits h
		JOIN sessions s ON s.rowid = h.rowid
	`, snippetFunc)

	var args []interface{}
	args = append(args, mainQuery)

	if opts.Author != "" {
		sqlQuery += " WHERE s.author LIKE ?"
		args = append(args, "%"+opts.Author+"%")
	}

	sqlQuery += " ORDER BY bm25_score ASC LIMIT ?"
	args = append(args, opts.Limit * 2) 

	rows, err := db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var fullNarrative string
		var ageDays float64
		var snippet string
		
		if err := rows.Scan(&r.ID, &r.Title, &snippet, &fullNarrative, &r.BM25, &ageDays); err != nil {
			continue
		}

		// Calculate Bonuses in Go for more flexibility
		
		// 1. Phrase Bonus: +3.0 if title contains exact phrase
		cleanPhrase := strings.Trim(phrase, "\"")
		if strings.Contains(strings.ToLower(r.Title), strings.ToLower(cleanPhrase)) {
			r.PhraseBonus = 3.0
		}

		// 2. Proximity (Near) Bonus: +1.5 if narrative contains all tokens within 20 chars of each other
		// Simplified NEAR check in Go: check if all tokens appear in narrative
		allPresent := true
		for _, t := range tokens {
			if !strings.Contains(strings.ToLower(fullNarrative), strings.ToLower(t)) {
				allPresent = false
				break
			}
		}
		if allPresent {
			r.NearBonus = 1.5
		}

		// 3. Recency Bonus: max 0.75, decays over 30 days
		r.RecencyBonus = 0.75 * (1.0 - (ageDays / 30.0))
		if r.RecencyBonus < 0 {
			r.RecencyBonus = 0
		}

		// 4. Entity Pair Bonus: +0.75 if multiple tokens appear together
		// (Already partially covered by Near bonus, but we can refine)
		r.EntityBonus = 0 

		// Final Score (BM25 is lower=better, bonuses reduce score)
		r.Score = r.BM25 - r.PhraseBonus - r.NearBonus - r.RecencyBonus - r.EntityBonus
		
		if opts.Preview {
			// Extract first 3 lines of narrative
			lines := strings.Split(fullNarrative, "\n")
			count := 3
			if len(lines) < 3 {
				count = len(lines)
			}
			r.Narrative = strings.Join(lines[:count], "\n")
		} else {
			r.Narrative = snippet
		}

		results = append(results, r)
	}

	// Final sort by our hybrid score (ascending = better)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score < results[j].Score
	})

	if len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}
