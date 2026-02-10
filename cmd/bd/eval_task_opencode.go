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
	Input  int `json:"input"`
	Output int `json:"output"`
	Total  int `json:"total"`
}

// OpenCodeTrace represents the aggregated trace from OpenCode CLI
type OpenCodeTrace struct {
	ToolCalls []ToolCall `json:"tool_calls"`
	Response  string     `json:"response"`
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

func runOpenCodeEval() {
	fmt.Printf("\n%s %s\n", "🧪", "BeadsLog Eval Mode - OpenCode A/B/C Testing")
	fmt.Println(strings.Repeat("-", 60))

	projHash := eval.GetProjectHash()

	// 0. PRE-FLIGHT JANITOR
	fmt.Print("Cleaning up stale artifacts from previous runs... ")
	eval.PruneProjectOrphans()
	fmt.Println("Done")

	// 0.5 STASH RECOVERY
	handleLeftoverStashes()

	// 1. AUTOMATIC STASH
	fmt.Print("Stashing your current work... ")
	stashRef, err := createStash()
	if err != nil {
		fmt.Printf("Failed: %v\n", err)
		return
	}
	if stashRef == "" {
		fmt.Println("Done (Nothing to stash)")
	} else {
		fmt.Printf("Done (Ref: %s)\n", stashRef)
	}

	// Restore stash on exit
	defer func() {
		if stashRef != "" {
			fmt.Print("Restoring your work... ")
			restoreStash(stashRef)
			fmt.Println("Done")
		}
	}()

	// OUTER LOOP: New Task
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("\nEnter task for agent: ")
		task, _ := reader.ReadString('\n')
		task = strings.TrimSpace(task)

		if task == "" {
			fmt.Println("Empty task, aborting")
			return
		}

		// INNER LOOP: Restart Task
		for {
			timestamp := fmt.Sprintf("%d", time.Now().Unix())

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
				if err := createEvalWorktree(path); err != nil {
					fmt.Printf("  Failed to create %s worktree: %v\n", name, err)
					eval.SafeCleanupABC(wtImplicit, wtExplicit, wtBase, tempDir)
					return
				}
				fmt.Printf("  %s Sandbox: Ready\n", name)
			}

			// 4. RUN TESTS
			fmt.Printf("\nRunning agent tests...\n")

			// Run 1: Implicit
			traceIdImplicit := "eval-" + timestamp + "-implicit"
			implicitResult := runOpenCodeTest(wtImplicit, task, traceIdImplicit, false)

			// Run 2: Explicit
			traceIdExplicit := "eval-" + timestamp + "-explicit"
			explicitTask := fmt.Sprintf("%s\n\n(CRITICAL PROTOCOL: Use 'bd devlog' tools FIRST to map the landscape, then verify findings with classic tools like grep/cat if needed.)", task)
			explicitResult := runOpenCodeTest(wtExplicit, explicitTask, traceIdExplicit, true)

			// Run 3: Base (No BD)
			traceIdBase := "eval-" + timestamp + "-base"
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
							huh.NewOption("Quit (Clean up & Restore)", "quit"),
							huh.NewOption("Restart test (Retry same task)", "restart"),
							huh.NewOption("Start a new test (New task)", "new"),
						).
						Value(&action),
				),
			)

			err = form.Run()
			if err != nil {
				action = "quit" // Handle ctrl+c
			}

			// Clean up current run
			fmt.Print("Cleaning up sandboxes... ")
			eval.SafeCleanupABC(wtImplicit, wtExplicit, wtBase, tempDir)
			fmt.Println("Done")

			if action == "quit" {
				fmt.Println("\n✅ Eval task complete.")
				return
			}
			
			if action == "new" {
				break // Break inner loop, continue outer loop (New Task)
			}
			
			if action == "restart" {
				continue // Continue inner loop (Same Task)
			}
		}
	}
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
			case "text":
				if text, ok := getEventText(event); ok && text != "" {
					msg = text
				}
			case "tool_use":
				tool, _ := getEventTool(event)
				args := getEventToolArgs(event)
				display := tool
				if tool == "bash" {
					display = fmt.Sprintf("bash(%s)", args)
				} else if args != "" {
					display = fmt.Sprintf("%s(%s)", tool, args)
				}
				msg = fmt.Sprintf("\n🛠️  [TOOL] %s\n", display)
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
		
		trace.Tokens.Total = trace.Tokens.Input + trace.Tokens.Output

	case "tokens":
		// Alternative key used in Method 4 guide
		if input, ok := event["input"].(float64); ok {
			trace.Tokens.Input += int(input)
		}
		if output, ok := event["output"].(float64); ok {
			trace.Tokens.Output += int(output)
		}
		trace.Tokens.Total = trace.Tokens.Input + trace.Tokens.Output

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
			}
		}
		trace.Tokens.Total = trace.Tokens.Input + trace.Tokens.Output
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
		name := formatToolName(c)
		out += fmt.Sprintf("  %d. %s\n", i+1, name)
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

func createEvalWorktree(path string) error {
	exe, err := os.Executable()
	if err != nil { return err }
	
	if err := exec.Command("git", "worktree", "add", "-d", path, "HEAD").Run(); err != nil {
		return err
	}
	
	cmdRm := exec.Command("git", "rm", "--cached", "-r", ".")
	cmdRm.Dir = path
	_ = cmdRm.Run()

	cmdRead := exec.Command("git", "read-tree", "HEAD")
	cmdRead.Dir = path
	_ = cmdRead.Run()

	cmdCheckout := exec.Command("git", "checkout-index", "-a", "--force")
	cmdCheckout.Dir = path
	_ = cmdCheckout.Run()

	cmdConfig := exec.Command("git", "config", "core.fileMode", "false")
	cmdConfig.Dir = path
	_ = cmdConfig.Run()

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
	if apiKey != "" {
		env = append(env, "GOOGLE_GENERATIVE_AI_API_KEY="+apiKey)
	}
	if beadsDir != "" {
		env = append(env, "BEADS_DIR="+beadsDir)
	}
	return env
}

func createStash() (string, error) {
	statusCmd := exec.Command("git", "status", "--porcelain")
	out, _ := statusCmd.Output()
	if len(strings.TrimSpace(string(out))) == 0 { return "", nil }
	cmd := exec.Command("git", "stash", "push", "-m", "bd eval auto-stash")
	if err := cmd.Run(); err != nil { return "", err }
	listCmd := exec.Command("git", "stash", "list", "--format=%h", "-n", "1")
	hash, err := listCmd.Output()
	return strings.TrimSpace(string(hash)), err
}

func restoreStash(ref string) { _ = exec.Command("git", "stash", "pop", ref).Run() }

func handleLeftoverStashes() {
	out, err := exec.Command("git", "stash", "list").Output()
	if err != nil || len(out) == 0 { return }
	lines := strings.Split(string(out), "\n")
	var leftovers []string
	for _, line := range lines {
		if strings.Contains(line, "bd eval auto-stash") {
			leftovers = append(leftovers, line)
		}
	}
	if len(leftovers) == 0 { return }
	fmt.Printf("\n⚠️  Found %d unrecovered stash(es) from previous evaluation runs:\n", len(leftovers))
	for _, l := range leftovers { fmt.Printf("  %s\n", l) }
	var action string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What would you like to do with these stashes?").
				Options(
					huh.NewOption("Restore (Pop) the most recent one", "pop"),
					huh.NewOption("Delete (Clear) ALL eval stashes", "clear"),
					huh.NewOption("Ignore and proceed", "ignore"),
				).
				Value(&action),
		),
	)
	if err := form.Run(); err != nil { return }
	switch action {
	case "pop":
		for i, line := range lines {
			if strings.Contains(line, "bd eval auto-stash") {
				ref := fmt.Sprintf("stash@{%d}", i)
				fmt.Printf("Restoring %s... ", ref)
				if err := exec.Command("git", "stash", "pop", ref).Run(); err != nil {
					fmt.Printf("Conflict or Error: %v\n", err)
				} else { fmt.Println("Done") }
				break
			}
		}
	case "clear":
		fmt.Print("Clearing eval stashes... ")
		for i := len(lines) - 1; i >= 0; i-- {
			if strings.Contains(lines[i], "bd eval auto-stash") {
				_ = exec.Command("git", "stash", "drop", fmt.Sprintf("stash@{%d}", i)).Run()
			}
		}
		fmt.Println("Done")
	}
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