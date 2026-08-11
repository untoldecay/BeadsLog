package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initGitRepo makes tmp a git repo so gitExcludePath resolves .git/info/exclude.
func initGitRepo(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Chdir(tmp)
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.name", "t"}, {"config", "user.email", "t@t"}} {
		if err := exec.Command("git", args...).Run(); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
	return tmp
}

func TestGitExcludeAddRemove(t *testing.T) {
	initGitRepo(t)
	patterns := []string{".beads/", "_rules/_devlog-solo/"}

	added, err := addGitExcludePatterns(patterns)
	if err != nil {
		t.Fatal(err)
	}
	if len(added) != 2 {
		t.Fatalf("expected 2 patterns added, got %d", len(added))
	}

	// Idempotent: a second add writes nothing new.
	added2, err := addGitExcludePatterns(patterns)
	if err != nil {
		t.Fatal(err)
	}
	if len(added2) != 0 {
		t.Errorf("re-add should be a no-op, got %v", added2)
	}

	path, _ := gitExcludePath()

	// A pre-existing user exclude must survive removal.
	body, _ := os.ReadFile(path) // #nosec G304
	if err := os.WriteFile(path, append([]byte("*.log\n"), body...), 0644); err != nil {
		t.Fatal(err)
	}

	if err := removeGitExcludePatterns(patterns); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(path) // #nosec G304
	got := string(after)
	for _, p := range patterns {
		if strings.Contains(got, p) {
			t.Errorf("pattern %q not removed:\n%s", p, got)
		}
	}
	if strings.Contains(got, "# Beads solo mode") {
		t.Errorf("solo header not removed:\n%s", got)
	}
	if !strings.Contains(got, "*.log") {
		t.Errorf("user's own exclude was clobbered:\n%s", got)
	}
}

func TestCopyThenPublishDevlogs(t *testing.T) {
	tmp := t.TempDir()
	team := filepath.Join(tmp, "team")
	solo := filepath.Join(tmp, "solo")
	if err := os.MkdirAll(team, 0750); err != nil {
		t.Fatal(err)
	}
	write := func(dir, name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(team, "2026-01-01_a.md", "team-a")
	write(team, "2026-01-02_b.md", "team-b")
	write(team, "notes.txt", "ignored") // non-md skipped

	// Continuity: carry the two team devlogs into the solo dir.
	if err := os.MkdirAll(solo, 0750); err != nil {
		t.Fatal(err)
	}
	carried, err := copyDevlogs(team, solo)
	if err != nil {
		t.Fatal(err)
	}
	if carried != 2 {
		t.Fatalf("expected 2 carried, got %d", carried)
	}

	// A new solo-only devlog is written locally.
	write(solo, "2026-02-01_secret.md", "solo-secret")

	// Publish: only the new file lands in team; carried copies are skipped.
	published, err := publishSoloDevlogs(solo, team)
	if err != nil {
		t.Fatal(err)
	}
	if published != 1 {
		t.Fatalf("expected 1 published, got %d", published)
	}
	if _, err := os.Stat(filepath.Join(team, "2026-02-01_secret.md")); err != nil {
		t.Errorf("new solo devlog not published: %v", err)
	}
}
