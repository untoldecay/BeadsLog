//go:build integration
// +build integration

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_Devlog_Show_List(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow CLI test in short mode")
	}

	tmpDir := setupCLITestDB(t)

	// 1. Setup Devlog Structure
	devlogDir := filepath.Join(tmpDir, "_rules", "_devlog")
	if err := os.MkdirAll(devlogDir, 0755); err != nil {
		t.Fatalf("Failed to create devlog dir: %v", err)
	}

	// 2. Create Session File
	sessionFile := "2026-01-01_test-session.md"
	sessionPath := filepath.Join(devlogDir, sessionFile)
	content := "# Test Session\n\nThis is a test session content."
	if err := os.WriteFile(sessionPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write session file: %v", err)
	}

	// 3. Create Index
	indexFile := filepath.Join(devlogDir, "_index.md")
	indexContent := `## Work Index

| Subject | Problems | Date | Devlog |
|---------|----------|------|--------|
| [feature] Test Session | Testing | 2026-01-01 | [2026-01-01_test-session.md](2026-01-01_test-session.md) |
`
	if err := os.WriteFile(indexFile, []byte(indexContent), 0644); err != nil {
		t.Fatalf("Failed to write index file: %v", err)
	}

	// 4. Run Sync
	// We need to configure devlog_dir first or let auto-discovery find it.
	// Since we are in tmpDir, we need to run sync inside tmpDir.
	// Also need to set devlog_dir config because we are not in a standard repo structure maybe?
	// Actually `setupCLITestDB` runs `init`.
	// Let's run `bd devlog init` to be safe, or just set the config.
	runBDInProcess(t, tmpDir, "config", "set", "devlog_dir", "_rules/_devlog")
	runBDInProcess(t, tmpDir, "devlog", "sync")

	// 5. Run List and Verify ID is present
	out := runBDInProcess(t, tmpDir, "devlog", "list")
	if !strings.Contains(out, "[sess-") {
		t.Errorf("Expected session ID in list output, got: %s", out)
	}
	
	// Extract ID
	// Output format: [Date] [ID] Type - Title
	// Example: [2026-01-01 00:00] [sess-123456] feature - [feature] Test Session
	start := strings.Index(out, "[sess-")
	if start == -1 {
		t.Fatalf("Could not find ID in output: %s", out)
	}
	end := strings.Index(out[start:], "]")
	id := out[start+1 : start+end] // sess-xxxxxx

	// 6. Run Show with ID
	out = runBDInProcess(t, tmpDir, "devlog", "show", id)
	if !strings.Contains(out, "This is a test session content") {
		t.Errorf("Show by ID failed. Expected content, got: %s", out)
	}

	// 7. Run Show with Date (Partial)
	out = runBDInProcess(t, tmpDir, "devlog", "show", "2026-01-01")
	// Since there is only one, it should show content directly
	if !strings.Contains(out, "This is a test session content") {
		t.Errorf("Show by Date failed. Expected content, got: %s", out)
	}

	// 8. Run Show with Local Formatted Timestamp (simulate copy-paste from list)
	// We first need to see what the list output *actually* produced for the timestamp
	// It depends on the machine's timezone, so we parse the list output.
	// Wait, out from step 6/7 is content. We need out from step 5.
	listOut := runBDInProcess(t, tmpDir, "devlog", "list")
	
	// Parse the first part "[YYYY-MM-DD HH:MM]"
	// Assuming it's the first bracketed thing
	tsStart := strings.Index(listOut, "[")
	tsEnd := strings.Index(listOut, "]")
	if tsStart != -1 && tsEnd != -1 {
		timestampStr := listOut[tsStart+1 : tsEnd]
		
		// Run show with this exact string
		showOut := runBDInProcess(t, tmpDir, "devlog", "show", timestampStr)
		if !strings.Contains(showOut, "This is a test session content") {
			t.Errorf("Show by Local Timestamp '%s' failed. Got: %s", timestampStr, showOut)
		}
	}
}
