package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SearchResultItem represents a search result for rendering
type SearchResultItem struct {
	ID        string
	Title     string
	Narrative string
	Reason    string
}

// RenderResultsWithContext renders the search results with header outside the table
func RenderResultsWithContext(query string, results []SearchResultItem, related []string, neighbors []string, width int) string {
	var sections []string

	// 1. Header
	header := fmt.Sprintf("🔍 Search: %q", query)
	sections = append(sections, TableHeaderStyle.Render(header))
	sections = append(sections, "") // Spacer

	// 2. Context
	var contextLines []string
	if len(related) > 0 {
		contextLines = append(contextLines, fmt.Sprintf("💡 Related: %s", strings.Join(related, ", ")))
	}
	if len(neighbors) > 0 {
		contextLines = append(contextLines, fmt.Sprintf("🔗 Impact:  %s", strings.Join(neighbors, ", ")))
	}
	if len(contextLines) > 0 {
		sections = append(sections, TableHintStyle.Render(strings.Join(contextLines, "\n")))
		sections = append(sections, "") // Spacer
	}

	// 3. Table (Results Only)

rows := [][]string{}
	for i, r := range results {
		// Truncate title
		maxTitleWidth := width - 25 // Approximate ID width + padding
		if maxTitleWidth < 10 {
			maxTitleWidth = 10
		}
		title := r.Title
		if len(title) > maxTitleWidth {
			title = title[:maxTitleWidth-3] + "..."
		}

				idCol := fmt.Sprintf("%d. [%s]", i+1, r.ID)
				rows = append(rows, []string{idCol, title})
			}
	t := NewSearchTable(width).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			return lipgloss.NewStyle().Padding(0, 1)
		})

	sections = append(sections, t.String())

	// 4. Footer
	sections = append(sections, fmt.Sprintf("  Found %d sessions", len(results)))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// RenderTypoCorrection renders the typo correction view
func RenderTypoCorrection(query, corrected string, results []SearchResultItem, width int) string {
	var sections []string

	// 1. Header
	header := fmt.Sprintf("🔍 Search: %q", query)
	sections = append(sections, TableHeaderStyle.Render(header))
	sections = append(sections, "") // Spacer

	// 2. Typo Warning
	sections = append(sections, TableWarningStyle.Render(fmt.Sprintf("  ⚠️ No exact matches. Did you mean: %s ⭐", corrected)))
	sections = append(sections, TableSuccessStyle.Render(fmt.Sprintf("  🔄 Auto-searching: %q...", corrected)))
	sections = append(sections, "") // Spacer

	// 3. Table (Results)

rows := [][]string{}
	for i, r := range results {
		if i >= 5 {
			break // Limit to 5 for typo preview
		}
		// Truncate title
		maxTitleWidth := width - 25
		if maxTitleWidth < 10 {
			maxTitleWidth = 10
		}
		title := r.Title
		if len(title) > maxTitleWidth {
			title = title[:maxTitleWidth-3] + "..."
		}

				idCol := fmt.Sprintf("%d. [%s]", i+1, r.ID)
				rows = append(rows, []string{idCol, title})
			}
	t := NewSearchTable(width).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			return lipgloss.NewStyle().Padding(0, 1)
		})

	sections = append(sections, t.String())

	// 4. Footer
	sections = append(sections, fmt.Sprintf("  Found %d sessions", len(results)))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// RenderNoResults renders the no results view
func RenderNoResults(query string, suggestions []string, width int) string {
	var sections []string

	// 1. Header
	header := fmt.Sprintf("🔍 Search: %q", query)
	sections = append(sections, TableHeaderStyle.Render(header))
	sections = append(sections, "") // Spacer

	// 2. Warning
	sections = append(sections, TableWarningStyle.Render("  ⚠️ No sessions found."))
	sections = append(sections, "") // Spacer

	// 3. Suggestions
	if len(suggestions) > 0 {
		sections = append(sections, TableHintStyle.Bold(true).Render("  💡 Try these:"))
		for _, s := range suggestions {
			sections = append(sections, TableHintStyle.Render(fmt.Sprintf("  • %s", s)))
		}
	} else {
		sections = append(sections, TableHintStyle.Render("  Consider broadening your search or checking for related terms."))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}