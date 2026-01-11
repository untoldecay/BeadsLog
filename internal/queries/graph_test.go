// Package queries provides tests for complex entity relationship queries.
package queries

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/steveyegge/beads/internal/types"
)

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create schema
	_, err = db.Exec(`
		CREATE TABLE issues (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'open',
			issue_type TEXT NOT NULL DEFAULT 'task',
			priority INTEGER NOT NULL DEFAULT 2,
			assignee TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE dependencies (
			issue_id TEXT NOT NULL,
			depends_on_id TEXT NOT NULL,
			type TEXT NOT NULL DEFAULT 'blocks',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_by TEXT NOT NULL,
			metadata TEXT DEFAULT '{}',
			thread_id TEXT DEFAULT '',
			PRIMARY KEY (issue_id, depends_on_id),
			FOREIGN KEY (issue_id) REFERENCES issues(id) ON DELETE CASCADE,
			FOREIGN KEY (depends_on_id) REFERENCES issues(id) ON DELETE CASCADE
		);

		CREATE INDEX idx_dependencies_issue ON dependencies(issue_id);
		CREATE INDEX idx_dependencies_depends_on ON dependencies(depends_on_id);
	`)
	if err != nil {
		db.Close()
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

// insertTestIssue inserts a test issue into the database.
func insertTestIssue(t *testing.T, db *sql.DB, id, title, status, issueType string, priority int) {
	_, err := db.Exec(`
		INSERT INTO issues (id, title, status, issue_type, priority, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))
	`, id, title, status, issueType, priority)
	if err != nil {
		t.Fatalf("failed to insert issue %s: %v", id, err)
	}
}

// insertTestDependency inserts a test dependency into the database.
func insertTestDependency(t *testing.T, db *sql.DB, issueID, dependsOnID, depType string) {
	_, err := db.Exec(`
		INSERT INTO dependencies (issue_id, depends_on_id, type, created_by)
		VALUES (?, ?, ?, 'test')
	`, issueID, dependsOnID, depType)
	if err != nil {
		t.Fatalf("failed to insert dependency %s -> %s: %v", issueID, dependsOnID, err)
	}
}

// TestGetEntityGraph_SingleNode tests graph with just the root node.
func TestGetEntityGraph_SingleNode(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Insert a single issue
	insertTestIssue(t, db, "issue-1", "Single Issue", "open", "task", 2)

	// Get graph with depth 0 (just root)
	graph, err := GetEntityGraph(ctx, db, "issue-1", 0, nil, 0)
	if err != nil {
		t.Fatalf("GetEntityGraph failed: %v", err)
	}

	// Verify graph structure
	if graph.RootID != "issue-1" {
		t.Errorf("expected RootID 'issue-1', got '%s'", graph.RootID)
	}

	if graph.Depth != 0 {
		t.Errorf("expected Depth 0, got %d", graph.Depth)
	}

	if len(graph.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(graph.Nodes))
	}

	if len(graph.Edges) != 0 {
		t.Errorf("expected 0 edges, got %d", len(graph.Edges))
	}

	// Verify stats
	if graph.Stats.TotalNodes != 1 {
		t.Errorf("expected TotalNodes 1, got %d", graph.Stats.TotalNodes)
	}

	if graph.Stats.TotalEdges != 0 {
		t.Errorf("expected TotalEdges 0, got %d", graph.Stats.TotalEdges)
	}

	if graph.Stats.MaxDepthReached != 0 {
		t.Errorf("expected MaxDepthReached 0, got %d", graph.Stats.MaxDepthReached)
	}
}

// TestGetEntityGraph_LinearChain tests a simple linear dependency chain.
func TestGetEntityGraph_LinearChain(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create a linear chain: issue-1 -> issue-2 -> issue-3
	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("issue-%d", i)
		insertTestIssue(t, db, id, fmt.Sprintf("Issue %d", i), "open", "task", 2)
	}

	// Add dependencies
	insertTestDependency(t, db, "issue-2", "issue-1", "blocks")
	insertTestDependency(t, db, "issue-3", "issue-2", "blocks")

	// Get graph with depth 2
	graph, err := GetEntityGraph(ctx, db, "issue-1", 2, nil, 0)
	if err != nil {
		t.Fatalf("GetEntityGraph failed: %v", err)
	}

	// Verify all nodes are included
	if len(graph.Nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(graph.Nodes))
	}

	// Verify edges
	if len(graph.Edges) != 2 {
		t.Errorf("expected 2 edges, got %d", len(graph.Edges))
	}

	// Verify depth levels
	if graph.Stats.MaxDepthReached != 2 {
		t.Errorf("expected MaxDepthReached 2, got %d", graph.Stats.MaxDepthReached)
	}

	// Verify nodes at each depth
	depth0Nodes := graph.GetNodesByDepth(0)
	if len(depth0Nodes) != 1 {
		t.Errorf("expected 1 node at depth 0, got %d", len(depth0Nodes))
	}

	depth1Nodes := graph.GetNodesByDepth(1)
	if len(depth1Nodes) != 1 {
		t.Errorf("expected 1 node at depth 1, got %d", len(depth1Nodes))
	}
}

// TestGetEntityGraph_TreeStructure tests a tree-like dependency structure.
func TestGetEntityGraph_TreeStructure(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create a tree structure:
	//       issue-1 (root)
	//      /    \
	// issue-2  issue-3
	//    |
	// issue-4
	insertTestIssue(t, db, "issue-1", "Root Issue", "open", "epic", 1)
	insertTestIssue(t, db, "issue-2", "Child 1", "open", "task", 2)
	insertTestIssue(t, db, "issue-3", "Child 2", "open", "task", 2)
	insertTestIssue(t, db, "issue-4", "Grandchild", "open", "task", 3)

	// Add dependencies
	insertTestDependency(t, db, "issue-2", "issue-1", "parent-child")
	insertTestDependency(t, db, "issue-3", "issue-1", "parent-child")
	insertTestDependency(t, db, "issue-4", "issue-2", "parent-child")

	// Get graph
	graph, err := GetEntityGraph(ctx, db, "issue-1", 3, nil, 0)
	if err != nil {
		t.Fatalf("GetEntityGraph failed: %v", err)
	}

	// Verify structure
	if len(graph.Nodes) != 4 {
		t.Errorf("expected 4 nodes, got %d", len(graph.Nodes))
	}

	if len(graph.Edges) != 3 {
		t.Errorf("expected 3 edges, got %d", len(graph.Edges))
	}

	// Verify depth distribution
	if graph.Stats.NodesByDepth[0] != 1 {
		t.Errorf("expected 1 node at depth 0, got %d", graph.Stats.NodesByDepth[0])
	}

	if graph.Stats.NodesByDepth[1] != 2 {
		t.Errorf("expected 2 nodes at depth 1, got %d", graph.Stats.NodesByDepth[1])
	}

	if graph.Stats.NodesByDepth[2] != 1 {
		t.Errorf("expected 1 node at depth 2, got %d", graph.Stats.NodesByDepth[2])
	}
}

// TestGetEntityGraph_BidirectionalEdges tests that edges are found in both directions.
func TestGetEntityGraph_BidirectionalEdges(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create issues
	insertTestIssue(t, db, "issue-1", "Issue 1", "open", "task", 2)
	insertTestIssue(t, db, "issue-2", "Issue 2", "open", "task", 2)

	// Add a single dependency
	insertTestDependency(t, db, "issue-2", "issue-1", "blocks")

	// Get graph from issue-1
	graph, err := GetEntityGraph(ctx, db, "issue-1", 1, nil, 0)
	if err != nil {
		t.Fatalf("GetEntityGraph failed: %v", err)
	}

	// Should find issue-2 as a dependent (downstream)
	if len(graph.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(graph.Nodes))
	}

	if len(graph.Edges) != 1 {
		t.Errorf("expected 1 edge, got %d", len(graph.Edges))
	}
}

// TestGetEntityGraph_EdgeTypeFilter tests filtering by dependency types.
func TestGetEntityGraph_EdgeTypeFilter(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create issues
	insertTestIssue(t, db, "issue-1", "Root", "open", "epic", 1)
	insertTestIssue(t, db, "issue-2", "Child 1", "open", "task", 2)
	insertTestIssue(t, db, "issue-3", "Child 2", "open", "task", 2)
	insertTestIssue(t, db, "issue-4", "Related", "open", "task", 2)

	// Add different dependency types
	insertTestDependency(t, db, "issue-2", "issue-1", "parent-child")
	insertTestDependency(t, db, "issue-3", "issue-1", "blocks")
	insertTestDependency(t, db, "issue-4", "issue-1", "related")

	// Get graph filtered by parent-child type only
	graph, err := GetEntityGraph(ctx, db, "issue-1", 1, []types.DependencyType{types.DepParentChild}, 0)
	if err != nil {
		t.Fatalf("GetEntityGraph failed: %v", err)
	}

	// Should only include parent-child edges
	if len(graph.Edges) != 1 {
		t.Errorf("expected 1 edge (parent-child only), got %d", len(graph.Edges))
	}

	if len(graph.Edges) > 0 && graph.Edges[0].Type != types.DepParentChild {
		t.Errorf("expected parent-child edge, got %s", graph.Edges[0].Type)
	}
}

// TestGetEntityGraph_DepthLimit tests that depth limit is respected.
func TestGetEntityGraph_DepthLimit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create a deep chain
	for i := 1; i <= 10; i++ {
		id := fmt.Sprintf("issue-%d", i)
		insertTestIssue(t, db, id, fmt.Sprintf("Issue %d", i), "open", "task", 2)
		if i > 1 {
			insertTestDependency(t, db, id, fmt.Sprintf("issue-%d", i-1), "blocks")
		}
	}

	// Get graph with depth 3
	graph, err := GetEntityGraph(ctx, db, "issue-1", 3, nil, 0)
	if err != nil {
		t.Fatalf("GetEntityGraph failed: %v", err)
	}

	// Should only traverse to depth 3
	if graph.Stats.MaxDepthReached != 3 {
		t.Errorf("expected MaxDepthReached 3, got %d", graph.Stats.MaxDepthReached)
	}

	// Should have 4 nodes (depth 0, 1, 2, 3)
	if len(graph.Nodes) != 4 {
		t.Errorf("expected 4 nodes (depth 0-3), got %d", len(graph.Nodes))
	}
}

// TestGetEntityGraph_NodeLimit tests that node limit is respected.
func TestGetEntityGraph_NodeLimit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create a structure with many nodes
	insertTestIssue(t, db, "issue-1", "Root", "open", "epic", 1)
	for i := 2; i <= 10; i++ {
		id := fmt.Sprintf("issue-%d", i)
		insertTestIssue(t, db, id, fmt.Sprintf("Child %d", i), "open", "task", 2)
		insertTestDependency(t, db, id, "issue-1", "parent-child")
	}

	// Get graph with node limit of 5
	graph, err := GetEntityGraph(ctx, db, "issue-1", 2, nil, 5)
	if err != nil {
		t.Fatalf("GetEntityGraph failed: %v", err)
	}

	// Should truncate at 5 nodes
	if len(graph.Nodes) != 5 {
		t.Errorf("expected 5 nodes (truncated), got %d", len(graph.Nodes))
	}

	if !graph.Stats.Truncated {
		t.Error("expected graph to be marked as truncated")
	}
}

// TestGetEntityGraph_CyclePrevention tests that cycles are handled correctly.
func TestGetEntityGraph_CyclePrevention(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create issues
	insertTestIssue(t, db, "issue-1", "Issue 1", "open", "task", 2)
	insertTestIssue(t, db, "issue-2", "Issue 2", "open", "task", 2)
	insertTestIssue(t, db, "issue-3", "Issue 3", "open", "task", 2)

	// Create a cycle: issue-1 -> issue-2 -> issue-3 -> issue-1
	insertTestDependency(t, db, "issue-2", "issue-1", "blocks")
	insertTestDependency(t, db, "issue-3", "issue-2", "blocks")
	insertTestDependency(t, db, "issue-1", "issue-3", "blocks")

	// Get graph - should not hang or fail
	graph, err := GetEntityGraph(ctx, db, "issue-1", 10, nil, 0)
	if err != nil {
		t.Fatalf("GetEntityGraph failed: %v", err)
	}

	// Should find all 3 nodes
	if len(graph.Nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(graph.Nodes))
	}

	// Should find 3 edges
	if len(graph.Edges) != 3 {
		t.Errorf("expected 3 edges, got %d", len(graph.Edges))
	}
}

// TestGetEntityGraph_EmptyDatabase tests behavior with non-existent root.
func TestGetEntityGraph_EmptyDatabase(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Try to get graph for non-existent issue
	_, err := GetEntityGraph(ctx, db, "non-existent", 1, nil, 0)
	if err == nil {
		t.Error("expected error for non-existent root, got nil")
	}
}

// TestGetEntityGraph_InvalidInputs tests validation of input parameters.
func TestGetEntityGraph_InvalidInputs(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	tests := []struct {
		name     string
		rootID   string
		maxDepth int
		wantErr  bool
	}{
		{"empty root ID", "", 1, true},
		{"negative depth", "issue-1", -1, true},
		{"depth too large", "issue-1", 25, true},
		{"valid inputs", "issue-1", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetEntityGraph(ctx, db, tt.rootID, tt.maxDepth, nil, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEntityGraph() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestEntityGraph_GetSubgraph tests subgraph filtering.
func TestEntityGraph_GetSubgraph(t *testing.T) {
	// Create a mock graph
	graph := &EntityGraph{
		RootID: "root",
		Depth:  2,
		Nodes: []*GraphNode{
			{Issue: &types.Issue{ID: "root", Title: "Root", Status: "open", IssueType: "epic"}, Depth: 0},
			{Issue: &types.Issue{ID: "node1", Title: "Node 1", Status: "open", IssueType: "task"}, Depth: 1},
			{Issue: &types.Issue{ID: "node2", Title: "Node 2", Status: "closed", IssueType: "bug"}, Depth: 1},
		},
		Edges: []*GraphEdge{
			{From: "root", To: "node1", Type: types.DepParentChild},
			{From: "root", To: "node2", Type: types.DepParentChild},
		},
	}

	// Filter by depth
	subgraph := graph.GetSubgraph(
		func(depth int) bool { return depth <= 0 },
		nil,
		nil,
	)

	if len(subgraph.Nodes) != 1 {
		t.Errorf("expected 1 node at depth 0, got %d", len(subgraph.Nodes))
	}

	// Filter by status
	subgraph = graph.GetSubgraph(
		nil,
		func(issue *types.Issue) bool { return issue.Status == types.StatusClosed },
		nil,
	)

	if len(subgraph.Nodes) != 1 {
		t.Errorf("expected 1 closed node, got %d", len(subgraph.Nodes))
	}

	if subgraph.Nodes[0].Issue.ID != "node2" {
		t.Errorf("expected node2, got %s", subgraph.Nodes[0].Issue.ID)
	}
}

// TestEntityGraph_FindShortestPath tests shortest path finding.
func TestEntityGraph_FindShortestPath(t *testing.T) {
	graph := &EntityGraph{
		RootID: "root",
		Nodes: []*GraphNode{
			{Issue: &types.Issue{ID: "a"}},
			{Issue: &types.Issue{ID: "b"}},
			{Issue: &types.Issue{ID: "c"}},
			{Issue: &types.Issue{ID: "d"}},
		},
		Edges: []*GraphEdge{
			{From: "a", To: "b"},
			{From: "b", To: "c"},
			{From: "a", To: "d"},
		},
	}

	// Path from a to c should be [a, b, c]
	path := graph.FindShortestPath("a", "c")
	if len(path) != 3 {
		t.Errorf("expected path length 3, got %d", len(path))
	}

	if path[0] != "a" || path[1] != "b" || path[2] != "c" {
		t.Errorf("expected path [a, b, c], got %v", path)
	}

	// Path from a to d should be [a, d]
	path = graph.FindShortestPath("a", "d")
	if len(path) != 2 {
		t.Errorf("expected path length 2, got %d", len(path))
	}
}

// TestEntityGraph_FindClusters tests connected components detection.
func TestEntityGraph_FindClusters(t *testing.T) {
	graph := &EntityGraph{
		RootID: "root",
		Nodes: []*GraphNode{
			{Issue: &types.Issue{ID: "a"}},
			{Issue: &types.Issue{ID: "b"}},
			{Issue: &types.Issue{ID: "c"}},
			{Issue: &types.Issue{ID: "d"}},
		},
		Edges: []*GraphEdge{
			{From: "a", To: "b"},
			{From: "c", To: "d"},
		},
	}

	clusters := graph.FindClusters()

	if len(clusters) != 2 {
		t.Errorf("expected 2 clusters, got %d", len(clusters))
	}

	// Check that each cluster has 2 nodes
	for _, cluster := range clusters {
		if len(cluster) != 2 {
			t.Errorf("expected cluster size 2, got %d", len(cluster))
		}
	}
}

// TestEntityGraph_Validate tests graph validation.
func TestEntityGraph_Validate(t *testing.T) {
	// Valid graph
	validGraph := &EntityGraph{
		RootID: "root",
		Nodes: []*GraphNode{
			{Issue: &types.Issue{ID: "a"}},
			{Issue: &types.Issue{ID: "b"}},
		},
		Edges: []*GraphEdge{
			{From: "a", To: "b"},
		},
	}

	if err := validGraph.Validate(); err != nil {
		t.Errorf("expected valid graph, got error: %v", err)
	}

	// Graph with duplicate nodes
	invalidGraph := &EntityGraph{
		RootID: "root",
		Nodes: []*GraphNode{
			{Issue: &types.Issue{ID: "a"}},
			{Issue: &types.Issue{ID: "a"}},
		},
		Edges: []*GraphEdge{},
	}

	if err := invalidGraph.Validate(); err == nil {
		t.Error("expected error for duplicate nodes, got nil")
	}
}

// TestEntityGraph_ToDOT tests DOT format generation.
func TestEntityGraph_ToDOT(t *testing.T) {
	graph := &EntityGraph{
		RootID: "root",
		Nodes: []*GraphNode{
			{Issue: &types.Issue{ID: "a", Title: "Issue A"}},
			{Issue: &types.Issue{ID: "b", Title: "Issue B"}},
		},
		Edges: []*GraphEdge{
			{From: "a", To: "b", Type: types.DepBlocks},
		},
	}

	dot := graph.ToDOT()

	// Check for expected DOT syntax
	if !contains(dot, "digraph EntityGraph") {
		t.Error("DOT output should contain 'digraph EntityGraph'")
	}

	if !contains(dot, "\"a\"") || !contains(dot, "\"b\"") {
		t.Error("DOT output should contain node IDs")
	}

	if !contains(dot, "->") {
		t.Error("DOT output should contain edge syntax")
	}
}

// TestEntityGraph_FilterMethods tests various filter methods.
func TestEntityGraph_FilterMethods(t *testing.T) {
	graph := &EntityGraph{
		RootID: "root",
		Depth:  3,
		Nodes: []*GraphNode{
			{Issue: &types.Issue{ID: "node1", IssueType: types.TypeTask, Status: types.StatusOpen}, Depth: 0},
			{Issue: &types.Issue{ID: "node2", IssueType: types.TypeBug, Status: types.StatusClosed}, Depth: 1},
			{Issue: &types.Issue{ID: "node3", IssueType: types.TypeFeature, Status: types.StatusOpen}, Depth: 2},
		},
		Edges: []*GraphEdge{
			{From: "node1", To: "node2", Type: types.DepBlocks},
			{From: "node2", To: "node3", Type: types.DepParentChild},
		},
	}

	// Test FilterByDepth
	depthFiltered := graph.FilterByDepth(1)
	if len(depthFiltered.Nodes) != 2 {
		t.Errorf("expected 2 nodes at depth <= 1, got %d", len(depthFiltered.Nodes))
	}

	// Test FilterByIssueType
	typeFiltered := graph.FilterByIssueType(types.TypeBug)
	if len(typeFiltered.Nodes) != 1 {
		t.Errorf("expected 1 bug node, got %d", len(typeFiltered.Nodes))
	}

	// Test FilterByEdgeType
	edgeFiltered := graph.FilterByEdgeType(types.DepBlocks)
	if len(edgeFiltered.Edges) != 1 {
		t.Errorf("expected 1 blocks edge, got %d", len(edgeFiltered.Edges))
	}
}

// TestEntityGraph_DegreeCalculations tests in-degree, out-degree, and total degree.
func TestEntityGraph_DegreeCalculations(t *testing.T) {
	graph := &EntityGraph{
		RootID: "root",
		Nodes: []*GraphNode{
			{Issue: &types.Issue{ID: "a"}},
			{Issue: &types.Issue{ID: "b"}},
			{Issue: &types.Issue{ID: "c"}},
		},
		Edges: []*GraphEdge{
			{From: "a", To: "b"},
			{From: "a", To: "c"},
			{From: "c", To: "b"},
		},
	}

	// Node 'a': out-degree 2, in-degree 0
	if graph.GetOutDegree("a") != 2 {
		t.Errorf("expected out-degree 2 for 'a', got %d", graph.GetOutDegree("a"))
	}

	if graph.GetInDegree("a") != 0 {
		t.Errorf("expected in-degree 0 for 'a', got %d", graph.GetInDegree("a"))
	}

	// Node 'b': out-degree 0, in-degree 2
	if graph.GetOutDegree("b") != 0 {
		t.Errorf("expected out-degree 0 for 'b', got %d", graph.GetOutDegree("b"))
	}

	if graph.GetInDegree("b") != 2 {
		t.Errorf("expected in-degree 2 for 'b', got %d", graph.GetInDegree("b"))
	}

	// Node 'c': out-degree 1, in-degree 1
	if graph.GetDegree("c") != 2 {
		t.Errorf("expected total degree 2 for 'c', got %d", graph.GetDegree("c"))
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
