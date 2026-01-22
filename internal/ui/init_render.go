package ui

import (
	"fmt"
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
	DevlogHooks         []string
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



	// 2. Component Table (Summary)

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



	// 3. Hierarchical Progress List

	// Helper for checkmark list

	checkList := func() *list.List {

		return list.New().

			Enumerator(func(_ list.Items, i int) string {

				return RenderPass("✓")

			}).

			EnumeratorStyle(lipgloss.NewStyle().MarginRight(1))

	}



	// 3a. Orchestration

	lOrch := checkList()

	lOrch.Item("Orchestration space: _rules/_orchestration")

	orchFiles := list.New().Enumerator(func(_ list.Items, i int) string { return RenderPass("✓") }).EnumeratorStyle(lipgloss.NewStyle().MarginRight(1))

	for _, f := range res.OrchestrationFiles {

		orchFiles.Item(f)

	}

	lOrch.Item(orchFiles)

	sections = append(sections, lOrch.String())



	// 3b. Agent Rules

	if len(res.AgentRules) > 0 {

		lAgent := checkList()

		lAgent.Item("Agent instructions: " + strings.Join(res.AgentRules, ", "))

		sections = append(sections, lAgent.String())

	}



	// 3c. Devlog

	lDev := checkList()

	lDev.Item(strings.TrimSpace("Devlog space: _rules/_devlog " + res.DevlogSpaceStatus))

	promptList := list.New().Enumerator(func(_ list.Items, i int) string { return RenderPass("✓") }).EnumeratorStyle(lipgloss.NewStyle().MarginRight(1))

	promptList.Item(strings.TrimSpace("_generate-devlog.md " + res.DevlogPromptStatus))

	lDev.Item(promptList)

	sections = append(sections, lDev.String())



	// 3d. Git Hooks

	allHooks := []string{}

	if res.HooksInstalled {

		allHooks = append(allHooks, "pre-commit", "prepare-commit-msg", "pre-push", "post-checkout")

	}

	allHooks = append(allHooks, res.DevlogHooks...)

	if len(allHooks) > 0 {

		lHooks := checkList()

		lHooks.Item("Git hooks installed: " + strings.Join(allHooks, ", "))

		sections = append(sections, lHooks.String())

	}



	sections = append(sections, "") // Spacer



	// 4. Setup Completion Table (Diagnostics)

	if len(res.DoctorIssues) > 0 {

		warnRows := [][]string{}

		for _, issue := range res.DoctorIssues {

			warnRows = append(warnRows, []string{"⚠", issue})

		}



		diagTable := table.New().

			Headers("!", "Setup Completion / Warnings").

			Rows(warnRows...).

			Border(lipgloss.RoundedBorder()).

			BorderStyle(lipgloss.NewStyle().Foreground(ColorWarn)).

			Width(width).

			StyleFunc(func(row, col int) lipgloss.Style {

				if row == table.HeaderRow {

					if col == 0 {

						return TableHeaderStyle.Width(3).Foreground(ColorWarn)

					}

					return TableHeaderStyle.Width(width - 3 - 3).Foreground(ColorWarn)

				}

				style := lipgloss.NewStyle().Padding(0, 1).Align(lipgloss.Left)

				if col == 0 {

					style = style.Foreground(ColorWarn).Bold(true)

				}

				return style

			})

		sections = append(sections, diagTable.String(), "")

	}



	// 5. Help (Quickfix)

	sections = append(sections, lipgloss.NewStyle().Bold(true).Render("Help & Diagnostics:"))

	sections = append(sections, "  • Run "+lipgloss.NewStyle().Foreground(ColorAccent).Render("bd doctor --fix")+" to resolve setup warnings.")

	sections = append(sections, "")



	// 6. Final Message

	nextStep := lipgloss.NewStyle().Foreground(ColorAccent).Bold(true).Render("bd onboard")

	finalMsg := fmt.Sprintf("Ready to roll! 🚀 Start your session by running %s to prime your agent.", nextStep)

	sections = append(sections, finalMsg)



	return lipgloss.JoinVertical(lipgloss.Left, sections...)

}
