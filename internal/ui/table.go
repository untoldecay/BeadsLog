package ui

import (
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

// EvalColumn represents a column in the comparative eval report
type EvalColumn struct {
	Title            string
	Methodology      string
	Tools            string
	TokenCost        string
	Context          string
	PrimaryFocus     string
	EfficiencyRatio  string
	BestUsedFor      string
	CriticalMissing  string
}

func RenderEvalReport(cols []EvalColumn, width int) string {
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(TableBorderStyle).
		Width(width)

	headers := []string{"", "1 - Only bd commands", "2 - Only brute force", "3 - Hybrid"}
	// Ensure we have exactly 3 columns of data even if empty
	displayCols := make([]EvalColumn, 3)
	for i := range displayCols {
		if i < len(cols) {
			displayCols[i] = cols[i]
		}
	}

	t.Headers(headers...)

	// Define rows
	rows := [][]string{
		{"Methodology", displayCols[0].Methodology, displayCols[1].Methodology, displayCols[2].Methodology},
		{"Tools", displayCols[0].Tools, displayCols[1].Tools, displayCols[2].Tools},
		{"Token Cost", displayCols[0].TokenCost, displayCols[1].TokenCost, displayCols[2].TokenCost},
		{"Primary Focus", displayCols[0].PrimaryFocus, displayCols[1].PrimaryFocus, displayCols[2].PrimaryFocus},
		{"Efficiency Ratio", displayCols[0].EfficiencyRatio, displayCols[1].EfficiencyRatio, displayCols[2].EfficiencyRatio},
		{"Best Used For", displayCols[0].BestUsedFor, displayCols[1].BestUsedFor, displayCols[2].BestUsedFor},
		{"Critical Missing", displayCols[0].CriticalMissing, displayCols[1].CriticalMissing, displayCols[2].CriticalMissing},
	}

	t.Rows(rows...)

	// Styling
	t.StyleFunc(func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)
		
		if row == 0 { // Headers
			return TableHeaderStyle
		}
		
		if col == 0 { // First column (Labels)
			return style.Bold(true).Foreground(ColorAccent)
		}

		// Plain background for the whole table content to make it stand out
		return style.Background(lipgloss.Color("#1a1a1a"))
	})

	return t.Render()
}
