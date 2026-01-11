package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/types"
)

// SearchOptions contains options for the search command
type SearchOptions struct {
	Query     string // Search query string
	Type      string // Filter by type (e.g., "event", "feature", "bug")
	Format    string // Output format: "table", "json", or "graph"
	Limit     int    // Maximum number of results to show
	IndexPath string // Path to index.md file (default: ./index.md)
	Depth     int    // Depth for related entity graph traversal
}

var searchOpts = &SearchOptions{}

// SearchResult represents a single search result with context
type SearchResult struct {
	Row          *IndexRow       `json:"row"`
	MatchScore   float64         `json:"match_score"`
	MatchedText  string          `json:"matched_text"`
	RelatedEntities []string     `json:"related_entities"`
	SessionInfo  *SessionInfo    `json:"session_info,omitempty"`
}

// SearchResultsOutput is the container for search results with graph context
type SearchResultsOutput struct {
	Query       string         `json:"query"`
	TotalMatches int           `json:"total_matches"`
	Results     []*SearchResult `json:"results"`
	RelatedGraph map[string][]string `json:"related_graph,omitempty"` // entity -> related entities
}

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Full-text search across sessions and narratives with graph context",
	Long: `Search across devlog entries with full-text search and entity graph context.

The search command:
  - Performs full-text search across titles and descriptions
  - Shows matching sessions/narratives with relevance scores
  - Displays related entities found in matching entries
  - Can show entity relationship graphs for context

Examples:
  devlog search "authentication"        # Search for authentication-related entries
  devlog search "bug" --type bug        # Search for bugs mentioning "bug"
  devlog search "API" --format json     # Output as JSON
  devlog search "database" --limit 5    # Show top 5 results
  devlog search "session" --depth 2     # Show related entity graph with depth 2`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().StringVarP(&searchOpts.Type, "type", "t", "", "Filter by type (e.g., event, feature, bug)")
	searchCmd.Flags().StringVarP(&searchOpts.Format, "format", "f", "table", "Output format: table, json, or graph")
	searchCmd.Flags().IntVarP(&searchOpts.Limit, "limit", "l", 0, "Maximum number of results to show (0 = unlimited)")
	searchCmd.Flags().StringVarP(&searchOpts.IndexPath, "index", "i", "./index.md", "Path to index.md file")
	searchCmd.Flags().IntVarP(&searchOpts.Depth, "depth", "d", 1, "Depth for related entity graph traversal")
}

// runSearch executes the search command
func runSearch(cmd *cobra.Command, args []string) error {
	// Get query from args or flag
	query := searchOpts.Query
	if len(args) > 0 && args[0] != "" {
		query = args[0]
	}

	if query == "" {
		return fmt.Errorf("search query is required")
	}

	searchOpts.Query = query

	// Try to read from index.md first
	rows, err := parseIndexMD(searchOpts.IndexPath)
	if err != nil {
		// Fall back to searching session events from issues
		return searchInSessions(query)
	}

	// Perform full-text search
	results := performFullTextSearch(rows, query, searchOpts.Type)

	// Apply limit if specified
	if searchOpts.Limit > 0 && len(results) > searchOpts.Limit {
		results = results[:searchOpts.Limit]
	}

	// Build entity graph for related context
	entityGraph := buildEntityGraph(rows)
	relatedGraph := buildRelatedGraph(results, entityGraph, searchOpts.Depth)

	output := &SearchResultsOutput{
		Query:         query,
		TotalMatches:  len(results),
		Results:       results,
		RelatedGraph:  relatedGraph,
	}

	// Output based on format
	switch searchOpts.Format {
	case "json":
		return outputSearchJSON(output)
	case "graph":
		return outputSearchGraph(output, entityGraph)
	case "table":
		return outputSearchTable(output)
	default:
		return fmt.Errorf("invalid format: %s (must be 'table', 'json', or 'graph')", searchOpts.Format)
	}
}

// performFullTextSearch performs case-insensitive full-text search with relevance scoring
func performFullTextSearch(rows []*IndexRow, query string, typeFilter string) []*SearchResult {
	queryLower := strings.ToLower(query)
	queryTerms := strings.Fields(queryLower)

	var results []*SearchResult

	for _, row := range rows {
		// Apply type filter if specified
		if typeFilter != "" {
			if !matchesTypeFilter(row, typeFilter) {
				continue
			}
		}

		// Calculate match score
		score, matchedText := calculateMatchScore(row, queryTerms)

		if score > 0 {
			// Extract related entities from this row
			relatedEntities := row.Entities

			result := &SearchResult{
				Row:            row,
				MatchScore:     score,
				MatchedText:    matchedText,
				RelatedEntities: relatedEntities,
			}
			results = append(results, result)
		}
	}

	// Sort by match score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].MatchScore > results[j].MatchScore
	})

	return results
}

// calculateMatchScore calculates a relevance score for a row based on query terms
func calculateMatchScore(row *IndexRow, queryTerms []string) (float64, string) {
	titleLower := strings.ToLower(row.Title)
	descLower := strings.ToLower(row.Description)

	var score float64
	var matchedTexts []string

	for _, term := range queryTerms {
		// Check title matches (higher weight)
		if strings.Contains(titleLower, term) {
			score += 10.0
			if len(matchedTexts) < 3 {
				matchedTexts = append(matchedTexts, fmt.Sprintf("title: %s", term))
			}
		}

		// Check description matches (medium weight)
		if strings.Contains(descLower, term) {
			score += 5.0
			if len(matchedTexts) < 3 {
				matchedTexts = append(matchedTexts, fmt.Sprintf("desc: %s", term))
			}
		}

		// Check entity matches (medium weight)
		for _, entity := range row.Entities {
			if strings.Contains(strings.ToLower(entity), term) {
				score += 3.0
				if len(matchedTexts) < 3 {
					matchedTexts = append(matchedTexts, fmt.Sprintf("entity: %s", entity))
				}
				break
			}
		}

		// Exact phrase match bonus
		fullQuery := strings.Join(queryTerms, " ")
		if strings.Contains(titleLower, fullQuery) || strings.Contains(descLower, fullQuery) {
			score += 15.0
		}
	}

	matchedText := strings.Join(matchedTexts, ", ")
	return score, matchedText
}

// matchesTypeFilter checks if a row matches the type filter
func matchesTypeFilter(row *IndexRow, typeFilter string) bool {
	typeFilterLower := strings.ToLower(typeFilter)

	// Check title
	if strings.Contains(strings.ToLower(row.Title), typeFilterLower) {
		return true
	}

	// Check description
	if strings.Contains(strings.ToLower(row.Description), typeFilterLower) {
		return true
	}

	// Check entities
	for _, entity := range row.Entities {
		if strings.Contains(strings.ToLower(entity), typeFilterLower) {
			return true
		}
	}

	return false
}

// buildRelatedGraph builds a map of related entities from search results
func buildRelatedGraph(results []*SearchResult, entityGraph map[string]*EntityNode, depth int) map[string][]string {
	related := make(map[string][]string)
	seen := make(map[string]bool)

	for _, result := range results {
		for _, entity := range result.RelatedEntities {
			if seen[entity] {
				continue
			}
			seen[entity] = true

			// Get related entities from graph
			if node, exists := entityGraph[entity]; exists {
				var relatedNames []string
				for _, ref := range node.RelatedTo {
					if depth <= 1 || ref.Strength >= 2 { // Only show strong relationships if depth is 1
						relatedNames = append(relatedNames, ref.Name)
					}
				}
				related[entity] = relatedNames
			}
		}
	}

	return related
}

// searchInSessions searches within session events when index.md is not available
func searchInSessions(query string) error {
	// Find .beads directory
	beadsDir := ".beads"
	issuesFile := filepath.Join(beadsDir, "issues.jsonl")

	// Check if issues.jsonl exists
	if _, err := os.Stat(issuesFile); os.IsNotExist(err) {
		return fmt.Errorf("no index.md found at %s and no issues.jsonl at %s", searchOpts.IndexPath, issuesFile)
	}

	// Read and parse issues.jsonl
	issues, err := readIssuesJSONL(issuesFile)
	if err != nil {
		return fmt.Errorf("failed to read issues: %w", err)
	}

	queryLower := strings.ToLower(query)
	queryTerms := strings.Fields(queryLower)

	var results []*SearchResult

	for _, issue := range issues {
		// Look for event-type issues
		if issue.IssueType != types.TypeEvent {
			continue
		}

		// Apply type filter if specified
		if searchOpts.Type != "" && !matchesIssueTypeFilter(issue, searchOpts.Type) {
			continue
		}

		// Calculate match score
		titleLower := strings.ToLower(issue.Title)
		descLower := strings.ToLower(issue.Description)

		var score float64
		var matchedTexts []string

		for _, term := range queryTerms {
			if strings.Contains(titleLower, term) {
				score += 10.0
				matchedTexts = append(matchedTexts, fmt.Sprintf("title: %s", term))
			}
			if strings.Contains(descLower, term) {
				score += 5.0
				matchedTexts = append(matchedTexts, fmt.Sprintf("desc: %s", term))
			}
		}

		if score > 0 {
			// Convert issue to IndexRow for consistent handling
			row := &IndexRow{
				Date:        issue.CreatedAt.Format("2006-01-02"),
				Title:       issue.Title,
				Description: issue.Description,
				Timestamp:   issue.CreatedAt,
			}

			// Extract entities from the issue
			row.Entities = extractEntities(issue.Title + " " + issue.Description)

			result := &SearchResult{
				Row:            row,
				MatchScore:     score,
				MatchedText:    strings.Join(matchedTexts, ", "),
				RelatedEntities: row.Entities,
				SessionInfo: &SessionInfo{
					ID:        issue.ID,
					Date:      issue.CreatedAt.Format("2006-01-02"),
					Title:     issue.Title,
					Status:    string(issue.Status),
					CreatedAt: issue.CreatedAt,
				},
			}
			results = append(results, result)
		}
	}

	// Sort by match score
	sort.Slice(results, func(i, j int) bool {
		return results[i].MatchScore > results[j].MatchScore
	})

	// Apply limit
	if searchOpts.Limit > 0 && len(results) > searchOpts.Limit {
		results = results[:searchOpts.Limit]
	}

	output := &SearchResultsOutput{
		Query:        query,
		TotalMatches: len(results),
		Results:      results,
	}

	// Output based on format
	switch searchOpts.Format {
	case "json":
		return outputSearchJSON(output)
	case "graph":
		return outputSearchGraph(output, nil)
	case "table":
		return outputSearchTable(output)
	default:
		return fmt.Errorf("invalid format: %s", searchOpts.Format)
	}
}

// matchesIssueTypeFilter checks if an issue matches the type filter
func matchesIssueTypeFilter(issue *types.Issue, typeFilter string) bool {
	typeFilterLower := strings.ToLower(typeFilter)

	if strings.Contains(strings.ToLower(issue.Title), typeFilterLower) {
		return true
	}
	if strings.Contains(strings.ToLower(issue.Description), typeFilterLower) {
		return true
	}

	return false
}

// outputSearchTable displays search results in table format
func outputSearchTable(output *SearchResultsOutput) error {
	if len(output.Results) == 0 {
		fmt.Printf("No results found for query: %s\n", output.Query)
		return nil
	}

	fmt.Printf("🔍 Search Results for: %s\n", output.Query)
	fmt.Printf("Found %d match(es)\n\n", output.TotalMatches)

	for i, result := range output.Results {
		row := result.Row
		fmt.Printf("%d. [%s] %s - %s\n", i+1, row.Date, row.Date, row.Title)
		fmt.Printf("   Score: %.1f | Matches: %s\n", result.MatchScore, result.MatchedText)

		if row.Description != "" {
			// Show excerpt of description
			excerpt := truncateString(row.Description, 100)
			fmt.Printf("   %s\n", excerpt)
		}

		if len(result.RelatedEntities) > 0 {
			fmt.Printf("   Entities: %s\n", strings.Join(result.RelatedEntities, ", "))
		}

		// Show related entities from graph if available
		if output.RelatedGraph != nil {
			for _, entity := range result.RelatedEntities {
				if related, ok := output.RelatedGraph[entity]; ok && len(related) > 0 {
					fmt.Printf("   → %s related to: %s\n", entity, strings.Join(related, ", "))
				}
			}
		}

		fmt.Println()
	}

	return nil
}

// outputSearchJSON displays search results in JSON format
func outputSearchJSON(output *SearchResultsOutput) error {
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// outputSearchGraph displays search results with entity relationship graph
func outputSearchGraph(output *SearchResultsOutput, entityGraph map[string]*EntityNode) error {
	if len(output.Results) == 0 {
		fmt.Printf("No results found for query: %s\n", output.Query)
		return nil
	}

	fmt.Printf("🔍 Search Results for: %s\n", output.Query)
	fmt.Printf("Found %d match(es)\n\n", output.TotalMatches)

	// Show results
	for i, result := range output.Results {
		row := result.Row
		fmt.Printf("%d. [%s] %s\n", i+1, row.Date, row.Title)
		fmt.Printf("   Score: %.1f\n", result.MatchScore)

		if len(result.RelatedEntities) > 0 {
			fmt.Printf("   Entities: %s\n", strings.Join(result.RelatedEntities, ", "))
		}

		fmt.Println()
	}

	// Show related entity graph
	if output.RelatedGraph != nil && len(output.RelatedGraph) > 0 {
		fmt.Println("📊 Related Entity Graph:")
		fmt.Println()

		for entity, related := range output.RelatedGraph {
			if len(related) > 0 {
				fmt.Printf("  %s\n", entity)
				for _, rel := range related {
					fmt.Printf("    └── %s\n", rel)
				}
				fmt.Println()
			}
		}
	}

	return nil
}
