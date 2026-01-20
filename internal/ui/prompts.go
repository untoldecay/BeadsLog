package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// PromptYesNo displays a yes/no question and returns the user's answer.
// It defaults to the `defaultYes` value if the user just presses Enter or in non-interactive mode.
func PromptYesNo(question string, defaultYes bool) bool {
	var input string
	var prompt string

	if defaultYes {
		prompt = fmt.Sprintf("%s [Y/n] ", question)
	} else {
		prompt = fmt.Sprintf("%s [y/N] ", question)
	}

	// In non-interactive mode (e.g., CI/script), return default
	if !IsTerminal() {
		fmt.Printf("%s (non-interactive, defaulting to %t)\n", prompt, defaultYes)
		return defaultYes
	}

	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		// On error (e.g., EOF), default
		fmt.Printf("(error reading input, defaulting to %t)\n", defaultYes)
		return defaultYes
	}

	input = strings.ToLower(strings.TrimSpace(line))

	if input == "y" || input == "yes" {
		return true
	}
	if input == "n" || input == "no" {
		return false
	}

	// Default if empty or invalid input
	return defaultYes
}

// Prompt for simple string input
func Prompt(question, defaultValue string) string {
	var input string
	prompt := fmt.Sprintf("%s (default: %q): ", question, defaultValue)

	if !IsTerminal() {
		fmt.Printf("%s (non-interactive, defaulting to %q)\n", prompt, defaultValue)
		return defaultValue
	}

	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("(error reading input, defaulting to %q)\n", defaultValue)
		return defaultValue
	}

	input = strings.TrimSpace(line)
	if input == "" {
		return defaultValue
	}
	return input
}
