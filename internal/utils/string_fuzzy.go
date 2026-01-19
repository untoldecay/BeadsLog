package utils

import "strings"

// FuzzyMatch checks if source is a fuzzy match of target.
// Characters in source must appear in target in the same order.
// Case-insensitive.
func FuzzyMatch(source, target string) bool {
	source = strings.ToLower(source)
	target = strings.ToLower(target)

	sourceRunes := []rune(source)
	targetRunes := []rune(target)

	sourceIdx := 0
	targetIdx := 0

	for sourceIdx < len(sourceRunes) && targetIdx < len(targetRunes) {
		if sourceRunes[sourceIdx] == targetRunes[targetIdx] {
			sourceIdx++
		}
		targetIdx++
	}

	return sourceIdx == len(sourceRunes)
}
