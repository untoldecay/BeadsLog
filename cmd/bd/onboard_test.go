package main

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/untoldecay/BeadsLog/internal/storage/memory"
)

func TestOnboardingGateFlow(t *testing.T) {
	tempDir := t.TempDir()
	
	// Change CWD to tempDir so _rules/_orchestration is created there
	oldCwd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldCwd)

	ctx := context.Background()
	store := memory.New("")

	t.Run("Stage 1: Restricted Onboarding", func(t *testing.T) {
		f := "GEMINI.md"
		legacyContent := "# Legacy Context\nSpecial rules."
		os.WriteFile(f, []byte(legacyContent), 0644)

		executeOnboard(ctx, store)

		// Check GEMINI.md has RESTRICTED bootloader
		content, _ := os.ReadFile(f)
		strContent := string(content)
		if !strings.Contains(strContent, "BEFORE ANYTHING ELSE") {
			t.Error("Restricted bootloader not installed")
		}
		if strings.Contains(strContent, "PROJECT_CONTEXT.md") {
			t.Error("Restricted bootloader should not link to PROJECT_CONTEXT.md")
		}

		// Check flag in DB
		finalized, _ := store.GetConfig(ctx, "onboarding_finalized")
		if finalized != "false" {
			t.Errorf("Expected onboarding_finalized=false, got %s", finalized)
		}
	})

	t.Run("Stage 2: Finalization via ready trigger", func(t *testing.T) {
		f := "GEMINI.md"
		
		// Simulate running bd ready
		finalizeOnboarding(ctx, store)

		// Check GEMINI.md has FULL bootloader
		content, _ := os.ReadFile(f)
		strContent := string(content)
		if strings.Contains(strContent, "SETUP IN PROGRESS") {
			t.Error("Full bootloader should not contain setup-in-progress warning")
		}
		if !strings.Contains(strContent, "PROJECT_CONTEXT.md") {
			t.Error("Full bootloader should link to PROJECT_CONTEXT.md")
		}

		// Check flag in DB
		finalized, _ := store.GetConfig(ctx, "onboarding_finalized")
		if finalized != "true" {
			t.Errorf("Expected onboarding_finalized=true, got %s", finalized)
		}
	})
}