package queries

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"github.com/untoldecay/BeadsLog/internal/utils"
)

type ResolvedEntity struct {
	ID   string
	Name string
}

type Suggestion struct {
	Name     string
	Distance int    // -1 if not applicable
	Type     string // "exact", "typo", "fuzzy"
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
// It uses fuzzy matching and Levenshtein distance for typo correction.
func SuggestEntities(ctx context.Context, db *sql.DB, term string) ([]Suggestion, error) {
	// First, try FTS/LIKE to get direct matches or close prefixes
	entities, err := ResolveEntities(ctx, db, term, 5) // Limit initial suggestions
	if err != nil {
		return nil, err
	}
	if len(entities) > 0 {
		var suggestions []Suggestion
		for _, e := range entities {
			suggestions = append(suggestions, Suggestion{Name: e.Name, Distance: 0, Type: "exact"})
		}
		return suggestions, nil // Return direct matches if any
	}

	// No direct matches, try typo correction (Levenshtein) and fuzzy matching
	allEntityNames, err := getAllEntityNames(ctx, db) // Get all entity names
	if err != nil {
		return nil, err
	}

	// Max Levenshtein distance of 2 for a suggestion (configurable)
	closestName, dist := findClosestEntity(term, allEntityNames, 2)
	if closestName != "" {
		return []Suggestion{{Name: closestName, Distance: dist, Type: "typo"}}, nil
	}

	// No Levenshtein match, try fuzzy matching (e.g., "mod" -> "managecolumnsmodal")
	var suggestions []Suggestion
	for _, candidate := range allEntityNames {
		if utils.FuzzyMatch(term, candidate) {
			suggestions = append(suggestions, Suggestion{Name: candidate, Distance: -1, Type: "fuzzy"})
		}
	}
	// Sort fuzzy suggestions alphabetically
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Name < suggestions[j].Name
	})
	if len(suggestions) > 0 {
		// Limit to top 5 fuzzy suggestions
		if len(suggestions) > 5 {
			suggestions = suggestions[:5]
		}
		return suggestions, nil
	}

	return nil, nil // No suggestions found
}

// getAllEntityNames retrieves all entity names from the database.
func getAllEntityNames(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SELECT name FROM entities ORDER BY name ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to query entity names: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan entity name: %w", err)
		}
		names = append(names, name)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating entity names: %w", err)
	}

	return names, nil
}

// findClosestEntity uses Levenshtein distance to find the closest entity name
func findClosestEntity(query string, candidates []string, maxDistance int) (string, int) {
	if query == "" || len(candidates) == 0 {
		return "", -1
	}

	closestName := ""
	minDistance := maxDistance + 1

	for _, candidate := range candidates {
		dist := utils.ComputeDistance(query, candidate)
		if dist < minDistance {
			minDistance = dist
			closestName = candidate
		}
	}

	if minDistance <= maxDistance {
		return closestName, minDistance
	}
	return "", -1
}