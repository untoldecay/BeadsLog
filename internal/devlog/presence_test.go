package devlog

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const testDir = "_rules/_devlog"

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func write(t *testing.T, root, name, content string) string {
	t.Helper()
	path := filepath.Join(root, testDir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func commit(t *testing.T, root, msg string) {
	t.Helper()
	run(t, root, "add", "-A")
	run(t, root, "commit", "-q", "--allow-empty", "-m", msg)
}

// newRepo returns a repo on "develop" with an older branch "old/feat".
// A.md exists only on develop.
func newRepo(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "main repo") // space: paths must survive it
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	run(t, root, "init", "-q", "-b", "develop")
	write(t, root, "base.md", "base")
	commit(t, root, "base")
	run(t, root, "branch", "old/feat")
	write(t, root, "A.md", "content A")
	commit(t, root, "A")
	return root
}

func at(root string) *Presence { return NewPresence(filepath.Join(root, testDir)) }

func expect(t *testing.T, p *Presence, name string, want State) {
	t.Helper()
	if got := p.Lookup(name); got != want {
		t.Errorf("Lookup(%s) = %d, want %d (err: %v)", name, got, want, p.Err())
	}
}

func TestPresenceBranchTips(t *testing.T) {
	root := newRepo(t)
	run(t, root, "switch", "-q", "old/feat")

	p := at(root)
	expect(t, p, "base.md", Here)
	expect(t, p, "A.md", Elsewhere)
	expect(t, p, "never.md", Absent)
	content, src, err := p.Read("A.md")
	if err != nil || string(content) != "content A" || src != "develop" {
		t.Errorf("Read = %q from %q, %v", content, src, err)
	}
	if r := p.Restriction(); r != "" {
		t.Errorf("main checkout should not be restricted: %s", r)
	}
}

// History does not count: committed then deleted everywhere is really gone.
func TestPresenceDeletedFromAllTips(t *testing.T) {
	root := newRepo(t)
	if err := os.Remove(filepath.Join(root, testDir, "A.md")); err != nil {
		t.Fatal(err)
	}
	expect(t, at(root), "A.md", Elsewhere) // deletion not committed yet
	commit(t, root, "delete A")
	expect(t, at(root), "A.md", Absent)
}

func TestPresenceRemoteBranch(t *testing.T) {
	root := newRepo(t)
	clone := filepath.Join(t.TempDir(), "clone")
	run(t, root, "clone", "-q", "-b", "old/feat", root, clone)
	expect(t, at(clone), "A.md", Elsewhere)
}

func TestPresenceLinkedWorktree(t *testing.T) {
	root := newRepo(t)
	wt := filepath.Join(t.TempDir(), "wt")
	run(t, root, "worktree", "add", "-q", wt, "old/feat")
	write(t, wt, "uncommitted.md", "draft")

	fromWT := at(wt)
	expect(t, fromWT, "A.md", Elsewhere)
	if fromWT.Restriction() == "" {
		t.Error("linked worktree must be restricted")
	}

	fromMain := at(root)
	expect(t, fromMain, "uncommitted.md", Elsewhere)
	if fromMain.Restriction() != "" {
		t.Errorf("main checkout restricted: %s", fromMain.Restriction())
	}

	// A commit on a detached worktree HEAD is on no branch.
	run(t, wt, "switch", "-q", "--detach")
	write(t, wt, "detached.md", "d")
	commit(t, wt, "detached")
	if err := os.Remove(filepath.Join(wt, testDir, "detached.md")); err != nil {
		t.Fatal(err)
	}
	expect(t, at(root), "detached.md", Elsewhere)
}

// A worktree directory that vanished must not break the scan.
func TestPresenceMissingWorktreeDir(t *testing.T) {
	root := newRepo(t)
	wt := filepath.Join(t.TempDir(), "wt")
	run(t, root, "worktree", "add", "-q", wt, "old/feat")
	if err := os.RemoveAll(wt); err != nil {
		t.Fatal(err)
	}
	expect(t, at(root), "never.md", Absent)
}

func TestPresenceBareRepoWorktree(t *testing.T) {
	root := newRepo(t)
	bare := filepath.Join(t.TempDir(), "repo.git")
	run(t, root, "clone", "-q", "--bare", root, bare)
	wt := filepath.Join(t.TempDir(), "wt")
	run(t, bare, "worktree", "add", "-q", wt, "old/feat")
	p := at(wt)
	expect(t, p, "A.md", Elsewhere)
	if r := p.Restriction(); r != "" {
		t.Errorf("a bare repo has no main checkout to defer to, got %q", r)
	}
}

func TestPresenceRedirectedToOtherRepo(t *testing.T) {
	rootA, rootB := newRepo(t), newRepo(t)
	for _, d := range []string{filepath.Join(rootA, ".beads"), filepath.Join(rootB, ".beads")} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}
	redirect := filepath.Join(rootA, ".beads", "redirect")
	if err := os.WriteFile(redirect, []byte(filepath.Join(rootB, ".beads")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if at(rootA).Restriction() == "" {
		t.Error("checkout using another repo's database must be restricted")
	}
	if r := at(rootB).Restriction(); r != "" {
		t.Errorf("owning repo restricted: %s", r)
	}
}

func TestPresenceOutsideGit(t *testing.T) {
	root := t.TempDir()
	write(t, root, "x.md", "x")
	p := at(root)
	expect(t, p, "x.md", Here)
	expect(t, p, "gone.md", Absent)
	if p.Restriction() != "" {
		t.Error("plain directory must not be restricted")
	}
}

// If git fails inside a repository, absence cannot be proven.
func TestPresenceGitFailureIsUnknown(t *testing.T) {
	root := newRepo(t)
	if err := os.WriteFile(filepath.Join(root, ".git", "HEAD"), []byte("garbage"), 0644); err != nil {
		t.Fatal(err)
	}
	p := at(root)
	expect(t, p, "A.md", Here)
	expect(t, p, "gone.md", Unknown)
}

// Hooks run with GIT_DIR exported; it must not redirect the scan.
func TestPresenceIgnoresHookEnvironment(t *testing.T) {
	root, other := newRepo(t), newRepo(t)
	run(t, root, "switch", "-q", "old/feat")
	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)
	p := at(root)
	expect(t, p, "A.md", Elsewhere)
	if !strings.HasSuffix(p.toplevel, "main repo") {
		t.Errorf("scanned the wrong repo: %s", p.toplevel)
	}
}

func TestPresenceDevlogDirMissingOnBranch(t *testing.T) {
	root := newRepo(t)
	run(t, root, "switch", "-q", "--orphan", "empty")
	run(t, root, "commit", "-q", "--allow-empty", "-m", "empty")
	expect(t, at(root), "A.md", Elsewhere)
}

func TestRowIsDead(t *testing.T) {
	dead := map[string]bool{"a.md": true}
	if !rowIsDead("| x | [a.md](a.md?id=sess-1) |", dead) || rowIsDead("| x | [ba.md](ba.md) |", dead) {
		t.Error("row matching is wrong")
	}
}
