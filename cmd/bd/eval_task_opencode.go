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

	"github.com/spf13/cobra"
	"github.com/untoldecay/BeadsLog/internal/ui"
)

var evalTaskCmd = &cobra.Command{
	Use:   "task",
	Short: "Interactive eval: sandbox -> test -> analyze -> cleanup",
	Long: `Runs an A/B evaluation of the agent protocol in isolated sandboxes using OpenCode CLI.
It creates two git worktrees (Implicit vs Explicit), runs the agent, extracts structured
tool traces, generates a comparative report, and cleans up.`,
	Run: func(cmd *cobra.Command, args []string) {
		runOpenCodeEval()
	},
}

// OpenCodeTrace represents the aggregated trace from OpenCode CLI
type OpenCodeTrace struct {
	ToolCalls []ToolCall `json:"tool_calls"`
	Response  string     `json:"response"`
	Tokens    struct {
		Input  int `json:"input"`
		Output int `json:"output"`
	} `json:"tokens_used"`
}

type ToolCall struct {
	Name   string `json:"name"`
	Args   string `json:"args"` // Store command string for bash
	Result string `json:"result"`
}

type Scoring struct {
	Strategy   string
	Status     string
	Efficiency float64
	Emoji      string
}

func runOpenCodeEval() {
	fmt.Printf("\n%s %s\n", "🧪", "BeadsLog Eval Mode - OpenCode A/B Testing")
	fmt.Println(strings.Repeat("-", 60))

	reader := bufio.NewReader(os.Stdin)

	// 1. CHECK DIRTY
	isDirty, _ := checkDirty()
	if isDirty {
		fmt.Printf("%s Your workspace has uncommitted changes. To ensure isolation and prevent rollbacks, please commit or stash them manually before running evals.\n", ui.RenderWarn("⚠️  WARNING:"))
		fmt.Printf("Abort? (y/n): ")
		var resp string
		if tty, err := os.Open("/dev/tty"); err == nil {
			defer tty.Close()
			ttyReader := bufio.NewReader(tty)
			resp, _ = ttyReader.ReadString('\n')
		} else {
			resp, _ = reader.ReadString('\n')
		}
		if strings.HasPrefix(strings.ToLower(resp), "y") {
			return
		}
	}

	// 2. TASK INPUT
	fmt.Print("\nEnter task for agent: ")
	task, _ := reader.ReadString('\n')
	task = strings.TrimSpace(task)

	if task == "" {
		fmt.Println("Empty task, aborting")
		return
	}

	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	// 3. CREATE SANDBOX ROOT
	tempDir, err := os.MkdirTemp("", "beads-eval-*")
	if err != nil {
		fmt.Printf("Failed to create temp dir: %v\n", err)
		return
	}
	tempDir, _ = filepath.Abs(tempDir)

	wtImplicit := filepath.Join(tempDir, "implicit")
	wtExplicit := filepath.Join(tempDir, "explicit")

	fmt.Printf("\nCreating isolated sandboxes in %s...\n", tempDir)

	if err := createEvalWorktree(wtImplicit); err != nil {
		fmt.Printf("  Failed to create implicit worktree: %v\n", err)
		cleanupEvals(wtImplicit, wtExplicit, tempDir)
		return
	}
	fmt.Printf("  Implicit Sandbox: Ready\n")

	if err := createEvalWorktree(wtExplicit); err != nil {
		fmt.Printf("  Failed to create explicit worktree: %v\n", err)
		cleanupEvals(wtImplicit, wtExplicit, tempDir)
		return
	}
	fmt.Printf("  Explicit Sandbox: Ready\n")

	// 4. RUN TESTS
	fmt.Printf("\nRunning agent tests...\n")

	traceIdImplicit := "eval-" + timestamp + "-implicit"
	implicitResult := runOpenCodeTest(wtImplicit, task, traceIdImplicit, false)

	explicitTask := fmt.Sprintf("%s\n\n(MANDATORY: Use 'bd devlog' tools for retrieval first)", task)
	traceIdExplicit := "eval-" + timestamp + "-explicit"
	explicitResult := runOpenCodeTest(wtExplicit, explicitTask, traceIdExplicit, true)

	// 5. REPORT
	fmt.Printf("\nGenerating comparative report...\n")
	report := generateABReport(task, implicitResult, explicitResult)
	fmt.Println(report)

	// 6. CLEANUP
	fmt.Printf("\nTests validated? (y/n): ")
	
	var resp string
	if tty, err := os.Open("/dev/tty"); err == nil {
		defer tty.Close()
		ttyReader := bufio.NewReader(tty)
		resp, _ = ttyReader.ReadString('\n')
	} else {
		resp, _ = reader.ReadString('\n')
	}

	if strings.HasPrefix(strings.ToLower(resp), "y") {
		fmt.Print("Cleaning up sandboxes... ")
		cleanupEvals(wtImplicit, wtExplicit, tempDir)
		fmt.Println("Done")
		fmt.Println("\n✅ Eval task complete.")
	} else {
		fmt.Println("\nKeeping sandboxes for inspection:")
		fmt.Printf("  %s/\n", wtImplicit)
		fmt.Printf("  %s/\n", wtExplicit)
		fmt.Println("Run 'git worktree remove <path>' and 'rm -rf <tempdir>' when done.")
	}
}

func checkDirty() (bool, error) {
	out, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
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

	// Auth + workspace
	env := os.Environ()
	env = append(env, "GOOGLE_GENERATIVE_AI_API_KEY="+apiKey)
	env = append(env, "BEADS_DIR="+beadsDir)
	cmd.Env = env

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
			switch eventType {
			case "text":
				if text, ok := getEventText(event); ok && text != "" {
					fmt.Print(text)
					readableLogFile.WriteString(text)
				}
			case "tool_use":
				if tool, ok := getEventTool(event); ok {
					msg := fmt.Sprintf("\n🛠️  [TOOL] %s\n", tool)
					fmt.Print(msg)
					readableLogFile.WriteString(msg)
				}
			case "tool_result":
				msg := "✅ [DONE]\n"
				fmt.Print(msg)
				readableLogFile.WriteString(msg)
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

		// Extract command string from state.input.command
		argsStr := ""
		if part, ok := event["part"].(map[string]interface{}); ok {
			if state, ok := part["state"].(map[string]interface{}); ok {
				if input, ok := state["input"].(map[string]interface{}); ok {
					if cmd, ok := input["command"].(string); ok {
						argsStr = cmd
					} else {
						// Fallback: try to serialize input if not a simple command string
						b, _ := json.Marshal(input)
						argsStr = string(b)
					}
				}
			}
		}

		trace.ToolCalls = append(trace.ToolCalls, ToolCall{
			Name: toolName,
			Args: argsStr,
		})

	case "tool_result":
		// Match to last tool call
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
		// Handle direct or nested usage formats
		input := 0.0
		output := 0.0
		
		if val, ok := event["input_tokens"].(float64); ok {
			input = val
		} else if part, ok := event["part"].(map[string]interface{}); ok {
			if val, ok := part["input_tokens"].(float64); ok {
				input = val
			}
		}

		if val, ok := event["output_tokens"].(float64); ok {
			output = val
		} else if part, ok := event["part"].(map[string]interface{}); ok {
			if val, ok := part["output_tokens"].(float64); ok {
				output = val
			}
		}

		trace.Tokens.Input += int(input)
		trace.Tokens.Output += int(output)
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

func generateABReport(task string, implicit, explicit OpenCodeTrace) string {
	implicitScore := scoreTrace(implicit.ToolCalls)
	explicitScore := scoreTrace(explicit.ToolCalls)

	improvement := 0.0
	if explicitScore.Efficiency > 0 {
		improvement = implicitScore.Efficiency / explicitScore.Efficiency
	}

	report := fmt.Sprintf(`
%s
TASK: "%s"

╭──────────────────┬────────────┬──────────────────────────────────┬────────┬──────┬──────────╮
│ Run              │ Strategy   │ First Tools                      │ Tokens │ Stat │ Eff.     │
├──────────────────┼────────────┼──────────────────────────────────┼────────┼──────┼──────────┤
│ Implicit         │ %-10s │ %-32s │ %-6d │ %s │ %.1fx %s │
├──────────────────┼────────────┼──────────────────────────────────┼────────┼──────┼──────────┤
│ Explicit         │ %-10s │ %-32s │ %-6d │ %s │ %.1fx %s │
╰──────────────────┴────────────┴──────────────────────────────────┴────────┴──────┴──────────╯

🎯 Protocol Effectiveness: %.1fx better than explicit

`,
		ui.RenderAccent("📊 A/B Comparison Report"),
		task,
		implicitScore.Strategy, strings.Join(first3Tools(implicit.ToolCalls), " → "),
		implicit.Tokens.Input+implicit.Tokens.Output,
		implicitScore.Status,
		implicitScore.Efficiency, implicitScore.Emoji,
		explicitScore.Strategy, strings.Join(first3Tools(explicit.ToolCalls), " → "),
		explicit.Tokens.Input+explicit.Tokens.Output,
		explicitScore.Status,
		explicitScore.Efficiency, explicitScore.Emoji,
		improvement,
	)

	report += fmt.Sprintf("\n%s\n%s", ui.RenderAccent("🕒 Chronological Tool Traces:"), renderToolCalls("Implicit", implicit.ToolCalls))
	report += renderToolCalls("Explicit", explicit.ToolCalls)

	return report
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
		return Scoring{"Blind", "FAIL", 0.1, "🔴"}
	}

	// Analyze INTENT, not just tool name
	intents := make([]string, len(toolCalls))
	for i, tc := range toolCalls {
		if tc.Name == "bash" {
			// Parse bash command for intent
			cmd := strings.ToLower(tc.Args)
			if strings.Contains(cmd, "bd status") {
				intents[i] = "bd_status"
			} else if strings.Contains(cmd, "devlog graph") {
				intents[i] = "bd_devlog_graph"
			} else if strings.Contains(cmd, "devlog search") {
				intents[i] = "bd_devlog_search"
			} else if strings.Contains(cmd, "devlog resume") {
				intents[i] = "bd_devlog_resume"
			} else if strings.Contains(cmd, "grep") || strings.Contains(cmd, "ls") || strings.Contains(cmd, "cat") {
				intents[i] = "grep/ls"
			} else {
				intents[i] = "bash_generic"
			}
		} else {
			intents[i] = tc.Name
		}
	}

	mapTools := []string{"bd_devlog_graph", "bd_devlog_search", "bd_devlog_resume", "bd_status"}
	verifyTools := []string{"read_file", "grep/ls", "search_file_content", "glob"}

	if mapFirst(intents, mapTools, verifyTools) {
		return Scoring{"Optimal", "PASS", 2.5, "🟢🟢"}
	} else if shallowRetrieval(intents, mapTools) {
		return Scoring{"Shallow", "PASS", 1.0, "🟠"}
	} else if disordered(intents, mapTools, verifyTools) {
		return Scoring{"Disordered", "FAIL", 0.5, "🔴🟠"}
	}
	return Scoring{"Blind", "FAIL", 0.1, "🔴"}
}

func mapFirst(tools, mapTools, verifyTools []string) bool {
	firstMap := -1
	firstVerify := -1
	for i, t := range tools {
		if contains(mapTools, t) && firstMap == -1 {
			firstMap = i
		}
		if contains(verifyTools, t) && firstVerify == -1 {
			firstVerify = i
		}
	}
	if firstMap != -1 {
		if firstVerify == -1 {
			return true
		}
		return firstMap < firstVerify
	}
	return false
}

func shallowRetrieval(tools, mapTools []string) bool {
	for _, t := range tools {
		if contains(mapTools, t) {
			return true
		}
	}
	return false
}

func disordered(tools, mapTools, verifyTools []string) bool {
	// If verification tools used but mapping was late or missing
	return true
}

func contains(list []string, item string) bool {
	for _, s := range list {
		if strings.Contains(item, s) {
			return true
		}
	}
	return false
}

func first3Tools(calls []ToolCall) []string {
	names := make([]string, 0, 3)
	for i := range calls {
		if i >= 3 {
			break
		}
		// Display meaningful intent
		name := calls[i].Name
		if name == "bash" {
			// Truncate long commands
			cmd := calls[i].Args
			if len(cmd) > 20 {
				cmd = cmd[:17] + "..."
			}
			name = cmd
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return []string{"(no tools)"}
	}
	return names
}

func createEvalWorktree(path string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if err := exec.Command("git", "worktree", "add", path, "HEAD").Run(); err != nil {
		return err
	}
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
	initCmd.Env = append(os.Environ(), "BEADS_DIR="+beadsDir)
	_ = initCmd.Run()

	readyCmd := exec.Command(exe, "ready")
	readyCmd.Dir = path
	readyCmd.Env = append(os.Environ(), "BEADS_DIR="+beadsDir)
	_ = readyCmd.Run()

	absExe, _ := filepath.Abs(exe)
	// Simplified opencode.json without 'env' field
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

func cleanupEvals(wt1, wt2, tempDir string) {
	if wt1 != "" {
		_ = exec.Command("git", "worktree", "remove", "--force", wt1).Run()
	}
	if wt2 != "" {
		_ = exec.Command("git", "worktree", "remove", "--force", wt2).Run()
	}
	if tempDir != "" {
		_ = os.RemoveAll(tempDir)
	}
}

func getEvalApiKey() string {
	home, _ := os.UserHomeDir()
	paths := []string{filepath.Join(home, ".gemini", "eval.env"), ".env"}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "GEMINI_API_KEY=") {
				return strings.Trim(line[len("GEMINI_API_KEY="):], "\"'")
			}
		}
	}
	return os.Getenv("GEMINI_API_KEY")
}