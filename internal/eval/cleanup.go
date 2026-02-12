package eval

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetProjectHash returns a stable hash of the current project path
func GetProjectHash() string {
	cwd, _ := os.Getwd()
	abs, _ := filepath.Abs(cwd)
	h := 0
	for _, char := range abs {
		h = 31*h + int(char)
	}
	if h < 0 {
		h = -h
	}
	return fmt.Sprintf("%x", h)
}

// PruneProjectOrphans removes stale eval directories belonging to this project
func PruneProjectOrphans() int {
	projHash := GetProjectHash()
	marker := "beads-eval-" + projHash + "-"
	
	tempBase := os.TempDir()
	entries, err := os.ReadDir(tempBase)
	if err != nil {
		return 0
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() && strings.Contains(entry.Name(), marker) {
			path := filepath.Join(tempBase, entry.Name())
			if err := os.RemoveAll(path); err == nil {
				count++
			}
		}
	}
	
	return count
}

// SafeCleanupABC cleans up specific sandbox directories and their parent
func SafeCleanupABC(wt1, wt2, wt3, tempDir string) {
	projHash := GetProjectHash()
	marker := "beads-eval-" + projHash + "-"

	// Since we use git init in a parent tempDir, 
	// removing tempDir is sufficient if it matches the marker.
	if tempDir != "" && strings.Contains(tempDir, marker) {
		_ = os.RemoveAll(tempDir)
	}
}
