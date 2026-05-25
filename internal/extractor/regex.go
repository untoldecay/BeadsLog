package extractor

import (
	"regexp"
	"strings"
)

type RegexExtractor struct {
	blacklist map[string]bool
}

func NewRegexExtractor() *RegexExtractor {
	bl := map[string]bool{
		"using":     true,
		"fixed":     true,
		"added":     true,
		"updated":   true,
		"deleted":   true,
		"component": true,
		"service":   true,
		"module":    true,
		"feature":   true,
		"initial":   true,
		"setup":     true,
		"working":   true,
		"problem":   true,
		"session":   true,
		"devlog":    true,
		"beads":     true,
		"beadslog":  true,
		"implemented": true,
		"improved":    true,
		"refactor":    true,
		"refactored":  true,
		"testing":     true,
		"tested":      true,
		"verification": true,
		"verified":     true,
		"error":        true,
		"failed":       true,
		"passed":       true,
		"issue":        true,
		"result":       true,
		"found":        true,
		"missing":      true,
		"logic":        true,
		"state":        true,
		"system":       true,
		"change":       true,
		"changes":      true,
	}
	return &RegexExtractor{
		blacklist: bl,
	}
}

func (r *RegexExtractor) Name() string {
	return "regex"
}

type pattern struct {
	re         *regexp.Regexp
	confidence float64
	minLen     int
}

func (r *RegexExtractor) Extract(text string) ([]Entity, []Relationship, error) {
	patterns := []pattern{
		// STRONG Patterns (0.8 confidence)
		{regexp.MustCompile(`(?i)\b(modal|hook|endpoint|migration|service|proxy_pass|upstream)\b`), 0.8, 4},
		{regexp.MustCompile(`nginx[\w-]*`), 0.8, 5},
		{regexp.MustCompile(`(use|api)[\w]+Service`), 0.8, 5},
		{regexp.MustCompile(`cloudron[\w]*`), 0.8, 5},
		{regexp.MustCompile(`mcp[\w]*`), 0.8, 3}, // mcp is short but strong
		{regexp.MustCompile(`(?i)\b[\w]*modal\b`), 0.8, 5},
		{regexp.MustCompile(`\d+\w+Modal`), 0.8, 5},
		{regexp.MustCompile(`use[A-Z][\w]+`), 0.8, 5}, // hooks like useSortable

		// WEAK Patterns (0.5 confidence)
		{regexp.MustCompile(`[A-Z][a-z]+(?:[A-Z][a-z]+)+`), 0.5, 5}, // Generic CamelCase
		{regexp.MustCompile(`[a-z]+-[a-z]+`), 0.5, 5},               // Generic kebab-case
	}

	seen := make(map[string]bool)
	var entities []Entity

	for _, p := range patterns {
		matches := p.re.FindAllString(text, -1)
		for _, match := range matches {
			if len(match) >= p.minLen {
				lowerMatch := strings.ToLower(match)
				if !r.blacklist[lowerMatch] && !seen[lowerMatch] {
					entities = append(entities, Entity{
						Name:       lowerMatch,
						Type:       "component", // Default type for regex matches
						Confidence: p.confidence,
						Source:     "regex",
					})
					seen[lowerMatch] = true
				}
			}
		}
	}

	// Also extract explicit relationships
	relationships := ExtractRelationships(text)

	return entities, relationships, nil
}

// ExtractRelationships extracts explicit relationships from text
// Pattern: "- EntityA -> EntityB (relationship)"
func ExtractRelationships(text string) []Relationship {
	relPat := regexp.MustCompile(`(?m)^\s*-\s+(.+?)\s+->\s+(.+?)(?:\s+\(([^)]+)\))?$`)
	matches := relPat.FindAllStringSubmatch(text, -1)

	var rels []Relationship
	for _, match := range matches {
		if len(match) >= 3 {
			relType := "depends_on"
			if len(match) > 3 && match[3] != "" {
				relType = strings.TrimSpace(match[3])
			}

			rels = append(rels, Relationship{
				FromEntity: strings.ToLower(strings.TrimSpace(match[1])),
				ToEntity:   strings.ToLower(strings.TrimSpace(match[2])),
				Type:       relType,
				Confidence: 1.0,      // Explicit manual edges are gold standard
				Source:     "manual", // Differentiate from inferred regex edges
			})
		}
	}
	return rels
}
