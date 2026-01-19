package queries

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/agnivade/levenshtein"
)

// GetAllEntityNames retrieves all entity names from the database.
// This is intended for use with fuzzy matching algorithms like Levenshtein
// where a full list of candidates is required.
// Callers should cache this list if performance is critical for repeated lookups.
func GetAllEntityNames(ctx context.Context, db *sql.DB) ([]string, error) {
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

// FindClosestEntity uses Levenshtein distance to find the closest entity name
// from a list of candidates. Returns the closest name and its distance.
// `maxDistance` specifies the maximum Levenshtein distance to consider a match.
func FindClosestEntity(query string, candidates []string, maxDistance int) (string, int) {
	if query == "" || len(candidates) == 0 {
		return "", -1
	}

	closestName := ""
	minDistance := maxDistance + 1 // Initialize with a value higher than maxDistance

	for _, candidate := range candidates {
		// Convert to lowercase for case-insensitive comparison
		dist := levenshtein.ComputeDistance(strings.ToLower(query), strings.ToLower(candidate))
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
