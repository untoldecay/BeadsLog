package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// TestImpactCmd tests the impact command functionality
func TestImpactCmd(t *testing.T) {
	// Create a temporary test index file
	tmpDir := t.TempDir()
	testIndexPath := filepath.Join(tmpDir, "test-index.md")

	testContent := `# Devlog

## 2024-01-15 - Implemented UserService
Added JWT-based authentication to the API.
Users can now login with email/password and receive tokens.
TODO: Add refresh token support.

## 2024-01-16 - Fixed UserService database bug
Fixed issue where connections were not being properly closed.
This was causing memory leaks in production.
Related to bd-123.

## 2024-01-17 - Added unit tests for UserService
Wrote comprehensive tests for user CRUD operations.
Coverage now at 85% for UserService.
MyFunction was refactored to support this.

## 2024-01-18 - Performance optimization
Optimized query performance by adding database indexes.
Search queries now 3x faster.
index-md-parser updated to handle larger files.
MyFunction tested again.
`

	if err := os.WriteFile(testIndexPath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test index file: %v", err)
	}

	// Test with an entity that has dependencies
	t.Run("Impact analysis for UserService", func(t *testing.T) {
		// Reset package-level variables
		impactDepth = 1

		// Parse index to get entities
		rows, err := parseIndexMD(testIndexPath)
		if err != nil {
			t.Fatalf("Failed to parse test index: %v", err)
		}

		// Build entity graph
		entityGraph := buildEntityGraph(rows)

		// Check that UserService exists
		targetNode, exists := entityGraph["UserService"]
		if !exists {
			t.Fatal("UserService not found in entity graph")
		}

		// Build dependency graph
		dependencyGraph := buildDependencyGraph(entityGraph)

		// Get dependencies
		dependencies := getDependencies("UserService", dependencyGraph, entityGraph)

		// UserService should have dependencies (entities mentioned with it)
		if len(dependencies) == 0 {
			t.Error("Expected UserService to have dependencies, but found none")
		}

		// Check that expected entities are in dependencies
		depNames := make(map[string]bool)
		for _, dep := range dependencies {
			depNames[dep.Name] = true
		}

		// TODO, bd-123, and MyFunction should be dependencies
		expectedDeps := []string{"TODO", "bd-123", "MyFunction"}
		foundAny := false
		for _, exp := range expectedDeps {
			if depNames[exp] {
				foundAny = true
				t.Logf("Found expected dependency: %s", exp)
			}
		}

		if !foundAny {
			t.Errorf("Expected to find at least one of %v in dependencies, but got: %v", expectedDeps, depNames)
		}

		t.Logf("UserService has %d dependencies: %v", len(dependencies), depNames)
		t.Logf("UserService appears in %d rows", len(targetNode.Rows))
	})

	// Test with an entity that has no direct dependencies
	t.Run("Impact analysis for entity with no dependencies", func(t *testing.T) {
		impactDepth = 1

		rows, err := parseIndexMD(testIndexPath)
		if err != nil {
			t.Fatalf("Failed to parse test index: %v", err)
		}

		entityGraph := buildEntityGraph(rows)
		dependencyGraph := buildDependencyGraph(entityGraph)

		// Find an entity that appears alone (if any)
		for entityName := range entityGraph {
			dependencies := getDependencies(entityName, dependencyGraph, entityGraph)
			// Just verify the function doesn't crash
			t.Logf("Entity %s has %d dependencies", entityName, len(dependencies))
			break
		}
	})
}

// TestBuildDependencyGraph tests the dependency graph building logic
func TestBuildDependencyGraph(t *testing.T) {
	rows := []*IndexRow{
		{
			Date:        "2024-01-15",
			Title:       "Implemented UserService",
			Description: "Added JWT-based authentication",
			Entities:    []string{"UserService", "JWT"},
		},
		{
			Date:        "2024-01-16",
			Title:       "Fixed UserService bug",
			Description: "Memory leak in production",
			Entities:    []string{"UserService", "bd-123"},
		},
		{
			Date:        "2024-01-17",
			Title:       "Added unit tests",
			Description: "UserService CRUD tests",
			Entities:    []string{"UserService", "TODO"},
		},
	}

	// Build entity graph
	entityGraph := buildEntityGraph(rows)

	// Build dependency graph
	dependencyGraph := buildDependencyGraph(entityGraph)

	// UserService should have entities that depend on it
	// JWT, bd-123, and TODO all "depend on" UserService (are mentioned with it)
	userServiceDeps, exists := dependencyGraph["UserService"]
	if !exists {
		t.Fatal("UserService not found in dependency graph")
	}

	if len(userServiceDeps) == 0 {
		t.Error("Expected UserService to have dependent entities")
	}

	// Check that JWT, bd-123, and TODO are in the dependencies
	depNames := make(map[string]bool)
	for _, dep := range userServiceDeps {
		depNames[dep.Name] = true
	}

	expectedDeps := []string{"JWT", "bd-123", "TODO"}
	for _, exp := range expectedDeps {
		if !depNames[exp] {
			t.Errorf("Expected %s to be a dependency of UserService, but it was not found", exp)
		}
	}

	t.Logf("UserService has %d dependencies: %v", len(userServiceDeps), depNames)
}

// TestGetDependencies tests the getDependencies function
func TestGetDependencies(t *testing.T) {
	rows := []*IndexRow{
		{
			Date:        "2024-01-15",
			Title:       "Implemented MyFunction",
			Description: "Added utility function",
			Entities:    []string{"MyFunction", "Utils"},
		},
		{
			Date:        "2024-01-16",
			Title:       "Used MyFunction in UserService",
			Description: "Refactored to use MyFunction",
			Entities:    []string{"MyFunction", "UserService"},
		},
	}

	entityGraph := buildEntityGraph(rows)
	dependencyGraph := buildDependencyGraph(entityGraph)

	// Test getting dependencies for MyFunction
	deps := getDependencies("MyFunction", dependencyGraph, entityGraph)

	// Utils and UserService should both depend on MyFunction
	if len(deps) == 0 {
		t.Error("Expected MyFunction to have dependencies")
	}

	// Verify dependencies are sorted by strength
	for i := 0; i < len(deps)-1; i++ {
		if deps[i].Strength < deps[i+1].Strength {
			t.Errorf("Dependencies not sorted by strength: %v", deps)
		}
	}

	t.Logf("MyFunction has %d dependencies", len(deps))
	for _, dep := range deps {
		t.Logf("  - %s (strength: %d)", dep.Name, dep.Strength)
	}
}
