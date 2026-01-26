package extractor

import (
	"context"
	"fmt"
	"time"
)

type Pipeline struct {
	extractors []Extractor
}

func NewPipeline() *Pipeline {
	return &Pipeline{
		extractors: []Extractor{
			NewRegexExtractor(),
		},
	}
}

// ExtractionResult contains all extracted information and metadata
type ExtractionResult struct {
	Entities      []Entity
	Relationships []Relationship
	Duration      time.Duration
	Extractor     string
}

func (p *Pipeline) Run(ctx context.Context, text string) (*ExtractionResult, error) {
	start := time.Now()
	
	// For now, we only have regex. Later we'll add Ollama and merging logic.
	// We'll iterate through extractors (currently just one)
	
	allEntities := make(map[string]Entity)
	
	for _, ext := range p.extractors {
		entities, err := ext.Extract(text)
		if err != nil {
			// Log error but continue with other extractors if any
			fmt.Printf("Error running extractor %s: %v\n", ext.Name(), err)
			continue
		}
		
		for _, e := range entities {
			if existing, ok := allEntities[e.Name]; ok {
				// Merge logic: keep higher confidence
				if e.Confidence > existing.Confidence {
					allEntities[e.Name] = e
				}
			} else {
				allEntities[e.Name] = e
			}
		}
	}
	
	resultEntities := make([]Entity, 0, len(allEntities))
	for _, e := range allEntities {
		resultEntities = append(resultEntities, e)
	}
	
	// Extract explicit relationships (always run regex for this for now)
	relationships := ExtractRelationships(text)
	
	return &ExtractionResult{
		Entities:      resultEntities,
		Relationships: relationships,
		Duration:      time.Since(start),
		Extractor:     "regex", // Placeholder until we have multi-extractor logic
	}, nil
}
