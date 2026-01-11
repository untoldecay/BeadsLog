package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// EntityStats contains statistics for a single entity
type EntityStats struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`         // "CamelCase", "kebab-case", "keyword", "issue-id"
	MentionCount int      `json:"mention_count"`
	FirstSeen    string   `json:"first_seen"`
	LastSeen     string   `json:"last_seen"`
	Contexts     []string `json:"contexts"` // Titles/contexts where entity appears
}

// EntitiesReport contains the full entities report
type EntitiesReport struct {
	TotalEntities int                    `json:"total_entities"`
	TotalMentions int                    `json:"total_mentions"`
	ByType        map[string]int         `json:"by_type"`
	Entities      []*EntityStats         `json:"entities"`
	SortedBy      string                 `json:"sorted_by"` // "mention_count"
}

var (
	entitiesFormat  string // Output format: "table" or "json"
	entitiesType    string // Filter by entity type
	entitiesLimit   int    // Limit number of entities shown
	entitiesMinimum int    // Minimum mention count to include
)

// entitiesCmd represents the entities command
var entitiesCmd = &cobra.Command{
	Use:   "entities",
	Short: "List all entities sorted by mention count",
	Long: `List all entities from index.md sorted by mention count with statistics.

This command analyzes all entries in index.md and displays:
  - Entity name and type (CamelCase, kebab-case, keyword, issue-id)
  - Mention count (frequency)
  - First and last seen dates
  - Contexts (entries where entity appears)

Entities are sorted by mention count (descending) by default.

Examples:
  devlog entities                    # Show all entities sorted by mentions
  devlog entities --type CamelCase   # Show only CamelCase entities
  devlog entities --min 3            # Show entities mentioned at least 3 times
  devlog entities --limit 10         # Show top 10 entities
  devlog entities --format json      # Output in JSON format`,
	RunE: runEntities,
}

func init() {
	entitiesCmd.Flags().StringVarP(&entitiesFormat, "format", "f", "table", "Output format: table or json")
	entitiesCmd.Flags().StringVarP(&entitiesType, "type", "t", "", "Filter by type (CamelCase, kebab-case, keyword, issue-id)")
	entitiesCmd.Flags().IntVarP(&entitiesLimit, "limit", "l", 0, "Limit number of entities shown (0 = unlimited)")
	entitiesCmd.Flags().IntVarP(&entitiesMinimum, "min", "m", 1, "Minimum mention count to include")
}

func runEntities(cmd *cobra.Command, args []string) error {
	// Default devlog path
	indexPath := "./index.md"
	if len(args) > 0 {
		indexPath = args[0]
	}

	// Parse the index file
	rows, err := parseIndexMD(indexPath)
	if err != nil {
		return fmt.Errorf("error parsing index.md: %w", err)
	}

	if len(rows) == 0 {
		fmt.Println("No entries found in index.md")
		return nil
	}

	// Build entity statistics
	report := buildEntitiesReport(rows)

	// Apply filters
	report = filterEntitiesReport(report)

	// Sort entities by mention count
	sortEntitiesByMentionCount(report.Entities)

	// Apply limit
	if entitiesLimit > 0 && len(report.Entities) > entitiesLimit {
		report.Entities = report.Entities[:entitiesLimit]
	}

	// Update totals after filtering
	report.TotalEntities = len(report.Entities)

	// Output based on format
	switch entitiesFormat {
	case "json":
		return outputEntitiesJSON(report)
	case "table":
		return outputEntitiesTable(report)
	default:
		return fmt.Errorf("invalid format: %s (must be 'table' or 'json')", entitiesFormat)
	}
}

// buildEntitiesReport creates a comprehensive report of all entities
func buildEntitiesReport(rows []*IndexRow) *EntitiesReport {
	// Track entities and their statistics
	entityMap := make(map[string]*EntityStats)
	typeCounts := make(map[string]int)
	totalMentions := 0

	for _, row := range rows {
		for _, entity := range row.Entities {
			// Initialize entity stats if not exists
			if _, exists := entityMap[entity]; !exists {
				entityType := getEntityType(entity)
				entityMap[entity] = &EntityStats{
					Name:      entity,
					Type:      entityType,
					Contexts:  []string{},
					FirstSeen: row.Date,
					LastSeen:  row.Date,
				}
				typeCounts[entityType]++
			}

			// Update statistics
			stats := entityMap[entity]
			stats.MentionCount++

			// Update first/last seen
			if row.Date < stats.FirstSeen {
				stats.FirstSeen = row.Date
			}
			if row.Date > stats.LastSeen {
				stats.LastSeen = row.Date
			}

			// Add context if not already present
			context := fmt.Sprintf("%s: %s", row.Date, row.Title)
			found := false
			for _, ctx := range stats.Contexts {
				if ctx == context {
					found = true
					break
				}
			}
			if !found {
				stats.Contexts = append(stats.Contexts, context)
			}

			totalMentions++
		}
	}

	// Convert map to slice
	var entities []*EntityStats
	for _, stats := range entityMap {
		entities = append(entities, stats)
	}

	return &EntitiesReport{
		TotalEntities: len(entities),
		TotalMentions: totalMentions,
		ByType:        typeCounts,
		Entities:      entities,
		SortedBy:      "mention_count",
	}
}

// filterEntitiesReport applies filters to the entities report
func filterEntitiesReport(report *EntitiesReport) *EntitiesReport {
	if entitiesType == "" && entitiesMinimum <= 1 {
		return report
	}

	var filtered []*EntityStats
	typeCounts := make(map[string]int)
	totalMentions := 0

	for _, entity := range report.Entities {
		// Filter by type
		if entitiesType != "" {
			if !strings.EqualFold(entity.Type, entitiesType) {
				continue
			}
		}

		// Filter by minimum mention count
		if entity.MentionCount < entitiesMinimum {
			continue
		}

		filtered = append(filtered, entity)
		typeCounts[entity.Type]++
		totalMentions += entity.MentionCount
	}

	return &EntitiesReport{
		TotalEntities: len(filtered),
		TotalMentions: totalMentions,
		ByType:        typeCounts,
		Entities:      filtered,
		SortedBy:      report.SortedBy,
	}
}

// sortEntitiesByMentionCount sorts entities by mention count (descending)
func sortEntitiesByMentionCount(entities []*EntityStats) {
	sort.Slice(entities, func(i, j int) bool {
		// Primary sort: by mention count (descending)
		if entities[i].MentionCount != entities[j].MentionCount {
			return entities[i].MentionCount > entities[j].MentionCount
		}
		// Secondary sort: by name (ascending)
		return entities[i].Name < entities[j].Name
	})
}

// outputEntitiesTable displays entities in table format
func outputEntitiesTable(report *EntitiesReport) error {
	if len(report.Entities) == 0 {
		fmt.Println("No entities found matching the criteria.")
		return nil
	}

	// Print header
	fmt.Println("📊 Entity Statistics Report")
	fmt.Println()

	// Print summary
	fmt.Printf("Total Entities: %d\n", report.TotalEntities)
	fmt.Printf("Total Mentions: %d\n", report.TotalMentions)
	fmt.Println()

	// Print breakdown by type
	if len(report.ByType) > 0 {
		fmt.Println("Breakdown by Type:")
		// Sort types by count
		types := make([]string, 0, len(report.ByType))
		for t := range report.ByType {
			types = append(types, t)
		}
		sort.Slice(types, func(i, j int) bool {
			return report.ByType[types[i]] > report.ByType[types[j]]
		})

		for _, t := range types {
			fmt.Printf("  %s: %d\n", t, report.ByType[t])
		}
		fmt.Println()
	}

	// Print entities table
	fmt.Println("Top Entities (by mention count):")
	fmt.Println()

	// Calculate column widths
	maxNameLen := 4
	maxTypeLen := 4
	for _, e := range report.Entities {
		if len(e.Name) > maxNameLen {
			maxNameLen = len(e.Name)
		}
		if len(e.Type) > maxTypeLen {
			maxTypeLen = len(e.Type)
		}
	}

	// Print header row
	fmt.Printf("  %-*s  %-*s  %7s  %12s  %12s  %s\n",
		maxNameLen, "Entity",
		maxTypeLen, "Type",
		"Mentions",
		"First Seen",
		"Last Seen",
		"Contexts")
	fmt.Printf("  %s  %s  %s  %s  %s  %s\n",
		strings.Repeat("-", maxNameLen),
		strings.Repeat("-", maxTypeLen),
		strings.Repeat("-", 7),
		strings.Repeat("-", 12),
		strings.Repeat("-", 12),
		strings.Repeat("-", 50))

	// Print entity rows
	for _, e := range report.Entities {
		// Truncate contexts if too long
		contextsStr := ""
		if len(e.Contexts) > 0 {
			contextsStr = fmt.Sprintf("[%d] %s", len(e.Contexts), truncateString(e.Contexts[0], 40))
			if len(e.Contexts) > 1 {
				contextsStr += fmt.Sprintf(" (+%d more)", len(e.Contexts)-1)
			}
		}

		fmt.Printf("  %-*s  %-*s  %7d  %12s  %12s  %s\n",
			maxNameLen, e.Name,
			maxTypeLen, e.Type,
			e.MentionCount,
			e.FirstSeen,
			e.LastSeen,
			contextsStr)
	}

	fmt.Println()

	return nil
}

// outputEntitiesJSON displays entities in JSON format
func outputEntitiesJSON(report *EntitiesReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// getEntityType determines the type of an entity
func getEntityType(entity string) string {
	// Check in order: keyword, issue-id, CamelCase, kebab-case
	if isKeyword(entity) {
		return "keyword"
	}
	if isIssueID(entity) {
		return "issue-id"
	}
	if isCamelCase(entity) {
		return "CamelCase"
	}
	if isKebabCase(entity) {
		return "kebab-case"
	}
	return "unknown"
}
