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

		// Prepare runs for the report
		runs := make([]ui.EvalRun, len(sessions))
		for i, s := range sessions {
			strategy := detectStrategy(s.Tools)
			cost := estimateCostSummary(s.Tools)
			
			run := ui.EvalRun{
				ID:        s.ID,
				Strategy:  strategy,
				Tools:     strings.Join(s.Tools, "\n"),
				TokenCost: fmt.Sprintf("%d", cost),
			}

			// Map strategy to qualitative attributes (from user provided table)
			switch strategy {
			case "Retrieval-First":
				run.Methodology = "Metadata Only"
				run.PrimaryFocus = "The \"Why\" & \"Where\"\n(Intent, Dependencies, Bugs)"
				run.EfficiencyRatio = "High Context / Low Detail\nGreat for understanding risk, but misses implementation specifics."
				run.BestUsedFor = "🟢 Onboarding\n🟢 Impact Analysis\n🟢 Architecture Review"
				run.CriticalMissing = "❌ The Code Reality\n(Doesn't know how it works now)"
				run.Pass = "PASS"
				run.Score = 100
			case "Brute-Force":
				run.Methodology = "Brute Force Code"
				run.PrimaryFocus = "The \"What\" & \"How\"\n(Current Logic, Syntax, Regex)"
				run.EfficiencyRatio = "High Detail / High Cost\nWasteful reading of irrelevant files to find the right ones."
				run.BestUsedFor = "🟢 Refactoring\n🟢 Debugging specific lines\n🟢 Writing Docs"
				run.CriticalMissing = "❌ The Hidden Context\n(Doesn't know why it was built)"
				run.Pass = "FAIL"
				run.Score = 0
			case "Hybrid":
				run.Methodology = "Indexed Verification"
				run.PrimaryFocus = "Synthesis\n(Intent + Reality)"
				run.EfficiencyRatio = "Maximum Efficiency\nLowest token cost for the highest quality, comprehensive insight."
				run.BestUsedFor = "👑 Master Reports\n👑 Complex Bug Fixes\n👑 Feature Planning"
				run.CriticalMissing = "✅ Nothing Critical\n(Complete picture)"
				run.Pass = "PASS"
				run.Score = 85
			default:
				run.Methodology = "Unknown"
				run.Pass = "FAIL"
				run.Score = 0
			}

			runs[i] = run
		}

		fmt.Println()
		fmt.Println(ui.RenderEvalReport(runs, ui.GetWidth()))
		fmt.Println()
	},
}

func detectStrategy(tools []string) string {
	if len(tools) == 0 {
		return "Unknown"
	}
	firstTool := strings.ToLower(tools[0])
	if strings.Contains(firstTool, "devlog") || strings.Contains(firstTool, "onboard") {
		return "Retrieval-First"
	}
	
	for _, t := range tools {
		if strings.Contains(strings.ToLower(t), "devlog") {
			return "Hybrid"
		}
	}
	return "Brute-Force"
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