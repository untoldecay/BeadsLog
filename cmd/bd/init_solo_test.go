package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestApplySoloConfig verifies that solo mode writes the three local-only
// settings to .beads/config.yaml: the declarative marker (sync-mode) and the
// two enforcement keys (no-push, daemon.auto-sync).
func TestApplySoloConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	beadsDir := filepath.Join(tmpDir, ".beads")
	if err := os.MkdirAll(beadsDir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := createConfigYaml(beadsDir, false); err != nil {
		t.Fatal(err)
	}

	if err := applySoloConfig(); err != nil {
		t.Fatalf("applySoloConfig failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(beadsDir, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(content)

	for _, want := range []string{
		`sync-mode: "local-only"`,
		`no-push: "true"`,
		`daemon.auto-sync: "false"`,
	} {
		// Values may be written quoted or bare depending on updateYamlKey;
		// accept both.
		bare := strings.ReplaceAll(want, `"`, "")
		if !strings.Contains(got, want) && !strings.Contains(got, bare) {
			t.Errorf("config.yaml missing %q\n--- content ---\n%s", want, got)
		}
	}
}
