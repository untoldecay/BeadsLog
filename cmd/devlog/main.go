package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "devlog",
	Short: "Devlog markdown parser and analyzer",
	Long: `devlog is a CLI tool for parsing and analyzing devlog markdown files.

It can:
  - Parse index.md files with entries in "## YYYY-MM-DD - Title" format
  - Extract entities (CamelCase, kebab-case, keywords, issue IDs)
  - Display entity relationship graphs
  - Show hierarchical connections between entities
  - Search across entries with full-text search and graph context`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
	},
	// Cobra configuration
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: false,
	},
	DisableFlagsInUseLine: true,
	SilenceUsage:  true,
	SilenceErrors: false,
}

func init() {
	rootCmd.AddCommand(importMDCmd)
	rootCmd.AddCommand(graphCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(entitiesCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(resumeCmd)
	rootCmd.AddCommand(impactCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
