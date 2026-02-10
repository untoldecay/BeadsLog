package eval

import (
	"fmt"
	"os"
	"os/exec"
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

// PruneProjectOrphans removes stale eval worktrees belonging to this project
func PruneProjectOrphans() int {
	projHash := GetProjectHash()
	marker := "beads-eval-" + projHash + "-"
	
	out, err := exec.Command("git", "worktree", "list", "--porcelain").Output()
	if err != nil {
		return 0
	}

	count := 0
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimPrefix(line, "worktree ")
			// Only remove if it belongs to this project's namespace
			if strings.Contains(path, marker) {
				if err := exec.Command("git", "worktree", "remove", "--force", path).Run(); err == nil {
					count++
					// Also try to remove the parent temp directory if it's an eval dir
					parent := filepath.Dir(path)
					if strings.Contains(filepath.Base(parent), marker) {
						_ = os.RemoveAll(parent)
					}
				}
			}
		}
	}
	
	_ = exec.Command("git", "worktree", "prune").Run()
	return count
}

// SafeCleanupABC cleans up specific worktrees and their parent directory
func SafeCleanupABC(wt1, wt2, wt3, tempDir string) {
	projHash := GetProjectHash()
	marker := "beads-eval-" + projHash + "-"

	for _, wt := range []string{wt1, wt2, wt3} {
		if wt != "" && strings.Contains(wt, marker) {
			_ = exec.Command("git", "worktree", "remove", "--force", wt).Run()
		}
	}
	
	if tempDir != "" && strings.Contains(tempDir, marker) {
		_ = os.RemoveAll(tempDir)
	}
	_ = exec.Command("git", "worktree", "prune").Run()
}
