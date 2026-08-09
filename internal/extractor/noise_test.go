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

func TestIsNoiseSingleCommonWords(t *testing.T) {
	for _, n := range []string{"master", "main", "step", "map", "using"} {
		if !IsNoise(n) {
			t.Errorf("IsNoise(%q) = false, want true (bare common word)", n)
		}
	}
	for _, n := range []string{"ollama", "nginx", "sessions", "beadslog"} {
		if IsNoise(n) {
			t.Errorf("IsNoise(%q) = true, want false", n)
		}
	}
}

func TestIsNoiseGenericProseNouns(t *testing.T) {
	// Bare generic nouns are prose noise (BeadsLog-4qu).
	for _, n := range []string{"config", "component", "service", "services",
		"technologies", "system", "module", "feature", "state"} {
		if !IsNoise(n) {
			t.Errorf("IsNoise(%q) = false, want true (generic prose noun)", n)
		}
	}
	// Specific compound names carrying an uncommon token must still survive.
	for _, n := range []string{"auth-service", "config.yaml", "drawerpanelview",
		"UserService", "payment-gateway"} {
		if IsNoise(n) {
			t.Errorf("IsNoise(%q) = true, want false (specific name must survive)", n)
		}
	}
}

func TestGroundExtraction(t *testing.T) {
	text := "BeadsLog was initialized to track the PaymentGateway rollout."
	entities := []Entity{
		{Name: "beadslog"},
		{Name: "paymentgateway"},
		{Name: "nginx"},        // few-shot echo — not in text
		{Name: "auth-service"}, // few-shot echo — not in text
	}
	rels := []Relationship{
		{FromEntity: "beadslog", ToEntity: "paymentgateway", Type: "tracks"},
		{FromEntity: "nginx", ToEntity: "auth-service", Type: "proxies_to"}, // echo
		{FromEntity: "beadslog", ToEntity: "nginx", Type: "uses"},           // half-grounded
	}

	ge, gr := groundExtraction(text, entities, rels)
	if len(ge) != 2 {
		t.Errorf("expected 2 grounded entities, got %d: %+v", len(ge), ge)
	}
	for _, e := range ge {
		if e.Name == "nginx" || e.Name == "auth-service" {
			t.Errorf("hallucinated entity survived grounding: %q", e.Name)
		}
	}
	if len(gr) != 1 || gr[0].Type != "tracks" {
		t.Errorf("expected only the grounded relationship, got %+v", gr)
	}
}
