package extractor

import (
	"context"
	"testing"
)

func TestPipeline(t *testing.T) {
	// Test regex-only (empty model)
	pipeline := NewPipeline("")
	text := `
This is a test session.
We fixed a bug in ManageColumnsModal.
The issue was related to useSortable hook.
Also changed nginx.conf settings.

- ManageColumnsModal -> useSortable (uses)
- nginx -> nginx.conf (configures)
`

	result, err := pipeline.Run(context.Background(), text)
	if err != nil {
		t.Fatalf("Pipeline.Run failed: %v", err)
	}

	// Verify Entities
	expectedEntities := map[string]bool{
		"managecolumnsmodal": false,
		"usesortable":        false,
		"nginx":              false,
	}

	for _, e := range result.Entities {
		if _, ok := expectedEntities[e.Name]; ok {
			expectedEntities[e.Name] = true
			if e.Confidence != 0.8 {
				t.Errorf("Expected confidence 0.8 for %s, got %f", e.Name, e.Confidence)
			}
			if e.Source != "regex" {
				t.Errorf("Expected source 'regex' for %s, got %s", e.Name, e.Source)
			}
		}
	}

	for name, found := range expectedEntities {
		if !found {
			t.Errorf("Expected entity %s not found", name)
		}
	}

	// Verify Relationships
	if len(result.Relationships) != 2 {
		t.Errorf("Expected 2 relationships, got %d", len(result.Relationships))
	}
}
