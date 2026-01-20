package ui

import (
	"fmt"
	"strings"

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

// RenderResultsWithContext renders the search results table with related entities
func RenderResultsWithContext(query string, results []SearchResultItem, related []string, neighbors []string, width int) string {
	rows := [][]string{}

	// Add context rows if we have any
	if len(related) > 0 {
		rows = append(rows, []string{"💡 Related entities:", strings.Join(related, ", ")})
	}
	if len(neighbors) > 0 {
		rows = append(rows, []string{"🔗 Graph neighbors:", strings.Join(neighbors, ", ")})
	}
	
	// Add summary row
	rows = append(rows, []string{fmt.Sprintf("Found %d sessions:", len(results)), ""})

	// Add result rows
	for i, r := range results {
		// Truncate title to fit width (approximate, simpler than full width calc)
		// Max width for title = width - ID width - padding
		maxTitleWidth := width - 20
		if maxTitleWidth < 10 {
			maxTitleWidth = 10
		}
		
		title := r.Title
		if len(title) > maxTitleWidth {
			title = title[:maxTitleWidth-3] + "..."
		}
		
		// ID column: "1. [session-id]"
		idCol := fmt.Sprintf("%d. [%s]", i+1, r.ID)
		rows = append(rows, []string{idCol, title})
		
		// Optional: Add narrative row if present? 
		// PRD didn't explicitly show narrative in table, but printSearchResults did.
		// For table view, maybe we keep it simple or add a third row?
		// PRD says: "Current search output uses manual... lines of code...".
		// The PRD implementation:
		// rows = append(rows, []string{fmt.Sprintf("%d. [%s]", i+1, s.Type), truncate(s.Title, width-20)})
		// It shows Type and Title.
		// My SearchResultItem has ID and Title. I'll stick to ID and Title for now.
	}

	return NewSearchTable(width).
		Headers("🔍 Search", fmt.Sprintf("%q", query)).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == table.HeaderRow:
				return TableHeaderStyle
			case row < len(related)+len(neighbors)+1: // Context rows
				return TableHintStyle
			default:
				return lipgloss.NewStyle().Padding(0, 1)
			}
		}).
		String()
}

// RenderTypoCorrection renders the typo correction table
func RenderTypoCorrection(query, corrected string, results []SearchResultItem, width int) string {
	rows := [][]string{
		{"⚠️ No exact matches.", fmt.Sprintf("Did you mean: %s ⭐", corrected)},
		{"🔄 Auto-searching:", fmt.Sprintf("%q...", corrected)},
		{fmt.Sprintf("Found %d sessions:", len(results)), ""},
	}

	for i, r := range results {
		if i >= 5 {
			break // Limit to 5 for typo preview
		}
		rows = append(rows, []string{fmt.Sprintf("%d. [%s]", i+1, r.ID), r.Title})
	}

	return NewSearchTable(width).
		Headers("🔍 Search", fmt.Sprintf("%q", query)).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch row {
			case table.HeaderRow:
				return TableHeaderStyle
			case 0:
				return TableWarningStyle
			case 1:
				return TableSuccessStyle
			default:
				return lipgloss.NewStyle().Padding(0, 1)
			}
		}).
		String()
}

// RenderNoResults renders the no results table with suggestions
func RenderNoResults(query string, suggestions []string, width int) string {
	rows := [][]string{
		{"⚠️ No sessions found.", ""},
		{"💡 Try these:", ""},
	}

	for _, s := range suggestions {
		rows = append(rows, []string{"  •", s})
	}

	return NewSearchTable(width).
		Headers("🔍 Search", fmt.Sprintf("%q", query)).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == table.HeaderRow:
				return TableHeaderStyle
			case row == 0:
				return TableWarningStyle
			case row == 1:
				return TableHintStyle.Bold(true)
			default:
				return TableHintStyle
			}
		}).
		String()
}
