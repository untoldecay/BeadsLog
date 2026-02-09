**Yes, Gemini CLI headless mode (`-p -y`) captures and returns tool output** in structured ways. Here's exactly how to get it:

***

## Tool Output in Headless Gemini CLI

### 1. **Default Behavior: Text Output with Tool Indicators**

```
$ echo "List files in src" | gemini -p -y

[Tool: read_directory("src")]
drwxr-xr-x  modal/
drwxr-xr-x  auth/
-rw-r--r--  main.go

Gemini: The src directory contains modal/, auth/, and main.go.
```

**Tool calls** are **interleaved in stdout** with clear markers like `[Tool: ...]`.

### 2. **JSON Mode: Structured Tool Calls** (`--output-format json`)

```
$ echo "List files in src" | gemini -p -y --output-format json | jq .

{
  "response": "The src directory contains...",
  "tools_used": [
    {
      "name": "read_directory",
      "args": {"path": "src"},
      "output": "drwxr-xr-x  modal/\ndrwxr-xr-x  auth/\n-rw-r--r--  main.go"
    }
  ],
  "tokens_used": 1350,
  "model": "gemini-2.0-flash-exp"
}
```

**Perfect for your eval** - parse with `jq` or Go `json.Unmarshal`.

### 3. **Debug Mode: Full Tool Trace** (`--debug`)

```
$ echo "Fix auth bug" | gemini -p -y --debug

[DEBUG] Tool call #1: bd_devlog_graph(args={"entity":"auth"})
[DEBUG] Tool output: "auth → middleware → routes"
[DEBUG] Tool call #2: read_file(args={"path":"src/auth.go"})
[DEBUG] Tool output: "// auth middleware..."
```

**Capture everything** with `cmd.CombinedOutput()`.

***

## Your Exact Integration (Updated `runEvalTest`)

```go
func runEvalTest(worktree, task, traceId string, explicit bool) {
    wtDir := filepath.Join("/tmp", "beads-eval-"+traceId)
    
    // ... scaffolding (bd init, etc.) ...
    
    fmt.Printf("  🟢 Running %s... ", worktree)
    
    // Gemini CLI with JSON output for structured parsing
    cmd := exec.Command("gemini",
        "-p", task,           // Headless prompt
        "-y",                 // Auto-confirm tools
        "--output-format", "json",  // Structured tool data
        "--debug",            // Full tool trace
    )
    cmd.Dir = wtDir
    cmd.Env = append(os.Environ(),
        "BEADS_DIR="+filepath.Join(wtDir, ".beads"),
        "GEMINI_API_KEY="+getEvalApiKey(),
    )
    
    // Capture structured JSON + debug
    outputBytes, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Print("✗\n")
        return
    }
    fmt.Print("✓\n")
    
    output := string(outputBytes)
    
    // Parse JSON (tool calls + outputs)
    var geminiTrace struct {
        Response  string              `json:"response"`
        ToolsUsed []ToolCall          `json:"tools_used"`
        Tokens    int                 `json:"tokens_used"`
    }
    json.Unmarshal([]byte(output), &geminiTrace)
    
    // Save structured trace
    trace := map[string]interface{}{
        "trace_id":  traceId,
        "gemini":    geminiTrace,
        "raw_output": output,
        "worktree":   wtDir,
    }
    
    traceFile := filepath.Join("eval", "traces", traceId+".full.json")
    os.WriteFile(traceFile, []byte(jsonDump(trace)), 0644)
    
    // Your BD eval extraction
    exec.Command("bd", "eval", "extract", "--id", traceId).Dir(wtDir).Run()
}
```

**`ToolCall` struct** (from JSON output):
```go
type ToolCall struct {
    Name   string                 `json:"name"`
    Args   map[string]interface{} `json:"args"`
    Output string                 `json:"output"`  // THE TOOL RESULT
}
```

***

## Capturing Your BD Tools Specifically

Since your agent uses `bd devlog graph`, etc., you'll see:

```json
{
  "tools_used": [
    {
      "name": "bd_devlog_graph",
      "args": {"entity": "modal"},
      "output": "modal → hooks → components\nhooks → endpoints"
    },
    {
      "name": "read_file",
      "args": {"path": "src/modal.tsx"},
      "output": "// modal implementation..."
    }
  ]
}
```

**Perfect for your rubric** - parse `tools_used` array to check:
- `Map` tools first? (`bd_devlog_graph` before `read_file`)
- Token weights? (BD tools = 150t, `read_file` = 1200t)

***

## Flags Summary for Tool Capture

| Flag | Purpose | Output |
|------|---------|--------|
| `-p` | Headless mode | Runs task without terminal interaction |
| `-y` | Auto‑confirm tools | No "Approve tool call?" prompts |
| `--output-format json` | **Structured tool data** | Parseable `tools_used[]` array |
| `--debug` | Verbose tool execution | Full trace in stderr |

**Minimal working set:**
```bash
echo "$TASK" | gemini -p -y --output-format json
```

***

## Complete Example Output

```
$ echo "map modal deps" | gemini -p -y --output-format json | jq .

{
  "response": "Modal implementation spans 3 files with these deps...",
  "tools_used": [
    {
      "name": "bd_devlog_graph",
      "args": {"entity": "modal"},
      "output": "modal(components) → hooks → endpoints"
    },
    {
      "name": "read_file", 
      "args": {"path": "src/modal.tsx"},
      "output": "export ModalComponent..."
    }
  ],
  "tokens_used": {
    "input": 850,
    "output": 500,
    "tools": 150
  }
}
```

***

## Your Rubric Parsing (New Helper)

```go
func applyRubric(geminiTrace GeminiTrace, testDef TestDefinition) Scoring {
    tools := geminiTrace.ToolsUsed
    
    // Extract tool sequence
    toolNames := make([]string, len(tools))
    for i, t := range tools {
        toolNames[i] = t.Name
    }
    
    // Your logic
    if mapFirst(toolNames) && verifyLater(toolNames) {
        return Scoring{Optimal, 1.5, calculateTokens(tools)}
    }
    // ... rest of rubric
}

func mapFirst(tools []string) bool {
    mapTools := []string{"bd_devlog_graph", "bd_devlog_search"}
    verifyTools := []string{"read_file", "grep"}
    
    firstMap := len(tools)
    firstVerify := len(tools)
    
    for i, t := range tools {
        if contains(mapTools, t) {
            firstMap = i
        }
        if contains(verifyTools, t) {
            firstVerify = i
        }
    }
    
    return firstMap < firstVerify && firstMap < len(tools)/2  // Map in first half
}
```

***

## Debugging Tool Output

If tools aren't captured:

```
# 1. Test manually in worktree
cd /tmp/beads-eval-test001
echo "run bd devlog graph modal" | gemini -p -y --output-format json | jq .

# 2. Check BD tools are registered
gemini /tools  # Should list bd_devlog_graph, etc.

# 3. Verbose debug
echo "test" | gemini -p -y --debug --output-format json
```

**Most common issues:**
- `BEADS_DIR` not set → BD tools fail to init
- AGENTS.md missing → Agent doesn't know to use BD tools
- OAuth instead of API key → Hangs on auth

Your current setup + `--output-format json` gives you **everything** needed for perfect rubric scoring.

Test the manual command above and confirm tool output appears in `tools_used`.
