package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/untoldecay/BeadsLog/internal/audit"
	"github.com/untoldecay/BeadsLog/internal/ui"
)

var evalTaskCmd = &cobra.Command{
	Use:   "task",
	Short: "Interactive eval: sandbox -> test -> analyze -> cleanup",
	Long: `Runs an A/B evaluation of the agent protocol in isolated sandboxes.
It creates two git worktrees (Implicit vs Explicit) in a temporary directory, 
runs the gemini agent using an API key (bypassing OAuth), extracts the 
audit logs into the main repository, generates a comparative report, 
and cleans up.`,
	Run: func(cmd *cobra.Command, args []string) {
		runEvalTask()
	},
}

func init() {
	evalCmd.AddCommand(evalTaskCmd)
}

func runEvalTask() {
	fmt.Printf("\n%s %s\n", "🧪", "BeadsLog Eval Mode - Interactive Testing")
	fmt.Println(strings.Repeat("-", 60))

	// 1. STASH current work
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

	// 2. GET TASK FROM USER
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nEnter task for agent: ")
	task, _ := reader.ReadString('\n')
	task = strings.TrimSpace(task)

	if task == "" {
		fmt.Println("Empty task, aborting")
		if stashRef != "" {
			restoreStash(stashRef)
		}
		return
	}

	// 3. CREATE SANDBOX ROOT
	tempDir, err := os.MkdirTemp("", "beads-eval-*")
	if err != nil {
		fmt.Printf("Failed to create temp dir: %v\n", err)
		if stashRef != "" { restoreStash(stashRef) }
		return
	}
	tempDir, _ = filepath.Abs(tempDir)

	wtImplicit := filepath.Join(tempDir, "implicit")
	wtExplicit := filepath.Join(tempDir, "explicit")

	fmt.Printf("\nCreating isolated sandboxes in %s...\n", tempDir)
	
	if err := createEvalWorktree(wtImplicit); err != nil {
		fmt.Printf("  Failed to create implicit worktree: %v\n", err)
		cleanupEvals(wtImplicit, wtExplicit, stashRef, tempDir)
		return
	}
	fmt.Printf("  Implicit Sandbox: Ready\n")

	if err := createEvalWorktree(wtExplicit); err != nil {
		fmt.Printf("  Failed to create explicit worktree: %v\n", err)
		cleanupEvals(wtImplicit, wtExplicit, stashRef, tempDir)
		return
	}
	fmt.Printf("  Explicit Sandbox: Ready\n")

	// 4. RUN TESTS
	fmt.Printf("\nRunning agent tests (streaming to eval/traces/)... \n")
	fmt.Printf("%s %s %s\n\n", ui.RenderAccent("💡 TIP:"), "Monitor progress in real-time by running:", ui.RenderAccent("tail -f eval/traces/*.log"))

	// Run 1: Implicit
	traceId1 := fmt.Sprintf("%d-implicit", time.Now().Unix())
	runEvalSession(wtImplicit, task, traceId1, false)

	// Run 2: Explicit
	traceId2 := fmt.Sprintf("%d-explicit", time.Now().Unix())
	explicitTask := task + " (MANDATORY: Use 'bd devlog' tools for retrieval first)"
	runEvalSession(wtExplicit, explicitTask, traceId2, true)

	fmt.Println("\nTests complete!")

	// 5. GENERATE REPORT
	fmt.Printf("\nGenerating comparative report...\n")
	report, err := generateABReport(traceId1, traceId2)
	if err != nil {
		fmt.Printf("  Failed to generate report: %v\n", err)
	} else {
		fmt.Println(report)
	}

	// 6. USER VALIDATION & CLEANUP
	fmt.Printf("\nTests validated? (y/n): ")
	response, _ := reader.ReadString('\n')
	
	if strings.HasPrefix(strings.ToLower(response), "y") {
		fmt.Print("Cleaning up sandboxes... ")
		cleanupEvals(wtImplicit, wtExplicit, "", tempDir)
		fmt.Println("Done")
	} else {
		fmt.Println("\nKeeping sandboxes for inspection:")
		fmt.Printf("  %s/\n", wtImplicit)
		fmt.Printf("  %s/\n", wtExplicit)
		fmt.Println("Run 'git worktree remove <path>' and 'rm -rf <tempdir>' when done.")
	}

	// 7. RESTORE ORIGINAL STATE
	if stashRef != "" {
		fmt.Print("Restoring your work... ")
		restoreStash(stashRef)
		fmt.Println("Done")
	}

	fmt.Println("\n✅ Eval task complete.")
}

func createStash() (string, error) {
	statusCmd := exec.Command("git", "status", "--porcelain")
	out, _ := statusCmd.Output()
	if len(strings.TrimSpace(string(out))) == 0 {
		return "", nil
	}

	cmd := exec.Command("git", "stash", "push", "-m", "bd eval auto-stash")
	if err := cmd.Run(); err != nil {
		return "", err
	}

	listCmd := exec.Command("git", "stash", "list", "--format=%h", "-n", "1")
	hash, err := listCmd.Output()
	return strings.TrimSpace(string(hash)), err
}

func restoreStash(ref string) {
	_ = exec.Command("git", "stash", "pop", ref).Run()
}

func createEvalWorktree(path string) error {
	exe, err := os.Executable()
	if err != nil { return err }
	exeDir := filepath.Dir(exe)

	if err := exec.Command("git", "worktree", "add", path, "HEAD").Run(); err != nil {
		return err
	}

	beadsDir := filepath.Join(path, ".beads")
	_ = os.RemoveAll(beadsDir)

	newPath := exeDir + string(os.PathListSeparator) + os.Getenv("PATH")

	// bd init
	initCmd := exec.Command(exe, "init", "--quiet", "--force")
	initCmd.Dir = path
	initCmd.Env = append(os.Environ(), "BEADS_DIR="+beadsDir, "PATH="+newPath)
	if out, err := initCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("bd init failed: %v, output: %s", err, string(out))
	}

	// bd ready
	readyCmd := exec.Command(exe, "ready")
	readyCmd.Dir = path
	readyCmd.Env = append(os.Environ(), "BEADS_DIR="+beadsDir, "PATH="+newPath)
	if out, err := readyCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("bd ready failed: %v, output: %s", err, string(out))
	}

	// bd eval start (Enables logging in sandbox)
	startCmd := exec.Command(exe, "eval", "start", "--quiet")
	startCmd.Dir = path
	startCmd.Env = append(os.Environ(), "BEADS_DIR="+beadsDir, "PATH="+newPath)
	if out, err := startCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("bd eval start failed: %v, output: %s", err, string(out))
	}

	return nil
}

func getEvalApiKey() string {
	home, _ := os.UserHomeDir()
	paths := []string{
		filepath.Join(home, ".gemini", "eval.env"),
		".env",
	}
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

// ProtocolFilterWriter suppresses protocol noise from logs
type ProtocolFilterWriter struct {
	w io.Writer
}

func (pw *ProtocolFilterWriter) Write(p []byte) (n int, err error) {
	s := string(p)
	if strings.Contains(s, "Hook system message: # BEADSLOG AGENTS.MD") {
		return len(p), nil
	}
	// Filter out common noise
	if strings.Contains(s, "Loading extension:") || 
	   strings.Contains(s, "Server '") || 
	   strings.Contains(s, "Hook registry initialized") ||
	   strings.Contains(s, "YOLO mode is enabled") {
		return len(p), nil
	}
	return pw.w.Write(p)
}

func runEvalSession(wtPath, task, traceId string, explicit bool) {
	fmt.Printf("  Executing %s... ", filepath.Base(wtPath))
	
	geminiPath, err := exec.LookPath("gemini")
	if err != nil {
		fmt.Println("Failed (gemini CLI missing)")
		return
	}

	apiKey := getEvalApiKey()
	if apiKey == "" {
		fmt.Println("Failed (GEMINI_API_KEY not found)")
		return
	}

	beadsDir := filepath.Join(wtPath, ".beads")
	exe, _ := os.Executable()
	exeDir := filepath.Dir(exe)
	newPath := exeDir + string(os.PathListSeparator) + os.Getenv("PATH")

	// Create log file early
	debugFile := filepath.Join("eval", "traces", traceId+".log")
	_ = os.MkdirAll(filepath.Dir(debugFile), 0755)
	logF, _ := os.Create(debugFile)
	defer logF.Close()

	env := os.Environ()
	env = append(env, "BEADS_DIR="+beadsDir)
	env = append(env, "PATH="+newPath)
	env = append(env, "GEMINI_API_KEY="+apiKey)
	env = append(env, "GEMINI_OAUTH_TOKEN=")
	env = append(env, "GEMINI_ACCESS_TOKEN=")
	env = append(env, "BD_EVAL_MODE=true")
	env = append(env, "BD_EVAL_SESSION_ID=eval-"+traceId)

	cmd := exec.Command(geminiPath, "-p", task, "-y", "--extensions", "")
	cmd.Dir = wtPath
	cmd.Env = env
	
	// Stream output to log file via filter
	filter := &ProtocolFilterWriter{w: logF}
	cmd.Stdout = filter
	cmd.Stderr = filter
	
	err = cmd.Run()
	
	// EXTRACT AUDIT LOG FROM SANDBOX
	extractAuditLog(wtPath, traceId, task)

	if err != nil {
		fmt.Printf("Failed (Error: %v). Log: %s\n", err, debugFile)
	} else {
		fmt.Print("Done\n")
	}
}

func extractAuditLog(wtPath, traceId, task string) {
	sandboxAuditPath := filepath.Join(wtPath, ".beads", "interactions.jsonl")
	if _, err := os.Stat(sandboxAuditPath); os.IsNotExist(err) {
		return
	}

	f, err := os.Open(sandboxAuditPath)
	if err != nil { return }
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry audit.Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}
		entry.EvalSessionID = "eval-" + traceId
		if entry.Extra == nil {
			entry.Extra = make(map[string]any)
		}
		entry.Extra["scenario_id"] = task
		_, _ = audit.Append(&entry)
	}
}

func generateABReport(id1, id2 string) (string, error) {
	time.Sleep(500 * time.Millisecond)
	exe, err := os.Executable()
	if err != nil { return "", err }
	cmd := exec.Command(exe, "eval", "report", "2")
	env := os.Environ()
	for i, e := range env {
		if strings.HasPrefix(e, "BEADS_DIR=") { env[i] = "BEADS_DIR=" }
	}
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func cleanupEvals(wt1, wt2, stashRef, tempDir string) {
	if wt1 != "" { _ = exec.Command("git", "worktree", "remove", "--force", wt1).Run() }
	if wt2 != "" { _ = exec.Command("git", "worktree", "remove", "--force", wt2).Run() }
	if tempDir != "" { _ = os.RemoveAll(tempDir) }
	if stashRef != "" { restoreStash(stashRef) }
}
