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
)

var evalTaskCmd = &cobra.Command{
	Use:   "task",
	Short: "Interactive eval: sandbox -> test -> analyze -> cleanup",
	Long: `Runs an A/B evaluation of the agent protocol in isolated sandboxes.
It creates two git worktrees (Implicit vs Explicit), runs the gemini agent,
generates a comparative report, and cleans up.`,
	Run: func(cmd *cobra.Command, args []string) {
		runEvalTask()
	},
}

func runEvalTask() {
	fmt.Printf("\n%s %s\n", "🧪", "BeadsLog Eval Mode - Interactive Testing")
	fmt.Println(strings.Repeat("-", 60))

	// 1. STASH current work (non-destructive)
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

	// 3. CREATE WORKTREES
	timestamp := time.Now().Unix()
	wtImplicit := fmt.Sprintf("eval-test_%d_implicit", timestamp)
	wtExplicit := fmt.Sprintf("eval-test_%d_explicit", timestamp)

	// Get absolute paths for worktrees (siblings to main repo)
	cwd, _ := os.Getwd()
	parentDir := filepath.Dir(cwd)
	wtImplicitPath := filepath.Join(parentDir, wtImplicit)
	wtExplicitPath := filepath.Join(parentDir, wtExplicit)

	fmt.Printf("\nCreating isolated sandboxes...\n")
	
	if err := createEvalWorktree(wtImplicitPath); err != nil {
		fmt.Printf("  Failed to create implicit worktree: %v\n", err)
		cleanupEvals(wtImplicitPath, wtExplicitPath, stashRef)
		return
	}
	fmt.Printf("  Implicit: %s (Ready)\n", wtImplicit)

	if err := createEvalWorktree(wtExplicitPath); err != nil {
		fmt.Printf("  Failed to create explicit worktree: %v\n", err)
		cleanupEvals(wtImplicitPath, wtExplicitPath, stashRef)
		return
	}
	fmt.Printf("  Explicit: %s (Ready)\n", wtExplicit)

	// 4. RUN TESTS
	fmt.Printf("\nRunning agent tests (this may take a minute)...\n")

	// Run 1: Implicit (Standard Protocol)
	traceId1 := fmt.Sprintf("%d-implicit", timestamp)
	runEvalSession(wtImplicitPath, task, traceId1, false)

	// Run 2: Explicit (Direct Tool Instructions)
	traceId2 := fmt.Sprintf("%d-explicit", timestamp)
	explicitTask := task + " (MANDATORY: Use 'bd devlog' tools for retrieval first)"
	runEvalSession(wtExplicitPath, explicitTask, traceId2, true)

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
		cleanupEvals(wtImplicitPath, wtExplicitPath, "")
		fmt.Println("Done")
	} else {
		fmt.Println("\nKeeping sandboxes for inspection:")
		fmt.Printf("  %s/\n", wtImplicitPath)
		fmt.Printf("  %s/\n", wtExplicitPath)
		fmt.Println("Run 'git worktree remove <path>' when done.")
	}

	// 7. RESTORE ORIGINAL STATE
	if stashRef != "" {
		fmt.Print("Restoring your work... ")
		restoreStash(stashRef)
		fmt.Println("Done")
	}

	fmt.Println("\nEval task complete.")
}

func createStash() (string, error) {
	// Check if dirty
	statusCmd := exec.Command("git", "status", "--porcelain")
	out, _ := statusCmd.Output()
	if len(strings.TrimSpace(string(out))) == 0 {
		return "", nil
	}

	cmd := exec.Command("git", "stash", "push", "-m", "bd eval auto-stash")
	if err := cmd.Run(); err != nil {
		return "", err
	}

	// Get latest stash hash
	listCmd := exec.Command("git", "stash", "list", "--format=%h", "-n", "1")
	hash, err := listCmd.Output()
	return strings.TrimSpace(string(hash)), err
}

func restoreStash(ref string) {
	_ = exec.Command("git", "stash", "pop", ref).Run()
}

func createEvalWorktree(path string) error {
	// Get current executable path
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exeDir := filepath.Dir(exe)

	// 1. git worktree add
	if err := exec.Command("git", "worktree", "add", path, "HEAD").Run(); err != nil {
		return err
	}

	// 2. Initialize Beads with Isolation
	// We MUST set BEADS_DIR to ensure it doesn't leak to main repo
	beadsDir := filepath.Join(path, ".beads")
	
	// Ensure we remove any inherited .beads from worktree checkout (if somehow tracked)
	_ = os.RemoveAll(beadsDir)

	// Copy AGENT.md to worktree root (ensures protocol is available)
	if data, err := os.ReadFile("AGENT.md"); err == nil {
		_ = os.WriteFile(filepath.Join(path, "AGENT.md"), data, 0644)
	}

	// Add current exe dir to PATH so it can find 'bd' if called simply
	newPath := exeDir + string(os.PathListSeparator) + os.Getenv("PATH")

	// bd init --force (ensures fresh DB even if redirect/files present)
	initCmd := exec.Command(exe, "init", "--quiet", "--force")
	initCmd.Dir = path
	initCmd.Env = append(os.Environ(), "BEADS_DIR="+beadsDir, "PATH="+newPath)
	out, err := initCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("bd init failed: %v, output: %s", err, string(out))
	}

	// bd ready (Unlocks Full Protocol)
	readyCmd := exec.Command(exe, "ready")
	readyCmd.Dir = path
	readyCmd.Env = append(os.Environ(), "BEADS_DIR="+beadsDir, "PATH="+newPath)
	if out, err := readyCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("bd ready failed: %v, output: %s", err, string(out))
	}

	return nil
}

func runEvalSession(wtPath, task, traceId string, explicit bool) {
	fmt.Printf("  Executing %s... ", filepath.Base(wtPath))
	
	// Check for gemini CLI
	geminiPath, err := exec.LookPath("gemini")
	if err != nil {
		fmt.Println("Failed (gemini CLI missing)")
		return
	}

	// Isolated session ID
	sessionId := fmt.Sprintf("eval-%s-%d", traceId, time.Now().UnixNano())
	
	// Set BEADS_DIR for the agent's subprocesses
	beadsDir := filepath.Join(wtPath, ".beads")
	
	// Add current exe dir to PATH
	exe, _ := os.Executable()
	exeDir := filepath.Dir(exe)
	newPath := exeDir + string(os.PathListSeparator) + os.Getenv("PATH")

	// Build environment
	env := os.Environ()
	env = append(env, "BEADS_DIR="+beadsDir)
	env = append(env, "PATH="+newPath)
	
	// Pass GEMINI_API_KEY explicitly
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		env = append(env, "GEMINI_API_KEY="+key)
	}

	// Configure Gemini CLI
	// -p: prompt
	// -y: automatically accept actions (YOLO)
	cmd := exec.Command(geminiPath,
		"-p", task,
		"-y",
	)
	cmd.Dir = wtPath
	cmd.Env = env
	
	output, err := cmd.CombinedOutput()
	
	// Save Trace
	saveTrace(traceId, wtPath, sessionId, string(output), err)

	if err != nil {
		fmt.Printf("Failed (Error: %v)\n", err)
		if len(output) > 0 {
			fmt.Printf("      Output Snippet: %s\n", string(output)[:200])
		}
	} else {
		fmt.Print("Done\n")
	}
}

func saveTrace(traceId, wtPath, sessionId, output string, runErr error) {
	status := "success"
	if runErr != nil {
		status = "error"
	}

	traceData := map[string]interface{}{
		"trace_id":   traceId,
		"session_id": sessionId,
		"worktree":   wtPath,
		"status":     status,
		"output":     output,
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	if runErr != nil {
		traceData["error"] = runErr.Error()
	}

	// Save to eval/traces/
	traceFile := filepath.Join("eval", "traces", traceId+".json")
	_ = os.MkdirAll(filepath.Dir(traceFile), 0755)
	
	data, _ := json.MarshalIndent(traceData, "", "  ")
	_ = os.WriteFile(traceFile, data, 0644)
}

func generateABReport(id1, id2 string) (string, error) {
	// Get current executable path
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	// Call existing bd eval report for the last 2 sessions
	cmd := exec.Command(exe, "eval", "report", "2")
	out, err := cmd.Output()
	return string(out), err
}

func cleanupEvals(wt1, wt2, stashRef string) {
	if wt1 != "" {
		_ = exec.Command("git", "worktree", "remove", "--force", wt1).Run()
	}
	if wt2 != "" {
		_ = exec.Command("git", "worktree", "remove", "--force", wt2).Run()
	}
	if stashRef != "" {
		restoreStash(stashRef)
	}
}
