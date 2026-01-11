package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
)

var (
	impactDepth int
)

// DependencyNode represents an entity that depends on the target entity
type DependencyNode struct {
	Name         string       // Name of the depending entity
	Rows         []*IndexRow  // Rows where this dependency was found
	SharedWith   []string     // Entities that co-occur with this dependency
	Strength     int          // Number of rows where the dependency exists
}

// impactCmd implements 'devlog impact [entity]'
// Shows what depends on the specified entity (reverse graph traversal)
var impactCmd = &cobra.Command{
	Use:   "impact [entity]",
	Short: "Show what depends on an entity (reverse graph)",
	Long: `Show what depends on the specified entity using reverse graph traversal.

This command performs upward traversal in the entity graph to show:
  - Which entities are mentioned together with the target entity
  - Which rows reference the target entity
  - Dependency chains (what depends on what depends on the target)
  - Strength of dependency based on co-occurrence count

An entity can be:
  - CamelCase identifiers (e.g., MyFunction, ClassName)
  - kebab-case identifiers (e.g., my-function, user-name)
  - Issue IDs (e.g., bd-123)
  - Keywords (e.g., TODO, FIXME, NOTE)

Examples:
  devlog impact MyFunction          # Show what depends on MyFunction
  devlog impact --depth 2 bd-123    # Show 2 levels of dependencies
  devlog impact --depth 1 TODO      # Show direct dependencies only`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		targetEntity := args[0]

		// Default devlog path
		indexPath := "index.md"

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

		// Check if target entity exists
		targetNode, exists := entityGraph[targetEntity]
		if !exists {
			fmt.Fprintf(os.Stderr, "Entity '%s' not found in devlog\n", targetEntity)
			fmt.Fprintf(os.Stderr, "Run 'devlog graph' without arguments to see all entities\n")
			os.Exit(1)
		}

		// Build reverse dependency graph (what depends on what)
		dependencyGraph := buildDependencyGraph(entityGraph)

		// Get entities that depend on the target
		dependencies := getDependencies(targetEntity, dependencyGraph, entityGraph)

		if len(dependencies) == 0 {
			fmt.Printf("\n📊 Impact Analysis: %s\n\n", targetEntity)
			fmt.Println("  No entities found that depend on this entity.")
			fmt.Println("\n  This entity appears in the following locations:")
			for _, row := range targetNode.Rows {
				fmt.Printf("    • %s: %s\n", row.Date, row.Title)
			}
			return
		}

		// Print impact analysis
		printImpactAnalysis(targetEntity, dependencies, dependencyGraph, entityGraph, rows)
	},
}

func init() {
	impactCmd.Flags().IntVarP(&impactDepth, "depth", "d", 1, "Maximum depth of dependency traversal")
}

// buildDependencyGraph creates a reverse graph showing what depends on what
// For each entity X, it stores which entities depend on X
func buildDependencyGraph(graph map[string]*EntityNode) map[string][]*DependencyNode {
	dependencyGraph := make(map[string][]*DependencyNode)

	// For each entity in the graph
	for entityName, node := range graph {
		// For each entity that this entity relates to
		for _, ref := range node.RelatedTo {
			// entityName depends on ref.Name
			// So in the reverse graph, ref.Name should have entityName as a dependent

			// Find the rows where entityName and ref.Name co-occur
			var sharedRows []*IndexRow
			for _, row := range node.Rows {
				hasRef := false
				for _, e := range row.Entities {
					if e == ref.Name {
						hasRef = true
						break
					}
				}
				if hasRef {
					sharedRows = append(sharedRows, row)
				}
			}

			// Collect other entities that co-occur in these rows
			sharedWithSet := make(map[string]bool)
			for _, row := range sharedRows {
				for _, e := range row.Entities {
					if e != entityName && e != ref.Name {
						sharedWithSet[e] = true
					}
				}
			}
			var sharedWith []string
			for e := range sharedWithSet {
				sharedWith = append(sharedWith, e)
			}
			sort.Strings(sharedWith)

			// Create or update the dependency node
			depNode := &DependencyNode{
				Name:       entityName,
				Rows:       sharedRows,
				SharedWith: sharedWith,
				Strength:   ref.Strength,
			}

			dependencyGraph[ref.Name] = append(dependencyGraph[ref.Name], depNode)
		}
	}

	return dependencyGraph
}

// getDependencies gets entities that depend on the target entity
func getDependencies(targetEntity string, dependencyGraph map[string][]*DependencyNode, entityGraph map[string]*EntityNode) []*DependencyNode {
	deps, exists := dependencyGraph[targetEntity]
	if !exists {
		return []*DependencyNode{}
	}

	// Sort by strength (descending)
	sort.Slice(deps, func(i, j int) bool {
		if deps[i].Strength != deps[j].Strength {
			return deps[i].Strength > deps[j].Strength
		}
		return deps[i].Name < deps[j].Name
	})

	return deps
}

// printImpactAnalysis prints the impact analysis for a target entity
func printImpactAnalysis(targetEntity string, dependencies []*DependencyNode, dependencyGraph map[string][]*DependencyNode, entityGraph map[string]*EntityNode, rows []*IndexRow) {
	fmt.Printf("\n📊 Impact Analysis: %s\n\n", targetEntity)

	// Show summary
	fmt.Printf("  %d entity/ies depend on %s\n\n", len(dependencies), targetEntity)

	// Group dependencies by type for better organization
	camelCase := []*DependencyNode{}
	kebabCase := []*DependencyNode{}
	keywords := []*DependencyNode{}
	issueIDs := []*DependencyNode{}

	for _, dep := range dependencies {
		if isCamelCase(dep.Name) {
			camelCase = append(camelCase, dep)
		} else if isKebabCase(dep.Name) {
			kebabCase = append(kebabCase, dep)
		} else if isKeyword(dep.Name) {
			keywords = append(keywords, dep)
		} else if isIssueID(dep.Name) {
			issueIDs = append(issueIDs, dep)
		}
	}

	// Print dependencies by type
	printDependencyGroup("CamelCase", camelCase, targetEntity, 0, dependencyGraph, entityGraph)
	printDependencyGroup("kebab-case", kebabCase, targetEntity, 0, dependencyGraph, entityGraph)
	printDependencyGroup("Keywords", keywords, targetEntity, 0, dependencyGraph, entityGraph)
	printDependencyGroup("Issue IDs", issueIDs, targetEntity, 0, dependencyGraph, entityGraph)

	// Show where the target entity appears
	targetNode := entityGraph[targetEntity]
	fmt.Printf("  %s appears in %d row(s):\n", targetEntity, len(targetNode.Rows))
	for _, row := range targetNode.Rows {
		fmt.Printf("    • %s: %s\n", row.Date, row.Title)
		if row.Description != "" {
			fmt.Printf("      %s\n", truncateString(row.Description, 80))
		}
	}

	fmt.Println()
}

// printDependencyGroup prints a group of dependencies
func printDependencyGroup(groupName string, dependencies []*DependencyNode, targetEntity string, currentDepth int, dependencyGraph map[string][]*DependencyNode, entityGraph map[string]*EntityNode) {
	if len(dependencies) == 0 {
		return
	}

	fmt.Printf("  %s (%d):\n", groupName, len(dependencies))

	indent := ""
	for i := 0; i < currentDepth; i++ {
		indent += "    "
	}

	for i, dep := range dependencies {
		prefix := "├── "
		if i == len(dependencies)-1 {
			prefix = "└── "
		}

		fmt.Printf("  %s%s%s", indent, prefix, dep.Name)
		fmt.Printf(" (%d co-occurrence%s)", dep.Strength, pluralS(dep.Strength))
		fmt.Println()

		// Show shared context
		if len(dep.SharedWith) > 0 && currentDepth == 0 {
			sharedIndent := indent + "    "
			fmt.Printf("%s    Shared with: %s\n", sharedIndent, formatEntityList(dep.SharedWith))
		}

		// Show rows where dependency exists
		if currentDepth == 0 && len(dep.Rows) > 0 {
			rowIndent := indent + "    "
			fmt.Printf("%s    Found in:\n", rowIndent)
			for j, row := range dep.Rows {
				if j >= 3 {
					fmt.Printf("%s      ... and %d more\n", rowIndent, len(dep.Rows)-j)
					break
				}
				fmt.Printf("%s      • %s: %s\n", rowIndent, row.Date, row.Title)
			}
		}

		// Recursively show transitive dependencies
		if currentDepth < impactDepth-1 {
			transitiveDeps := getDependencies(dep.Name, dependencyGraph, entityGraph)
			if len(transitiveDeps) > 0 {
				// Filter out the target entity to avoid cycles
				filteredDeps := []*DependencyNode{}
				for _, td := range transitiveDeps {
					if td.Name != targetEntity {
						filteredDeps = append(filteredDeps, td)
					}
				}

				if len(filteredDeps) > 0 {
					// Group transitive dependencies
					camelCase := []*DependencyNode{}
					kebabCase := []*DependencyNode{}
					keywords := []*DependencyNode{}
					issueIDs := []*DependencyNode{}

					for _, td := range filteredDeps {
						if isCamelCase(td.Name) {
							camelCase = append(camelCase, td)
						} else if isKebabCase(td.Name) {
							kebabCase = append(kebabCase, td)
						} else if isKeyword(td.Name) {
							keywords = append(keywords, td)
						} else if isIssueID(td.Name) {
							issueIDs = append(issueIDs, td)
						}
					}

					printDependencyGroup("", camelCase, targetEntity, currentDepth+1, dependencyGraph, entityGraph)
					printDependencyGroup("", kebabCase, targetEntity, currentDepth+1, dependencyGraph, entityGraph)
					printDependencyGroup("", keywords, targetEntity, currentDepth+1, dependencyGraph, entityGraph)
					printDependencyGroup("", issueIDs, targetEntity, currentDepth+1, dependencyGraph, entityGraph)
				}
			}
		}
	}

	if currentDepth == 0 {
		fmt.Println()
	}
}

// formatEntityList formats a list of entities for display
func formatEntityList(entities []string) string {
	if len(entities) == 0 {
		return "none"
	}
	if len(entities) <= 3 {
		return fmt.Sprintf("%s", entities)
	}
	return fmt.Sprintf("%s, and %d more", entities[:3], len(entities)-3)
}
