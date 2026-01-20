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
