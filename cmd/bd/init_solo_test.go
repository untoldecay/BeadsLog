package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestApplySoloConfig verifies solo mode writes the three local-only settings
// (declarative marker sync-mode + enforcement keys no-push/daemon.auto-sync)
// and, critically, that invisible mode (local=true) routes them to the
// git-excluded config.local.yaml — never the committed config.yaml — so solo
// state can't leak to the team (BeadsLog-9vd).
func TestApplySoloConfig(t *testing.T) {
	wantKeys := []string{
		`sync-mode: "local-only"`,
		`no-push: "true"`,
		`daemon.auto-sync: "false"`,
	}
	assertHasKeys := func(t *testing.T, got string) {
		t.Helper()
		for _, want := range wantKeys {
			// Values may be quoted or bare depending on updateYamlKey; accept both.
			bare := strings.ReplaceAll(want, `"`, "")
			if !strings.Contains(got, want) && !strings.Contains(got, bare) {
				t.Errorf("missing %q\n--- content ---\n%s", want, got)
			}
		}
	}

	t.Run("committed writes to config.yaml", func(t *testing.T) {
		beadsDir := setupBeadsDir(t)
		if err := applySoloConfig(false); err != nil {
			t.Fatalf("applySoloConfig(false) failed: %v", err)
		}
		content, err := os.ReadFile(filepath.Join(beadsDir, "config.yaml"))
		if err != nil {
			t.Fatal(err)
		}
		assertHasKeys(t, string(content))
		if _, err := os.Stat(filepath.Join(beadsDir, "config.local.yaml")); err == nil {
			t.Error("config.local.yaml should not exist in committed mode")
		}
	})

	t.Run("local writes to config.local.yaml only", func(t *testing.T) {
		beadsDir := setupBeadsDir(t)
		if err := applySoloConfig(true); err != nil {
			t.Fatalf("applySoloConfig(true) failed: %v", err)
		}
		local, err := os.ReadFile(filepath.Join(beadsDir, "config.local.yaml"))
		if err != nil {
			t.Fatalf("config.local.yaml not written: %v", err)
		}
		assertHasKeys(t, string(local))
		// The committed config.yaml must NOT carry any solo settings — that's
		// the whole point of the local override.
		committed, err := os.ReadFile(filepath.Join(beadsDir, "config.yaml"))
		if err != nil {
			t.Fatal(err)
		}
		// Only active (uncommented) lines count — the template ships commented
		// examples like "# daemon.auto-sync: false".
		for _, line := range strings.Split(string(committed), "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			for _, leak := range []string{"local-only", "no-push:", "daemon.auto-sync:"} {
				if strings.Contains(trimmed, leak) {
					t.Errorf("config.yaml leaked active solo setting: %q", trimmed)
				}
			}
		}
	})
}

func setupBeadsDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	beadsDir := filepath.Join(tmpDir, ".beads")
	if err := os.MkdirAll(beadsDir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := createConfigYaml(beadsDir, false); err != nil {
		t.Fatal(err)
	}
	return beadsDir
}
