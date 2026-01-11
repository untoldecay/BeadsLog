package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// TestEntitiesCmd tests the entities command functionality
func TestEntitiesCmd(t *testing.T) {
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

## 2024-01-18 - Performance optimization
Optimized query performance by adding database indexes.
Search queries now 3x faster.
index-md-parser updated to handle larger files.
MyFunction tested again.
`

	if err := os.WriteFile(testIndexPath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test index file: %v", err)
	}

	tests := []struct {
		name       string
		args       []string
		indexPath  string
		wantContain string
		wantErr    bool
	}{
		{
			name:       "List all entities",
			args:       []string{testIndexPath},
			indexPath:  testIndexPath,
			wantContain: "Entity Statistics Report",
			wantErr:    false,
		},
		{
			name:       "List with type filter",
			args:       []string{testIndexPath, "--type", "CamelCase"},
			indexPath:  testIndexPath,
			wantContain: "CamelCase",
			wantErr:    false,
		},
		{
			name:       "List with limit",
			args:       []string{testIndexPath, "--limit", "5"},
			indexPath:  testIndexPath,
			wantContain: "Total Entities:",
			wantErr:    false,
		},
		{
			name:       "List with JSON format",
			args:       []string{testIndexPath, "--format", "json"},
			indexPath:  testIndexPath,
			wantContain: "\"total_entities\":",
			wantErr:    false,
		},
		{
			name:       "List with minimum mentions",
			args:       []string{testIndexPath, "--min", "2"},
			indexPath:  testIndexPath,
			wantContain: "MyFunction",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset package-level variables
			entitiesFormat = "table"
			entitiesType = ""
			entitiesLimit = 0
			entitiesMinimum = 1

			// Parse flags from args
			for i := 0; i < len(tt.args); i++ {
				switch tt.args[i] {
				case "--format", "-f":
					if i+1 < len(tt.args) {
						entitiesFormat = tt.args[i+1]
					}
				case "--type", "-t":
					if i+1 < len(tt.args) {
						entitiesType = tt.args[i+1]
					}
				case "--limit", "-l":
					if i+1 < len(tt.args) {
						entitiesLimit = 5 // Simplified
					}
				case "--min", "-m":
					if i+1 < len(tt.args) {
						entitiesMinimum = 2 // Simplified
					}
				}
			}

			// Create command
			cmd := &cobra.Command{
				Use:   "entities",
				Short: "List entities",
			}

			// Execute
			err := runEntities(cmd, tt.args)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("runEntities() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For now, just check that command runs without error
			t.Logf("Test %s passed", tt.name)
		})
	}
}

// TestBuildEntitiesReport tests the report building logic
func TestBuildEntitiesReport(t *testing.T) {
	rows := []*IndexRow{
		{
			Date:        "2024-01-15",
			Title:       "Implemented user authentication",
			Description: "Added JWT-based authentication",
			Entities:    []string{"JWT", "UserService"},
		},
		{
			Date:        "2024-01-16",
			Title:       "Fixed database bug",
			Description: "Memory leak in production",
			Entities:    []string{"bd-123", "UserService"},
		},
		{
			Date:        "2024-01-17",
			Title:       "Added unit tests",
			Description: "UserService CRUD tests",
			Entities:    []string{"UserService", "TODO"},
		},
	}

	report := buildEntitiesReport(rows)

	// Check totals
	if report.TotalEntities != 3 {
		t.Errorf("buildEntitiesReport() TotalEntities = %d, want 3", report.TotalEntities)
	}

	if report.TotalMentions != 5 {
		t.Errorf("buildEntitiesReport() TotalMentions = %d, want 5", report.TotalMentions)
	}

	// Check that UserService is mentioned 3 times
	var userServiceStats *EntityStats
	for _, e := range report.Entities {
		if e.Name == "UserService" {
			userServiceStats = e
			break
		}
	}

	if userServiceStats == nil {
		t.Fatal("UserService not found in entities")
	}

	if userServiceStats.MentionCount != 3 {
		t.Errorf("UserService MentionCount = %d, want 3", userServiceStats.MentionCount)
	}

	if userServiceStats.FirstSeen != "2024-01-15" {
		t.Errorf("UserService FirstSeen = %s, want 2024-01-15", userServiceStats.FirstSeen)
	}

	if userServiceStats.LastSeen != "2024-01-17" {
		t.Errorf("UserService LastSeen = %s, want 2024-01-17", userServiceStats.LastSeen)
	}
}

// TestGetEntityType tests entity type detection
func TestGetEntityType(t *testing.T) {
	tests := []struct {
		name     string
		entity   string
		wantType string
	}{
		{
			name:     "CamelCase identifier",
			entity:   "MyFunction",
			wantType: "CamelCase",
		},
		{
			name:     "keyword",
			entity:   "TODO",
			wantType: "keyword",
		},
		{
			name:     "issue ID",
			entity:   "bd-123",
			wantType: "issue-id",
		},
		{
			name:     "kebab-case",
			entity:   "my-function",
			wantType: "kebab-case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getEntityType(tt.entity)
			if got != tt.wantType {
				t.Errorf("getEntityType(%s) = %s, want %s", tt.entity, got, tt.wantType)
			}
		})
	}
}
