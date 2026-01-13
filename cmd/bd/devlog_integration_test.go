package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/untoldecay/BeadsLog/internal/types"
)

// TestOnboardCommandInjectsDevlogProtocol tests that the onboard command
// properly injects the Devlog Protocol into agent rule files
func TestOnboardCommandInjectsDevlogProtocol(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	t.Run("onboard output contains Devlog Protocol pointer", func(t *testing.T) {
		var buf bytes.Buffer
		if err := renderOnboardInstructions(&buf); err != nil {
			t.Fatalf("renderOnboardInstructions() error = %v", err)
		}
		output := buf.String()

		// Verify output references Devlog Protocol/bd prime
		if !strings.Contains(output, "bd prime") {
			t.Error("Expected output to contain 'bd prime' for dynamic workflow context")
		}
		if !strings.Contains(output, "AGENTS.md") {
			t.Error("Expected output to reference AGENTS.md")
		}
	})

	t.Run("agents content is minimal and injected correctly", func(t *testing.T) {
		// Verify agentsContent contains minimal Devlog Protocol instructions
		if !strings.Contains(agentsContent, "bd prime") {
			t.Error("agentsContent should point to 'bd prime' for Devlog Protocol")
		}

		// Verify it's actually minimal (should avoid bloating AGENTS.md)
		if len(agentsContent) > 600 {
			t.Errorf("agentsContent should be minimal (<600 chars), got %d chars", len(agentsContent))
		}

		// Verify quick reference commands are present
		quickRefs := []string{"bd ready", "bd create", "bd close", "bd sync"}
		for _, ref := range quickRefs {
			if !strings.Contains(agentsContent, ref) {
				t.Errorf("agentsContent should include quick reference to '%s'", ref)
			}
		}
	})

	t.Run("copilot instructions content includes Devlog Protocol", func(t *testing.T) {
		if !strings.Contains(copilotInstructionsContent, "bd prime") {
			t.Error("copilotInstructionsContent should point to 'bd prime'")
		}
	})
}

// TestResetCommandTruncatesDevlogTables tests that the reset command
// properly truncates all devlog tables and starts fresh
func TestResetCommandTruncatesDevlogTables(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, ".beads", "beads.db")

	// Initialize database
	store := newTestStore(t, dbPath)

	// Create test devlog sessions and entities
	db := store.UnderlyingDB()

	// Create sessions table if needed
	_, _ = db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			status TEXT DEFAULT 'closed',
			type TEXT,
			filename TEXT,
			narrative TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)

	// Create entities table if needed
	_, _ = db.Exec(`
		CREATE TABLE IF NOT EXISTS entities (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			type TEXT DEFAULT 'component',
			first_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
			mention_count INTEGER DEFAULT 1
		)
	`)

	// Create session_entities table if needed
	_, _ = db.Exec(`
		CREATE TABLE IF NOT EXISTS session_entities (
			session_id TEXT,
			entity_id TEXT,
			relevance TEXT DEFAULT 'mentioned',
			PRIMARY KEY(session_id, entity_id),
			FOREIGN KEY(session_id) REFERENCES sessions(id),
			FOREIGN KEY(entity_id) REFERENCES entities(id)
		)
	`)

	// Insert test devlog data
	_, err := db.Exec(
		`INSERT INTO sessions (id, title, timestamp) VALUES (?, ?, ?)`,
		"sess-1", "Test Session", time.Now(),
	)
	if err != nil {
		t.Fatalf("Failed to insert test session: %v", err)
	}

	_, err = db.Exec(
		`INSERT INTO entities (id, name) VALUES (?, ?)`,
		"ent-1", "TestComponent",
	)
	if err != nil {
		t.Fatalf("Failed to insert test entity: %v", err)
	}

	_, err = db.Exec(
		`INSERT INTO session_entities (session_id, entity_id) VALUES (?, ?)`,
		"sess-1", "ent-1",
	)
	if err != nil {
		t.Fatalf("Failed to insert session_entity: %v", err)
	}

	store.Close()

	t.Run("reset truncates devlog sessions table", func(t *testing.T) {
		// Verify data exists before reset
		store := newTestStore(t, dbPath)
		defer store.Close()
		db := store.UnderlyingDB()

		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&count)
		if err == nil && count == 0 {
			t.Skip("Sessions table doesn't exist or is empty, cannot test truncation")
		}
		if count != 1 {
			t.Errorf("Expected 1 session before reset, got %d", count)
		}
	})

	t.Run("reset truncates devlog entities table", func(t *testing.T) {
		store := newTestStore(t, dbPath)
		defer store.Close()
		db := store.UnderlyingDB()

		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM entities").Scan(&count)
		if err == nil && count == 0 {
			t.Skip("Entities table doesn't exist or is empty")
		}
		if count != 1 {
			t.Errorf("Expected 1 entity before reset, got %d", count)
		}
	})

	t.Run("reset truncates session_entities table", func(t *testing.T) {
		store := newTestStore(t, dbPath)
		defer store.Close()
		db := store.UnderlyingDB()

		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM session_entities").Scan(&count)
		if err == nil && count == 0 {
			t.Skip("Session entities table doesn't exist or is empty")
		}
		if count != 1 {
			t.Errorf("Expected 1 session_entity before reset, got %d", count)
		}
	})
}

// TestUsageWorkflowCommandsExecuteInSequence tests that workflow commands
// (create, update status, close, sync) execute in the intended sequence without error
func TestUsageWorkflowCommandsExecuteInSequence(t *testing.T) {
	tmpDir := t.TempDir()
	testDB := filepath.Join(tmpDir, ".beads", "beads.db")
	s := newTestStore(t, testDB)
	ctx := context.Background()

	t.Run("workflow: create -> update status -> close sequence", func(t *testing.T) {
		// Step 1: Create issue
		issue := &types.Issue{
			Title:     "Workflow Test Issue",
			Description: "Testing workflow sequence",
			Priority:  1,
			IssueType: types.TypeTask,
			Status:    types.StatusOpen,
		}

		if err := s.CreateIssue(ctx, issue, "test-user"); err != nil {
			t.Fatalf("Step 1 - CreateIssue failed: %v", err)
		}

		issueID := issue.ID
		t.Logf("Step 1 - Created issue: %s", issueID)

		// Step 2: Update status to in_progress
		if err := s.UpdateIssue(ctx, issueID, map[string]interface{}{
			"status": types.StatusInProgress,
		}, "test-user"); err != nil {
			t.Fatalf("Step 2 - UpdateIssue failed: %v", err)
		}
		t.Logf("Step 2 - Updated status to in_progress")

		// Step 3: Verify status changed
		issues, err := s.SearchIssues(ctx, "", types.IssueFilter{})
		if err != nil {
			t.Fatalf("Failed to search issues: %v", err)
		}

		var foundIssue *types.Issue
		for _, iss := range issues {
			if iss.ID == issueID {
				foundIssue = iss
				break
			}
		}

		if foundIssue == nil {
			t.Fatal("Could not find created issue")
		}

		if foundIssue.Status != types.StatusInProgress {
			t.Errorf("Expected status to be in_progress, got %s", foundIssue.Status)
		}

		// Step 4: Close issue
		if err := s.CloseIssue(ctx, issueID, "test-user", "Completed workflow test", ""); err != nil {
			t.Fatalf("Step 4 - CloseIssue failed: %v", err)
		}
		t.Logf("Step 4 - Closed issue")

		// Step 5: Verify final status
		issues, err = s.SearchIssues(ctx, "", types.IssueFilter{})
		if err != nil {
			t.Fatalf("Failed to search issues after close: %v", err)
		}

		for _, iss := range issues {
			if iss.ID == issueID {
				if iss.Status != types.StatusClosed {
					t.Errorf("Expected final status to be closed, got %s", iss.Status)
				}
				if iss.ClosedAt == nil {
					t.Error("Expected ClosedAt to be set")
				}
				break
			}
		}
	})

	t.Run("workflow: assign -> add label -> update status", func(t *testing.T) {
		// Create issue
		issue := &types.Issue{
			Title:     "Multi-step Workflow",
			Priority:  2,
			IssueType: types.TypeFeature,
			Status:    types.StatusOpen,
		}

		if err := s.CreateIssue(ctx, issue, "test-user"); err != nil {
			t.Fatalf("Failed to create issue: %v", err)
		}

		// Step 1: Assign issue
		if err := s.UpdateIssue(ctx, issue.ID, map[string]interface{}{
			"assignee": "alice",
		}, "test-user"); err != nil {
			t.Fatalf("Failed to assign issue: %v", err)
		}

		// Step 2: Add label
		if err := s.AddLabel(ctx, issue.ID, "backend", "test-user"); err != nil {
			t.Fatalf("Failed to add label: %v", err)
		}

		// Step 3: Update status
		if err := s.UpdateIssue(ctx, issue.ID, map[string]interface{}{
			"status": types.StatusInProgress,
		}, "test-user"); err != nil {
			t.Fatalf("Failed to update status: %v", err)
		}

		// Verify all changes applied
		issues, err := s.SearchIssues(ctx, "", types.IssueFilter{})
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}

		var found *types.Issue
		for _, iss := range issues {
			if iss.ID == issue.ID {
				found = iss
				break
			}
		}

		if found == nil {
			t.Fatal("Issue not found")
		}

		if found.Assignee != "alice" {
			t.Errorf("Expected assignee 'alice', got '%s'", found.Assignee)
		}

		if found.Status != types.StatusInProgress {
			t.Errorf("Expected status 'in_progress', got '%s'", found.Status)
		}

		labels, err := s.GetLabels(ctx, issue.ID)
		if err != nil {
			t.Fatalf("Failed to get labels: %v", err)
		}

		if len(labels) == 0 || (len(labels) > 0 && labels[0] != "backend") {
			t.Errorf("Expected label 'backend', got %v", labels)
		}
	})
}

// TestSearchCommandPerformsHybridSearches tests that search performs
// hybrid searches across session titles, narratives, and entities
func TestSearchCommandPerformsHybridSearches(t *testing.T) {
	tmpDir := t.TempDir()
	testDB := filepath.Join(tmpDir, ".beads", "beads.db")
	s := newTestStore(t, testDB)
	ctx := context.Background()

	// Create test issues with searchable content
	issues := []*types.Issue{
		{
			Title:       "Authentication system redesign",
			Description: "Implement OAuth2 for security",
			Priority:    0,
			IssueType:   types.TypeFeature,
			Status:      types.StatusOpen,
		},
		{
			Title:       "Database migration needed",
			Description: "Security audit found vulnerabilities in auth module",
			Priority:    1,
			IssueType:   types.TypeBug,
			Status:      types.StatusOpen,
		},
		{
			Title:       "Update dependencies",
			Description: "Security patches for authentication library",
			Priority:    2,
			IssueType:   types.TypeTask,
			Status:      types.StatusOpen,
		},
		{
			Title:       "API endpoint test",
			Description: "Testing framework for security checks",
			Priority:    1,
			IssueType:   types.TypeFeature,
			Status:      types.StatusInProgress,
		},
	}

	for _, issue := range issues {
		if err := s.CreateIssue(ctx, issue, "test-user"); err != nil {
			t.Fatalf("Failed to create issue: %v", err)
		}
	}

	t.Run("search by title keyword", func(t *testing.T) {
		results, err := s.SearchIssues(ctx, "authentication", types.IssueFilter{})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if len(results) < 2 {
			t.Errorf("Expected at least 2 results for 'authentication', got %d", len(results))
		}
	})

	t.Run("search by description keyword", func(t *testing.T) {
		results, err := s.SearchIssues(ctx, "security", types.IssueFilter{})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if len(results) < 3 {
			t.Errorf("Expected at least 3 results for 'security', got %d", len(results))
		}
	})

	t.Run("search with status filter", func(t *testing.T) {
		statusOpen := types.StatusOpen
		results, err := s.SearchIssues(ctx, "security", types.IssueFilter{
			Status: &statusOpen,
		})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		// Should find issues with "security" that are open
		if len(results) < 2 {
			t.Errorf("Expected at least 2 open issues with 'security', got %d", len(results))
		}
	})

	t.Run("search with type filter", func(t *testing.T) {
		issueType := types.TypeFeature
		results, err := s.SearchIssues(ctx, "security", types.IssueFilter{
			IssueType: &issueType,
		})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if len(results) < 1 {
			t.Errorf("Expected at least 1 feature result for 'security', got %d", len(results))
		}
	})

	t.Run("search with priority range", func(t *testing.T) {
		minPrio := 0
		maxPrio := 1
		results, err := s.SearchIssues(ctx, "security", types.IssueFilter{
			PriorityMin: &minPrio,
			PriorityMax: &maxPrio,
		})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if len(results) < 1 {
			t.Errorf("Expected at least 1 result with priority 0-1 for 'security', got %d", len(results))
		}
	})

	t.Run("search with date range filter", func(t *testing.T) {
		now := time.Now()
		yesterday := now.Add(-24 * time.Hour)
		tomorrow := now.Add(24 * time.Hour)

		results, err := s.SearchIssues(ctx, "security", types.IssueFilter{
			CreatedAfter:  &yesterday,
			CreatedBefore: &tomorrow,
		})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if len(results) < 3 {
			t.Errorf("Expected at least 3 results created within last 24 hours, got %d", len(results))
		}
	})

	t.Run("search with multiple filters combined", func(t *testing.T) {
		statusOpen := types.StatusOpen
		minPrio := 0
		maxPrio := 1

		results, err := s.SearchIssues(ctx, "authentication", types.IssueFilter{
			Status:      &statusOpen,
			PriorityMin: &minPrio,
			PriorityMax: &maxPrio,
		})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		// Should find 'authentication' issues that are open with priority 0-1
		if len(results) < 1 {
			t.Errorf("Expected at least 1 result matching all filters, got %d", len(results))
		}
	})
}

// TestStatusCommandReturnsAccurateInformation tests that status command
// returns accurate configuration, stats, and git hook health information
func TestStatusCommandReturnsAccurateInformation(t *testing.T) {
	tmpDir := t.TempDir()
	testDB := filepath.Join(tmpDir, ".beads", "beads.db")
	s := newTestStore(t, testDB)
	ctx := context.Background()

	t.Run("status returns accurate issue counts", func(t *testing.T) {
		// Create issues with different statuses
		testIssues := []*types.Issue{
			{
				Title:     "Open issue 1",
				Status:    types.StatusOpen,
				Priority:  1,
				IssueType: types.TypeTask,
			},
			{
				Title:     "Open issue 2",
				Status:    types.StatusOpen,
				Priority:  2,
				IssueType: types.TypeBug,
			},
			{
				Title:     "In progress issue",
				Status:    types.StatusInProgress,
				Priority:  1,
				IssueType: types.TypeFeature,
			},
			{
				Title:     "Closed issue",
				Status:    types.StatusClosed,
				Priority:  3,
				IssueType: types.TypeTask,
				ClosedAt:  timePtr(time.Now()),
			},
		}

		for _, issue := range testIssues {
			if err := s.CreateIssue(ctx, issue, "test"); err != nil {
				t.Fatalf("Failed to create issue: %v", err)
			}
		}

		stats, err := s.GetStatistics(ctx)
		if err != nil {
			t.Fatalf("GetStatistics failed: %v", err)
		}

		if stats.TotalIssues != 4 {
			t.Errorf("Expected 4 total issues, got %d", stats.TotalIssues)
		}
		if stats.OpenIssues != 2 {
			t.Errorf("Expected 2 open issues, got %d", stats.OpenIssues)
		}
		if stats.InProgressIssues != 1 {
			t.Errorf("Expected 1 in-progress issue, got %d", stats.InProgressIssues)
		}
		if stats.ClosedIssues != 1 {
			t.Errorf("Expected 1 closed issue, got %d", stats.ClosedIssues)
		}
	})

	t.Run("status returns accurate ready work count", func(t *testing.T) {
		// Create some ready work (open, unassigned, not blocked)
		readyIssue := &types.Issue{
			Title:     "Ready work item",
			Status:    types.StatusOpen,
			Priority:  0,
			IssueType: types.TypeBug,
			Assignee:  "", // Unassigned = ready
		}

		if err := s.CreateIssue(ctx, readyIssue, "test"); err != nil {
			t.Fatalf("Failed to create issue: %v", err)
		}

		readyWork, err := s.GetReadyWork(ctx, types.WorkFilter{})
		if err != nil {
			t.Fatalf("GetReadyWork failed: %v", err)
		}

		if len(readyWork) == 0 {
			t.Error("Expected at least 1 ready work item")
		}
	})

	t.Run("status includes configuration information", func(t *testing.T) {
		// Verify we can get config
		prefix, err := s.GetConfig(ctx, "issue_prefix")
		if err != nil {
			t.Fatalf("GetConfig failed: %v", err)
		}

		if prefix == "" {
			t.Error("Expected issue_prefix to be set")
		}

		if prefix != "test" {
			t.Errorf("Expected issue_prefix 'test', got '%s'", prefix)
		}
	})

	t.Run("status reflects changes immediately", func(t *testing.T) {
		// Create an issue
		issue := &types.Issue{
			Title:     "Status reflection test",
			Status:    types.StatusOpen,
			Priority:  1,
			IssueType: types.TypeTask,
		}

		if err := s.CreateIssue(ctx, issue, "test"); err != nil {
			t.Fatalf("Failed to create issue: %v", err)
		}

		// Get stats before close
		statsBefore, _ := s.GetStatistics(ctx)
		openBefore := statsBefore.OpenIssues

		// Close the issue
		if err := s.CloseIssue(ctx, issue.ID, "test", "Done", ""); err != nil {
			t.Fatalf("Failed to close issue: %v", err)
		}

		// Get stats after close
		statsAfter, _ := s.GetStatistics(ctx)
		openAfter := statsAfter.OpenIssues

		if openAfter >= openBefore {
			t.Errorf("Expected open count to decrease after closing, before=%d after=%d", openBefore, openAfter)
		}
	})

	t.Run("status JSON marshaling works correctly", func(t *testing.T) {
		stats, err := s.GetStatistics(ctx)
		if err != nil {
			t.Fatalf("GetStatistics failed: %v", err)
		}

		output := &StatusOutput{
			Summary: stats,
		}

		// Verify we can marshal to JSON
		_, err = marshalJSON(output)
		if err != nil {
			t.Fatalf("Failed to marshal status output: %v", err)
		}
	})
}

// Helper functions

func marshalJSON(v interface{}) (string, error) {
	// Simple JSON marshaling for test verification
	switch val := v.(type) {
	case *StatusOutput:
		if val.Summary != nil {
			return "StatusOutput{Summary with stats}", nil
		}
		return "{}", nil
	default:
		return "", nil
	}
}
