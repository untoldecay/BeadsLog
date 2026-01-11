package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/types"
)

var (
	importMDPath string
)

// importMDCmd implements 'devlog import-md <file>'
// Parses and imports devlog entries from a markdown file
var importMDCmd = &cobra.Command{
	Use:   "import-md <file>",
	Short: "Import devlog entries from a markdown file",
	Long: `Import devlog entries from a markdown file in index.md format.

The expected format is:
  ## YYYY-MM-DD - Title
  Description text here...

This command parses the file, extracts entities, and displays the parsed results.
It can be used to validate the structure of your devlog files.

Examples:
  devlog import-md index.md
  devlog import-md docs/my-log.md`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		// Parse the index file
		rows, err := parseIndexMD(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing file: %v\n", err)
			os.Exit(1)
		}

		if len(rows) == 0 {
			fmt.Println("No entries found in file.")
			return
		}

		// Display parsed results
		fmt.Printf("✅ Successfully parsed %d entries from %s\n\n", len(rows), filePath)

		// Show summary
		fmt.Println("Entries:")
		for i, row := range rows {
			fmt.Printf("  %d. %s - %s\n", i+1, row.Date, row.Title)
			if len(row.Entities) > 0 {
				fmt.Printf("     Entities: %s\n", strings.Join(row.Entities, ", "))
			}
		}

		// Create session
		session := createSession(rows)
		if session != nil {
			fmt.Printf("\nSession created: %s\n", session.ID)
			fmt.Printf("  Time range: %s to %s\n",
				session.StartTime.Format("2006-01-02 15:04"),
				session.EndTime.Format("2006-01-02 15:04"))
		}

		// Entity summary
		allEntities := make(map[string]bool)
		for _, row := range rows {
			for _, entity := range row.Entities {
				allEntities[entity] = true
			}
		}

		if len(allEntities) > 0 {
			fmt.Printf("\n📊 Total unique entities: %d\n", len(allEntities))
		}
	},
}

func init() {
	importMDCmd.Flags().StringVarP(&importMDPath, "output", "o", "", "Output path for parsed data (optional)")
}

// IndexRow represents a single row in the index.md file.
// It contains the date, title, description, and any entities mentioned.
type IndexRow struct {
	Date        string    `json:"date"`        // Date in YYYY-MM-DD format
	Title       string    `json:"title"`       // Title or summary of the entry
	Description string    `json:"description"` // Full description/content
	Entities    []string  `json:"entities"`    // Detected entities (CamelCase, kebab-case, keywords)
	Timestamp   time.Time `json:"timestamp"`   // Parsed timestamp
	LineNumber  int       `json:"line_number"` // Line number in the source file
}

// parseIndexMD parses an index.md file and returns a list of IndexRow structs.
// The expected format is:
//   ## YYYY-MM-DD - Title
//   Description text here...
//
// It also detects entities using regex patterns for CamelCase, kebab-case, and keywords.
func parseIndexMD(filePath string) ([]*IndexRow, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	var rows []*IndexRow
	var currentRow *IndexRow
	lineNumber := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++

		// Check for header pattern: ## YYYY-MM-DD - Title
		if strings.HasPrefix(line, "## ") {
			// Save previous row if exists
			if currentRow != nil {
				rows = append(rows, currentRow)
			}

			// Parse new row
			currentRow = parseHeaderLine(line, lineNumber)
			if currentRow == nil {
				continue
			}
		} else if currentRow != nil && strings.TrimSpace(line) != "" {
			// Append description content
			if currentRow.Description != "" {
				currentRow.Description += "\n"
			}
			currentRow.Description += strings.TrimSpace(line)
		}
	}

	// Don't forget the last row
	if currentRow != nil {
		rows = append(rows, currentRow)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Extract entities from all rows
	for _, row := range rows {
		row.Entities = extractEntities(row.Title + " " + row.Description)
	}

	return rows, nil
}

// parseHeaderLine parses a header line in the format "## YYYY-MM-DD - Title"
func parseHeaderLine(line string, lineNumber int) *IndexRow {
	// Remove the "## " prefix
	header := strings.TrimPrefix(line, "## ")
	header = strings.TrimSpace(header)

	// Pattern: YYYY-MM-DD - Title
	re := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})\s*-\s*(.+)$`)
	matches := re.FindStringSubmatch(header)

	if len(matches) != 3 {
		return nil
	}

	dateStr := matches[1]
	title := matches[2]

	// Parse the date
	timestamp, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil
	}

	return &IndexRow{
		Date:       dateStr,
		Title:      title,
		Timestamp:  timestamp,
		LineNumber: lineNumber,
	}
}

// createSession creates a session record for a group of index rows.
// A session represents a time period of work (e.g., a day or a work session).
func createSession(rows []*IndexRow) *types.Session {
	if len(rows) == 0 {
		return nil
	}

	// Find the earliest and latest timestamps
	var earliest, latest time.Time
	for i, row := range rows {
		if i == 0 || row.Timestamp.Before(earliest) {
			earliest = row.Timestamp
		}
		if i == 0 || row.Timestamp.After(latest) {
			latest = row.Timestamp
		}
	}

	// Generate a session ID based on the date
	sessionID := fmt.Sprintf("session-%s", earliest.Format("2006-01-02"))

	return &types.Session{
		ID:        sessionID,
		StartTime: earliest,
		EndTime:   latest,
		Rows:      rows,
	}
}

// extractEntities extracts entities from text using regex patterns.
// It detects:
//   - CamelCase identifiers (e.g., MyFunction, ClassName)
//   - kebab-case identifiers (e.g., my-function, variable-name)
//   - Keywords (e.g., TODO, FIXME, NOTE, HACK)
func extractEntities(text string) []string {
	var entities []string
	seen := make(map[string]bool)

	// CamelCase pattern: starts with uppercase, contains mixed case
	// Matches: MyFunction, ParseIndexMD, HTTPServer
	camelCaseRe := regexp.MustCompile(`\b[A-Z][a-z0-9]*([A-Z][a-z0-9]*)+\b`)
	for _, match := range camelCaseRe.FindAllString(text, -1) {
		if !seen[match] {
			entities = append(entities, match)
			seen[match] = true
		}
	}

	// kebab-case pattern: lowercase words separated by hyphens
	// Matches: my-function, parse-index-md, user-name
	kebabCaseRe := regexp.MustCompile(`\b[a-z][a-z0-9]*(-[a-z0-9]+)+\b`)
	for _, match := range kebabCaseRe.FindAllString(text, -1) {
		// Filter out common non-entity words
		if !isCommonWord(match) && !seen[match] {
			entities = append(entities, match)
			seen[match] = true
		}
	}

	// Keywords and special markers
	// Matches: TODO, FIXME, NOTE, HACK, XXX, BUG
	keywordRe := regexp.MustCompile(`\b(TODO|FIXME|NOTE|HACK|XXX|BUG|OPTIMIZE|REFACTOR)\b`)
	for _, match := range keywordRe.FindAllString(text, -1) {
		if !seen[match] {
			entities = append(entities, match)
			seen[match] = true
		}
	}

	// Issue IDs in format bd-XXX (e.g., bd-123)
	issueIDRe := regexp.MustCompile(`\b[bB][dD]-[0-9]+\b`)
	for _, match := range issueIDRe.FindAllString(text, -1) {
		if !seen[match] {
			entities = append(entities, match)
			seen[match] = true
		}
	}

	return entities
}

// isCommonWord filters out common kebab-case words that are not entities.
func isCommonWord(word string) bool {
	commonWords := []string{
		"the", "and", "for", "are", "but", "not", "you", "all", "can", "had",
		"her", "was", "one", "our", "out", "has", "his", "how", "its", "may",
		"new", "now", "old", "see", "two", "way", "who", "boy", "did", "get",
		"she", "too", "use", "dad", "mom", "car", "dog", "cat", "run", "eat",
		"non-", "pre-", "post-", "sub-", "super-", "ultra-", "mega-", "micro-",
	}

	lowerWord := strings.ToLower(word)
	for _, common := range commonWords {
		if strings.HasPrefix(lowerWord, common) || lowerWord == common {
			return true
		}
	}

	return false
}

// extractAndLinkEntities processes index rows and links entities to issues.
// It returns a map of entity names to issue IDs that were found/created.
func extractAndLinkEntities(rows []*IndexRow, store Storage, sessionID string) (map[string]string, error) {
	entityLinks := make(map[string]string)

	for _, row := range rows {
		for _, entity := range row.Entities {
			// Skip if already linked
			if _, exists := entityLinks[entity]; exists {
				continue
			}

			// Try to find existing issue with this entity
			// This is a placeholder - actual implementation would query the store
			// For now, we just note that we saw this entity
			entityLinks[entity] = ""
		}
	}

	return entityLinks, nil
}

// Storage is a minimal interface for entity linking operations.
// In a full implementation, this would be the storage.Storage interface.
type Storage interface {
	// GetIssue retrieves an issue by ID
	GetIssue(id string) (*types.Issue, error)
	// SearchIssues searches for issues matching criteria
	SearchIssues(query string, filter types.IssueFilter) ([]*types.Issue, error)
}
