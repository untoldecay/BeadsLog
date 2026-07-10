package queries

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/untoldecay/BeadsLog/internal/types"
)

type SearchResult struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Date      string  `json:"date"`
	Narrative string  `json:"narrative"` // Snippet or Preview
	Score     float64 `json:"score"`
	Reason    string  `json:"reason"`    // e.g. "text match", "entity:Modal"
	
	// Explanation fields (for --explain)
	BM25        float64 `json:"bm25"`
	PhraseBonus float64 `json:"phrase_bonus"`
	NearBonus   float64 `json:"near_bonus"`
	EntityBonus  float64 `json:"entity_bonus"`
	RecencyBonus float64 `json:"recency_bonus"`

	IsLowConfidence bool `json:"is_low_confidence"`

	// Lifecycle fields
	LifecycleStatus types.LifecycleState `json:"lifecycle_status"`
	StatusReason    string               `json:"status_reason"`
	IsValidated     bool                 `json:"is_validated"`
	Branch          string               `json:"branch"`
	Author          string               `json:"author"`
	AuthorEmail     string               `json:"author_email"`
	Agent           string               `json:"agent"`
}

type SearchOptions struct {
	Query         string
	Author        string
	CurrentBranch string
	Limit         int
	Strict        bool // If true, only BM25 on sessions, no expansion
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
		entityQuery := `
			SELECT COALESCE(e.preferred_name, ef.name) 
			FROM entities_fts ef
			JOIN entities e ON ef.rowid = e.rowid
			WHERE ef.name MATCH ? 
			LIMIT 5
		`
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

// sessionsLinkedToEntities returns the set of session IDs graph-linked to an
// entity whose name matches one of the query tokens, the raw query, or the
// separator-stripped join of all tokens ("payment-gateway" → "paymentgateway").
func sessionsLinkedToEntities(ctx context.Context, db *sql.DB, rawQuery string, tokens []string) map[string]bool {
	linked := make(map[string]bool)
	if len(tokens) == 0 {
		return linked
	}
	candidates := make([]string, 0, len(tokens)+2)
	for _, t := range tokens {
		candidates = append(candidates, strings.ToLower(strings.TrimSuffix(t, "*")))
	}
	candidates = append(candidates,
		strings.ToLower(strings.TrimSpace(rawQuery)),
		strings.ToLower(strings.Join(tokens, "")))

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(candidates)), ",")
	args := make([]interface{}, len(candidates))
	for i, c := range candidates {
		args[i] = c
	}
	// Match canonical names AND registry aliases, so merged variant
	// spellings ("PaymentGateway" → payment-gateway) still earn the bonus.
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT se.session_id
		FROM session_entities se
		JOIN entities e ON se.entity_id = e.id
		WHERE e.name IN (`+placeholders+`)
		UNION
		SELECT DISTINCT se.session_id
		FROM session_entities se
		JOIN entity_aliases a ON se.entity_id = a.canonical_id
		WHERE a.alias_name IN (`+placeholders+`)`, append(args, args...)...)
	if err != nil {
		return linked
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			linked[id] = true
		}
	}
	return linked
}

func executeRankedSearch(ctx context.Context, db *sql.DB, phrase, near, mainQuery string, tokens []string, opts SearchOptions) ([]SearchResult, error) {
	// snippet column is index 2 (narrative)
	snippetFunc := "snippet(sessions_fts, 2, '<b>', '</b>', '...', 128)"

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
			s.timestamp,
			h.snippet,
			s.narrative,
			h.bm25_score,
			(julianday('now') - julianday(s.timestamp)) as age_days,
			COALESCE(s.branch, 'N/A'),
			COALESCE(s.author, 'Unknown'),
			COALESCE(s.author_email, ''),
			COALESCE(s.agent, 'Unknown'),
			bs.state,
			bs.short_reason,
			COALESCE(bc.is_merged, 0),
			COALESCE(bc.is_deleted, 0)
		FROM hits h
		JOIN sessions s ON s.rowid = h.rowid
		LEFT JOIN branch_states bs ON (bs.scope_type = 'branch' AND (bs.scope_ref = s.branch OR bs.scope_ref = REPLACE(s.branch, ' (local)', ''))) 
			 OR (bs.scope_type = 'session' AND bs.scope_ref = s.id)
			 OR (bs.scope_type = 'entity' AND s.id IN (SELECT session_id FROM session_entities WHERE entity_id = bs.scope_ref))
			 OR (bs.full_reason_ref = s.id)
		LEFT JOIN branch_cache bc ON bc.branch_name = s.branch OR bc.branch_name = REPLACE(s.branch, ' (local)', '')
		WHERE s.is_ghost = 0
	`, snippetFunc)

	var args []interface{}
	args = append(args, mainQuery)

	if opts.Author != "" {
		sqlQuery += " AND s.author LIKE ?"
		args = append(args, "%"+opts.Author+"%")
	}

	sqlQuery += " ORDER BY bm25_score ASC LIMIT ?"
	args = append(args, opts.Limit * 2) 

	rows, err := db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entityLinked := sessionsLinkedToEntities(ctx, db, opts.Query, tokens)

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var fullNarrative, timestamp string
		var ageDays float64
		var snippet string
		var state sql.NullString
		var reason sql.NullString
		var isMerged, isDeleted int
		
		if err := rows.Scan(&r.ID, &r.Title, &timestamp, &snippet, &fullNarrative, &r.BM25, &ageDays, &r.Branch, &r.Author, &r.AuthorEmail, &r.Agent, &state, &reason, &isMerged, &isDeleted); err != nil {
			continue
		}

		// Lifecycle Logic
		cleanBranch := strings.TrimSuffix(r.Branch, " (local)")
		if state.Valid {
			r.LifecycleStatus = types.LifecycleState(state.String)
			r.StatusReason = reason.String
		} else {
			// Automatic derivation
			if isMerged == 1 {
				r.LifecycleStatus = types.StateActive
				r.IsValidated = true
			} else if isDeleted == 1 {
				r.LifecycleStatus = types.StatePaused
			} else if opts.CurrentBranch != "" && cleanBranch == opts.CurrentBranch {
				r.LifecycleStatus = types.StateActive
			} else if cleanBranch != "main" && cleanBranch != "develop" && cleanBranch != "master" {
				// Off-branch and not validated -> effectively paused
				r.LifecycleStatus = types.StatePaused
			} else {
				r.LifecycleStatus = types.StateActive
			}
		}

		// Format Date
		r.Date = timestamp
		if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
			r.Date = t.Format("2006-01-02")
		} else if len(timestamp) >= 10 {
			r.Date = timestamp[:10]
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

		// 4. Entity Bonus: +0.75 if the session is graph-linked to an entity
		// matching a query token (stronger signal than raw text presence)
		if entityLinked[r.ID] {
			r.EntityBonus = 0.75
		}

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

type ProximityWarning struct {
	State        types.LifecycleState
	ScopeType    types.ScopeType
	ScopeRef     string
	ShortReason  string
	FullReasonID string
}

// CheckProximityWarnings checks if the given scope overlaps with any paused or abandoned states
func CheckProximityWarnings(ctx context.Context, db *sql.DB, scopeType types.ScopeType, scopeRef string) ([]ProximityWarning, error) {
	query := `
		SELECT state, scope_type, scope_ref, short_reason, full_reason_ref
		FROM branch_states
		WHERE (scope_type = ? AND scope_ref = ?)
		   OR (scope_type = 'branch' AND scope_ref = (SELECT branch FROM sessions WHERE id = ? OR filename LIKE ? LIMIT 1))
	`
	
	rows, err := db.QueryContext(ctx, query, scopeType, scopeRef, scopeRef, "%"+scopeRef+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var warnings []ProximityWarning
	for rows.Next() {
		var w ProximityWarning
		if err := rows.Scan(&w.State, &w.ScopeType, &w.ScopeRef, &w.ShortReason, &w.FullReasonID); err == nil {
			warnings = append(warnings, w)
		}
	}
	return warnings, nil
}
