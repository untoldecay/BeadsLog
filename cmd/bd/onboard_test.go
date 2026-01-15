package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInjectProtocol(t *testing.T) {
	tempDir := t.TempDir()
	protocol := "## Devlog Protocol (MANDATORY)\n1. Do this.\n2. Do that."

	t.Run("Empty file (Clean Slate)", func(t *testing.T) {
		f := filepath.Join(tempDir, "empty.md")
		os.WriteFile(f, []byte("   \n"), 0644) // Whitespace only

		injectProtocol(f, protocol)

		content, _ := os.ReadFile(f)
		if string(content) != protocol {
			t.Errorf("Expected content to be just protocol, got:\n%q", string(content))
		}
	})

	t.Run("File with old trigger", func(t *testing.T) {
		f := filepath.Join(tempDir, "old_trigger.md")
		initial := "Some intro.\nBEFORE ANYTHING ELSE: run 'bd devlog onboard'\nSome footer."
		os.WriteFile(f, []byte(initial), 0644)

		injectProtocol(f, protocol)

		content, _ := os.ReadFile(f)
		strContent := string(content)

		if strings.Contains(strContent, "BEFORE ANYTHING ELSE") {
			t.Error("Old trigger should be removed")
		}
		if !strings.Contains(strContent, "## Devlog Protocol") {
			t.Error("Protocol should be injected")
		}
		if !strings.Contains(strContent, "Some intro.") {
			t.Error("Original content should be preserved")
		}
	})

	t.Run("File with new trigger", func(t *testing.T) {
		f := filepath.Join(tempDir, "new_trigger.md")
		initial := "Start.\nBEFORE ANYTHING ELSE: run 'bd onboard'\nEnd."
		os.WriteFile(f, []byte(initial), 0644)

		injectProtocol(f, protocol)

		content, _ := os.ReadFile(f)
		strContent := string(content)

		if strings.Contains(strContent, "BEFORE ANYTHING ELSE") {
			t.Error("New trigger should be removed")
		}
		if !strings.Contains(strContent, "## Devlog Protocol") {
			t.Error("Protocol should be injected")
		}
	})

	t.Run("Idempotency", func(t *testing.T) {
		f := filepath.Join(tempDir, "idempotent.md")
		initial := "Existing content.\n" + protocol
		os.WriteFile(f, []byte(initial), 0644)

		injectProtocol(f, protocol)

		content, _ := os.ReadFile(f)
		if string(content) != initial {
			t.Errorf("Content changed despite protocol being present. Got:\n%q", string(content))
		}
	})

	t.Run("Append to existing", func(t *testing.T) {
		f := filepath.Join(tempDir, "append.md")
		initial := "Existing content."
		os.WriteFile(f, []byte(initial), 0644)

		injectProtocol(f, protocol)

		content, _ := os.ReadFile(f)
		expected := initial + "\n" + protocol
		if string(content) != expected {
			t.Errorf("Expected appended content. Got:\n%q", string(content))
		}
	})
}