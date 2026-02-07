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

		// Header
		fmt.Printf("\n%-25s | %-5s | %-6s | %-15s | %s\n", "Session ID", "Pass", "Score", "Tokens (Est)", "Strategy")
		fmt.Println(strings.Repeat("-", 90))

		for _, s := range sessions {
			passStr := "FAIL"
			score := 0
			strategy := "Mixed"
			if len(s.Tools) > 0 {
				// Simple heuristic: if first tool is devlog, it's retrieval-first
				firstTool := strings.ToLower(s.Tools[0])
				if strings.Contains(firstTool, "devlog") || strings.Contains(firstTool, "onboard") {
					strategy = "Retrieval-First"
					score = 100
					passStr = "PASS"
				} else {
					strategy = "Brute-Force"
					score = 50 // Partial if they used it later, 0 if never
					for _, t := range s.Tools {
						if strings.Contains(strings.ToLower(t), "devlog") {
							strategy = "Hybrid"
							break
						}
					}
					if strategy == "Brute-Force" {
						score = 0
					}
				}
			}

			fmt.Printf("%-25s | %-5s | %-6d | %-15d | %s\n", s.ID, passStr, score, estimateCostSummary(s.Tools), strategy)
		}
		fmt.Println()
	},
}

type evalSession struct {
	ID         string
	ScenarioID string
	Tools      []string
	Timestamp  time.Time
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
				Timestamp: entry.CreatedAt,
			}
			sessionMap[entry.EvalSessionID] = s
			sessionOrder = append(sessionOrder, entry.EvalSessionID)
		}

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

func init() {
	evalCmd.AddCommand(evalStartCmd)
	evalCmd.AddCommand(evalNextCmd)
	evalCmd.AddCommand(evalReportCmd)
	evalCmd.AddCommand(evalStopCmd)
	rootCmd.AddCommand(evalCmd)
}