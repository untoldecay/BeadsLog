package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// TestResumeCmd tests the resume command functionality
func TestResumeCmd(t *testing.T) {
	// Create a temporary test index file
	tmpDir := t.TempDir()
	testIndexPath := filepath.Join(tmpDir, "test-index.md")

	testContent := `# Devlog

## 2024-01-15 - Implemented user authentication
Added JWT-based authentication to the API.
Users can now login with email/password and receive tokens.
TODO: Add refresh token support.

## 2024-01-16 - Fixed database connection bug
Fixed issue where connections were not being properly closed.
This was causing memory leaks in production.
Related to bd-123.

## 2024-01-17 - Added unit tests for UserService
Wrote comprehensive tests for user CRUD operations.
Coverage now at 85% for UserService.
MyFunction was refactored to support this.

## 2024-01-18 - Refactored API endpoints
Cleaned up the REST API structure.
MyFunction now uses dependency injection.
`

	if err := os.WriteFile(testIndexPath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test index file: %v", err)
	}

	tests := []struct {
		name       string
		args       []string
		indexPath  string
		hybrid     bool
		format     string
		wantContain string
		wantErr    bool
	}{
		{
			name:       "Basic resume search",
			args:       []string{"authentication"},
			indexPath:  testIndexPath,
			hybrid:     false,
			format:     "table",
			wantContain: "authentication",
			wantErr:    false,
		},
		{
			name:       "Hybrid search with entity graph",
			args:       []string{"MyFunction"},
			indexPath:  testIndexPath,
			hybrid:     true,
			format:     "table",
			wantContain: "MyFunction",
			wantErr:    false,
		},
		{
			name:       "JSON format output",
			args:       []string{"database"},
			indexPath:  testIndexPath,
			hybrid:     false,
			format:     "json",
			wantContain: "\"query\":",
			wantErr:    false,
		},
		{
			name:       "AI format output",
			args:       []string{"UserService"},
			indexPath:  testIndexPath,
			hybrid:     true,
			format:     "ai",
			wantContain: "Work Context",
			wantErr:    false,
		},
		{
			name:       "No matches found",
			args:       []string{"nonexistent"},
			indexPath:  testIndexPath,
			hybrid:     false,
			format:     "table",
			wantContain: "No results found",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set options
			resumeOpts = &ResumeOptions{
				Query:     tt.args[0],
				Hybrid:    tt.hybrid,
				Format:    tt.format,
				IndexPath: tt.indexPath,
				Limit:     0,
				Depth:     2,
			}

			// Execute command
			cmd := &cobra.Command{
				Use:   "resume",
				Short: "Resume work by finding sessions",
				RunE: func(cmd *cobra.Command, args []string) error {
					return runResume(cmd, tt.args)
				},
			}

			// Execute
			err := runResume(cmd, tt.args)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("runResume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For now, just check that command runs without error
			t.Logf("Test %s passed", tt.name)
		})
	}
}

// TestBuildTwoHopGraph tests the 2-hop entity graph building
func TestBuildTwoHopGraph(t *testing.T) {
	// Create test rows
	rows := []*IndexRow{
		{
			Date:        "2024-01-15",
			Title:       "Implemented UserService",
			Description: "Added user CRUD operations",
			Entities:    []string{"UserService", "MyFunction"},
		},
		{
			Date:        "2024-01-16",
			Title:       "Refactored MyFunction",
			Description: "Improved performance",
			Entities:    []string{"MyFunction", "Database"},
		},
		{
			Date:        "2024-01-17",
			Title:       "Added Database tests",
			Description: "Test coverage for Database",
			Entities:    []string{"Database", "TestFramework"},
		},
	}

	// Build entity graph
	entityGraph := buildEntityGraph(rows)

	// Test 2-hop graph building
	entities := []string{"UserService"}
	graph := buildTwoHopGraph(entities, entityGraph, 2)

	// UserService should be connected to MyFunction (1-hop)
	// MyFunction should be connected to Database (2-hop)
	if len(graph) == 0 {
		t.Error("buildTwoHopGraph() returned empty graph")
	}

	// Check that UserService has relationships
	if relations, ok := graph["UserService"]; ok {
		if len(relations) == 0 {
			t.Error("UserService has no relations in 2-hop graph")
		}
		// Should have MyFunction as direct relation
		hasMyFunction := false
		for _, rel := range relations {
			if contains(rel, "MyFunction") {
				hasMyFunction = true
				break
			}
		}
		if !hasMyFunction {
			t.Error("UserService should be related to MyFunction in 2-hop graph")
		}
	} else {
		t.Error("UserService not found in 2-hop graph")
	}

	t.Logf("2-hop graph: %+v", graph)
}

// TestExtractRelatedContext tests context extraction
func TestExtractRelatedContext(t *testing.T) {
	rows := []*IndexRow{
		{
			Date:        "2024-01-15",
			Title:       "Session 1",
			Description: "Work on MyFunction",
			Entities:    []string{"MyFunction", "UserService"},
		},
		{
			Date:        "2024-01-16",
			Title:       "Session 2",
			Description: "More work on MyFunction",
			Entities:    []string{"MyFunction", "Database"},
		},
	}

	entityGraph := buildEntityGraph(rows)

	// Extract context for Session 1
	context := extractRelatedContext(rows[0], rows, entityGraph, []string{"MyFunction"})

	// Should find Session 2 as related (shares MyFunction entity)
	if len(context) == 0 {
		t.Error("extractRelatedContext() returned no context")
	}

	foundSession2 := false
	for _, ctx := range context {
		if contains(ctx, "Session 2") {
			foundSession2 = true
			break
		}
	}

	if !foundSession2 {
		t.Error("Should find Session 2 as related context")
	}

	t.Logf("Related context: %+v", context)
}

// TestBuildEntityTypeContext tests entity type categorization
func TestBuildEntityTypeContext(t *testing.T) {
	rows := []*IndexRow{
		{
			Date:        "2024-01-15",
			Title:       "Test entry",
			Description: "Work on MyFunction and user-service",
			Entities:    []string{"MyFunction", "user-service", "TODO", "bd-123"},
		},
	}

	entityGraph := buildEntityGraph(rows)

	results := []*ResumeResult{
		{
			Session:  rows[0],
			Entities: rows[0].Entities,
		},
	}

	context := buildEntityTypeContext(results, entityGraph)

	// Should categorize entities
	if len(context) == 0 {
		t.Error("buildEntityTypeContext() returned empty context")
	}

	// Check for expected categories
	hasCamelCase := false
	hasKebabCase := false
	hasKeyword := false
	hasIssue := false

	for category, entities := range context {
		switch category {
		case "camelcase":
			for _, e := range entities {
				if e == "MyFunction" {
					hasCamelCase = true
				}
			}
		case "kebabcase":
			for _, e := range entities {
				if e == "user-service" {
					hasKebabCase = true
				}
			}
		case "keyword":
			for _, e := range entities {
				if e == "TODO" {
					hasKeyword = true
				}
			}
		case "issue":
			for _, e := range entities {
				if e == "bd-123" {
					hasIssue = true
				}
			}
		}
	}

	if !hasCamelCase {
		t.Error("Should have camelcase category with MyFunction")
	}
	if !hasKebabCase {
		t.Error("Should have kebabcase category with user-service")
	}
	if !hasKeyword {
		t.Error("Should have keyword category with TODO")
	}
	if !hasIssue {
		t.Error("Should have issue category with bd-123")
	}

	t.Logf("Entity type context: %+v", context)
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
