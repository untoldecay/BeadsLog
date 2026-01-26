package extractor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
)

type OllamaExtractor struct {
	client *api.Client
	model  string
}

func NewOllamaExtractor(model string) (*OllamaExtractor, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to create ollama client: %w", err)
	}
	
	// Check if model is empty, default to llama3.2:3b per PRD if not specified
	if model == "" {
		model = "llama3.2:3b"
	}

	return &OllamaExtractor{
		client: client,
		model:  model,
	}, nil
}

func (o *OllamaExtractor) Name() string {
	return "ollama"
}

// Available checks if Ollama is running and reachable
func (o *OllamaExtractor) Available(ctx context.Context) bool {
	// Simple check by listing tags or version
	// We'll use a short timeout
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// List models as a health check
	_, err := o.client.List(ctx)
	return err == nil
}

type ollamaResponse struct {
	Entities []struct {
		Name json.RawMessage `json:"name"`
		Type string          `json:"type"`
	} `json:"entities"`
}

func (o *OllamaExtractor) Extract(text string) ([]Entity, error) {
	ctx := context.Background()

	// Check availability first to avoid long timeouts if service is down
	if !o.Available(ctx) {
		return nil, fmt.Errorf("ollama service not available")
	}

	prompt := fmt.Sprintf(`
You are an entity extractor for a Go/React/PostgreSQL codebase.

From this devlog session, extract a flat list of entities.
Focus on: 
- Components (Modal, Service, Hook, etc.)
- Config files (nginx.conf, env vars)
- Services (cloudron, mcp, postgres)
- Technologies (pgvector, redis, jwt)

RULES:
1. Output ONLY a valid JSON object.
2. The object MUST have exactly one key: "entities".
3. "entities" MUST be an array of objects.
4. Each entity object MUST have EXACTLY two string fields: "name" and "type".
5. "name" must be a single string (NOT an array).
6. DO NOT include headers, descriptions, or explanations.
7. DO NOT group entities into sub-objects.

Devlog:
%s

Required Output Format:
{
  "entities": [
    {"name": "nginx", "type": "config"},
    {"name": "proxy_buffering", "type": "nginx_setting"}
  ]
}
`, text)

	// Set options for deterministic output if possible
	req := &api.GenerateRequest{
		Model:  o.model,
		Prompt: prompt,
		Format: json.RawMessage(`"json"`), // Force JSON mode
		Stream: new(bool), // false
	}
	*req.Stream = false

	var respText string
	err := o.client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		respText = resp.Response
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("ollama generation failed: %w", err)
	}

	// Parse JSON
	var parsed ollamaResponse
	// Clean up potential markdown code blocks if Ollama adds them despite JSON mode
	cleanedJSON := cleanJSON(respText)
	
	if err := json.Unmarshal([]byte(cleanedJSON), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse ollama json: %w (response: %s)", err, respText)
	}

	var entities []Entity
	for _, e := range parsed.Entities {
		var name string
		// Handle cases where LLM might provide an array of names instead of string
		if err := json.Unmarshal(e.Name, &name); err != nil {
			// Try unmarshaling as array and taking first
			var names []string
			if err2 := json.Unmarshal(e.Name, &names); err2 == nil && len(names) > 0 {
				for _, n := range names {
					entities = append(entities, Entity{
						Name:       strings.ToLower(n),
						Type:       e.Type,
						Confidence: 1.0,
						Source:     "ollama",
					})
				}
				continue
			}
			// Skip entry if we can't parse name at all
			continue
		}

		if name == "" || len(name) < 2 {
			continue
		}

		entities = append(entities, Entity{
			Name:       strings.ToLower(name),
			Type:       e.Type,
			Confidence: 1.0, // Ollama extracted entities have high confidence
			Source:     "ollama",
		})
	}

	return entities, nil
}

func cleanJSON(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
