package main

import (
	"os"
	"strings"
	"testing"
)

func TestMigrateAndInjectProtocol(t *testing.T) {
	tempDir := t.TempDir()
	
	// Change CWD to tempDir so _rules/_orchestration is created there
	oldCwd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldCwd)

	t.Run("Migration of legacy content", func(t *testing.T) {
		f := "GEMINI.md"
		legacyContent := "# Legacy Context\nSpecial rules."
		os.WriteFile(f, []byte(legacyContent), 0644)

		migrateAndInjectProtocol(f)

		// Check GEMINI.md has bootloader
		content, _ := os.ReadFile(f)
		if !strings.Contains(string(content), "BeadsLog Agent Protocol") {
			t.Error("Bootloader not installed in GEMINI.md")
		}

		// Check legacy content moved to PROJECT_CONTEXT.md
		contextPath := "_rules/_orchestration/PROJECT_CONTEXT.md"
		contextContent, err := os.ReadFile(contextPath)
		if err != nil {
			t.Fatalf("PROJECT_CONTEXT.md not created: %v", err)
		}
		if !strings.Contains(string(contextContent), legacyContent) {
			t.Error("Legacy content not found in PROJECT_CONTEXT.md")
		}
	})

	t.Run("Idempotency", func(t *testing.T) {
		f := "CLAUDE.md"
		migrateAndInjectProtocol(f) // First run
		
		content1, _ := os.ReadFile(f)
		migrateAndInjectProtocol(f) // Second run
		
		content2, _ := os.ReadFile(f)
		if string(content1) != string(content2) {
			t.Error("Onboarding not idempotent")
		}
	})
}
