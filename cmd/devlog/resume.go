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

// ResumeOptions contains options for the resume command
type ResumeOptions struct {
	Query     string // Search query string
	Hybrid    bool   // Use hybrid matching (sessions + entity graph)
	Format    string // Output format: "table", "json", or "ai"
	Limit     int    // Maximum number of results to show
	IndexPath string // Path to index.md file (default: ./index.md)
	Depth     int    // Depth for entity graph traversal (default: 2 for 2-hop)
}

var resumeOpts = &ResumeOptions{}

// ResumeResult represents a single resume result with context
type ResumeResult struct {
	Session       *IndexRow       `json:"session"`
	MatchScore    float64         `json:"match_score"`
	MatchedText   string          `json:"matched_text"`
	Entities      []string        `json:"entities"`
	RelatedGraph  map[string][]string `json:"related_graph,omitempty"` // 2-hop entity relationships
	Context       []string        `json:"context,omitempty"`           // Related entries for context
}

// ResumeOutput is the container for resume results
type ResumeOutput struct {
	Query       string           `json:"query"`
	TotalMatches int             `json:"total_matches"`
	Results     []*ResumeResult  `json:"results"`
	EntityTypeContext map[string][]string `json:"entity_type_context,omitempty"` // Entity type categorization
}

// resumeCmd represents the resume command
var resumeCmd = &cobra.Command{
	Use:   "resume [query]",
	Short: "Resume work by finding matching sessions with entity graph context",
	Long: `Resume work by searching sessions with hybrid matching combining:
  - Full-text search across session titles and descriptions
  - 2-hop entity graph traversal for rich context
  - Related entities and their connections

This command is designed for AI agent consumption, providing structured
context about previous work sessions.

The --hybrid flag enables advanced matching:
  1. Matches sessions by query terms
  2. Extracts entities from matching sessions
  3. Traverses 2 hops in entity graph for related context
  4. Returns structured output with entity relationships

Examples:
  devlog resume "authentication"                    # Find authentication sessions
  devlog resume "API" --hybrid                      # Hybrid search with entity graph
  devlog resume "database" --format json            # Output as JSON for AI
  devlog resume "session" --limit 3 --depth 2       # Top 3 results with 2-hop context`,
	Args: cobra.MaximumNArgs(1),
	RunE: runResume,
}

func init() {
	resumeCmd.Flags().StringVarP(&resumeOpts.Query, "query", "q", "", "Search query string")
	resumeCmd.Flags().BoolVarP(&resumeOpts.Hybrid, "hybrid", "H", false, "Use hybrid matching (sessions + entity graph)")
	resumeCmd.Flags().StringVarP(&resumeOpts.Format, "format", "f", "table", "Output format: table, json, or ai")
	resumeCmd.Flags().IntVarP(&resumeOpts.Limit, "limit", "l", 0, "Maximum number of results to show (0 = unlimited)")
	resumeCmd.Flags().StringVarP(&resumeOpts.IndexPath, "index", "i", "./index.md", "Path to index.md file")
	resumeCmd.Flags().IntVarP(&resumeOpts.Depth, "depth", "d", 2, "Depth for entity graph traversal (2 = 2-hop)")
}

// runResume executes the resume command
func runResume(cmd *cobra.Command, args []string) error {
	// Get query from args or flag
	query := resumeOpts.Query
	if len(args) > 0 && args[0] != "" {
		query = args[0]
	}

	if query == "" {
		return fmt.Errorf("search query is required")
	}

	resumeOpts.Query = query

	// Try to read from index.md first
	rows, err := parseIndexMD(resumeOpts.IndexPath)
	if err != nil {
		// Fall back to searching session events from issues
		return resumeFromSessions(query)
	}

	// Build entity graph for context
	entityGraph := buildEntityGraph(rows)

	// Perform hybrid search if requested
	var results []*ResumeResult
	if resumeOpts.Hybrid {
		results = performHybridSearch(rows, entityGraph, query, resumeOpts.Depth)
	} else {
		results = performBasicResume(rows, query)
	}

	// Apply limit if specified
	if resumeOpts.Limit > 0 && len(results) > resumeOpts.Limit {
		results = results[:resumeOpts.Limit]
	}

	// Build entity type context
	entityTypeContext := buildEntityTypeContext(results, entityGraph)

	output := &ResumeOutput{
		Query:            query,
		TotalMatches:     len(results),
		Results:          results,
		EntityTypeContext: entityTypeContext,
	}

	// Output based on format
	switch resumeOpts.Format {
	case "json":
		return outputResumeJSON(output)
	case "ai":
		return outputResumeAI(output)
	case "table":
		return outputResumeTable(output)
	default:
		return fmt.Errorf("invalid format: %s (must be 'table', 'json', or 'ai')", resumeOpts.Format)
	}
}

// performHybridSearch combines session matching with 2-hop entity graph context
func performHybridSearch(rows []*IndexRow, entityGraph map[string]*EntityNode, query string, depth int) []*ResumeResult {
	queryLower := strings.ToLower(query)
	queryTerms := strings.Fields(queryLower)

	var results []*ResumeResult

	// First pass: find matching sessions
	for _, row := range rows {
		score, matchedText := calculateMatchScore(row, queryTerms)

		if score > 0 {
			result := &ResumeResult{
				Session:     row,
				MatchScore:  score,
				MatchedText: matchedText,
				Entities:    row.Entities,
			}
			results = append(results, result)
		}
	}

	// Second pass: enhance with 2-hop entity graph context
	for _, result := range results {
		result.RelatedGraph = buildTwoHopGraph(result.Entities, entityGraph, depth)
		result.Context = extractRelatedContext(result.Session, rows, entityGraph, result.Entities)
	}

	// Sort by match score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].MatchScore > results[j].MatchScore
	})

	return results
}

// performBasicResume performs simple session matching without entity graph
func performBasicResume(rows []*IndexRow, query string) []*ResumeResult {
	queryLower := strings.ToLower(query)
	queryTerms := strings.Fields(queryLower)

	var results []*ResumeResult

	for _, row := range rows {
		score, matchedText := calculateMatchScore(row, queryTerms)

		if score > 0 {
			result := &ResumeResult{
				Session:     row,
				MatchScore:  score,
				MatchedText: matchedText,
				Entities:    row.Entities,
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

// buildTwoHopGraph builds a 2-hop entity relationship graph
func buildTwoHopGraph(entities []string, entityGraph map[string]*EntityNode, maxDepth int) map[string][]string {
	graph := make(map[string][]string)
	seen := make(map[string]bool)

	for _, entity := range entities {
		if seen[entity] {
			continue
		}
		seen[entity] = true

		// Get direct relationships (1-hop)
		if node, exists := entityGraph[entity]; exists {
			var related []string

			// Add 1-hop relationships
			for _, ref := range node.RelatedTo {
				if !seen[ref.Name] {
					related = append(related, ref.Name)
					seen[ref.Name] = true

					// Add 2-hop relationships if depth allows
					if maxDepth >= 2 {
						if relatedNode, exists := entityGraph[ref.Name]; exists {
							for _, hop2Ref := range relatedNode.RelatedTo {
								if !seen[hop2Ref.Name] && hop2Ref.Name != entity {
									related = append(related, fmt.Sprintf("%s→%s", ref.Name, hop2Ref.Name))
									seen[hop2Ref.Name] = true
								}
							}
						}
					}
				}
			}

			graph[entity] = related
		}
	}

	return graph
}

// extractRelatedContext extracts related entries for additional context
func extractRelatedContext(session *IndexRow, rows []*IndexRow, entityGraph map[string]*EntityNode, entities []string) []string {
	context := []string{}
	seen := make(map[string]bool)

	// Find other sessions that share entities with this session
	for _, entity := range entities {
		if node, exists := entityGraph[entity]; exists {
			for _, row := range node.Rows {
				// Skip the current session
				if row.Date == session.Date && row.Title == session.Title {
					continue
				}

				// Create context key
				key := fmt.Sprintf("%s:%s", row.Date, row.Title)
				if !seen[key] {
					context = append(context, fmt.Sprintf("[%s] %s", row.Date, row.Title))
					seen[key] = true
				}
			}
		}
	}

	return context
}

// buildEntityTypeContext categorizes entities by type for AI consumption
func buildEntityTypeContext(results []*ResumeResult, entityGraph map[string]*EntityNode) map[string][]string {
	context := make(map[string][]string)
	seen := make(map[string]bool)

	for _, result := range results {
		for _, entity := range result.Entities {
			if seen[entity] {
				continue
			}
			seen[entity] = true

			// Categorize entity
			var category string
			if isCamelCase(entity) {
				category = "camelcase"
			} else if isKebabCase(entity) {
				category = "kebabcase"
			} else if isKeyword(entity) {
				category = "keyword"
			} else if isIssueID(entity) {
				category = "issue"
			} else {
				category = "other"
			}

			context[category] = append(context[category], entity)
		}
	}

	return context
}

// resumeFromSessions searches within session events when index.md is not available
func resumeFromSessions(query string) error {
	beadsDir := ".beads"
	issuesFile := filepath.Join(beadsDir, "issues.jsonl")

	if _, err := os.Stat(issuesFile); os.IsNotExist(err) {
		return fmt.Errorf("no index.md found at %s and no issues.jsonl at %s", resumeOpts.IndexPath, issuesFile)
	}

	issues, err := readIssuesJSONL(issuesFile)
	if err != nil {
		return fmt.Errorf("failed to read issues: %w", err)
	}

	queryLower := strings.ToLower(query)
	queryTerms := strings.Fields(queryLower)

	var results []*ResumeResult

	for _, issue := range issues {
		if issue.IssueType != types.TypeEvent {
			continue
		}

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
			row := &IndexRow{
				Date:        issue.CreatedAt.Format("2006-01-02"),
				Title:       issue.Title,
				Description: issue.Description,
				Timestamp:   issue.CreatedAt,
			}
			row.Entities = extractEntities(issue.Title + " " + issue.Description)

			result := &ResumeResult{
				Session:     row,
				MatchScore:  score,
				MatchedText: strings.Join(matchedTexts, ", "),
				Entities:    row.Entities,
			}
			results = append(results, result)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].MatchScore > results[j].MatchScore
	})

	if resumeOpts.Limit > 0 && len(results) > resumeOpts.Limit {
		results = results[:resumeOpts.Limit]
	}

	output := &ResumeOutput{
		Query:        query,
		TotalMatches: len(results),
		Results:      results,
	}

	switch resumeOpts.Format {
	case "json":
		return outputResumeJSON(output)
	case "ai":
		return outputResumeAI(output)
	case "table":
		return outputResumeTable(output)
	default:
		return fmt.Errorf("invalid format: %s", resumeOpts.Format)
	}
}

// outputResumeTable displays resume results in table format
func outputResumeTable(output *ResumeOutput) error {
	if len(output.Results) == 0 {
		fmt.Printf("No results found for query: %s\n", output.Query)
		return nil
	}

	fmt.Printf("🔄 Resume Results for: %s\n", output.Query)
	fmt.Printf("Found %d match(es)\n\n", output.TotalMatches)

	for i, result := range output.Results {
		session := result.Session
		fmt.Printf("%d. [%s] %s\n", i+1, session.Date, session.Title)
		fmt.Printf("   Score: %.1f | Matches: %s\n", result.MatchScore, result.MatchedText)

		if session.Description != "" {
			excerpt := truncateString(session.Description, 100)
			fmt.Printf("   %s\n", excerpt)
		}

		if len(result.Entities) > 0 {
			fmt.Printf("   Entities: %s\n", strings.Join(result.Entities, ", "))
		}

		// Show related graph if available
		if len(result.RelatedGraph) > 0 {
			fmt.Printf("   Related Graph (2-hop):\n")
			for entity, related := range result.RelatedGraph {
				if len(related) > 0 {
					fmt.Printf("     %s → %s\n", entity, strings.Join(related, ", "))
				}
			}
		}

		// Show related context
		if len(result.Context) > 0 {
			fmt.Printf("   Related Sessions:\n")
			for _, ctx := range result.Context {
				fmt.Printf("     • %s\n", ctx)
			}
		}

		fmt.Println()
	}

	return nil
}

// outputResumeJSON displays resume results in JSON format
func outputResumeJSON(output *ResumeOutput) error {
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// outputResumeAI displays resume results in AI-optimized format
func outputResumeAI(output *ResumeOutput) error {
	if len(output.Results) == 0 {
		fmt.Printf("# No Context Found\n\nQuery: %s\n", output.Query)
		return nil
	}

	fmt.Printf("# Work Context: %s\n\n", output.Query)
	fmt.Printf("Found %d relevant session(s)\n\n", output.TotalMatches)

	// Show entity type context
	if len(output.EntityTypeContext) > 0 {
		fmt.Println("## Entity Types")
		for category, entities := range output.EntityTypeContext {
			if len(entities) > 0 {
				fmt.Printf("- %s: %s\n", category, strings.Join(entities, ", "))
			}
		}
		fmt.Println()
	}

	// Show each result with full context
	for i, result := range output.Results {
		session := result.Session

		fmt.Printf("## Session %d: %s\n", i+1, session.Title)
		fmt.Printf("**Date:** %s\n", session.Date)
		fmt.Printf("**Relevance:** %.1f/10\n", result.MatchScore)
		fmt.Printf("**Matched Terms:** %s\n", result.MatchedText)

		if session.Description != "" {
			fmt.Printf("\n**Description:**\n%s\n", session.Description)
		}

		if len(result.Entities) > 0 {
			fmt.Printf("\n**Entities:** %s\n", strings.Join(result.Entities, ", "))
		}

		if len(result.RelatedGraph) > 0 {
			fmt.Printf("\n**Entity Graph (2-hop):**\n")
			for entity, related := range result.RelatedGraph {
				if len(related) > 0 {
					fmt.Printf("- %s → %s\n", entity, strings.Join(related, ", "))
				}
			}
		}

		if len(result.Context) > 0 {
			fmt.Printf("\n**Related Sessions:**\n")
			for _, ctx := range result.Context {
				fmt.Printf("- %s\n", ctx)
			}
		}

		fmt.Println("\n---\n")
	}

	return nil
}
