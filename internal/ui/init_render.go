package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

const InitLogo = `
▛▀▖        ▌   ▜       
▙▄▘▞▀▖▝▀▖▞▀▌▞▀▘▐ ▞▀▖▞▀▌
▌ ▌▛▀ ▞▀▌▌ ▌▝▀▖▐ ▌ ▌▚▄▌
▀▀ ▝▀▘▝▀▘▝▀▘▀▀  ▘▝▀ ▗▄▘
`

// RenderInitLogo returns the stylized ASCII logo for initialization
func RenderInitLogo() string {
	lines := strings.Split(strings.TrimSpace(InitLogo), "\n")
	var styledLines []string
	for _, line := range lines {
		styledLines = append(styledLines, RenderAccent(line))
	}
	return strings.Join(styledLines, "\n") + "\n"
}

// RenderInitReport generates a professional Lipgloss report for the init command
func RenderInitReport(res InitResult, width int) string {
	var sections []string

	// 1. Success Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPass).
		Render("✓ bd Initialized Successfully")
	sections = append(sections, header)

	// 2. Summary Block
	var summary strings.Builder
	summary.WriteString(RenderPass("✓") + " Configuration:\n")
	summary.WriteString(fmt.Sprintf("  %s Database: %s\n", RenderPass("✓"), res.DBPath))
	summary.WriteString(fmt.Sprintf("  %s Issue Prefix: %s\n", RenderPass("✓"), res.Prefix))
	if res.RepoID != "" {
		summary.WriteString(fmt.Sprintf("  %s Repository ID: %s\n", RenderPass("✓"), res.RepoID[:8]))
	}
	if res.CloneID != "" {
		summary.WriteString(fmt.Sprintf("  %s Clone ID: %s", RenderPass("✓"), res.CloneID))
	}

	allHooks := []string{}
	if res.HooksInstalled {
		allHooks = append(allHooks, "pre-commit", "prepare-commit-msg", "pre-push", "post-checkout")
	}
	allHooks = append(allHooks, res.DevlogHooks...)
	if len(allHooks) > 0 {
		summary.WriteString(fmt.Sprintf("\n%s Git hooks: %s", RenderPass("✓"), strings.Join(allHooks, ", ")))
	}
	
	sections = append(sections, summary.String())

	// 3. Setup Completion / Warnings
	if len(res.DoctorIssues) > 0 {
		var warnLines []string
		warnLines = append(warnLines, lipgloss.NewStyle().Bold(true).Foreground(ColorWarn).Render("⚠ SETUP INCOMPLETE / WARNINGS"))
		
		for _, issue := range res.DoctorIssues {
			warnLines = append(warnLines, lipgloss.NewStyle().Foreground(ColorWarn).Render("• ")+issue)
		}
		
		doctorCmd := lipgloss.NewStyle().Foreground(ColorAccent).Bold(true).Render("bd doctor --fix")
		warnLines = append(warnLines, "", "Run "+doctorCmd+" to resolve these issues.")

		warnBlock := lipgloss.NewStyle().
			Background(lipgloss.Color("#141414")).
			Padding(0, 1).
			Width(width).
			Render(strings.Join(warnLines, "\n"))
		
		sections = append(sections, "", warnBlock)
	}

	// 4. Final Message
	nextStep := lipgloss.NewStyle().Foreground(ColorAccent).Bold(true).Render("onboard")
	finalMsg := fmt.Sprintf("\nReady✨. Start your coding agent and initiate chat by saying : %s", nextStep)
	sections = append(sections, finalMsg)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}