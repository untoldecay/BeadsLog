package beads

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/untoldecay/BeadsLog/internal/git"
)

func mkBeads(t *testing.T, dir string, files ...string) string {
	t.Helper()
	b := filepath.Join(dir, ".beads")
	if err := os.MkdirAll(b, 0755); err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(b, f), []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return b
}

func setRedirect(t *testing.T, beadsDir, target string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(beadsDir, RedirectFileName), []byte(target+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@example.com"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func inDir(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("BEADS_DIR", "")
	t.Setenv("BEADS_DB", "")
	t.Chdir(dir)
	git.ResetCaches()
	t.Cleanup(git.ResetCaches)
}

func samePath(a, b string) bool {
	ra, _ := filepath.EvalSymlinks(a)
	rb, _ := filepath.EvalSymlinks(b)
	return ra == rb && ra != ""
}

func TestResolveRedirectForms(t *testing.T) {
	root := t.TempDir()
	target := mkBeads(t, filepath.Join(root, "a"), "issues.jsonl")
	local := mkBeads(t, filepath.Join(root, "w", "x", "y"))

	for name, path := range map[string]string{
		"absolute":                      target,
		"relative to project root":      "../../../a/.beads",
		"legacy: relative to .beads":    "../../../../a/.beads",
		"comment and blank lines first": "# note\n\n" + target,
	} {
		setRedirect(t, local, path)
		got, has, err := ResolveRedirect(local)
		if err != nil || !has || !samePath(got, target) {
			t.Errorf("%s: got %q, %v, %v", name, got, has, err)
		}
	}

	setRedirect(t, local, "../nowhere/.beads")
	if _, has, err := ResolveRedirect(local); !has || err == nil {
		t.Errorf("broken redirect: has=%v err=%v", has, err)
	}
	if got := FollowRedirect(local); got != local {
		t.Errorf("FollowRedirect on a broken redirect = %q, want original", got)
	}
}

// With both interpretations present, one with a project beats an empty directory.
func TestResolveRedirectPrefersProject(t *testing.T) {
	root := t.TempDir()
	target := mkBeads(t, filepath.Join(root, "a", "b"), "issues.jsonl")
	local := mkBeads(t, filepath.Join(root, "a", "w"))
	mkBeads(t, filepath.Join(root, "b")) // empty decoy at <project root>/../../b
	setRedirect(t, local, "../../b/.beads")
	if got, _, err := ResolveRedirect(local); err != nil || !samePath(got, target) {
		t.Errorf("got %q, %v; want %q", got, err, target)
	}
}

// Handoff 133ebe11: a worktree of clone A redirected to checkout B must use B
// for both the beads directory and the database, whatever A contains.
func TestWorktreeRedirectWinsForDirAndDatabase(t *testing.T) {
	for _, cloneFiles := range [][]string{{"config.yaml", "issues.jsonl"}, {"config.yaml", "beads.db"}} {
		root := t.TempDir()
		b := mkBeads(t, filepath.Join(root, "B"), "beads.db", "config.yaml")
		a := filepath.Join(root, "A")
		mkBeads(t, a, cloneFiles...)
		gitRun(t, a, "init", "-q")
		gitRun(t, a, "commit", "-q", "--allow-empty", "-m", "x")
		wt := filepath.Join(root, "w", "wt")
		gitRun(t, a, "worktree", "add", "-q", wt)
		setRedirect(t, mkBeads(t, wt), b)

		inDir(t, wt)
		if got := FindBeadsDir(); !samePath(got, b) {
			t.Errorf("clone has %v: FindBeadsDir = %q, want %q", cloneFiles, got, b)
		}
		if got := FindDatabasePath(); !samePath(got, filepath.Join(b, "beads.db")) {
			t.Errorf("clone has %v: FindDatabasePath = %q", cloneFiles, got)
		}
		info := GetRedirectInfo()
		if !info.IsRedirected || !samePath(info.TargetDir, b) {
			t.Errorf("GetRedirectInfo = %+v", info)
		}
		if err := CheckRedirect(); err != nil {
			t.Errorf("CheckRedirect: %v", err)
		}
		// The main repository itself is unaffected.
		inDir(t, a)
		if got := FindBeadsDir(); !samePath(got, filepath.Join(a, ".beads")) {
			t.Errorf("main repo FindBeadsDir = %q", got)
		}
	}
}

func TestCheckRedirectBroken(t *testing.T) {
	root := t.TempDir()
	a := filepath.Join(root, "A")
	mkBeads(t, a, "beads.db", "config.yaml")
	gitRun(t, a, "init", "-q")
	gitRun(t, a, "commit", "-q", "--allow-empty", "-m", "x")
	wt := filepath.Join(root, "wt")
	gitRun(t, a, "worktree", "add", "-q", wt)

	inDir(t, wt)
	if err := CheckRedirect(); err != nil {
		t.Errorf("no redirect: %v", err)
	}
	setRedirect(t, mkBeads(t, wt), "/nonexistent/.beads")
	if err := CheckRedirect(); err == nil {
		t.Error("broken redirect must be an error")
	}
	inDir(t, a)
	if err := CheckRedirect(); err != nil {
		t.Errorf("main repo must not be affected: %v", err)
	}

	// A broken redirect in a parent outside the repository is unrelated.
	setRedirect(t, mkBeads(t, root), "/nonexistent/.beads")
	plain := filepath.Join(root, "plain")
	if err := os.MkdirAll(plain, 0755); err != nil {
		t.Fatal(err)
	}
	gitRun(t, plain, "init", "-q")
	inDir(t, plain)
	if err := CheckRedirect(); err != nil {
		t.Errorf("redirect above the repository must be ignored: %v", err)
	}
}
