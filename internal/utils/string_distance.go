package utils

import "strings"

// ComputeDistance computes the Levenshtein distance between two strings.
// It is case-insensitive.
func ComputeDistance(s1, s2 string) int {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Compute distances
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			min := matrix[i-1][j] + 1             // deletion
			if ins := matrix[i][j-1] + 1; ins < min { // insertion
				min = ins
			}
			if sub := matrix[i-1][j-1] + cost; sub < min { // substitution
				min = sub
			}
			matrix[i][j] = min
		}
	}

	return matrix[len(s1)][len(s2)]
}
