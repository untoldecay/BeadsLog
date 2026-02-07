package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/untoldecay/BeadsLog/internal/audit"
	"github.com/untoldecay/BeadsLog/internal/config"
	"github.com/untoldecay/BeadsLog/internal/ui"
)

var evalCmd = &cobra.Command{
	Use:   "eval",
	Short: "Agent evaluation and performance reporting",
	Long: `Manage agent evaluation sessions and generate performance reports.
Eval mode attaches unique Session IDs to command logs for comparative analysis.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var evalStartCmd = &cobra.Command{
	Use:   "start [scenario_id]",
	Short: "Enable evaluation mode and start a new session",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scenarioID := ""
		if len(args) > 0 {
			scenarioID = args[0]
		}

		// Generate a simple session ID (timestamp + random)
		sessionID := fmt.Sprintf("eval-%d-%d", time.Now().Unix(), os.Getpid())
		
		_ = config.SetYamlConfig("eval.mode", "true")
		_ = config.SetYamlConfig("eval.session_id", sessionID)
		if scenarioID != "" {
			_ = config.SetYamlConfig("eval.scenario_id", scenarioID)
		}

		fmt.Printf("\n%s %s\n", ui.RenderPass("📊"), ui.RenderBold("Evaluation Mode Enabled"))
		fmt.Printf("Session ID: %s\n", ui.RenderAccent(sessionID))
		if scenarioID != "" {
			fmt.Printf("Scenario:   %s\n", ui.RenderAccent(scenarioID))
		}
		
		fmt.Println("\nHow to test:")
		fmt.Println("1. Run your agent task normally.")
		fmt.Println("2. When done, run 'bd eval report' to see performance.")
		fmt.Println("3. Use 'bd eval next' to start a fresh comparison run.")
		fmt.Println("4. Run 'bd eval stop' to finish and clean up.")
	},
}

var evalNextCmd = &cobra.Command{
	Use:   "next [scenario_id]",
	Short: "Rotate session ID for a fresh comparison run",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !config.GetBool("eval.mode") {
			fmt.Println("Error: Eval mode is not active. Run 'bd eval start' first.")
			os.Exit(1)
		}

		scenarioID := config.GetString("eval.scenario_id")
		if len(args) > 0 {
			scenarioID = args[0]
		}

		// Rotate ID
		sessionID := fmt.Sprintf("eval-%d-%d", time.Now().Unix(), os.Getpid())
		_ = config.SetYamlConfig("eval.session_id", sessionID)
		if scenarioID != "" {
			_ = config.SetYamlConfig("eval.scenario_id", scenarioID)
		}

		fmt.Printf("%s Started new eval session: %s\n", ui.RenderPass("✓"), ui.RenderAccent(sessionID))
	},
}

var evalStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Disable evaluation mode and clean up",
	Run: func(cmd *cobra.Command, args []string) {
		_ = config.SetYamlConfig("eval.mode", "false")
		_ = config.SetYamlConfig("eval.session_id", "")
		_ = config.SetYamlConfig("eval.scenario_id", "")

		fmt.Printf("%s Evaluation mode disabled. Traces preserved in interactions.jsonl.\n", ui.RenderPass("✓"))
	},
}

var evalReportCmd = &cobra.Command{
	Use:   "report [N]",
	Short: "Generate comparative performance report for the last N sessions",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		n := 1
		if len(args) > 0 {
			val, err := strconv.Atoi(args[0])
			if err == nil && val > 0 {
				n = val
			}
		}

		fmt.Printf("Analyzing the last %d eval sessions...\n", n)
		
		sessions, err := extractEvalSessions(n)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error extracting sessions: %v\n", err)
			os.Exit(1)
		}

		if len(sessions) == 0 {
			fmt.Println("No eval sessions found in audit log.")
			return
		}

		// Group sessions by prompt similarity
		groups := clusterSessions(sessions)

		for _, g := range groups {
			fmt.Printf("\nTASK GROUP: %q (Matched %d runs)\n", g.Name, len(g.Sessions))
			
			// Calculate group average tokens for baseline
			totalTokens := 0
			for _, s := range g.Sessions {
				totalTokens += estimateCostSummary(s.Tools)
			}
			avgTokens := float64(totalTokens) / float64(len(g.Sessions))
			if avgTokens == 0 { avgTokens = 1 } // Prevent div by zero

			// Prepare runs for the report
			runs := make([]ui.EvalRun, len(g.Sessions))
			for i, s := range g.Sessions {
				strategy, weight, emoji := detectRefinedStrategy(s.Tools)
				cost := estimateCostSummary(s.Tools)
				
				duration := "0s"
				if s.EndTime.After(s.StartTime) {
					duration = s.EndTime.Sub(s.StartTime).Round(time.Second).String()
				}

				// Calculate Weighted Efficiency
				// Efficiency = (Strategy Weight) * (Baseline Cost / Actual Cost)
				costFactor := avgTokens / float64(cost)
				if cost == 0 { costFactor = 0 } // Avoid infinity if cost is 0
				efficiencyScore := weight * costFactor
				
				effStr := fmt.Sprintf("%.1fx %s", efficiencyScore, emoji)

				run := ui.EvalRun{
					ID:        s.ID,
					Strategy:  strategy,
					Tools:     strings.Join(s.Tools, "\n"),
					TokenCost: fmt.Sprintf("%d", cost),
					TimeSpent: duration,
					Efficiency: effStr,
				}

				// Map strategy to qualitative attributes (simplified for row layout)
				if strings.Contains(strategy, "Optimal") {
					run.Pass = "PASS"
					run.Score = 100
				} else if strings.Contains(strategy, "Shallow") {
					run.Pass = "PASS"
					run.Score = 85
				} else if strings.Contains(strategy, "Disordered") {
					run.Pass = "FAIL"
					run.Score = 50
				} else {
					run.Pass = "FAIL"
					run.Score = 0
				}

				runs[i] = run
			}
			fmt.Println(ui.RenderEvalReport(runs, ui.GetWidth()))
		}
		fmt.Println()
	},
}

func detectRefinedStrategy(tools []string) (string, float64, string) {
	if len(tools) == 0 {
		return "Unknown", 0.0, "🔴"
	}

	firstTool := strings.ToLower(tools[0])
	hasDevlogFirst := strings.Contains(firstTool, "devlog") || strings.Contains(firstTool, "onboard")
	
	hasVerify := false
	hasMap := false
	mapIndex := -1
	verifyIndex := -1

	for i, t := range tools {
		t = strings.ToLower(t)
		if strings.Contains(t, "devlog") {
			hasMap = true
			if mapIndex == -1 { mapIndex = i }
		}
		if strings.Contains(t, "read_file") || strings.Contains(t, "grep") || strings.Contains(t, "cat") {
			hasVerify = true
			if verifyIndex == -1 { verifyIndex = i }
		}
	}

	if hasDevlogFirst {
		if hasVerify {
			return "Optimal (Mastery)", 1.5, "🟢🟢🟢"
		}
		return "Shallow", 0.8, "🟠"
	}

	if hasMap {
		// Mapped, but not first (Disordered)
		return "Disordered", 0.5, "🔴🟠"
	}

	return "Blind", 0.1, "🔴"
}

func estimateCostSummary(trace []string) int {
	total := 0
	for _, cmd := range trace {
		cmd = strings.ToLower(cmd)
		if strings.Contains(cmd, "grep") || strings.Contains(cmd, "search_file_content") {
			total += 2500 
		} else if strings.Contains(cmd, "read_file") {
			total += 1200 
		} else if strings.Contains(cmd, "devlog") {
			total += 150 
		} else {
			total += 300 
		}
	}
	return total
}

type evalSession struct {
	ID         string
	ScenarioID string
	Tools      []string
	StartTime  time.Time
	EndTime    time.Time
}

func extractEvalSessions(n int) ([]evalSession, error) {
	p, err := audit.Path()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(p); os.IsNotExist(err) {
		return nil, nil
	}

	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sessionMap := make(map[string]*evalSession)
	var sessionOrder []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry audit.Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		if entry.EvalSessionID == "" {
			continue
		}

		s, ok := sessionMap[entry.EvalSessionID]
		if !ok {
			s = &evalSession{
				ID:        entry.EvalSessionID,
				StartTime: entry.CreatedAt,
			}
			sessionMap[entry.EvalSessionID] = s
			sessionOrder = append(sessionOrder, entry.EvalSessionID)
		}

		s.EndTime = entry.CreatedAt

		if entry.Kind == "tool_call" {
			s.Tools = append(s.Tools, entry.ToolName)
		}
	}

	if len(sessionOrder) > n {
		sessionOrder = sessionOrder[len(sessionOrder)-n:]
	}

	var results []evalSession
	for _, id := range sessionOrder {
		results = append(results, *sessionMap[id])
	}

	return results, nil
}

type EvalGroup struct {
	Name     string
	Sessions []evalSession
}

func clusterSessions(sessions []evalSession) []EvalGroup {
	var groups []EvalGroup

	for _, s := range sessions {
		// 1. Try to find existing group by ID match
		found := false
		for i, g := range groups {
			// If session has specific ScenarioID, match exactly
			if s.ScenarioID != "" && g.Name == s.ScenarioID {
				groups[i].Sessions = append(groups[i].Sessions, s)
				found = true
				break
			}
			
			// Fuzzy match on prompt if available (assuming ScenarioID might be the prompt)
			// If ScenarioID is "A1", we don't fuzzy match "A1".
			// But if user did `bd eval start "Fix auth bug"`, then ScenarioID is the prompt.
			// Let's assume ScenarioID holds the prompt/intent.
			
			if s.ScenarioID != "" && len(s.ScenarioID) > 5 {
				dist := Levenshtein(strings.ToLower(g.Name), strings.ToLower(s.ScenarioID))
				// Similarity threshold: distance < 30% of length
				threshold := len(s.ScenarioID) / 3
				if dist <= threshold {
					groups[i].Sessions = append(groups[i].Sessions, s)
					found = true
					break
				}
			}
		}

		if !found {
			name := s.ScenarioID
			if name == "" {
				name = "Untitled Session"
			}
			groups = append(groups, EvalGroup{
				Name:     name,
				Sessions: []evalSession{s},
			})
		}
	}
	return groups
}

func init() {
	evalCmd.AddCommand(evalStartCmd)
	evalCmd.AddCommand(evalNextCmd)
	evalCmd.AddCommand(evalReportCmd)
	evalCmd.AddCommand(evalStopCmd)
	rootCmd.AddCommand(evalCmd)
}

// Levenshtein calculates the Levenshtein distance between two strings
func Levenshtein(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	n, m := len(r1), len(r2)
	if n == 0 {
		return m
	}
	if m == 0 {
		return n
	}
	matrix := make([][]int, n+1)
	for i := range matrix {
		matrix[i] = make([]int, m+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}
	return matrix[n][m]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}