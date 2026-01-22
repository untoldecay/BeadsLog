package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
)

// InitResult aggregates all information from the initialization process

type InitResult struct {

	// Database info

	DBPath  string

	Prefix  string

	RepoID  string

	CloneID string



	// Step results

	OrchestrationFiles   []string

	DevlogSpaceStatus    string

	DevlogPromptStatus   string

	AgentRules           []string

	DevlogHooks          []string

	HooksInstalled       bool

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



	// 2. Hierarchical Progress List

	// Helper for checkmark list

	checkList := func() *list.List {

		return list.New().

			Enumerator(func(_ list.Items, i int) string {

				return RenderPass("✓")

			}).

			EnumeratorStyle(lipgloss.NewStyle().MarginRight(1))

	}



	// 2a. Configuration List (Replacing Table)

	lConfig := checkList()

	lConfig.Item("Configuration:")

	configSubList := list.New().Enumerator(func(_ list.Items, i int) string { return RenderPass("✓") }).EnumeratorStyle(lipgloss.NewStyle().MarginRight(1))

	configSubList.Item(fmt.Sprintf("Database: %s", res.DBPath))

	configSubList.Item(fmt.Sprintf("Issue Prefix: %s", res.Prefix))

	configSubList.Item(fmt.Sprintf("Next IDs: %s-<hash>", res.Prefix))

	if res.RepoID != "" {

		configSubList.Item(fmt.Sprintf("Repository ID: %s", res.RepoID[:8]))

	}

	if res.CloneID != "" {

		configSubList.Item(fmt.Sprintf("Clone ID: %s", res.CloneID))

	}

	lConfig.Item(configSubList)

	sections = append(sections, lConfig.String())



	// 2b. Orchestration

	lOrch := checkList()

	lOrch.Item("Orchestration space: _rules/_orchestration")

	orchFiles := list.New().Enumerator(func(_ list.Items, i int) string { return RenderPass("✓") }).EnumeratorStyle(lipgloss.NewStyle().MarginRight(1))

	for _, f := range res.OrchestrationFiles {

		orchFiles.Item(f)

	}

	lOrch.Item(orchFiles)

	sections = append(sections, lOrch.String())



	// 2c. Agent Rules

	if len(res.AgentRules) > 0 {

		lAgent := checkList()

		lAgent.Item("Agent instructions: " + strings.Join(res.AgentRules, ", "))

		sections = append(sections, lAgent.String())

	}



	// 2d. Devlog

	lDev := checkList()

	lDev.Item(strings.TrimSpace("Devlog space: _rules/_devlog " + res.DevlogSpaceStatus))

	promptList := list.New().Enumerator(func(_ list.Items, i int) string { return RenderPass("✓") }).EnumeratorStyle(lipgloss.NewStyle().MarginRight(1))

	promptList.Item(strings.TrimSpace("_generate-devlog.md " + res.DevlogPromptStatus))

	lDev.Item(promptList)

	sections = append(sections, lDev.String())



	// 2e. Git Hooks

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



	// 3. Setup Completion / Warnings (Background Color)

	if len(res.DoctorIssues) > 0 {

		// Use a high-visibility background for diagnostics

		warnStyle := lipgloss.NewStyle().

			Background(lipgloss.Color("#2a2a2a")).

			Padding(1, 2).

			Width(width)



		var warnLines []string

		warnLines = append(warnLines, lipgloss.NewStyle().Bold(true).Foreground(ColorWarn).Render("⚠ SETUP INCOMPLETE / WARNINGS"))

		

		// Use a list inside the warning area

		diagList := list.New().Enumerator(func(_ list.Items, i int) string {

			return lipgloss.NewStyle().Foreground(ColorWarn).Render("•")

		}).EnumeratorStyle(lipgloss.NewStyle().MarginRight(1))

		

		for _, issue := range res.DoctorIssues {

			diagList.Item(issue)

		}

		warnLines = append(warnLines, diagList.String())

		

		// Embed doctor command at the end

		doctorCmd := lipgloss.NewStyle().Foreground(ColorAccent).Bold(true).Render("bd doctor --fix")

		warnLines = append(warnLines, "", "Run "+doctorCmd+" to resolve these issues.")



		sections = append(sections, warnStyle.Render(strings.Join(warnLines, "\n")), "")

	}



	// 4. Help (Quickfix) - Only if no warnings (to avoid redundancy)

	if len(res.DoctorIssues) == 0 {

		sections = append(sections, lipgloss.NewStyle().Bold(true).Render("Help & Diagnostics:"))

		sections = append(sections, "  • Run "+lipgloss.NewStyle().Foreground(ColorAccent).Render("bd doctor")+" for system health check.")

		sections = append(sections, "")

	}



	// 5. Final Message

	nextStep := lipgloss.NewStyle().Foreground(ColorAccent).Bold(true).Render("bd onboard")

	finalMsg := fmt.Sprintf("Ready to roll! 🚀 Start your session by running %s to prime your agent.", nextStep)

	sections = append(sections, finalMsg)



	return lipgloss.JoinVertical(lipgloss.Left, sections...)

}


