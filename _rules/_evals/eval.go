package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

type Scenario struct {
	ID                  string   `json:"id"`
	Prompt              string   `json:"prompt"`
	Category            string   `json:"category"`
	PassCriteria        string   `json:"pass_criteria"`
	RequiredAny         []string `json:"required_any,omitempty"`
	RequiredAll         []string `json:"required_all,omitempty"`
	Sequence            []string `json:"sequence,omitempty"`
	AntiPatterns        []string `json:"anti_patterns,omitempty"`
	MaxStepIndex        int      `json:"max_step_index,omitempty"`
	MinStepIndexFromEnd int      `json:"min_step_index_from_end,omitempty"`
}

type EvalResult struct {
	Timestamp       string   `json:"timestamp"`
	ScenarioID      string   `json:"scenario_id"`
	Prompt          string   `json:"prompt"`
	Category        string   `json:"category"`
	Pass            bool     `json:"pass"`
	Score           int      `json:"score"`
	ToolCalls       []string `json:"tool_calls"`
	TokensUsed      int      `json:"tokens_used"`
	EstimatedTokens int      `json:"estimated_tokens"`
	SavingsVsVanilla int     `json:"savings_vs_vanilla"`
	FailedReason    string   `json:"failed_reason,omitempty"`
}

type CostState struct {
	TotalTokensUsed int `json:"total_tokens_used"`
}

func getTotalTokens() int {
	data, err := os.ReadFile(".beads/cost_state.json")
	if err != nil {
		return 0
	}
	var state CostState
	if err := json.Unmarshal(data, &state); err != nil {
		return 0
	}
	return state.TotalTokensUsed
}

// estimateCost calculates projected token usage based on tool types
func estimateCost(trace []string) int {
	total := 0
	for _, cmd := range trace {
		cmd = strings.ToLower(cmd)
		if strings.Contains(cmd, "grep") || strings.Contains(cmd, "search_file_content") {
			total += 2500 // Reading context of many files
		} else if strings.Contains(cmd, "ls -r") || strings.Contains(cmd, "list_directory") {
			total += 800 // Large directory trees
		} else if strings.Contains(cmd, "read_file") {
			total += 1200 // Average file content
		} else if strings.HasPrefix(cmd, "bd devlog") {
			total += 150 // Focused metadata retrieval
		} else if strings.HasPrefix(cmd, "bd") {
			total += 200 // Standard issue query
		} else {
			total += 300 // Generic tool call
		}
	}
	return total
}

func main() {
	scenariosPath := flag.String("scenarios", "_rules/_evals/scenarios.json", "Path to scenarios.json")
	tracePath := flag.String("trace", "", "Path to trace log file")
	scenarioID := flag.String("id", "", "Scenario ID to evaluate")
	costStateStart := flag.Int("tokens-start", -1, "Token count before task")
	costStateEnd := flag.Int("tokens-end", -1, "Token count after task")
	flag.Parse()

	if *tracePath == "" || *scenarioID == "" {
		fmt.Println("Usage: go run eval.go --id <ID> --trace <path> [--tokens-start <n> --tokens-end <n>]")
		os.Exit(1)
	}

	scenarios, err := loadScenarios(*scenariosPath)
	if err != nil {
		fmt.Printf("Error loading scenarios: %v\n", err)
		os.Exit(1)
	}

	var scenario *Scenario
	for _, s := range scenarios {
		if s.ID == *scenarioID {
			scenario = &s
			break
		}
	}

	if scenario == nil {
		fmt.Printf("Scenario %s not found\n", *scenarioID)
		os.Exit(1)
	}

	trace, err := loadTrace(*tracePath)
	if err != nil {
		fmt.Printf("Error loading trace: %v\n", err)
		os.Exit(1)
	}

	startTokens := *costStateStart
	if startTokens == -1 {
		startTokens = getTotalTokens()
	}

	endTokens := *costStateEnd
	if endTokens == -1 {
		endTokens = getTotalTokens()
	}

	result := scoreTrace(scenario, trace)
	result.Timestamp = time.Now().Format(time.RFC3339)
	result.TokensUsed = endTokens - startTokens
	if result.TokensUsed < 0 {
		result.TokensUsed = 0
	}

	// Token Impact Analysis
	result.EstimatedTokens = estimateCost(trace)
	
	// A typical Vanilla trace for Scenario 1 would be: grep + ls + read_file ~= 4500
	vanillaBaseline := 4500 
	result.SavingsVsVanilla = vanillaBaseline - result.EstimatedTokens
	if result.SavingsVsVanilla < 0 && result.Score < 100 {
		// If they used brute force, savings is 0 or negative
		result.SavingsVsVanilla = 0
	}

	// Output JSONL
	output, _ := json.Marshal(result)
	fmt.Println(string(output))
}

func loadScenarios(path string) ([]Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s []Scenario
	err = json.Unmarshal(data, &s)
	return s, err
}

func loadTrace(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}

func scoreTrace(s *Scenario, trace []string) EvalResult {
	res := EvalResult{
		ScenarioID: s.ID,
		Prompt:     s.Prompt,
		Category:   s.Category,
		ToolCalls:  trace,
		Pass:       true,
		Score:      100,
	}

	// 1. Required Any
	if len(s.RequiredAny) > 0 {
		found := false
		firstIdx := -1
		for _, req := range s.RequiredAny {
			for i, cmd := range trace {
				if strings.Contains(cmd, req) {
					found = true
					if firstIdx == -1 || i < firstIdx {
						firstIdx = i
					}
					break
				}
			}
		}
		if !found {
			res.Pass = false
			res.Score -= 50
			res.FailedReason += fmt.Sprintf("Missing required tool (any of %v). ", s.RequiredAny)
		} else if s.MaxStepIndex > 0 && firstIdx >= s.MaxStepIndex {
			res.Pass = false
			res.Score -= 25
			res.FailedReason += fmt.Sprintf("Required tool used too late (step %d, max %d). ", firstIdx+1, s.MaxStepIndex)
		}
	}

	// 2. Required All
	if len(s.RequiredAll) > 0 {
		for _, req := range s.RequiredAll {
			found := false
			for _, cmd := range trace {
				if strings.Contains(cmd, req) {
					found = true
					break
				}
			}
			if !found {
				res.Pass = false
				res.Score -= 30
				res.FailedReason += fmt.Sprintf("Missing required tool: %s. ", req)
			}
		}
	}

	// 3. Sequence
	if len(s.Sequence) > 0 {
		lastIdx := -1
		for _, step := range s.Sequence {
			found := false
			for i := lastIdx + 1; i < len(trace); i++ {
				if strings.Contains(trace[i], step) {
					lastIdx = i
					found = true
					break
				}
			}
			if !found {
				res.Pass = false
				res.Score -= 20
				res.FailedReason += fmt.Sprintf("Sequence break: expected '%s' after previous step. ", step)
				break
			}
		}
	}

	// 4. Anti-Patterns
	if len(s.AntiPatterns) > 0 {
		for _, anti := range s.AntiPatterns {
			for i, cmd := range trace {
				if strings.Contains(cmd, anti) {
					// Check if any required tool was used BEFORE this anti-pattern
					firstReqIdx := -1
					for _, req := range append(s.RequiredAny, s.RequiredAll...) {
						for j, c := range trace {
							if strings.Contains(c, req) {
								if firstReqIdx == -1 || j < firstReqIdx {
									firstReqIdx = j
								}
							}
						}
					}
					if firstReqIdx == -1 || i < firstReqIdx {
						res.Pass = false
						res.Score -= 40
						res.FailedReason += fmt.Sprintf("Anti-pattern '%s' used before/without retrieval. ", anti)
						break
					}
				}
			}
		}
	}

	if res.Score < 0 {
		res.Score = 0
	}
	if res.Score < 80 {
		res.Pass = false
	}

	return res
}
