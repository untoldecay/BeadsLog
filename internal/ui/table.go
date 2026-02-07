package ui

import (
	"fmt"

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
	Methodology      string
	Tools            string
	TokenCost        string
	PrimaryFocus     string
	EfficiencyRatio  string
	BestUsedFor      string
	CriticalMissing  string
	Pass            string
	Score           int
}

func RenderEvalReport(runs []EvalRun, width int) string {
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(TableBorderStyle).
		Width(width)

	headers := []string{"Attribute"}
	for i, r := range runs {
		// Truncate long session IDs for display
		displayID := r.ID
		if len(displayID) > 12 {
			displayID = displayID[:12]
		}
		headers = append(headers, fmt.Sprintf("Run %d\n(%s)", i+1, displayID))
	}
	t.Headers(headers...)

	// Define attributes to show as rows
	attrs := []string{
		"Methodology",
		"Tools",
		"Token Cost",
		"Primary Focus",
		"Efficiency Ratio",
		"Best Used For",
		"Critical Missing",
		"Status",
	}

	var rows [][]string
	for _, attr := range attrs {
		row := []string{attr}
		for _, run := range runs {
			val := ""
			switch attr {
			case "Methodology":
				val = run.Methodology
			case "Tools":
				val = run.Tools
			case "Token Cost":
				val = run.TokenCost
			case "Primary Focus":
				val = run.PrimaryFocus
			case "Efficiency Ratio":
				val = run.EfficiencyRatio
			case "Best Used For":
				val = run.BestUsedFor
			case "Critical Missing":
				val = run.CriticalMissing
			case "Status":
				val = fmt.Sprintf("%s (%d/100)", run.Pass, run.Score)
			}
			row = append(row, val)
		}
		rows = append(rows, row)
	}

	t.Rows(rows...)

	// Styling
	t.StyleFunc(func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)
		
		if row == 0 { // Headers
			return TableHeaderStyle.Height(2)
		}
		
		if col == 0 { // First column (Labels)
			return style.Bold(true).Foreground(ColorAccent)
		}

		// Background for data cells
		return style.Background(lipgloss.Color("#1a1a1a"))
	})

	return t.Render()
}
