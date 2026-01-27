package extractor

import (
	"context"
	"fmt"
	"time"
)

type Pipeline struct {
	extractors []Extractor
}

func NewPipeline(ollamaModel string) *Pipeline {
	extractors := []Extractor{
		NewRegexExtractor(),
	}

	// Add Ollama if model provided (and let it fail gracefully if service is down)
	if ollamaModel != "" {
		if ollama, err := NewOllamaExtractor(ollamaModel); err == nil {
			extractors = append(extractors, ollama)
		} else {
			fmt.Printf("Warning: Failed to initialize Ollama extractor: %v\n", err)
		}
	}

	return &Pipeline{
		extractors: extractors,
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
	
	allEntities := make(map[string]Entity)
	allRelationships := make([]Relationship, 0)
	usedExtractors := make([]string, 0)
	
	for _, ext := range p.extractors {
		entities, relationships, err := ext.Extract(text)
		if err != nil {
			// Log error but continue with other extractors
			// Only verbose log to avoid noise if Ollama is just offline
			continue
		}
		
		usedExtractors = append(usedExtractors, ext.Name())

		// Merge Entities
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
		
		// Merge Relationships (simple append for now, duplicates handled by DB UNIQUE constraint or ignored)
		allRelationships = append(allRelationships, relationships...)
	}
	
	resultEntities := make([]Entity, 0, len(allEntities))
	for _, e := range allEntities {
		resultEntities = append(resultEntities, e)
	}
	
	// Note: We don't call ExtractRelationships separately anymore, it's part of RegexExtractor
	// which is always in the pipeline.
	
	extractorName := "regex"
	if len(usedExtractors) > 0 {
		// e.g. "regex,ollama" or just "regex"
		// If ollama was used, prioritize showing it
		for _, name := range usedExtractors {
			if name == "ollama" {
				extractorName = "ollama+regex"
				break
			}
		}
	}
	
	return &ExtractionResult{
		Entities:      resultEntities,
		Relationships: allRelationships,
		Duration:      time.Since(start),
		Extractor:     extractorName,
	}, nil
}
