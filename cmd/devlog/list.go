package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/types"
)

// ListOptions contains options for the list command
type ListOptions struct {
	Type      string // Filter by type (e.g., "event", "feature", "bug")
	Format    string // Output format: "table" or "json"
	Limit     int    // Maximum number of entries to show
	IndexPath string // Path to index.md file (default: ./index.md)
}

var listOpts = &ListOptions{}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List devlog entries with optional filtering",
	Long: `List devlog entries from index.md or from session events.

Supports filtering by type and various output formats.
The display matches the original index.md structure with dates and titles.`,
	RunE: runList,
}

func init() {
	listCmd.Flags().StringVarP(&listOpts.Type, "type", "t", "", "Filter by type (e.g., event, feature, bug)")
	listCmd.Flags().StringVarP(&listOpts.Format, "format", "f", "table", "Output format: table or json")
	listCmd.Flags().IntVarP(&listOpts.Limit, "limit", "l", 0, "Maximum number of entries to show (0 = unlimited)")
	listCmd.Flags().StringVarP(&listOpts.IndexPath, "index", "i", "./index.md", "Path to index.md file")
}

// runList executes the list command
func runList(cmd *cobra.Command, args []string) error {
	// Try to read from index.md first
	rows, err := parseIndexMD(listOpts.IndexPath)
	if err != nil {
		// Fall back to querying session events from issues
		return listFromSessions()
	}

	// Filter by type if specified
	filtered := filterRowsByType(rows, listOpts.Type)

	// Apply limit if specified
	if listOpts.Limit > 0 && len(filtered) > listOpts.Limit {
		filtered = filtered[:listOpts.Limit]
	}

	// Output based on format
	switch listOpts.Format {
	case "json":
		return outputJSON(filtered)
	case "table":
		return outputTable(filtered)
	default:
		return fmt.Errorf("invalid format: %s (must be 'table' or 'json')", listOpts.Format)
	}
}

// filterRowsByType filters index rows by type
func filterRowsByType(rows []*IndexRow, typeFilter string) []*IndexRow {
	if typeFilter == "" {
		return rows
	}

	var filtered []*IndexRow
	for _, row := range rows {
		// Check if the type is mentioned in entities or description
		if strings.Contains(strings.ToLower(row.Title), strings.ToLower(typeFilter)) ||
			strings.Contains(strings.ToLower(row.Description), strings.ToLower(typeFilter)) {
			filtered = append(filtered, row)
		}
		// Check entities
		for _, entity := range row.Entities {
			if strings.Contains(strings.ToLower(entity), strings.ToLower(typeFilter)) {
				filtered = append(filtered, row)
				break
			}
		}
	}

	return filtered
}

// outputTable displays entries in table format matching index.md structure
func outputTable(rows []*IndexRow) error {
	if len(rows) == 0 {
		fmt.Println("No entries found.")
		return nil
	}

	// Sort by date (newest first)
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Timestamp.After(rows[j].Timestamp)
	})

	// Display in index.md format
	fmt.Println("# Devlog")
	fmt.Println()

	for _, row := range rows {
		fmt.Printf("## %s - %s\n", row.Date, row.Title)
		if row.Description != "" {
			fmt.Printf("%s\n", row.Description)
		}
		if len(row.Entities) > 0 {
			fmt.Printf("\nEntities: %s\n", strings.Join(row.Entities, ", "))
		}
		fmt.Println()
	}

	return nil
}

// outputJSON displays entries in JSON format
func outputJSON(rows []*IndexRow) error {
	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// listFromSessions queries session events from issues.jsonl
// This is used when index.md is not available
func listFromSessions() error {
	// Find .beads directory
	beadsDir := ".beads"
	issuesFile := filepath.Join(beadsDir, "issues.jsonl")

	// Check if issues.jsonl exists
	if _, err := os.Stat(issuesFile); os.IsNotExist(err) {
		return fmt.Errorf("no index.md found at %s and no issues.jsonl at %s", listOpts.IndexPath, issuesFile)
	}

	// Read and parse issues.jsonl to find session events
	issues, err := readIssuesJSONL(issuesFile)
	if err != nil {
		return fmt.Errorf("failed to read issues: %w", err)
	}

	// Filter for session events
	var sessions []*types.Issue
	for _, issue := range issues {
		// Look for event-type issues that represent sessions
		if issue.IssueType == types.TypeEvent &&
			(issue.Status == types.StatusClosed || issue.Status == types.StatusOpen) {
			sessions = append(sessions, issue)
		}
	}

	// Filter by type if specified
	if listOpts.Type != "" {
		var filtered []*types.Issue
		for _, session := range sessions {
			if strings.Contains(strings.ToLower(session.Title), strings.ToLower(listOpts.Type)) ||
				strings.Contains(strings.ToLower(session.Description), strings.ToLower(listOpts.Type)) {
				filtered = append(filtered, session)
			}
		}
		sessions = filtered
	}

	// Apply limit if specified
	if listOpts.Limit > 0 && len(sessions) > listOpts.Limit {
		sessions = sessions[:listOpts.Limit]
	}

	// Sort by created date (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].CreatedAt.After(sessions[j].CreatedAt)
	})

	// Output based on format
	switch listOpts.Format {
	case "json":
		data, err := json.MarshalIndent(sessions, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	case "table":
		return outputSessionsTable(sessions)
	default:
		return fmt.Errorf("invalid format: %s", listOpts.Format)
	}
}

// outputSessionsTable displays sessions in table format
func outputSessionsTable(sessions []*types.Issue) error {
	if len(sessions) == 0 {
		fmt.Println("No session entries found.")
		return nil
	}

	// Display in a format similar to index.md
	fmt.Println("# Devlog Sessions")
	fmt.Println()

	for _, session := range sessions {
		date := session.CreatedAt.Format("2006-01-02")
		fmt.Printf("## %s - %s\n", date, session.Title)

		if session.Description != "" {
			// Format description with proper line breaks
			lines := strings.Split(session.Description, "\n")
			for _, line := range lines {
				if line != "" {
					fmt.Printf("%s\n", line)
				}
			}
		}

		// Show metadata
		fmt.Printf("\nID: %s | Status: %s | Created: %s\n",
			session.ID,
			session.Status,
			session.CreatedAt.Format("2006-01-02 15:04:05"))

		if session.CreatedBy != "" {
			fmt.Printf("Created by: %s\n", session.CreatedBy)
		}

		if session.CloseReason != "" {
			fmt.Printf("Reason: %s\n", session.CloseReason)
		}

		fmt.Println()
	}

	return nil
}

// readIssuesJSONL reads issues from a JSONL file
func readIssuesJSONL(path string) ([]*types.Issue, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var issues []*types.Issue
	decoder := json.NewDecoder(file)

	for decoder.More() {
		var issue types.Issue
		if err := decoder.Decode(&issue); err != nil {
			// Skip malformed lines
			continue
		}
		issues = append(issues, &issue)
	}

	return issues, nil
}

// SessionInfo represents a simplified session view for listing
type SessionInfo struct {
	ID        string    `json:"id"`
	Date      string    `json:"date"`
	Title     string    `json:"title"`
	Type      string    `json:"type,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
