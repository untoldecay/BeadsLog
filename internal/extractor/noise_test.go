package extractor

import (
	"context"
	"strings"
	"testing"
)

func TestIsNoise(t *testing.T) {
	noise := []string{"map it step", "42", "-foo", "bar-", ".hidden", "it", ""}
	for _, n := range noise {
		if !IsNoise(n) {
			t.Errorf("IsNoise(%q) = false, want true", n)
		}
	}
	legit := []string{"ollama-extractor", "huh.select", "auth service", "AlphaService", "nginx-proxy", "multi-layered"}
	for _, n := range legit {
		if IsNoise(n) {
			t.Errorf("IsNoise(%q) = true, want false", n)
		}
	}
}

func TestKebabPatternNoMidWordTruncation(t *testing.T) {
	r := NewRegexExtractor()
	entities, _, err := r.Extract("Multi-layered bootstrap flows require care.")
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	for _, e := range entities {
		if strings.EqualFold(e.Name, "ulti-layered") {
			t.Errorf("mid-word truncation artifact extracted: %q", e.Name)
		}
	}
	found := false
	for _, e := range entities {
		if strings.EqualFold(e.Name, "Multi-layered") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected whole word 'Multi-layered' to be extracted, got: %v", entities)
	}
}

func TestPipelineFiltersNoise(t *testing.T) {
	p := NewPipeline("") // regex only
	text := `Working on AlphaService integration.

### Architectural Relationships
- map it step -> AlphaService (uses)
- AlphaComponent -> AlphaService (calls)
`
	result, err := p.Run(context.Background(), text)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	for _, r := range result.Relationships {
		if strings.EqualFold(r.FromEntity, "map it step") {
			t.Errorf("noise endpoint survived pipeline filter: %+v", r)
		}
	}
	kept := false
	for _, r := range result.Relationships {
		if r.FromEntity == "AlphaComponent" && r.ToEntity == "AlphaService" {
			kept = true
		}
	}
	if !kept {
		t.Errorf("legit relationship was dropped: %+v", result.Relationships)
	}
}
