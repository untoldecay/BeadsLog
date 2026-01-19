package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Search Styles
	searchBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorMuted).
		Padding(0, 1).
		Margin(1, 0)

	searchTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorAccent)

	searchContextStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(ColorMuted).
		Padding(0, 0).
		MarginTop(0)

	searchSuggestionStyle = lipgloss.NewStyle().
		Foreground(ColorPass).
		Bold(true)
)

// SearchViewModel holds data for rendering the search result box
type SearchViewModel struct {
	Query           string
	TypoCorrection  string // "nginx"
	TypoDistance    int    // 1
	AutoSearching   bool
	Suggestions     []string
	RelatedEntities []string
	GraphNeighbors  []string
	ResultsCount    int
	NoResults       bool
}

// RenderSearchBox renders the search context box
func RenderSearchBox(vm SearchViewModel) string {
	var sections []string

	// 1. Header: 🔍 Search: "query"
	header := fmt.Sprintf("🔍 Search: %q", vm.Query)
	sections = append(sections, searchTitleStyle.Render(header))

	// 2. Context Section
	var contextLines []string

	// Typo Correction
	if vm.TypoCorrection != "" {
		msg := fmt.Sprintf("⚠️  No exact matches. Did you mean: %s ⭐", searchSuggestionStyle.Render(vm.TypoCorrection))
		contextLines = append(contextLines, msg)
		if vm.AutoSearching {
			contextLines = append(contextLines, fmt.Sprintf("🔄 Auto-searching %q...", vm.TypoCorrection))
		}
	}

	// No Results & Suggestions
	if vm.NoResults && len(vm.Suggestions) > 0 && vm.TypoCorrection == "" {
		contextLines = append(contextLines, "⚠️  No sessions found.")
		contextLines = append(contextLines, "💡 Try these:")
		for _, s := range vm.Suggestions {
			contextLines = append(contextLines, fmt.Sprintf("  • %s", s))
		}
	}

	// Related Entities
	if len(vm.RelatedEntities) > 0 {
		contextLines = append(contextLines, fmt.Sprintf("💡 Related: %s", strings.Join(vm.RelatedEntities, ", ")))
	}

	// Graph Neighbors
	if len(vm.GraphNeighbors) > 0 {
		contextLines = append(contextLines, fmt.Sprintf("🔗 Graph neighbors: %s", strings.Join(vm.GraphNeighbors, ", ")))
	}

	// Results Summary
	if vm.ResultsCount > 0 {
		// Add separator if we have previous context
		if len(contextLines) > 0 {
			// contextLines = append(contextLines, RenderSeparator()) // Too wide?
		}
		contextLines = append(contextLines, fmt.Sprintf("Found %d sessions:", vm.ResultsCount))
	} else if vm.NoResults && len(vm.Suggestions) == 0 && vm.TypoCorrection == "" {
		contextLines = append(contextLines, "⚠️  No sessions found.")
		contextLines = append(contextLines, "Consider broadening your search or checking for related terms.")
	}

	if len(contextLines) > 0 {
		// Join context lines and wrap in a section style (top border)
		contextBlock := strings.Join(contextLines, "\n")
		// Only add border if header exists (always does)
		sections = append(sections, searchContextStyle.Render(contextBlock))
	}

	return searchBoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, sections...))
}
