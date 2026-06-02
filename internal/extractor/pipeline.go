package extractor

import (
	"context"
	"fmt"
	"os"
	"strings"
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
	allRelationships := make(map[string]Relationship)
	usedExtractors := make([]string, 0)
	
	for _, ext := range p.extractors {
		entities, relationships, err := ext.Extract(text)
		if err != nil {
			// Log warning to stderr so user knows why AI failed
			fmt.Fprintf(os.Stderr, "Warning: Extractor %s failed: %v\n", ext.Name(), err)
			continue
		}
		
		usedExtractors = append(usedExtractors, ext.Name())

		// Merge Entities
		for _, e := range entities {
			key := strings.ToLower(e.Name)
			if existing, ok := allEntities[key]; ok {
				// Merge logic: keep higher confidence
				if e.Confidence > existing.Confidence {
					allEntities[key] = e
				}
			} else {
				allEntities[key] = e
			}
		}
		
		// Merge Relationships
		for _, r := range relationships {
			key := strings.ToLower(fmt.Sprintf("%s|%s|%s", r.FromEntity, r.ToEntity, r.Type))
			if existing, ok := allRelationships[key]; ok {
				if r.Confidence > existing.Confidence {
					allRelationships[key] = r
				}
			} else {
				allRelationships[key] = r
			}
		}
	}
	
	resultEntities := make([]Entity, 0, len(allEntities))
	for _, e := range allEntities {
		resultEntities = append(resultEntities, e)
	}

	resultRelationships := make([]Relationship, 0, len(allRelationships))
	for _, r := range allRelationships {
		resultRelationships = append(resultRelationships, r)
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
		Relationships: resultRelationships,
		Duration:      time.Since(start),
		Extractor:     extractorName,
	}, nil
}
