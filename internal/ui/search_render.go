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
	Date      string
	Narrative string
	Reason    string
	
	// Explanation fields
	Score        float64
	BM25         float64
	PhraseBonus  float64
	NearBonus    float64
	EntityBonus  float64
	RecencyBonus float64
	
	IsLowConfidence bool

	// Lifecycle fields
	LifecycleStatus string
	StatusReason    string
	IsValidated     bool
	Author          string
	AuthorEmail     string
	Agent           string
	Branch          string
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
func RenderResultsWithContext(query string, results []SearchResultItem, related []string, neighbors []string, width int, strategy string, explain bool) string {
	var sections []string

	// 1. Header
	headerText := fmt.Sprintf("🔍 Search: %q", query)
	if strategy == "fallback" {
		headerText += " (Fallback Mode)"
	}
	sections = append(sections, TableHeaderStyle.Render(headerText))
	sections = append(sections, "") // Spacer

	// 2. Context Tables (only if not fallback or if they have items)
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
			// ID column
			idCol := fmt.Sprintf("%d. [%s]", i+1, r.ID)
			if r.Date != "" {
				idCol += "\n" + lipgloss.NewStyle().Foreground(ColorMuted).Render(r.Date)
			}
			
			// Content column (Title + Metadata + Narrative + Explain)
			title := r.Title
			if r.LifecycleStatus == "paused" {
				title = lipgloss.NewStyle().Foreground(ColorWarn).Render("[⏸ PAUSED]") + " " + title
			} else if r.LifecycleStatus == "abandoned" {
				title = lipgloss.NewStyle().Foreground(ColorFail).Render("[🚫 ABANDONED]") + " " + title
			} else if r.LifecycleStatus == "ongoing" {
				title = lipgloss.NewStyle().Foreground(ColorMuted).Render("[🔄 ONGOING]") + " " + title
			} else if r.IsValidated {
				title = lipgloss.NewStyle().Foreground(ColorPass).Render("[🟢 VALIDATED]") + " " + title
			}

			var contentLines []string

			if explain {
				displayScore := -r.Score 
				contentLines = append(contentLines, fmt.Sprintf("%s (Relevance: %.2f)", title, displayScore))
			} else {
				contentLines = append(contentLines, title)
			}

			// Add Metadata (Author, Agent, Branch)
			metaParts := []string{fmt.Sprintf("   👤 %s", lipgloss.NewStyle().Foreground(ColorAccent).Render(r.Author))}
			if r.Agent != "" && r.Agent != "Unknown" {
				metaParts = append(metaParts, fmt.Sprintf("🤖 %s", lipgloss.NewStyle().Foreground(ColorPass).Render(r.Agent)))
			}
			if r.Branch != "" && r.Branch != "N/A" {
				metaParts = append(metaParts, fmt.Sprintf(" %s", lipgloss.NewStyle().Foreground(ColorMuted).Render(r.Branch)))
			}
			contentLines = append(contentLines, strings.Join(metaParts, "  "))

			if r.StatusReason != "" {
				reasonStyle := lipgloss.NewStyle().Foreground(ColorWarn).Italic(true)
				if r.LifecycleStatus == "abandoned" {
					reasonStyle = lipgloss.NewStyle().Foreground(ColorFail).Italic(true)
				}
				contentLines = append(contentLines, reasonStyle.Render("   Reason: "+r.StatusReason))
			}
			
			if r.Narrative != "" {
				// For snippets (like FTS <b> highlights), we keep them as is but ensure they are indented
				// and don't break the table.
				narrative := r.Narrative
				
				// Highlighted snippets: convert <b> to lipgloss bold/color
				// We do this surgically to avoid messing up the rest of the string
				if strings.Contains(narrative, "<b>") {
					narrative = strings.ReplaceAll(narrative, "<b>", lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(""))
					narrative = strings.ReplaceAll(narrative, "</b>", lipgloss.NewStyle().Bold(false).Foreground(lipgloss.NoColor{}).Render(""))
				}

				// Handle multi-line previews by indenting each line
				lines := strings.Split(narrative, "\n")
				lineCount := 0
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line == "" {
						continue
					}
					// Truncate line if too long for table
					maxLineLen := width - 18
					if len(line) > maxLineLen {
						line = line[:maxLineLen-3] + "..."
					}
					contentLines = append(contentLines, "   "+line)
					lineCount++
					if lineCount >= 3 { // Limit to 3 lines of snippet for cleaner list
						break
					}
				}
			}
			
			if explain {
				displayBM25 := -r.BM25
				explainStr := fmt.Sprintf("   Base: %.2f | Phrase: +%.1f | Near: +%.1f | Recency: +%.1f | Entity: +%.1f", 
					displayBM25, r.PhraseBonus, r.NearBonus, r.RecencyBonus, r.EntityBonus)
				contentLines = append(contentLines, lipgloss.NewStyle().Foreground(ColorMuted).Render(explainStr))
			}
			
			rows = append(rows, []string{idCol, strings.Join(contentLines, "\n")})
		}

		titleText := fmt.Sprintf("📄 Found %d sessions", len(results))
		if strategy == "fallback" {
			titleText += " (Low Confidence Matches)"
		}

		t := table.New().
			Headers(titleText, "").
			Border(lipgloss.RoundedBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(ColorMuted)).
			Width(width).
			Rows(rows...).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return TableHeaderStyle
				}
				
				style := lipgloss.NewStyle().Padding(0, 1).Align(lipgloss.Left)
				if col == 0 {
					style = style.Width(15).Foreground(ColorAccent)
				} else {
					style = style.Width(width - 17)
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
				break
			}
			idCol := fmt.Sprintf("%d. [%s]", i+1, r.ID)
			rows = append(rows, []string{idCol, r.Title})
		}

		t := table.New().
			Headers(fmt.Sprintf("📄 Found %d sessions", len(results)), "").
			Border(lipgloss.RoundedBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(ColorMuted)).
			Width(width).
			Rows(rows...).
			StyleFunc(func(row, col int) lipgloss.Style {
				if row == table.HeaderRow {
					return TableHeaderStyle
				}
				style := lipgloss.NewStyle().Padding(0, 1).Align(lipgloss.Left)
				if col == 0 {
					style = style.Width(15).Foreground(ColorAccent)
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

// RenderEntitiesTable renders a list of entities and their counts
func RenderEntitiesTable(title string, entities [][]string, width int) string {
	return table.New().
		Headers(title, "Count").
		Rows(entities...).
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(ColorMuted)).
		Width(width).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return TableHeaderStyle.Width(width/2 - 1)
			}
			style := lipgloss.NewStyle().Padding(0, 1)
			if col == 1 {
				style = style.Align(lipgloss.Right).Foreground(ColorMuted)
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
