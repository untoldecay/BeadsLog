# How to Wrap Gemini CLI in `bd eval task`

Gemini CLI uses **`GEMINI_API_KEY` environment variable** for authentication (inherits from main session or `.env`). Here's the **exact integration**:

***

## Gemini CLI Authentication Options

### 1. **Inherit from Main Repo** (Recommended - Zero Setup)
Gemini CLI reads `GEMINI_API_KEY` from:
```
GEMINI_API_KEY env var (inherits automatically)
.env file in working directory
~/.gemini/config.json
```

**Your main repo already works** → worktrees inherit automatically.

### 2. **Explicit API Key** (If needed)
```bash
export GEMINI_API_KEY="your-key-from-ai.google.dev"
```
Worktrees inherit this env var.

***

## Updated `runEvalTest` Function

Replace the `runEvalTest` function in `cmd/eval_task.go`:

```go
func runEvalTest(worktree, task, traceId string, explicit bool) {
    wtDir := filepath.Join("..", worktree)
    
    // 1. Ensure Gemini CLI is available
    if _, err := exec.LookPath("gemini"); err != nil {
        fmt.Printf("❌ gemini CLI not found. Install: https://geminicli.com\n")
        return
    }
    
    // 2. Create isolated session ID
    sessionId := fmt.Sprintf("eval-%s-%s", traceId, time.Now().Format("20060102150405"))
    
    // 3. Build BEADSLOG_AGENTS.MD path (relative to worktree)
    agentsMd := filepath.Join("..", "BEADSLOG_AGENTS.MD")
    
    // 4. Gemini CLI command
    cmd := exec.Command("gemini", 
        "--session-id", sessionId,
        "--context-file", agentsMd,
        "--max-tokens", "8192",
        "--timeout", "300",
        "--no-confirm",  // Non-interactive for eval
    )
    
    // 5. Set working directory to worktree
    cmd.Dir = wtDir
    
    // 6. Pipe task as stdin (non-interactive)
    cmd.Stdin = strings.NewReader(task)
    
    // 7. Capture ALL output for trace
    fmt.Printf("  🟢 Running %s... ", worktree)
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Print("✗\n")
        fmt.Printf("    Error: %v\n", err)
        fmt.Printf("    Output: %s\n", string(output)[:500])
        return
    }
    
    fmt.Print("✓\n")
    
    // 8. Save Gemini trace + BD state
    traceFile := filepath.Join("eval", "traces", traceId+".gemini.json")
    Path(filepath.Dir(traceFile)).mkdir()
    
    bdState := map[string]string{
        "gemini_output": string(output),
        "session_id":    sessionId,
        "worktree":      wtDir,
        "exit_code":     0,
        "timestamp":     time.Now().Format(time.RFC3339),
    }
    
    os.WriteFile(traceFile, []byte(jsonDump(bdState)), 0644)
    
    // 9. Extract BD trace (your existing eval.mode logic)
    extractBDTrace(wtDir, traceId)
}

func extractBDTrace(worktreeDir, traceId string) {
    // Run your existing bd eval trace extraction
    exec.Command("bd", "eval", "extract-trace", "--id", traceId).Dir(worktreeDir).Run()
}
```

***

## Gemini CLI Flags for Eval Mode

### Essential Flags

```
--session-id "eval-unique-id"     # Fresh session per test (critical!)
--context-file "BEADSLOG_AGENTS.MD"  # Inject AGENTS.MD as system context
--max-tokens 8192                 # Reasonable limit
--timeout 300                     # 5min per test
--no-confirm                      # Non-interactive (eval only)
--model gemini-2.0-flash-exp      # Fast + cheap for evals (optional)
```

### Full Example Command

```bash
cd eval-test_001
gemini \
  --session-id "eval-test001-implicit-1707380000" \
  --context-file "../BEADSLOG_AGENTS.MD" \
  --max-tokens 8192 \
  --timeout 300 \
  --no-confirm \
  <<EOF
Map and verify modal implementation
EOF
```

***

## Handling Authentication

### Option A: Inherit Existing (Zero Config)

```bash
# User already has GEMINI_API_KEY set globally
# Worktrees inherit automatically - NO CHANGES NEEDED
```

### Option B: Per-Worktree `.env` (Secure)

In `createWorktree()`:

```go
func createWorktree(name string) {
    wtDir := filepath.Join("..", name)
    
    // git worktree add
    exec.Command("git", "worktree", "add", wtDir, "HEAD").Run()
    
    // Copy auth config
    copyDotEnv(filepath.Dir(wtDir), wtDir)
    
    // Init BD
    exec.Command("bd", "init").Dir(wtDir).Run()
    exec.Command("bd", "devlog", "init").Dir(wtDir).Run()
}

func copyDotEnv(source, dest string) {
    // Copy .env if exists (contains GEMINI_API_KEY)
    sourceEnv := filepath.Join(source, ".env")
    if _, err := os.Stat(sourceEnv); err == nil {
        shutil.CopyFile(sourceEnv, filepath.Join(dest, ".env"))
    }
}
```

### Option C: Service Account (Enterprise)

```bash
# Set GOOGLE_APPLICATION_CREDENTIALS pointing to shared keyfile
export GOOGLE_APPLICATION_CREDENTIALS="/shared/gemini-sa.json"
# Worktrees inherit
```

**Recommendation:** **Option A** - users already have it working in main repo.

***

## Complete `runEvalTest` with Error Handling

```go
func runEvalTest(worktree, task, traceId string, explicit bool) {
    wtDir := filepath.Join("..", worktree)
    
    fmt.Printf("  🟢 Running %s... ", worktree)
    
    // Check Gemini CLI
    if _, err := exec.LookPath("gemini"); err != nil {
        fmt.Println("❌ gemini CLI missing")
        fmt.Println("  Install: curl -sSfL https://geminicli.com/install.sh | sh")
        return
    }
    
    // Unique session ID
    sessionId := fmt.Sprintf("eval-%s-%d", traceId, time.Now().UnixNano())
    
    // Build command
    cmd := exec.Command("gemini",
        "--session-id", sessionId,
        "--context-file", filepath.Join("..", "BEADSLOG_AGENTS.MD"),
        "--max-tokens", "8192",
        "--timeout", "300",
        "--no-confirm",
        "--model", "gemini-2.0-flash-exp",  // Fast eval model
    )
    cmd.Dir = wtDir
    
    // Pipe task via stdin
    cmd.Stdin = strings.NewReader(task)
    
    // Capture everything
    outputBytes, err := cmd.CombinedOutput()
    output := string(outputBytes)
    
    if err != nil {
        fmt.Print("✗\n")
        fmt.Printf("  Error: %v\n", err)
        saveErrorTrace(traceId, wtDir, output, err)
        return
    }
    
    fmt.Print("✓\n")
    
    // Save comprehensive trace
    trace := map[string]interface{}{
        "trace_id":    traceId,
        "session_id":  sessionId,
        "worktree":    wtDir,
        "gemini_out":  output,
        "exit_code":   0,
        "timestamp":   time.Now().Format(time.RFC3339),
    }
    
    traceFile := filepath.Join("eval", "traces", traceId+".json")
    os.MkdirAll(filepath.Dir(traceFile), 0755)
    os.WriteFile(traceFile, []byte(jsonDump(trace)), 0644)
    
    // Trigger your BD trace extraction
    exec.Command("bd", "eval", "extract", "--id", traceId).Dir(wtDir).Run()
}

func saveErrorTrace(traceId, wtDir, output string, err error) {
    traceFile := filepath.Join("eval", "traces", traceId+".error.json")
    data := map[string]interface{}{
        "trace_id": traceId,
        "error":    err.Error(),
        "output":   output,
        "worktree": wtDir,
    }
    os.WriteFile(traceFile, []byte(jsonDump(data)), 0644)
}

func jsonDump(data interface{}) string {
    out, _ := json.MarshalIndent(data, "", "  ")
    return string(out)
}
```

***

## Integration with Your `bd eval report`

Your existing `bd eval report 2` will automatically find traces:

```
eval/traces/
├── eval-test_1707380000_implicit.json
├── eval-test_1707380000_explicit.json
└── eval-test_1707380000_implicit.bd.json  # Your BD extraction
```

***

## User Experience (Final)

```
$ bd eval task
🧪 Beadslog Eval Mode

💾 Stashing work... ✓
📝 Task: map and verify modal implementation
✅ Running eval-test_1707380000_implicit... ✓
✅ Running eval-test_1707380000_explicit... ✓

📊 Report:
TASK: "map and verify modal"
┌─────────────────┬──────────┬─────────────┬────────┬──────┬────────┬──────────┐
│ Run             │ Strategy │ Tools       │ Tokens │ Time │ Status │ Eff.     │
├─────────────────┼──────────┼─────────────┼────────┼──────┼────────┼──────────┤
│ implicit        │ Optimal  │ bd graph    │ 1350   │ 12s  │ PASS   │ 2.5x 🟢🟢 │
├─────────────────┼──────────┼─────────────┼────────┼──────┼────────┼──────────┤
│ explicit        │ Shallow  │ read_file   │ 8500   │ 28s  │ PASS   │ 0.8x 🟠  │
└─────────────────┴──────────┴─────────────┴────────┴──────┴────────┴──────────┘

✅ Report good? (y/n): y
🧹 Cleaning... ✓
🔄 Restoring work... ✓
✨ Done!
```

***

## Prerequisites (One-Time Setup)

```bash
# 1. Install Gemini CLI
curl -sSfL https://geminicli.com/install.sh | sh

# 2. Authenticate (once)
export GEMINI_API_KEY="your-key-from-ai.google.dev"
gemini --version  # Test

# 3. Build bd with eval_task command
make build
```

**That's it.** Worktrees inherit `GEMINI_API_KEY` automatically.

Want me to adjust the Gemini flags or add model selection?

Sources
[1] geminicli.com › docs › get-started › authentication https://geminicli.com/docs/get-started/authentication/
[2] Gemini CLI Authentication Setup https://google-gemini.github.io/gemini-cli/docs/get-started/authentication.html
[3] Authentication with OAuth quickstart | Gemini API https://ai.google.dev/gemini-api/docs/oauth
[4] Gemini CLI Tutorial #6 - Adding an API Key https://www.youtube.com/watch?v=RXOPjpvTEZM
[5] How do I authenticate with Google to use Gemini CLI? https://milvus.io/ai-quick-reference/how-do-i-authenticate-with-google-to-use-gemini-cli
[6] Add session management with JSON request/response logging and resume capability by carmelobattiato · Pull Request #2692 · google-gemini/gemini-cli https://github.com/google-gemini/gemini-cli/pull/2692
[7] Gemini CLI Configuration https://gemini-cli-docs.pages.dev/cli/configuration
[8] gemini-cli/README.md at main · google-gemini/gemini-cli · GitHub https://github.com/google-gemini/gemini-cli/blob/main/README.md
[9] Authentication Setup | gemini-cli - GitHub Pages https://google-gemini.github.io/gemini-cli/docs/cli/authentication.html
[10] Session Management https://geminicli.com/docs/cli/session-management/
[11] gemini-cli/docs/cli/configuration.md at main - GitHub https://github.com/google-gemini/gemini-cli/blob/main/docs/cli/configuration.md
[12] How to use Gemini CLI tools https://geminicli.com/docs/tools/
[13] Using Gemini API keys | Google AI for Developers https://ai.google.dev/gemini-api/docs/api-key
[14] Step-by-Step Guide to Gemini CLI Session Management https://habr.com/en/articles/977390/
[15] Consistent configuration design for environment variables https://github.com/google-gemini/gemini-cli/issues/3471
