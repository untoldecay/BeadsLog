package main

import (
	"os"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/types"
)

func TestParseIndexMD(t *testing.T) {
	// Create a temporary test file
	content := `# Test Devlog

## 2024-01-15 - First Entry
This is a test entry for parsing.
It should detect MyFunction and parse-index-md entities.

## 2024-01-16 - Second Entry
Another entry with different entities.
TODO: Add more features here.
`

	// Write to temp file
	tmpFile := "/tmp/test-devlog-index.md"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Parse the file
	rows, err := parseIndexMD(tmpFile)
	if err != nil {
		t.Fatalf("parseIndexMD failed: %v", err)
	}

	// Verify we got 2 rows
	if len(rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(rows))
	}

	// Check first row
	if rows[0].Date != "2024-01-15" {
		t.Errorf("Expected date '2024-01-15', got '%s'", rows[0].Date)
	}
	if rows[0].Title != "First Entry" {
		t.Errorf("Expected title 'First Entry', got '%s'", rows[0].Title)
	}

	// Check entities were detected
	if len(rows[0].Entities) == 0 {
		t.Error("Expected entities to be detected, got none")
	}

	// Verify MyFunction was detected
	foundMyFunction := false
	foundKebabCase := false
	for _, entity := range rows[0].Entities {
		if entity == "MyFunction" {
			foundMyFunction = true
		}
		if entity == "parse-index-md" {
			foundKebabCase = true
		}
	}
	if !foundMyFunction {
		t.Error("Expected to find 'MyFunction' entity")
	}
	if !foundKebabCase {
		t.Error("Expected to find 'parse-index-md' entity")
	}
}

func TestExtractEntities(t *testing.T) {
	text := "MyFunction calls parse-index-md to process data. TODO: Add error handling for HTTPServer."

	entities := extractEntities(text)

	// Check for CamelCase
	foundCamelCase := false
	for _, e := range entities {
		if e == "MyFunction" || e == "HTTPServer" {
			foundCamelCase = true
		}
	}
	if !foundCamelCase {
		t.Error("Expected to detect CamelCase entities")
	}

	// Check for kebab-case
	foundKebabCase := false
	for _, e := range entities {
		if e == "parse-index-md" {
			foundKebabCase = true
		}
	}
	if !foundKebabCase {
		t.Error("Expected to detect kebab-case entities")
	}

	// Check for keywords
	foundKeyword := false
	for _, e := range entities {
		if e == "TODO" {
			foundKeyword = true
		}
	}
	if !foundKeyword {
		t.Error("Expected to detect TODO keyword")
	}
}

func TestCreateSession(t *testing.T) {
	now := time.Now()
	rows := []*IndexRow{
		{
			Date:      "2024-01-15",
			Title:     "Entry 1",
			Timestamp: now.Add(-2 * time.Hour),
		},
		{
			Date:      "2024-01-15",
			Title:     "Entry 2",
			Timestamp: now,
		},
	}

	session := createSession(rows)

	if session == nil {
		t.Fatal("createSession returned nil")
	}

	if session.ID != "session-2024-01-15" {
		t.Errorf("Expected session ID 'session-2024-01-15', got '%s'", session.ID)
	}

	// Verify time bounds
	if !session.StartTime.Equal(rows[0].Timestamp) {
		t.Error("StartTime not set correctly")
	}

	if !session.EndTime.Equal(rows[1].Timestamp) {
		t.Error("EndTime not set correctly")
	}
}

func TestParseHeaderLine(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		expectTitle string
		expectDate  string
		expectNil   bool
	}{
		{
			name:        "Valid header",
			line:        "## 2024-01-15 - My Title",
			expectTitle: "My Title",
			expectDate:  "2024-01-15",
			expectNil:   false,
		},
		{
			name:        "Header with extra spaces",
			line:        "##   2024-01-16   -   Another Title  ",
			expectTitle: "Another Title",
			expectDate:  "2024-01-16",
			expectNil:   false,
		},
		{
			name:      "Invalid format",
			line:      "## Invalid Header",
			expectNil: true,
		},
		{
			name:      "Missing date",
			line:      "## - Title only",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := parseHeaderLine(tt.line, 1)
			if tt.expectNil {
				if row != nil {
					t.Errorf("Expected nil, got %+v", row)
				}
				return
			}

			if row == nil {
				t.Fatal("Expected row, got nil")
			}

			if row.Title != tt.expectTitle {
				t.Errorf("Expected title '%s', got '%s'", tt.expectTitle, row.Title)
			}

			if row.Date != tt.expectDate {
				t.Errorf("Expected date '%s', got '%s'", tt.expectDate, row.Date)
			}
		})
	}
}
