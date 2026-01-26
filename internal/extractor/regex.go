package extractor

import (
	"regexp"
	"strings"
)

type RegexExtractor struct{}

func NewRegexExtractor() *RegexExtractor {
	return &RegexExtractor{}
}

func (r *RegexExtractor) Name() string {
	return "regex"
}

func (r *RegexExtractor) Extract(text string) ([]Entity, error) {
	entityPatterns := []*regexp.Regexp{
		regexp.MustCompile(`[A-Z][a-z]+(?:[A-Z][a-z]+)+`), // CamelCase (e.g. ManageColumnsModal)
		regexp.MustCompile(`(?i)(modal|hook|endpoint|migration|service)`), // Keywords
		regexp.MustCompile(`[a-z]+-[a-z]+`), // kebab-case (e.g. mcp-sse)
		
		// Expanded patterns from PRD
		regexp.MustCompile(`nginx[\w-]*`),
		regexp.MustCompile(`(?i)modal[\w]*`),
		regexp.MustCompile(`(use|api)[\w]+Service`),
		regexp.MustCompile(`cloudron[\w]*`),
		regexp.MustCompile(`mcp[\w]*`),
		regexp.MustCompile(`proxy_[\w]*`),
		regexp.MustCompile(`\d+\w+Modal`),
		regexp.MustCompile(`use[\w]+`), // Covered useSortable
	}

	seen := make(map[string]bool)
	var entities []Entity

	for _, pat := range entityPatterns {
		matches := pat.FindAllString(text, -1)
		for _, match := range matches {
			if len(match) > 3 {
				lowerMatch := strings.ToLower(match)
				if !seen[lowerMatch] {
					entities = append(entities, Entity{
						Name:       lowerMatch,
						Type:       "component", // Default type for regex matches
						Confidence: 0.8,         // Regex matches have high confidence but less than manual/LLM
						Source:     "regex",
					})
					seen[lowerMatch] = true
				}
			}
		}
	}
	return entities, nil
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
			})
		}
	}
	return rels
}
