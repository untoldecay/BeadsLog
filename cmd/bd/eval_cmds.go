package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/untoldecay/BeadsLog/internal/audit"
	"github.com/untoldecay/BeadsLog/internal/config"
	"github.com/untoldecay/BeadsLog/internal/eval"
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

		fmt.Printf("%s Evaluation mode disabled.\n", ui.RenderPass("✓"))
	},
}

var evalCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Prune orphan eval worktrees and temporary directories",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Pruning eval artifacts for this project...\n")
		count := eval.PruneProjectOrphans()
		if count == 0 {
			fmt.Println("No orphan eval worktrees found.")
		} else {
			fmt.Printf("\n✨ Cleaned up %d artifacts.\n", count)
		}
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

		groups := clusterSessions(sessions)

		for _, g := range groups {
			fmt.Printf("\nTASK GROUP: %q (Matched %d runs)\n", g.Name, len(g.Sessions))
			
			// Find base run (Base/Blind) to serve as reference
			baseTokens := 0
			for _, s := range g.Sessions {
				if strings.Contains(strings.ToLower(s.ScenarioID), "restriction") || strings.Contains(strings.ToLower(s.ScenarioID), "base") {
					baseTokens = s.TotalTokens
					break
				}
			}
			if baseTokens == 0 { baseTokens = 1 } 

			runs := make([]ui.EvalRun, len(g.Sessions))
			for i, s := range g.Sessions {
				strategy, logic, bonus, emoji := detectRefinedStrategyFull(s.Tools)
				
				duration := "0s"
				if s.EndTime.After(s.StartTime) {
					duration = s.EndTime.Sub(s.StartTime).Round(time.Second).String()
				}

				isBase := strings.Contains(strings.ToLower(s.ScenarioID), "restriction")
				
				var effStr string
				if isBase {
					effStr = "1.0x ⚪"
					logic = "-"
					strategy = "Base"
				} else {
					costFactor := float64(baseTokens) / float64(s.TotalTokens)
					if s.TotalTokens == 0 { costFactor = 0 } 
					efficiencyScore := bonus * costFactor
					effStr = fmt.Sprintf("%.1fx %s", efficiencyScore, emoji)
				}

				run := ui.EvalRun{
					ID:        s.ID,
					Strategy:  logic, 
					Tools:     strings.Join(s.Tools, "\n"),
					TokenCost: fmt.Sprintf("%d", s.TotalTokens),
					TimeSpent: duration,
					Efficiency: effStr,
				}

				if isBase {
					run.Pass = "REF"
					run.Score = 0
				} else if strings.Contains(strategy, "Optimal") {
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

func detectRefinedStrategyFull(tools []string) (string, string, float64, string) {
	if len(tools) == 0 {
		return "Unknown", "No tools used", 0.0, "🔴"
	}

	intents := make([]string, len(tools))
	for i, t := range tools {
		tLower := strings.ToLower(t)
		if strings.Contains(tLower, "bd status") || strings.Contains(tLower, "bd onboard") || strings.Contains(tLower, "bd devlog sync") {
			intents[i] = "bd_hydration"
		} else if strings.Contains(tLower, "devlog graph") || strings.Contains(tLower, "devlog search") || strings.Contains(tLower, "devlog resume") || strings.Contains(tLower, "devlog entities") {
			intents[i] = "bd_mapping"
		} else if strings.Contains(tLower, "read_file") || strings.Contains(tLower, "grep") || strings.Contains(tLower, "glob") || strings.Contains(tLower, "ls") || strings.Contains(tLower, "cat") || strings.Contains(tLower, "find") {
			intents[i] = "verification"
		} else {
			intents[i] = "other"
		}
	}

	hasHydration := sliceContainsStrict(intents, "bd_hydration")
	hasMapping := sliceContainsStrict(intents, "bd_mapping")
	hasVerification := sliceContainsStrict(intents, "verification")

	// Find first occurrences
	firstHydration, firstMapping, firstVerify := -1, -1, -1
	for i, t := range intents {
		if t == "bd_hydration" && firstHydration == -1 { firstHydration = i }
		if t == "bd_mapping" && firstMapping == -1 { firstMapping = i }
		if t == "verification" && firstVerify == -1 { firstVerify = i }
	}

	// Base Case
	if !hasHydration && !hasMapping {
		return "Base", "Pure Brute-Force", 0.1, "🔴"
	}

	// Optimal: Hydration -> Mapping -> Verification
	if hasHydration && hasMapping {
		if firstVerify == -1 || firstMapping < firstVerify {
			if hasVerification {
				return "Optimal", "Map FIRST + Verify LATER", 1.5, "🟢🟢🟢"
			}
			return "Shallow", "Map ONLY (Risk)", 0.8, "🟠"
		}
		return "Disordered", "Verify BEFORE Map", 0.5, "🔴🟠"
	}

	// Disordered: Missing Hydration
	if hasMapping && !hasHydration {
		return "Disordered", "Map WITHOUT Hydration", 0.5, "🔴"
	}

	return "Disordered", "Verify BEFORE Map", 0.5, "🔴🟠"
}

type evalSession struct {
	ID          string
	ScenarioID  string
	Tools       []string
	TotalTokens int
	StartTime   time.Time
	EndTime     time.Time
}

func extractEvalSessions(n int) ([]evalSession, error) {
	historyFile := filepath.Join("eval", "history.jsonl")
	if _, err := os.Stat(historyFile); err == nil {
		f, err := os.Open(historyFile)
		if err != nil { return nil, err }
		defer f.Close()

		var results []evalSession
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			var entry audit.Entry
			if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil { continue }
			
			if entry.Kind != "eval_run" { continue }

			s := evalSession{
				ID:         entry.EvalSessionID,
				StartTime:  entry.CreatedAt,
				EndTime:    entry.CreatedAt,
				ScenarioID: entry.Prompt,
			}

			if tokens, ok := entry.Extra["tokens"].(map[string]any); ok {
				if total, ok := tokens["total"].(float64); ok {
					s.TotalTokens = int(total)
				}
			}

			if tc, ok := entry.Extra["tool_calls"].([]any); ok {
				for _, t := range tc {
					if name, ok := t.(string); ok {
						s.Tools = append(s.Tools, name)
					} else if tMap, ok := t.(map[string]any); ok {
						if name, ok := tMap["name"].(string); ok {
							s.Tools = append(s.Tools, name)
						}
					}
				}
			}
			results = append(results, s)
		}
		
		if len(results) > n {
			results = results[len(results)-n:]
		}
		return results, nil
	}

	return nil, nil
}

type EvalGroup struct {
	Name     string
	Sessions []evalSession
}

func clusterSessions(sessions []evalSession) []EvalGroup {
	var groups []EvalGroup

	for _, s := range sessions {
		found := false
		
		cleanName := s.ScenarioID
		cleanName = strings.Split(cleanName, "\n\n(CRITICAL PROTOCOL")[0]
		cleanName = strings.Split(cleanName, "\n\n(MANDATORY")[0]
		cleanName = strings.Split(cleanName, "\n\n(RESTRICTION")[0]
		cleanName = strings.Split(cleanName, "\n\n(USE BD COMMANDS")[0]
		cleanName = strings.TrimSpace(cleanName)

		if cleanName == "" {
			cleanName = "Untitled Session"
		}

		for i, g := range groups {
			if g.Name == cleanName {
				groups[i].Sessions = append(groups[i].Sessions, s)
				found = true
				break
			}
			
			dist := Levenshtein(strings.ToLower(g.Name), strings.ToLower(cleanName))
			threshold := len(cleanName) / 3
			if dist <= threshold {
				groups[i].Sessions = append(groups[i].Sessions, s)
				found = true
				break
			}
		}

		if !found {
			groups = append(groups, EvalGroup{
				Name:     cleanName,
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
	evalCmd.AddCommand(evalCleanCmd)
	evalCmd.AddCommand(evalTaskCmd)
	evalCmd.AddCommand(evalBenchCmd)
	rootCmd.AddCommand(evalCmd)
}

func Levenshtein(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	n, m := len(r1), len(r2)
	if n == 0 { return m }
	if m == 0 { return n }
	matrix := make([][]int, n+1)
	for i := range matrix {
		matrix[i] = make([]int, m+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] { matrix[0][j] = j }
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] { cost = 1 }
			matrix[i][j] = min(
				matrix[i-1][j]+1,
				matrix[i][j-1]+1,
				matrix[i-1][j-1]+cost,
			)
		}
	}
	return matrix[n][m]
}

func min(a, b, c int) int {
	if a < b {
		if a < c { return a }
		return c
	}
	if b < c { return b }
	return c
}