package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// SearchResultItem represents a search result for rendering
type SearchResultItem struct {
	ID        string
	Title     string
	Narrative string
	Reason    string
}

// renderSingleTable renders a simple list into a 1-column table with a header
func renderSingleTable(title string, items []string, width int) string {
	if len(items) == 0 {
		return ""
	}

	rows := [][]string{}
	for i, item := range items {
		rows = append(rows, []string{fmt.Sprintf("%d. %s", i+1, item)})
	}

	return table.New().
		Headers(title).
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(ColorMuted)).
		Width(width).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return TableHeaderStyle.Width(width - 2)
			}
			return lipgloss.NewStyle().Padding(0, 1).Align(lipgloss.Left)
		}).
		String()
}

// RenderResultsWithContext renders the search results with headers and tables
func RenderResultsWithContext(query string, results []SearchResultItem, related []string, neighbors []string, width int) string {
	var sections []string

	// 1. Header
	header := fmt.Sprintf("🔍 Search: %q", query)
	sections = append(sections, TableHeaderStyle.Render(header))
	sections = append(sections, "") // Spacer

	// 2. Context Tables
	if relatedTable := renderSingleTable("💡 Related Entities (Matched via FTS)", related, width); relatedTable != "" {
		sections = append(sections, relatedTable)
		sections = append(sections, "") // Spacer
	}

	if neighborsTable := renderSingleTable("🔗 Graph Neighbors (Impact)", neighbors, width); neighborsTable != "" {
		sections = append(sections, neighborsTable)
		sections = append(sections, "") // Spacer
	}

	// 3. Results Table
	if len(results) > 0 {
		rows := [][]string{}
		for i, r := range results {
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

		t := table.New().
			Headers(fmt.Sprintf("📄 Found %d sessions", len(results))).
			Border(lipgloss.RoundedBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(ColorMuted)).
			Width(width).
			Rows(rows...).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return TableHeaderStyle.Width(width - 2)
				}
				// Column 0 (ID) gets fixed width, Column 1 (Title) takes rest
				style := lipgloss.NewStyle().Padding(0, 1).Align(lipgloss.Left)
				if col == 0 {
					style = style.Width(20)
				}
				return style
			})

		sections = append(sections, t.String())
	}

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
	sections = append(sections, TableWarningStyle.Render(fmt.Sprintf("  ⚠️  No exact matches. Did you mean: %s ⭐", corrected)))
	sections = append(sections, TableSuccessStyle.Render(fmt.Sprintf("  🔄 Auto-searching: %q...", corrected)))
	sections = append(sections, "") // Spacer

	// 3. Results Table
	if len(results) > 0 {
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

		t := table.New().
			Headers(fmt.Sprintf("📄 Found %d sessions", len(results))).
			Border(lipgloss.RoundedBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(ColorMuted)).
			Width(width).
			Rows(rows...).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return TableHeaderStyle.Width(width - 2)
				}
				style := lipgloss.NewStyle().Padding(0, 1).Align(lipgloss.Left)
				if col == 0 {
					style = style.Width(20)
				}
				return style
			})

		sections = append(sections, t.String())
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// RenderImpactTable renders the impact of an entity using a Lipgloss table
func RenderImpactTable(target string, deps []string, width int) string {
	title := fmt.Sprintf("💥 Impact Analysis: [%s]", target)
	
	if len(deps) == 0 {
		return table.New().
			Headers(title).
			Rows([]string{"(No known dependencies)"}).
			Border(lipgloss.RoundedBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(ColorMuted)).
			Width(width).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return TableHeaderStyle.Width(width - 2)
				}
				return lipgloss.NewStyle().Padding(0, 1).Align(lipgloss.Left).Foreground(ColorMuted)
			}).
			String()
	}

	rows := [][]string{}
	for _, dep := range deps {
		rows = append(rows, []string{dep})
	}

	return table.New().
		Headers(title).
		Rows(rows...).
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(ColorMuted)).
		Width(width).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return TableHeaderStyle.Width(width - 2)
			}
			return lipgloss.NewStyle().Padding(0, 1).Align(lipgloss.Left)
		}).
		String()
}

// RenderEntitiesTable renders a list of entities and their mention counts
func RenderEntitiesTable(entities [][]string, width int) string {
	return table.New().
		Headers("Top Entities", "Mentions").
		Rows(entities...).
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(ColorMuted)).
		Width(width).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				// Each header cell needs to be half width - 1 for border
				return TableHeaderStyle.Width(width/2 - 1)
			}
			style := lipgloss.NewStyle().Padding(0, 1)
			if col == 1 {
				style = style.Align(lipgloss.Right)
			} else {
				style = style.Align(lipgloss.Left)
			}
			return style
		}).
		String()
}

// RenderNoResults renders the no results view
func RenderNoResults(query string, suggestions []string, width int) string {
	var sections []string

	// 1. Header
	header := fmt.Sprintf("🔍 Search: %q", query)
	sections = append(sections, TableHeaderStyle.Render(header))
	sections = append(sections, "") // Spacer

	// 2. Warning
	sections = append(sections, TableWarningStyle.Render("  ⚠️  No sessions found."))
	sections = append(sections, "") // Spacer

	// 3. Suggestions Table
	if len(suggestions) > 0 {
		sections = append(sections, renderSingleTable("💡 Suggestions (Did you mean?)", suggestions, width))
	} else {
		sections = append(sections, TableHintStyle.Render("  Consider broadening your search or checking for related terms."))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
