package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// Table Styles
var (
	TableHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorAccent).
		Align(lipgloss.Center)

	TableWarningStyle = lipgloss.NewStyle().
		Foreground(ColorWarn)

	TableSuccessStyle = lipgloss.NewStyle().
		Foreground(ColorPass)

	TableHintStyle = lipgloss.NewStyle().
		Foreground(ColorMuted)

	TableBorderStyle = lipgloss.NewStyle().
		Foreground(ColorMuted)
)

// NewSearchTable creates a new table with default search styling
func NewSearchTable(width int) *table.Table {
	return table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(TableBorderStyle).
		Width(width)
}

// EvalRun represents a single session's data for the report
type EvalRun struct {
	ID              string
	Strategy        string
	Tools           string
	TokenCost       string
	TimeSpent       string
	Pass            string
	Score           int
	Efficiency      string // 🔴, 🟠, 🟢
}

func RenderEvalReport(runs []EvalRun, width int) string {
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderRow(true).
		BorderStyle(TableBorderStyle).
		Width(width)

	headers := []string{"Run", "Strategy", "Tools Used", "Tokens", "Time", "Status", "Eff."}
	t.Headers(headers...)

	var rows [][]string
	for i, r := range runs {
		// Use the last part of the ID for display to ensure uniqueness
		displayID := r.ID
		if parts := strings.Split(r.ID, "-"); len(parts) >= 2 {
			// Show last 2 parts, e.g. 1770485816-28014
			displayID = parts[len(parts)-2] + "-" + parts[len(parts)-1]
		}
		if len(displayID) > 15 {
			// If still too long, take the last 15
			displayID = displayID[len(displayID)-15:]
		}
		
		rows = append(rows, []string{
			fmt.Sprintf("%d\n(%s)", i+1, displayID),
			r.Strategy,
			r.Tools,
			r.TokenCost,
			r.TimeSpent,
			fmt.Sprintf("%s\n(%d/100)", r.Pass, r.Score),
			r.Efficiency,
		})
	}

	t.Rows(rows...)

	// Styling
	t.StyleFunc(func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)
		
		if row == 0 { // Headers
			return TableHeaderStyle
		}
		
		// Status coloring
		// row is 1-based when including headers in the style func context
		dataRow := row - 1
		if dataRow >= 0 && dataRow < len(rows) {
			if col == 5 { // Status column
				if strings.Contains(rows[dataRow][5], "PASS") {
					style = style.Foreground(ColorPass)
				} else {
					style = style.Foreground(ColorWarn)
				}
			}
		}

		// Plain background for the whole table content to make it stand out
		return style.Background(lipgloss.Color("#1a1a1a"))
	})

	return t.Render()
}
