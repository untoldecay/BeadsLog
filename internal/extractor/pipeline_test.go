package extractor

import (
	"context"
	"strings"
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
		lowerName := strings.ToLower(e.Name)
		if _, ok := expectedEntities[lowerName]; ok {
			expectedEntities[lowerName] = true
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

func TestRelationshipPrioritization(t *testing.T) {
	// We want to verify that manual relationships (confidence 1.0)
	// outrank inferred ones if they share the same key.
	
	pipeline := NewPipeline("")
	text := `
- AuthCore -> BillingAPI (uses)
`
	result, err := pipeline.Run(context.Background(), text)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	found := false
	for _, r := range result.Relationships {
		if strings.ToLower(r.FromEntity) == "authcore" && strings.ToLower(r.ToEntity) == "billingapi" {
			found = true
			if r.Source != "manual" || r.Confidence != 1.0 {
				t.Errorf("Expected manual relationship with 1.0 confidence, got %s and %f", r.Source, r.Confidence)
			}
		}
	}
	if !found {
		t.Error("Relationship AuthCore -> BillingAPI not found")
	}
}
