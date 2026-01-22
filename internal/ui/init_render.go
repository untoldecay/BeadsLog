package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/charmbracelet/lipgloss/table"
)

// InitResult aggregates all information from the initialization process
type InitResult struct {
	// Database info
	DBPath string
	Prefix string

	// Step results
	OrchestrationFiles []string
	DevlogSpaceStatus   string
	DevlogPromptStatus  string
	AgentRules          []string
	HooksInstalled      bool
	MergeDriverInstalled bool

	// Diagnostics
	DoctorIssues []string

	// Next steps
	QuickstartCommands []string
}

// RenderInitReport generates a professional Lipgloss report for the init command
func RenderInitReport(res InitResult, width int) string {
	var sections []string

	// 1. Success Header (Minimal)
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPass).
		Render("✓ bd Initialized Successfully")
	sections = append(sections, header, "")

	// 2. Hierarchical Progress List (using lipgloss/list)
	// Outer list uses checkmarks
	l := list.New().
		Enumerator(func(_ list.Items, i int) string {
			return RenderPass("✓")
		}).
		EnumeratorStyle(lipgloss.NewStyle().MarginRight(1))

	// Orchestration
	orchList := list.New().Enumerator(func(_ list.Items, i int) string {
		return RenderPass("✓")
	}).EnumeratorStyle(lipgloss.NewStyle().MarginRight(1))
	
	for _, f := range res.OrchestrationFiles {
		orchList.Item(f)
	}
	l.Item("Orchestration space: _rules/_orchestration")
	l.Item(orchList)

	// Devlog
	l.Item(strings.TrimSpace("Devlog space: _rules/_devlog " + res.DevlogSpaceStatus))
	l.Item(strings.TrimSpace("Devlog prompt: _rules/_devlog/_generate-devlog.md " + res.DevlogPromptStatus))

	// Agent Rules
	if len(res.AgentRules) > 0 {
		l.Item("Agent instruction: " + strings.Join(res.AgentRules, ", "))
	}

	sections = append(sections, l.String(), "")

	// 3. Setup Details Table (Summary)
	detailsRows := [][]string{
		{"Database", res.DBPath},
		{"Issue Prefix", res.Prefix},
		{"Next IDs", res.Prefix + "-<hash>"},
	}

	summaryTable := table.New().
		Headers("Component", "Configuration").
		Rows(detailsRows...).
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(ColorMuted)).
		Width(width).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				if col == 0 {
					return TableHeaderStyle.Width(20)
				}
				return TableHeaderStyle.Width(width - 20 - 3)
			}
			style := lipgloss.NewStyle().Padding(0, 1).Align(lipgloss.Left)
			if col == 0 {
				style = style.Bold(true).Foreground(ColorAccent)
			}
			return style
		})
	
	sections = append(sections, summaryTable.String(), "")

	// 4. Automation Status
	autoStatus := []string{}
	if res.HooksInstalled {
		autoStatus = append(autoStatus, RenderPass("✓")+" Git Hooks Installed")
	}
	if res.MergeDriverInstalled {
		autoStatus = append(autoStatus, RenderPass("✓")+" Merge Driver Configured")
	}
	
	if len(autoStatus) > 0 {
		sections = append(sections, strings.Join(autoStatus, "  "), "")
	}

	// 5. Diagnostics (Doctor Issues)
	if len(res.DoctorIssues) > 0 {
		warnBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorWarn).
			Padding(0, 1).
			Width(width - 2)

		var warnContent []string
		warnContent = append(warnContent, lipgloss.NewStyle().Bold(true).Foreground(ColorWarn).Render("⚠ Setup Incomplete / Warnings:"))
		for _, issue := range res.DoctorIssues {
			warnContent = append(warnContent, "  • "+issue)
		}
		warnContent = append(warnContent, "", "Run "+lipgloss.NewStyle().Foreground(ColorAccent).Render("bd doctor --fix")+" to resolve.")
		
		sections = append(sections, warnBox.Render(strings.Join(warnContent, "\n")),"")
	}

	// 6. Next Steps
	if len(res.QuickstartCommands) > 0 {
		sections = append(sections, lipgloss.NewStyle().Bold(true).Render("Next Steps:"))
		for _, cmd := range res.QuickstartCommands {
			sections = append(sections, "  • "+lipgloss.NewStyle().Foreground(ColorAccent).Render(cmd))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}