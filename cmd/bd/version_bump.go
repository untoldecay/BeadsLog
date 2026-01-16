package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var versionBumpCmd = &cobra.Command{
	Use:   "bump [major|minor|patch]",
	Short: "Bump the version number in cmd/bd/version.go",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		part := "patch"
		if len(args) > 0 {
			part = args[0]
		}

		path := "cmd/bd/version.go"
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
			return
		}

		// Regex to find Version = "X.Y.Z"
		re := regexp.MustCompile(`Version\s*=\s*"(\d+)\.(\d+)\.(\d+)"`);
		matches := re.FindSubmatch(content)
		if matches == nil {
			fmt.Println("Error: Could not find version string in cmd/bd/version.go")
			return
		}

		major, _ := strconv.Atoi(string(matches[1]))
		minor, _ := strconv.Atoi(string(matches[2]))
		patch, _ := strconv.Atoi(string(matches[3]))

		oldVer := fmt.Sprintf("%d.%d.%d", major, minor, patch)

		switch strings.ToLower(part) {
		case "major":
			major++
			minor = 0
			patch = 0
		case "minor":
			minor++
			patch = 0
		case "patch":
			patch++
		default:
			fmt.Printf("Unknown bump type: %s. Use major, minor, or patch.\n", part)
			return
		}

		newVer := fmt.Sprintf("%d.%d.%d", major, minor, patch)
		
		// Replace in content
		newContent := strings.Replace(string(content), oldVer, newVer, 1)
		
		if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			return
		}

		fmt.Printf("Bumped version: %s -> %s\n", oldVer, newVer)
	},
}

func init() {
	versionCmd.AddCommand(versionBumpCmd)
}
