package channel

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCompare(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int
	}{
		{"0.66.0", "0.66.0-dev.1", 1},
		{"0.66.0-dev.1", "0.66.0", -1},
		{"0.66.0-dev.1", "0.65.0", 1},
		{"0.66.0-dev.10", "0.66.0-dev.9", 1},
		{"v0.66.0-dev.1", "0.66.0-dev.1", 0},
		{"1.0", "1.0.0", 0},
		{"1.1", "1.0.5", 1},
		{"1.2.3.4", "1.2.3.3", 1}, // not semver
		{"", "0.1.0", -1},
	} {
		if got := Compare(tc.a, tc.b); got != tc.want {
			t.Errorf("Compare(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestLatest(t *testing.T) {
	tags := []string{"v0.64.0", "v0.65.0", "v0.66.0-dev.2", "v0.66.0-dev.10", "v0.67.0-rc.1", "v9", "vnext"}
	if got := Latest(Stable, tags); got != "0.65.0" {
		t.Errorf("stable = %q", got)
	}
	if got := Latest(Dev, tags); got != "0.66.0-dev.10" {
		t.Errorf("dev = %q", got)
	}
	// A dev machine moves to the release once it exists.
	if got := Latest(Dev, append(tags, "v0.66.0")); got != "0.66.0" {
		t.Errorf("dev after release = %q", got)
	}
}

func TestFetchLatest(t *testing.T) {
	repo := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q"},
		{"-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "--allow-empty", "-m", "x"},
		{"tag", "v0.65.0"}, {"tag", "v0.66.0-dev.1"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	t.Setenv("BD_RELEASE_SOURCE", repo)
	if got, err := FetchLatest(Stable); err != nil || got != "0.65.0" {
		t.Errorf("stable = %q, %v", got, err)
	}
	if got, err := FetchLatest(Dev); err != nil || got != "0.66.0-dev.1" {
		t.Errorf("dev = %q, %v", got, err)
	}
	t.Setenv("BD_RELEASE_SOURCE", filepath.Join(repo, "missing"))
	if _, err := FetchLatest(Stable); err == nil {
		t.Error("unreachable source must be an error")
	}
}

func TestCurrentAndSave(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "cfg"))
	t.Setenv("AppData", filepath.Join(home, "appdata"))

	if got := Current("0.65.0"); got != Stable {
		t.Errorf("default = %q", got)
	}
	if got := Current("0.66.0-dev.1"); got != Dev {
		t.Errorf("a dev build should follow dev, got %q", got)
	}
	if err := Save("beta"); err == nil {
		t.Error("unknown channel accepted")
	}
	if err := Save(Stable); err != nil {
		t.Fatal(err)
	}
	if got := Current("0.66.0-dev.1"); got != Stable {
		t.Errorf("an explicit choice must win, got %q", got)
	}
}
