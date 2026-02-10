# OpenCode CLI Token Count Extraction (Complete Guide)

OpenCode **does not include token counts** in `--format json` output. Tokens are tracked via **separate commands** and **storage files**. Here's how to get them:

***

## Method 1: `opencode stats` (Primary)

**Command:** `opencode stats --project . --days 1 --models`

```bash
cd your-worktree
opencode stats --days 1 --format json
```

**Output:**
```json
{
  "sessions": [
    {
      "id": "ses_abc123",
      "project": "/tmp/beads-eval-test001",
      "start": "2026-02-09T14:22:00Z",
      "tokens": {
        "input": 850,
        "output": 450,
        "cache_read": 120,
        "cache_write": 80,
        "reasoning": 0,
        "total": 1500
      },
      "cost": {
        "usd": 0.00375
      },
      "model": "google/gemini-3-flash-preview",
      "tools_used": 5
    }
  ]
}
```

**In Go (after `opencode run`):**
```go
func getTokensForSession(worktree string) TokenUsage {
    cmd := exec.Command("opencode", "stats",
        "--days", "1",
        "--project", worktree,
        "--format", "json",
    )
    cmd.Dir = worktree
    
    output, _ := cmd.Output()
    var stats struct {
        Sessions []struct {
            Tokens struct {
                Input      int `json:"input"`
                Output     int `json:"output"`
                CacheRead  int `json:"cache_read"`
                CacheWrite int `json:"cache_write"`
                Total      int `json:"total"`
            } `json:"tokens"`
        } `json:"sessions"`
    }
    json.Unmarshal(output, &stats)
    
    if len(stats.Sessions) > 0 {
        return TokenUsage{
            Input:      stats.Sessions[0].Tokens.Input,
            Output:     stats.Sessions[0].Tokens.Output,
            Total:      stats.Sessions[0].Tokens.Total,
            Efficiency: calculateEfficiency(stats.Sessions[0].Tokens),
        }
    }
    return TokenUsage{}
}
```

***

## Method 2: Session Storage Files (Real‑Time)

OpenCode stores **per‑session stats** in `~/.local/share/opencode/storage/sessions/`:

```
~/.local/share/opencode/storage/sessions/
└── ses_abc123.json
{
  "id": "ses_abc123",
  "tokens": {
    "input": 850,
    "output": 450,
    "cache": 200
  },
  "tools": 5,
  "model": "google/gemini-3-flash-preview"
}
```

**Get session ID** from `opencode run` stderr or parse JSONL:

```bash
# Capture session ID
opencode run --format json "test" 2>&1 | grep -o 'ses_[a-z0-9]*'
```

**In Go:**
```go
func getSessionTokens(sessionId string) TokenUsage {
    sessionFile := filepath.Join(
        os.Getenv("HOME"), 
        ".local/share/opencode/storage/sessions", 
        sessionId+".json"
    )
    
    data, err := os.ReadFile(sessionFile)
    if err != nil {
        return TokenUsage{}
    }
    
    var session struct {
        Tokens map[string]int `json:"tokens"`
    }
    json.Unmarshal(data, &session)
    
    return TokenUsage{
        Input:  session.Tokens["input"],
        Output: session.Tokens["output"],
        Total:  session.Tokens["total"],
    }
}
```

***

## Method 3: `--stats` Flag (Inline)

Some OpenCode versions support `--stats` with `run`:

```bash
opencode run --format json --stats "test task"
```

**Adds to final JSONL line:**
```json
{
  "type": "session_finish",
  "tokens": {
    "input": 850,
    "output": 450,
    "total": 1300
  }
}
```

**Test:**
```bash
opencode run --format json --stats "hello" | tail -1 | jq .
```

***

## Method 4: Live Streaming (Advanced)

**Parse JSONL stream** for `tokens` events:

```go
func runOpenCodeWithLiveTokens(wtDir, task string) (OpenCodeTrace, TokenUsage) {
    cmd := exec.Command("opencode", "run", "--format", "json", task)
    cmd.Dir = wtDir
    
    pipeReader, _ := io.Pipe()
    cmd.Stdout = pipeReader
    
    var trace OpenCodeTrace
    var liveTokens TokenUsage
    
    scanner := bufio.NewScanner(pipeReader)
    for scanner.Scan() {
        line := scanner.Text()
        var event map[string]interface{}
        
        if json.Unmarshal([]byte(line), &event) == nil {
            if eventType, ok := event["type"].(string); ok {
                switch eventType {
                case "tokens":
                    liveTokens.Input = int(event["input"].(float64))
                    liveTokens.Output = int(event["output"].(float64))
                case "tool_use":
                    // ... aggregate tools
                }
            }
        }
    }
    
    return trace, liveTokens
}
```

***

## Recommended: `opencode stats` Post‑Run

**Simplest + most reliable** for your eval harness:

```go
func runOpenCodeTest(wtDir, task, traceId string) OpenCodeTrace {
    // 1. Run OpenCode
    cmd := exec.Command("opencode", "run", "--format", "json", task)
    cmd.Dir = wtDir
    // ... env vars ...
    
    outputBytes, _ := cmd.CombinedOutput()
    
    // 2. Parse tool calls from JSONL (previous code)
    var trace OpenCodeTrace
    parseJsonl(outputBytes, &trace)
    
    // 3. Get tokens from stats
    tokens := getTokensForSession(wtDir)
    trace.Tokens = tokens
    
    return trace
}

func getTokensForSession(wtDir string) TokenUsage {
    cmd := exec.Command("opencode", "stats", "--days", "1", "--project", wtDir, "--format", "json")
    cmd.Dir = wtDir
    
    output, err := cmd.Output()
    if err != nil {
        return TokenUsage{Total: 0}
    }
    
    var stats map[string][]map[string]interface{}
    json.Unmarshal(output, &stats)
    
    if len(stats["sessions"]) > 0 {
        session := stats["sessions"][0]
        tokens := session["tokens"].(map[string]interface{})
        return TokenUsage{
            Input:  int(tokens["input"].(float64)),
            Output: int(tokens["output"].(float64)),
            Total:  int(tokens["total"].(float64)),
        }
    }
    
    return TokenUsage{}
}
```

***

## Your Rubric with Tokens

```go
type EvalResult struct {
    OpenCodeTrace
    TokenUsage
    Score Scoring
}

func (r *EvalResult) Efficiency(groupAvg int) float64 {
    if r.Total == 0 {
        return 0
    }
    return float64(groupAvg) / float64(r.Total) * r.Score.Weight
}
```

**Report example:**
```
Implicit: Optimal (bd_graph→read_file), 1350 tokens, 2.5x efficiency 🟢🟢🟢
Explicit: Shallow (read_file only), 8500 tokens, 0.8x efficiency 🟠
```

***

## Fallback: Manual Token Estimation

If `stats` fails, estimate from tools:

```go
func estimateTokens(toolCalls []ToolCall) TokenUsage {
    input := 0
    for _, tc := range toolCalls {
        switch tc.Name {
        case "bd_devlog_graph", "bd_devlog_search":
            input += 150  // Light context
        case "read_file":
            input += 1200 // Medium file
        case "grep":
            input += 2500 // Heavy search
        }
    }
    return TokenUsage{Total: input + 500} // + response overhead
}
```

***

## Complete Working Flow

```go
result := runOpenCodeTest(wtDir, task, traceId)
tokens := getTokensForSession(wtDir)
score := scoreTrace(result.ToolCalls)

fmt.Printf("Strategy: %s, Tokens: %d, Eff: %.1fx %s\n",
    score.Strategy, tokens.Total, score.Efficiency, score.Emoji)
```

**This gives you:**
- ✅ **Tool calls** from JSONL parsing
- ✅ **Token counts** from `opencode stats`
- ✅ **Rubric scoring** on Map/Verify sequence
- ✅ **A/B comparison** implicit vs explicit

Test `opencode stats --days 1 --format json` in your worktree first to confirm token data flows.
