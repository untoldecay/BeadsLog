package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/untoldecay/BeadsLog/internal/audit"
	"github.com/untoldecay/BeadsLog/internal/eval"
	"github.com/untoldecay/BeadsLog/internal/ui"
)

var evalTaskCmd = &cobra.Command{
	Use:   "task",
	Short: "Interactive eval: sandbox -> test -> analyze -> cleanup",
	Long: `Runs an A/B/C evaluation of the agent protocol in isolated sandboxes using OpenCode CLI.
It compares three runs:
1. Implicit: Standard protocol enforcement via AGENT.md.
2. Explicit: Direct instructions to follow the 'Map First, Verify Later' pattern.
3. Base: Forbids all 'bd' commands to establish a baseline for inefficiency.`,
	Run: func(cmd *cobra.Command, args []string) {
		runOpenCodeEval()
	},
}

func init() {
	evalCmd.AddCommand(evalTaskCmd)
}

// TokenUsage tracks token consumption
type TokenUsage struct {
	Input     int `json:"input"`
	Output    int `json:"output"`
	Reasoning int `json:"reasoning"`
	Total     int `json:"total"`
}

// OpenCodeTrace represents the aggregated trace from OpenCode CLI
type OpenCodeTrace struct {
	ToolCalls []ToolCall `json:"tool_calls"`
	Response  string     `json:"response"`
	Reasoning string     `json:"reasoning"`
	Tokens    TokenUsage `json:"tokens_used"`
}

type ToolCall struct {
	Name   string `json:"name"`
	Args   string `json:"args"` // Store command string for bash
	Result string `json:"result"`
}

type Scoring struct {
	Strategy   string
	LogicDesc  string
	Status     string
	Bonus      float64
	Emoji      string
}

func getToolCategoryIcon(name, args string) string {
	name = strings.ToLower(name)
	args = strings.ToLower(args)

	// ● BD Search/Mapping: bd devlog graph, bd devlog search, bd devlog entities
	if strings.Contains(args, "devlog graph") || strings.Contains(args, "devlog search") || strings.Contains(args, "devlog entities") {
		return "●"
	}

	// ◐ BD Hydration: bd onboard, bd status, bd devlog sync
	if strings.Contains(args, "bd onboard") || strings.Contains(args, "bd status") || strings.Contains(args, "bd devlog sync") {
		return "◐"
	}

	// ⚪ Vanilla: grep, read_file, ls, glob, bash (generic)
	return "⚪"
}

func runOpenCodeEval() {
	fmt.Printf("\n%s %s\n", "🧪", "BeadsLog Eval Mode - OpenCode A/B/C Testing")
	fmt.Println(strings.Repeat("-", 60))

	projHash := eval.GetProjectHash()

	// 0. PRE-FLIGHT JANITOR
	fmt.Print("Cleaning up stale artifacts from previous runs... ")
	eval.PruneProjectOrphans()
	fmt.Println("Done")

	// 1. BACKUP CURRENT WORK (Invisible Snapshot)
	timestamp := time.Now().Format("20060102-150405")
	backupBranch := fmt.Sprintf("eval-backup-%s", timestamp)

	fmt.Printf("Creating invisible snapshot on branch %s... ", backupBranch)
	if err := createBackupBranch(backupBranch); err != nil {
		fmt.Printf("Failed: %v\n", err)
		return
	}
	fmt.Println("Done (Workspace remains as-is)")

	// OUTER LOOP: New Task
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter task for agent: ")
		task, _ := reader.ReadString('\n')
		task = strings.TrimSpace(task)

		if task == "" {
			fmt.Println("Empty task, aborting")
			break
		}

		// INNER LOOP: Restart Task
		for {
			runTimestamp := fmt.Sprintf("%d", time.Now().Unix())

			// 3. CREATE SANDBOX ROOTS
			tempDir, err := os.MkdirTemp("", "beads-eval-"+projHash+"-*")
			if err != nil {
				fmt.Printf("Failed to create temp dir: %v\n", err)
				return
			}
			tempDir, _ = filepath.Abs(tempDir)

			wtImplicit := filepath.Join(tempDir, "implicit")
			wtExplicit := filepath.Join(tempDir, "explicit")
			wtBase := filepath.Join(tempDir, "base")

			fmt.Printf("\nCreating isolated sandboxes in %s...\n", tempDir)

			for name, path := range map[string]string{"Implicit": wtImplicit, "Explicit": wtExplicit, "Base": wtBase} {
				// Initialize sandbox from the BACKUP branch so it has latest code
				if err := createEvalWorktree(path, backupBranch); err != nil {
					fmt.Printf("  Failed to create %s worktree: %v\n", name, err)
					eval.SafeCleanupABC(wtImplicit, wtExplicit, wtBase, tempDir)
					return
				}
				fmt.Printf("  %s Sandbox: Ready (via %s)\n", name, backupBranch)
			}

			// 4. RUN TESTS
			traceIdImplicit := "eval-" + runTimestamp + "-implicit"
			implicitResult := runOpenCodeTest(wtImplicit, task, traceIdImplicit, false)

			time.Sleep(2 * time.Second) // Settle time

			traceIdExplicit := "eval-" + runTimestamp + "-explicit"
			explicitTask := fmt.Sprintf("%s\n\n(CRITICAL PROTOCOL: Use 'bd devlog' tools FIRST to map the landscape, then verify findings with classic tools like grep/cat if needed.)", task)
			explicitResult := runOpenCodeTest(wtExplicit, explicitTask, traceIdExplicit, true)

			time.Sleep(2 * time.Second) // Settle time

			traceIdBase := "eval-" + runTimestamp + "-base"
			baseTask := fmt.Sprintf("%s\n\n(RESTRICTION: Do NOT use any 'bd' or 'beads' commands. Rely only on standard Unix tools like grep, find, ls, etc.)", task)
			baseResult := runOpenCodeTest(wtBase, baseTask, traceIdBase, false)

			// 5. REPORT
			fmt.Printf("\nGenerating comparative report...\n")
			report := generateABCReport(task, implicitResult, explicitResult, baseResult)
			fmt.Println(report)
			
			// Show paths to logs
			fmt.Printf("\n✨ Logs from this run are available at:\n")
			fmt.Printf("  - _rules/_evals/results/%s.log\n", traceIdImplicit)
			fmt.Printf("  - _rules/_evals/results/%s.log\n", traceIdExplicit)
			fmt.Printf("  - _rules/_evals/results/%s.log\n", traceIdBase)

			// 6. INTERACTIVE MENU
			var action string
			fmt.Println() 
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Test is now done.").
						Options(
							huh.NewOption("Quit (Delete backup branch)", "quit_clean"),
							huh.NewOption("Quit (KEEP backup branch for rescue)", "quit_keep"),
							huh.NewOption("Restart test (Retry same task)", "restart"),
							huh.NewOption("Start a new test (New task)", "new"),
						).
						Value(&action),
				),
			)

			err = form.Run()
			if err != nil {
				action = "quit_clean" // Handle ctrl+c
			}

			// Clean up current run
			fmt.Print("Cleaning up sandboxes... ")
			eval.SafeCleanupABC(wtImplicit, wtExplicit, wtBase, tempDir)
			fmt.Println("Done")

			if action == "quit_clean" || action == "quit_keep" {
				if action == "quit_clean" {
					fmt.Printf("\nDeleting backup branch %s... ", backupBranch)
					_ = exec.Command("git", "branch", "-D", backupBranch).Run()
					fmt.Println("Done.")
				} else {
					fmt.Printf("\nKeeping backup branch %s for rescue.\n", backupBranch)
					fmt.Printf("  Rescue command: git reset --hard %s\n", backupBranch)
				}
				fmt.Println("✅ Eval task complete.")
				return
			}
			
			if action == "new" {
				break 
			}
			
			if action == "restart" {
				continue 
			}
		}
	}
}

func getEvalCurrentBranch() string {
	out, _ := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	return strings.TrimSpace(string(out))
}

func createBackupBranch(name string) error {
	// 1. Stage everything temporarily to capture dirty state
	_ = exec.Command("git", "add", ".").Run()

	// 2. Write tree object from index
	out, err := exec.Command("git", "write-tree").Output()
	if err != nil {
		_ = exec.Command("git", "reset").Run() // Cleanup index on fail
		return fmt.Errorf("git write-tree: %w", err)
	}
	treeID := strings.TrimSpace(string(out))

	// 3. Create commit object (parent is HEAD)
	headOut, _ := exec.Command("git", "rev-parse", "HEAD").Output()
	headID := strings.TrimSpace(string(headOut))

	commitOut, err := exec.Command("git", "commit-tree", treeID, "-p", headID, "-m", "chore(eval): safety snapshot for task evaluation").Output()
	if err != nil {
		_ = exec.Command("git", "reset").Run()
		return fmt.Errorf("git commit-tree: %w", err)
	}
	commitID := strings.TrimSpace(string(commitOut))

	// 4. Create the branch pointing to this commit
	if err := exec.Command("git", "branch", name, commitID).Run(); err != nil {
		_ = exec.Command("git", "reset").Run()
		return fmt.Errorf("git branch creation: %w", err)
	}

	// 5. Reset index back to normal (workspace stays dirty)
	_ = exec.Command("git", "reset").Run()
	
	return nil
}

func runOpenCodeTest(wtDir, task, traceId string, explicit bool) OpenCodeTrace {
	fmt.Printf("  Executing %s... ", filepath.Base(wtDir))

	opencodePath, err := exec.LookPath("opencode")
	if err != nil {
		fmt.Println("Failed (opencode CLI missing)")
		return OpenCodeTrace{}
	}

	apiKey := getEvalApiKey()
	if apiKey == "" {
		fmt.Println("Failed (GEMINI_API_KEY missing)")
		return OpenCodeTrace{}
	}

	beadsDir := filepath.Join(wtDir, ".beads")

	// OpenCode CLI headless run
	cmd := exec.Command(opencodePath, "run",
		"--format", "json",
		"--model", "google/gemini-flash-latest",
		task,
	)
	cmd.Dir = wtDir

	// Auth + workspace + HOME isolation
	cmd.Env = getSandboxEnv(wtDir, apiKey, beadsDir)

	// Use pipe to capture stdout for streaming parse
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Failed to create pipe: %v\n", err)
		return OpenCodeTrace{}
	}
	// stderr to file for debugging
	errFile, _ := os.Create(filepath.Join("eval", "traces", traceId+".error.txt"))
	cmd.Stderr = errFile

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed start: %v\n", err)
		return OpenCodeTrace{}
	}

	// Capture raw stream for debugging
	rawStreamFile, _ := os.Create(filepath.Join("eval", "traces", traceId+".stream.jsonl"))
	defer rawStreamFile.Close()

	// Human-readable log export
	readableLogFile, _ := os.Create(filepath.Join("eval", "traces", traceId+".readable.log"))
	defer readableLogFile.Close()

	// Prepend original query to log
	readableLogFile.WriteString(fmt.Sprintf("QUERY: %s\n", task))
	readableLogFile.WriteString(strings.Repeat("=", 60) + "\n\n")

	fmt.Printf("\n--- [%s] START ---\n", traceId)

	var trace OpenCodeTrace
	var streamContent []string
	
	scanner := bufio.NewScanner(cmdReader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "{") {
			continue // Skip logs/empty/headers
		}

		rawStreamFile.WriteString(line + "\n") // Save raw line

		var event map[string]interface{}
		if err := json.Unmarshal([]byte(line), &event); err == nil {
			aggregateEvent(&trace, event)

			// REAL-TIME VISIBILITY
			eventType, _ := event["type"].(string)
			var msg string
			switch eventType {
			case "reasoning", "thought":
				if text, ok := getEventText(event); ok && text != "" {
					msg = fmt.Sprintf("\n💭 [THINKING] %s\n", text)
				}
			case "text":
				if text, ok := getEventText(event); ok && text != "" {
					msg = text
				}
			case "tool_use":
				tool, _ := getEventTool(event)
				args := getEventToolArgs(event)
				icon := getToolCategoryIcon(tool, args)
				display := tool
				if tool == "bash" {
					display = fmt.Sprintf("bash(%s)", args)
				} else if args != "" {
					display = fmt.Sprintf("%s(%s)", tool, args)
				}
				msg = fmt.Sprintf("\n%s  [TOOL] %s\n", icon, display)
			case "tool_result":
				msg = "✅ [DONE]\n"
			}

			if msg != "" {
				readableLogFile.WriteString(msg)
				streamContent = append(streamContent, strings.Split(msg, "\n")...)
				
				// Keep only last 10 lines for display
				displayLines := streamContent
				if len(displayLines) > 10 {
					displayLines = displayLines[len(displayLines)-10:]
				}
				
				// Clear current block and print viewport-like output
				// We use ANSI codes to go up 11 lines and clear
				fmt.Print("\033[11A\033[J") 
				fmt.Printf("--- [%s] LIVE STREAM (LAST 10 LINES) ---\n", traceId)
				for _, l := range displayLines {
					fmt.Println(l)
				}
			}
		}
	}

	cmd.Wait()
	errFile.Close()
	fmt.Printf("\n--- [%s] END ---\n", traceId)
	fmt.Print("Done\n")

	// Save structured trace
	traceFile := filepath.Join("eval", "traces", traceId+".json")
	_ = os.MkdirAll(filepath.Dir(traceFile), 0755)
	formattedJSON, _ := json.MarshalIndent(trace, "", "  ")
	_ = os.WriteFile(traceFile, formattedJSON, 0644)

	// EXPORT to _rules/_evals/results/
	exportDir := filepath.Join("_rules", "_evals", "results")
	_ = os.MkdirAll(exportDir, 0755)
	exportFile := filepath.Join(exportDir, fmt.Sprintf("%s.log", traceId))
	if data, err := os.ReadFile(filepath.Join("eval", "traces", traceId+".readable.log")); err == nil {
		_ = os.WriteFile(exportFile, data, 0644)
	}

	// ARCHIVE to History (for 'bd eval report')
	archiveTrace(traceId, task, trace)

	return trace
}

func aggregateEvent(trace *OpenCodeTrace, event map[string]interface{}) {
	eventType, ok := event["type"].(string)
	if !ok {
		return
	}

	switch eventType {
	case "tool_use":
		toolName, _ := getEventTool(event)
		argsStr := getEventToolArgs(event)

		trace.ToolCalls = append(trace.ToolCalls, ToolCall{
			Name: toolName,
			Args: argsStr,
		})

	case "tool_result":
		if len(trace.ToolCalls) > 0 {
			result := ""
			if part, ok := event["part"].(map[string]interface{}); ok {
				if r, ok := part["output"].(string); ok {
					result = r
				}
			} else if r, ok := event["result"].(string); ok {
				result = r
			}
			trace.ToolCalls[len(trace.ToolCalls)-1].Result = result
		}

	case "text":
		if text, ok := getEventText(event); ok {
			trace.Response += text + "\n"
		}

	case "reasoning", "thought":
		if text, ok := getEventText(event); ok {
			trace.Reasoning += text + "\n"
		}

	case "usage":
		// METHOD 4: Capture tokens from stream
		if val, ok := event["input_tokens"].(float64); ok {
			trace.Tokens.Input += int(val)
		} else if part, ok := event["part"].(map[string]interface{}); ok {
			if val, ok := part["input_tokens"].(float64); ok {
				trace.Tokens.Input += int(val)
			}
		}

		if val, ok := event["output_tokens"].(float64); ok {
			trace.Tokens.Output += int(val)
		} else if part, ok := event["part"].(map[string]interface{}); ok {
			if val, ok := part["output_tokens"].(float64); ok {
				trace.Tokens.Output += int(val)
			}
		}

		if val, ok := event["reasoning_tokens"].(float64); ok {
			trace.Tokens.Reasoning += int(val)
		} else if part, ok := event["part"].(map[string]interface{}); ok {
			if val, ok := part["reasoning_tokens"].(float64); ok {
				trace.Tokens.Reasoning += int(val)
			}
		}
		
		trace.Tokens.Total = trace.Tokens.Input + trace.Tokens.Output + trace.Tokens.Reasoning

	case "tokens":
		// Alternative key used in Method 4 guide
		if input, ok := event["input"].(float64); ok {
			trace.Tokens.Input += int(input)
		}
		if output, ok := event["output"].(float64); ok {
			trace.Tokens.Output += int(output)
		}
		if reasoning, ok := event["reasoning"].(float64); ok {
			trace.Tokens.Reasoning += int(reasoning)
		}
		trace.Tokens.Total = trace.Tokens.Input + trace.Tokens.Output + trace.Tokens.Reasoning

	case "step_finish":
		// Captured from real OpenCode logs
		if part, ok := event["part"].(map[string]interface{}); ok {
			if tokens, ok := part["tokens"].(map[string]interface{}); ok {
				if input, ok := tokens["input"].(float64); ok {
					trace.Tokens.Input += int(input)
				}
				if output, ok := tokens["output"].(float64); ok {
					trace.Tokens.Output += int(output)
				}
				if reasoning, ok := tokens["reasoning"].(float64); ok {
					trace.Tokens.Reasoning += int(reasoning)
				}
			}
		}
		trace.Tokens.Total = trace.Tokens.Input + trace.Tokens.Output + trace.Tokens.Reasoning
	}
}

func getEventText(event map[string]interface{}) (string, bool) {
	if part, ok := event["part"].(map[string]interface{}); ok {
		if t, ok := part["text"].(string); ok {
			return t, true
		}
	} else if t, ok := event["text"].(string); ok {
		return t, true
	}
	return "", false
}

func getEventTool(event map[string]interface{}) (string, bool) {
	if part, ok := event["part"].(map[string]interface{}); ok {
		if t, ok := part["tool"].(string); ok {
			return t, true
		}
	} else if t, ok := event["tool"].(string); ok {
		return t, true
	}
	return "", false
}

func getEventToolArgs(event map[string]interface{}) string {
	if part, ok := event["part"].(map[string]interface{}); ok {
		if state, ok := part["state"].(map[string]interface{}); ok {
			if input, ok := state["input"].(map[string]interface{}); ok {
				if cmd, ok := input["command"].(string); ok {
					return cmd
				}
				b, _ := json.Marshal(input)
				return string(b)
			}
		}
	}
	return ""
}

func generateABCReport(task string, implicit, explicit, base OpenCodeTrace) string {
	implicitScore := scoreTrace(implicit.ToolCalls)
	explicitScore := scoreTrace(explicit.ToolCalls)
	
	// Base is reference
	baseTokens := base.Tokens.Total
	if baseTokens == 0 { baseTokens = 1 }

	implicitEff := calculateEfficiency(implicitScore.Bonus, implicit.Tokens.Total, baseTokens)
	explicitEff := calculateEfficiency(explicitScore.Bonus, explicit.Tokens.Total, baseTokens)

	report := fmt.Sprintf(`
%s
TASK: "%s"

╭──────────────────┬──────────────────────────────┬────────┬──────┬──────────╮
│ Run              │ Logic                        │ Tokens │ Stat │ Eff.     │
├──────────────────┼──────────────────────────────┼────────┼──────┼──────────┤
│ Implicit         │ %-28s │ %-6d │ %s │ %.1fx %s │
├──────────────────┼──────────────────────────────┼────────┼──────┼──────────┤
│ Explicit         │ %-28s │ %-6d │ %s │ %.1fx %s │
├──────────────────┼──────────────────────────────┼────────┼──────┼──────────┤
│ Base             │ -                            │ %-6d │ REF  │ 1.0x ⚪   │
╰──────────────────┴──────────────────────────────┴────────┴──────┴──────────╯

🎯 Protocol Effectiveness: %.1fx better than base
(See docs/EVAL.md for scoring rubric)

`,
		ui.RenderAccent("📊 A/B/C Comparison Report"),
		task,
		implicitScore.LogicDesc, implicit.Tokens.Total, implicitScore.Status, implicitEff, implicitScore.Emoji,
		explicitScore.LogicDesc, explicit.Tokens.Total, explicitScore.Status, explicitEff, explicitScore.Emoji,
		baseTokens,
		float64(baseTokens)/float64(implicit.Tokens.Total),
	)

	report += fmt.Sprintf("\n%s\n%s", ui.RenderAccent("🕒 Chronological Tool Traces:"), renderToolCalls("Implicit", implicit.ToolCalls))
	report += renderToolCalls("Explicit", explicit.ToolCalls)
	report += renderToolCalls("Base", base.ToolCalls)

	return report
}

func calculateEfficiency(bonus float64, tokens, base int) float64 {
	if tokens == 0 { return 0.0 }
	return bonus * (float64(base) / float64(tokens))
}

func renderToolCalls(label string, calls []ToolCall) string {
	if len(calls) == 0 {
		return fmt.Sprintf("\n[%s] (no tools used)\n", label)
	}

	out := fmt.Sprintf("\n[%s]:\n", label)
	for i, c := range calls {
		icon := getToolCategoryIcon(c.Name, c.Args)
		name := formatToolName(c)
		out += fmt.Sprintf("  %d. %s %s\n", i+1, icon, name)
	}
	return out
}

func formatToolName(c ToolCall) string {
	name := c.Name
	if name == "bash" {
		cmd := c.Args
		if len(cmd) > 60 {
			cmd = cmd[:57] + "..."
		}
		name = fmt.Sprintf("bash(%s)", cmd)
	}
	return name
}

func scoreTrace(toolCalls []ToolCall) Scoring {
	if len(toolCalls) == 0 {
		return Scoring{"Blind", "Pure Brute-Force", "FAIL", 0.1, "🔴"}
	}

	intents := make([]string, len(toolCalls))
	for i, tc := range toolCalls {
		if tc.Name == "bash" {
			cmd := strings.ToLower(tc.Args)
			if strings.Contains(cmd, "bd status") || strings.Contains(cmd, "bd onboard") || strings.Contains(cmd, "bd devlog sync") {
				intents[i] = "bd_hydration"
			} else if strings.Contains(cmd, "devlog graph") || strings.Contains(cmd, "devlog search") || strings.Contains(cmd, "devlog resume") || strings.Contains(cmd, "devlog entities") {
				intents[i] = "bd_mapping"
			} else if strings.Contains(cmd, "grep") || strings.Contains(cmd, "ls") || strings.Contains(cmd, "cat") || strings.Contains(cmd, "find") {
				intents[i] = "verification"
			} else {
				intents[i] = "bash_generic"
			}
		} else if strings.Contains(tc.Name, "read_file") || strings.Contains(tc.Name, "grep") || strings.Contains(tc.Name, "glob") {
			intents[i] = "verification"
		} else {
			intents[i] = tc.Name
		}
	}

	hasHydration := sliceContainsStrict(intents, "bd_hydration")
	hasMapping := sliceContainsStrict(intents, "bd_mapping")
	hasVerification := sliceContainsStrict(intents, "verification")

	firstHydration, firstMapping, firstVerify := -1, -1, -1
	for i, t := range intents {
		if t == "bd_hydration" && firstHydration == -1 { firstHydration = i }
		if t == "bd_mapping" && firstMapping == -1 { firstMapping = i }
		if t == "verification" && firstVerify == -1 { firstVerify = i }
	}

	if !hasHydration && !hasMapping {
		return Scoring{"Base", "Pure Brute-Force", "FAIL", 0.1, "🔴"}
	}

	if hasHydration && hasMapping {
		if firstVerify == -1 || firstMapping < firstVerify {
			if hasVerification {
				return Scoring{"Optimal", "Map FIRST + Verify LATER", "PASS", 1.5, "🟢🟢🟢"}
			}
			return Scoring{"Shallow", "Map ONLY (Risk)", "PASS", 0.8, "🟠"}
		}
		return Scoring{"Disordered", "Verify BEFORE Map", "FAIL", 0.5, "🔴🟠"}
	}

	if hasMapping && !hasHydration {
		return Scoring{"Disordered", "Map WITHOUT Hydration", "FAIL", 0.5, "🔴"}
	}

	return Scoring{"Disordered", "Verify BEFORE Map", "FAIL", 0.5, "🔴🟠"}
}

func sliceContainsStrict(list []string, item string) bool {
	for _, s := range list {
		if s == item { return true }
	}
	return false
}

func createEvalWorktree(path, sourceBranch string) error {
	exe, err := os.Executable()
	if err != nil { return err }

	// 1. Fresh temp git repo (NO worktree sharing)
	_ = os.RemoveAll(path)
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	if err := exec.Command("git", "init", "--template=", path).Run(); err != nil {
		return err
	}

	// 2. Hard reset to sourceBranch (no history)
	mainRepo, _ := os.Getwd()
	cmdRemote := exec.Command("git", "remote", "add", "origin", mainRepo)
	cmdRemote.Dir = path
	_ = cmdRemote.Run()

	cmdFetch := exec.Command("git", "fetch", "origin", sourceBranch)
	cmdFetch.Dir = path
	_ = cmdFetch.Run()

	cmdCheckout := exec.Command("git", "checkout", "--force", "-B", "eval-head", "FETCH_HEAD")
	cmdCheckout.Dir = path
	_ = cmdCheckout.Run()

	// 3. Clean index + working tree
	cmdRm := exec.Command("git", "rm", "--cached", "-r", ".")
	cmdRm.Dir = path
	_ = cmdRm.Run()

	cmdRead := exec.Command("git", "read-tree", "HEAD")
	cmdRead.Dir = path
	_ = cmdRead.Run()

	cmdCheckoutIndex := exec.Command("git", "checkout-index", "-a", "--force")
	cmdCheckoutIndex.Dir = path
	_ = cmdCheckoutIndex.Run()

	// 4. Disable Git autorevert & Hooks
	cmdConfig := exec.Command("git", "config", "core.fileMode", "false")
	cmdConfig.Dir = path
	_ = cmdConfig.Run()
	
	cmdConfigHooks := exec.Command("git", "config", "core.hooksPath", "/dev/null")
	cmdConfigHooks.Dir = path
	_ = cmdConfigHooks.Run()

	_ = os.RemoveAll(filepath.Join(path, "_sandbox"))
	agentPath := filepath.Join(path, "AGENTS.md")
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		agentPath = filepath.Join(path, "AGENT.md")
	}
	if content, err := os.ReadFile(agentPath); err == nil {
		override := "# SANDBOX OVERRIDE (REQUIRED FIRST STEPS)\n\n" +
			"⚠️ **YOU ARE IN A FRESH SANDBOX.**\n\n" +
			"1. **CHECK STATUS**: Run `bd status`.\n" +
			"2. **IF NOT INITIALIZED**: Run `bd onboard` IMMEDIATELY.\n" +
			"3. **THEN RESUME**: Run `bd devlog resume`.\n\n" +
			"⛔ **DO NOT RUN `bd prime`**.\n\n---\n\n"
		_ = os.WriteFile(agentPath, append([]byte(override), content...), 0644)
	}
	beadsDir := filepath.Join(path, ".beads")
	_ = os.RemoveAll(beadsDir)

	initCmd := exec.Command(exe, "init", "--quiet", "--force")
	initCmd.Dir = path
	initCmd.Env = getSandboxEnv(path, "", beadsDir)
	_ = initCmd.Run()

	readyCmd := exec.Command(exe, "ready")
	readyCmd.Dir = path
	readyCmd.Env = getSandboxEnv(path, "", beadsDir)
	_ = readyCmd.Run()

	absExe, _ := filepath.Abs(exe)
	// Disable OpenCode git features to prevent auto-revert
	defaultConfig := `{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "beads": {
      "type": "local",
      "command": ["` + absExe + `", "mcp"],
      "enabled": true
    }
  }
}`
	_ = os.WriteFile(filepath.Join(path, "opencode.json"), []byte(defaultConfig), 0644)
	return nil
}

func getEvalApiKey() string {
	home, _ := os.UserHomeDir()
	paths := []string{filepath.Join(home, ".gemini", "eval.env"), ".env"}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil { continue }
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "GEMINI_API_KEY=") {
				return strings.Trim(line[len("GEMINI_API_KEY="):], "\"'")
			}
		}
	}
	return os.Getenv("GEMINI_API_KEY")
}

func getSandboxEnv(homeDir string, apiKey string, beadsDir string) []string {
	env := os.Environ()
	// Override HOME to sandbox root to isolate global config/registry
	env = append(env, "HOME="+homeDir)
	
	// Prevent OpenCode from walking up to parent .git or config
	env = append(env, "OPENCODE_NO_PARENT_CONFIG=1")
	env = append(env, "OPENCODE_DISABLE_GIT=1")
	
	if apiKey != "" {
		env = append(env, "GOOGLE_GENERATIVE_AI_API_KEY="+apiKey)
	}
	if beadsDir != "" {
		env = append(env, "BEADS_DIR="+beadsDir)
	}
	return env
}

func archiveTrace(traceId, task string, trace OpenCodeTrace) {
	var toolTrace []string
	for _, tc := range trace.ToolCalls {
		name := tc.Name
		if name == "bash" { name = fmt.Sprintf("bash(%s)", tc.Args) }
		toolTrace = append(toolTrace, name)
	}
	entry := audit.Entry{
		Kind:          "eval_run",
		EvalSessionID: traceId,
		CreatedAt:     time.Now().UTC(),
		Prompt:        task,
		Response:      trace.Response,
		Extra: map[string]any{
			"tokens":     trace.Tokens,
			"tool_calls": toolTrace,
			"reasoning":  trace.Reasoning,
		},
	}
	historyFile := filepath.Join("eval", "history.jsonl")
	_ = os.MkdirAll(filepath.Dir(historyFile), 0755)
	f, err := os.OpenFile(historyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil { return }
	defer f.Close()
	data, _ := json.Marshal(entry)
	f.Write(data)
	f.Write([]byte("\n"))
}