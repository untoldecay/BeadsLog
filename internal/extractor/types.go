package extractor

// Entity represents an extracted entity
type Entity struct {
	Name       string
	Type       string
	Confidence float64
	Source     string // "regex" or "ollama"
}

// Extractor is the interface for entity extraction strategies
type Extractor interface {
	Extract(text string) ([]Entity, error)
	Name() string
}

// Relationship represents a dependency between two entities
type Relationship struct {
	FromEntity string
	ToEntity   string
	Type       string
}
