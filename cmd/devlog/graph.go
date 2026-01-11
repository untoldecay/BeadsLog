package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	graphDepth int
)

// EntityNode represents an entity and its relationships in the graph
type EntityNode struct {
	Name      string       // Entity name (e.g., "MyFunction", "bd-123")
	Rows      []*IndexRow  // Rows where this entity appears
	RelatedTo []*EntityRef // Entities referenced in the same rows
}

// EntityRef represents a reference to another entity
type EntityRef struct {
	Name     string
	Row      *IndexRow
	Strength int // Number of co-occurrences
}

// graphCmd implements 'devlog graph [entity] --depth N'
// Displays hierarchical tree output showing entity relationships
var graphCmd = &cobra.Command{
	Use:   "graph [entity]",
	Short: "Display entity relationship graph",
	Long: `Display a hierarchical tree of entity relationships from your devlog.

An entity can be:
  - CamelCase identifiers (e.g., MyFunction, ClassName)
  - kebab-case identifiers (e.g., my-function, user-name)
  - Issue IDs (e.g., bd-123)
  - Keywords (e.g., TODO, FIXME, NOTE)

The graph shows:
  - Which rows contain the entity
  - Related entities (co-mentioned in the same rows)
  - Hierarchical tree with proper indentation
  - Depth control to limit traversal

Examples:
  devlog graph MyFunction          # Show relationships for MyFunction
  devlog graph --depth 2 bd-123    # Show 2 levels of relationships
  devlog graph --depth 1 TODO      # Show TODO items and direct relationships`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Default devlog path
		indexPath := "index.md"
		if len(args) > 0 && args[0] != "" {
			// Check if arg is a file path or entity name
			if _, err := os.Stat(args[0]); err == nil {
				indexPath = args[0]
			}
		}

		// Parse the index file
		rows, err := parseIndexMD(indexPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing index.md: %v\n", err)
			os.Exit(1)
		}

		if len(rows) == 0 {
			fmt.Println("No entries found in index.md")
			return
		}

		// Build entity graph
		entityGraph := buildEntityGraph(rows)

		// If no entity specified, show all entities
		var targetEntity string
		if len(args) > 0 {
			// Check if last arg is entity name (not file path)
			potentialEntity := args[len(args)-1]
			_, isFile := os.Stat(potentialEntity)
			if isFile != nil {
				// Not a file, treat as entity
				targetEntity = potentialEntity
			}
		}

		if targetEntity != "" {
			// Show graph for specific entity
			node, exists := entityGraph[targetEntity]
			if !exists {
				fmt.Fprintf(os.Stderr, "Entity '%s' not found in devlog\n", targetEntity)
				fmt.Fprintf(os.Stderr, "Run 'devlog graph' without arguments to see all entities\n")
				os.Exit(1)
			}
			printEntityGraph(node, entityGraph, graphDepth, 0)
		} else {
			// Show all entities
			printAllEntities(entityGraph, rows)
		}
	},
}

func init() {
	graphCmd.Flags().IntVarP(&graphDepth, "depth", "d", 2, "Maximum depth of relationship traversal")
}

// buildEntityGraph creates a map of entity names to their nodes
func buildEntityGraph(rows []*IndexRow) map[string]*EntityNode {
	graph := make(map[string]*EntityNode)
	entityRefs := make(map[string]map[string]int) // entity -> related entity -> count

	// First pass: collect all entities and their rows
	for _, row := range rows {
		for _, entity := range row.Entities {
			if _, exists := graph[entity]; !exists {
				graph[entity] = &EntityNode{
					Name: entity,
					Rows: []*IndexRow{},
				}
				entityRefs[entity] = make(map[string]int)
			}
			graph[entity].Rows = append(graph[entity].Rows, row)
		}
	}

	// Second pass: build relationships
	for _, row := range rows {
		entities := row.Entities
		for i, entity1 := range entities {
			for j, entity2 := range entities {
				if i == j {
					continue
				}
				// entity1 is related to entity2
				entityRefs[entity1][entity2]++
			}
		}
	}

	// Third pass: convert refs to EntityRef structs
	for entityName, refs := range entityRefs {
		node := graph[entityName]
		for relatedName, count := range refs {
			// Find a row where both entities appear
			var sharedRow *IndexRow
			for _, row := range node.Rows {
				hasRelated := false
				for _, e := range row.Entities {
					if e == relatedName {
						hasRelated = true
						break
					}
				}
				if hasRelated {
					sharedRow = row
					break
				}
			}

			node.RelatedTo = append(node.RelatedTo, &EntityRef{
				Name:     relatedName,
				Row:      sharedRow,
				Strength: count,
			})
		}
	}

	return graph
}

// printEntityGraph prints a hierarchical tree for a single entity
func printEntityGraph(node *EntityNode, graph map[string]*EntityNode, depth int, currentDepth int) {
	if depth > 0 && currentDepth >= depth {
		return
	}

	indent := ""
	for i := 0; i < currentDepth; i++ {
		indent += "│   "
	}

	// Print entity header
	if currentDepth == 0 {
		fmt.Printf("\n📊 Entity Graph: %s\n\n", node.Name)
		fmt.Printf("  Found in %d row(s):\n", len(node.Rows))
		for _, row := range node.Rows {
			fmt.Printf("    • %s: %s\n", row.Date, row.Title)
			if row.Description != "" {
				fmt.Printf("      %s\n", truncateString(row.Description, 80))
			}
		}
	} else {
		fmt.Printf("%s├── %s", indent, node.Name)
		if len(node.Rows) > 0 {
			fmt.Printf(" (%d row%s)", len(node.Rows), pluralS(len(node.Rows)))
		}
		fmt.Println()
	}

	// Print related entities
	if len(node.RelatedTo) > 0 && (depth == 0 || currentDepth < depth-1) {
		if currentDepth == 0 {
			fmt.Printf("\n  Related entities:\n")
		}

		// Sort by strength (number of co-occurrences)
		sortedRefs := sortEntityRefs(node.RelatedTo)

		for i, ref := range sortedRefs {
			isLast := i == len(sortedRefs)-1
			prefix := "├── "
			if isLast {
				prefix = "└── "
			}

			relatedNode, exists := graph[ref.Name]
			if !exists {
				continue
			}

			if currentDepth == 0 {
				fmt.Printf("  %s%s (%d co-occurrence%s)", prefix, ref.Name, ref.Strength, pluralS(ref.Strength))
				fmt.Println()
			}

			// Recursively print related entities
			newIndent := indent
			if currentDepth == 0 {
				newIndent = "    "
			} else {
				if isLast {
					newIndent += "    "
				} else {
					newIndent += "│   "
				}
			}

			printEntityGraphRecursive(relatedNode, graph, depth, currentDepth+1, newIndent)
		}
	}

	fmt.Println()
}

// printEntityGraphRecursive handles recursive printing with proper indentation
func printEntityGraphRecursive(node *EntityNode, graph map[string]*EntityNode, depth int, currentDepth int, indent string) {
	if depth > 0 && currentDepth >= depth {
		return
	}

	// Print this entity
	fmt.Printf("%s%s", indent, node.Name)
	if len(node.Rows) > 0 {
		fmt.Printf(" (%d row%s)", len(node.Rows), pluralS(len(node.Rows)))
	}
	fmt.Println()

	// Recursively print related entities
	if len(node.RelatedTo) > 0 && (depth == 0 || currentDepth < depth-1) {
		sortedRefs := sortEntityRefs(node.RelatedTo)

		for i, ref := range sortedRefs {
			isLast := i == len(sortedRefs)-1
			prefix := "├── "
			if isLast {
				prefix = "└── "
			}

			relatedNode, exists := graph[ref.Name]
			if !exists {
				continue
			}

			newIndent := indent
			if isLast {
				newIndent += "    "
			} else {
				newIndent += "│   "
			}

			fmt.Printf("%s%s%s (%d co-occurrence%s)\n", newIndent, prefix, ref.Name, ref.Strength, pluralS(ref.Strength))

			// Recurse
			printEntityGraphRecursive(relatedNode, graph, depth, currentDepth+1, newIndent)
		}
	}
}

// printAllEntities prints a summary of all entities found
func printAllEntities(graph map[string]*EntityNode, rows []*IndexRow) {
	fmt.Printf("\n📊 Entities Found: %d\n\n", len(graph))

	// Group entities by type (CamelCase, kebab-case, keywords, issue IDs)
	camelCase := []string{}
	kebabCase := []string{}
	keywords := []string{}
	issueIDs := []string{}

	for name := range graph {
		if isCamelCase(name) {
			camelCase = append(camelCase, name)
		} else if isKebabCase(name) {
			kebabCase = append(kebabCase, name)
		} else if isKeyword(name) {
			keywords = append(keywords, name)
		} else if isIssueID(name) {
			issueIDs = append(issueIDs, name)
		}
	}

	printEntityGroup("CamelCase", camelCase, graph)
	printEntityGroup("kebab-case", kebabCase, graph)
	printEntityGroup("Keywords", keywords, graph)
	printEntityGroup("Issue IDs", issueIDs, graph)

	fmt.Printf("\nTotal: %d entries parsed\n", len(rows))
	fmt.Println("Use 'devlog graph <entity>' to see detailed relationships")
}

func printEntityGroup(groupName string, entities []string, graph map[string]*EntityNode) {
	if len(entities) == 0 {
		return
	}

	fmt.Printf("  %s (%d):\n", groupName, len(entities))
	for _, name := range entities {
		node := graph[name]
		fmt.Printf("    • %s", name)
		if len(node.Rows) > 0 {
			fmt.Printf(" (%d row%s)", len(node.Rows), pluralS(len(node.Rows)))
		}
		fmt.Println()
	}
	fmt.Println()
}

// Helper functions

func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func sortEntityRefs(refs []*EntityRef) []*EntityRef {
	// Simple bubble sort by strength (descending)
	sorted := make([]*EntityRef, len(refs))
	copy(sorted, refs)

	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Strength < sorted[j+1].Strength {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

func isCamelCase(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Check for at least one uppercase letter followed by lowercase
	hasUpper := false
	hasLower := false
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			hasUpper = true
		}
		if r >= 'a' && r <= 'z' {
			hasLower = true
		}
	}
	return hasUpper && hasLower
}

func isKebabCase(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Contains hyphen and starts with lowercase
	for i, r := range s {
		if r == '-' && i > 0 && i < len(s)-1 {
			return s[0] >= 'a' && s[0] <= 'z'
		}
	}
	return false
}

func isKeyword(s string) bool {
	keywords := []string{"TODO", "FIXME", "NOTE", "HACK", "XXX", "BUG", "OPTIMIZE", "REFACTOR"}
	for _, kw := range keywords {
		if s == kw {
			return true
		}
	}
	return false
}

func isIssueID(s string) bool {
	// Matches bd-XXX or BD-XXX pattern
	if len(s) < 4 {
		return false
	}
	prefix := s[:3]
	return (prefix == "bd-" || prefix == "BD-" || prefix == "Bd-" || prefix == "bD-")
}
