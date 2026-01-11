package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// ShowOptions contains options for the show command
type ShowOptions struct {
	IndexPath string // Path to index.md file (default: ./index.md)
}

var showOpts = &ShowOptions{}

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show [date|filename]",
	Short: "Show full devlog entry content",
	Long: `Show full devlog entry content by date or filename.

Reads and displays the complete markdown content from linked files,
combining metadata from database with narrative from filesystem.

Arguments:
  date      - Show entry for a specific date (YYYY-MM-DD format)
  filename  - Show entry from a specific markdown file

Examples:
  devlog show 2024-01-15
  devlog show 2024-01-15.md
  devlog show entries/my-feature.md`,
	RunE: runShow,
}

func init() {
	showCmd.Flags().StringVarP(&showOpts.IndexPath, "index", "i", "./index.md", "Path to index.md file")
}

// runShow executes the show command
func runShow(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires a date or filename argument\n\nUsage: devlog show [date|filename]")
	}

	target := args[0]

	// Determine if target is a date or filename
	if isDate(target) {
		return showByDate(target)
	}

	// Treat as filename
	return showByFilename(target)
}

// isDate checks if the string matches YYYY-MM-DD format
func isDate(s string) bool {
	// Check for YYYY-MM-DD format
	dateRe := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	return dateRe.MatchString(s)
}

// showByDate displays entries for a specific date
func showByDate(date string) error {
	// Parse the index.md file
	rows, err := parseIndexMD(showOpts.IndexPath)
	if err != nil {
		return fmt.Errorf("failed to parse index.md: %w", err)
	}

	// Find matching entries
	var matched []*IndexRow
	for _, row := range rows {
		if row.Date == date {
			matched = append(matched, row)
		}
	}

	if len(matched) == 0 {
		return fmt.Errorf("no entry found for date: %s", date)
	}

	// Display the entries
	for _, row := range matched {
		displayEntry(row)
	}

	return nil
}

// showByFilename displays entry from a specific markdown file
func showByFilename(filename string) error {
	// If no extension, add .md
	if !strings.HasSuffix(filename, ".md") {
		filename = filename + ".md"
	}

	// Check if filename is a relative path
	if !filepath.IsAbs(filename) {
		// Try current directory first
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			// Try relative to index.md directory
			indexDir := filepath.Dir(showOpts.IndexPath)
			filename = filepath.Join(indexDir, filename)
		}
	}

	// Read the file content
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Display the content
	fmt.Printf("# %s\n\n", filename)
	fmt.Println(string(content))

	return nil
}

// displayEntry displays a single index row with full content
func displayEntry(row *IndexRow) {
	// Display header
	fmt.Printf("## %s - %s\n\n", row.Date, row.Title)

	// Display description/content
	if row.Description != "" {
		fmt.Printf("%s\n\n", row.Description)
	}

	// Display entities if any
	if len(row.Entities) > 0 {
		fmt.Printf("**Entities:** %s\n\n", strings.Join(row.Entities, ", "))
	}

	// Display metadata
	fmt.Printf("---\n")
	fmt.Printf("Date: %s\n", row.Date)
	fmt.Printf("Line: %d\n", row.LineNumber)
	fmt.Println()
}
