// Package queries provides complex queries for traversing and analyzing entity relationships.
package queries

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"

	"github.com/steveyegge/beads/internal/types"
)

// EntityGraph represents a graph of entities and their relationships.
// Used for visualizing and analyzing entity connections up to a specified depth.
type EntityGraph struct {
	// RootID is the starting entity ID for the graph
	RootID string `json:"root_id"`

	// Depth is the maximum depth traversed from the root
	Depth int `json:"depth"`

	// Nodes contains all entities in the graph
	Nodes []*GraphNode `json:"nodes"`

	// Edges contains all relationships between entities
	Edges []*GraphEdge `json:"edges"`

	// Stats provides summary statistics about the graph
	Stats *GraphStats `json:"stats,omitempty"`
}

// GraphNode represents an entity in the graph with its relationship metadata.
type GraphNode struct {
	// Issue is the entity data
	Issue *types.Issue `json:"issue"`

	// Depth is the distance from the root entity (0 = root)
	Depth int `json:"depth"`

	// Path is the sequence of entity IDs from root to this node
	Path []string `json:"path,omitempty"`
}

// GraphEdge represents a relationship between two entities.
type GraphEdge struct {
	// From is the source entity ID
	From string `json:"from"`

	// To is the target entity ID
	To string `json:"to"`

	// Type is the dependency type
	Type types.DependencyType `json:"type"`

	// Metadata contains optional edge metadata
	Metadata string `json:"metadata,omitempty"`
}

// GraphStats provides summary statistics about the entity graph.
type GraphStats struct {
	// TotalNodes is the total number of nodes in the graph
	TotalNodes int `json:"total_nodes"`

	// TotalEdges is the total number of edges in the graph
	TotalEdges int `json:"total_edges"`

	// NodesByDepth counts nodes at each depth level
	NodesByDepth map[int]int `json:"nodes_by_depth"`

	// EdgeTypes counts edges by dependency type
	EdgeTypes map[types.DependencyType]int `json:"edge_types"`

	// MaxDepthReached is the actual maximum depth reached (may be less than requested)
	MaxDepthReached int `json:"max_depth_reached"`

	// Truncated indicates if the graph was truncated due to depth or size limits
	Truncated bool `json:"truncated"`
}

// GetEntityGraph retrieves a graph of entity relationships starting from rootID.
// Uses recursive SQL CTE to traverse relationships up to the specified depth.
//
// Parameters:
//   - ctx: Context for the database operation
//   - db: Database connection (can be *sql.DB or *sql.Tx)
//   - rootID: The starting entity ID
//   - maxDepth: Maximum depth to traverse (0 = just root, 1 = immediate neighbors, etc.)
//   - edgeTypes: Optional filter for dependency types (empty = all types)
//   - maxNodes: Maximum number of nodes to return (0 = unlimited, for safety)
//
// Returns:
//   - *EntityGraph: The complete graph structure with nodes, edges, and statistics
//   - error: Database or parsing errors
//
// The graph includes both upstream (dependencies) and downstream (dependents) relationships,
// providing a complete view of the entity's relationship network.
func GetEntityGraph(
	ctx context.Context,
	db databaseConnection,
	rootID string,
	maxDepth int,
	edgeTypes []types.DependencyType,
	maxNodes int,
) (*EntityGraph, error) {
	// Validate inputs
	if rootID == "" {
		return nil, fmt.Errorf("rootID cannot be empty")
	}
	if maxDepth < 0 {
		return nil, fmt.Errorf("maxDepth must be non-negative")
	}
	if maxDepth > 20 {
		return nil, fmt.Errorf("maxDepth cannot exceed 20 (to prevent excessive queries)")
	}

	// Build the recursive CTE query
	query := buildGraphCTE(rootID, maxDepth, edgeTypes)

	// Execute the query
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute graph query: %w", err)
	}
	defer rows.Close()

	// Parse results into graph structure
	graph := &EntityGraph{
		RootID: rootID,
		Depth:  maxDepth,
		Nodes:  make([]*GraphNode, 0),
		Edges:  make([]*GraphEdge, 0),
	}

	// Track unique nodes and edges to avoid duplicates
	nodeMap := make(map[string]*GraphNode)
	edgeMap := make(map[string]*GraphEdge)
	depthMap := make(map[int]int)
	edgeTypeMap := make(map[types.DependencyType]int)
	maxDepthReached := 0

	for rows.Next() {
		var (
			issueID         sql.NullString
			title           sql.NullString
			status          sql.NullString
			issueType       sql.NullString
			priority        sql.NullInt64
			assignee        sql.NullString
			createdAt       sql.NullTime
			updatedAt       sql.NullTime
			depth           sql.NullInt64
			path            sql.NullString
			relatedID       sql.NullString
			edgeType        sql.NullString
			edgeMetadata    sql.NullString
		)

		if err := rows.Scan(
			&issueID, &title, &status, &issueType, &priority,
			&assignee, &createdAt, &updatedAt,
			&depth, &path, &relatedID, &edgeType, &edgeMetadata,
		); err != nil {
			return nil, fmt.Errorf("failed to scan graph row: %w", err)
		}

		// Track maximum depth reached
		d := int(depth.Int64)
		if d > maxDepthReached {
			maxDepthReached = d
		}
		depthMap[d]++

		// Create node if not exists
		if issueID.Valid && issueID.String != "" {
			if _, exists := nodeMap[issueID.String]; !exists {
				// Parse path
				var pathList []string
				if path.Valid && path.String != "" {
					pathList = parsePath(path.String)
				}

				node := &GraphNode{
					Issue: &types.Issue{
						ID:        issueID.String,
						Title:     title.String,
						Status:    types.Status(status.String),
						IssueType: types.IssueType(issueType.String),
						Priority:  int(priority.Int64),
						Assignee:  assignee.String,
						CreatedAt: createdAt.Time,
						UpdatedAt: updatedAt.Time,
					},
					Depth: d,
					Path:  pathList,
				}
				nodeMap[issueID.String] = node
				graph.Nodes = append(graph.Nodes, node)

				// Check node limit
				if maxNodes > 0 && len(nodeMap) >= maxNodes {
					graph.Stats = &GraphStats{
						TotalNodes:      len(nodeMap),
						TotalEdges:      len(edgeMap),
						NodesByDepth:    depthMap,
						EdgeTypes:       edgeTypeMap,
						MaxDepthReached: maxDepthReached,
						Truncated:       true,
					}
					return graph, nil
				}
			}
		}

		// Create edge if we have a related entity
		if relatedID.Valid && relatedID.String != "" && edgeType.Valid && edgeType.String != "" {
			edgeKey := fmt.Sprintf("%s->%s:%s", issueID.String, relatedID.String, edgeType.String)
			if _, exists := edgeMap[edgeKey]; !exists {
				edge := &GraphEdge{
					From:     issueID.String,
					To:       relatedID.String,
					Type:     types.DependencyType(edgeType.String),
					Metadata: edgeMetadata.String,
				}
				edgeMap[edgeKey] = edge
				graph.Edges = append(graph.Edges, edge)
				edgeTypeMap[types.DependencyType(edgeType.String)]++
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating graph rows: %w", err)
	}

	// Build statistics
	graph.Stats = &GraphStats{
		TotalNodes:      len(nodeMap),
		TotalEdges:      len(edgeMap),
		NodesByDepth:    depthMap,
		EdgeTypes:       edgeTypeMap,
		MaxDepthReached: maxDepthReached,
		Truncated:       false,
	}

	return graph, nil
}

// databaseConnection is an interface that both *sql.DB and *sql.Tx implement.
// This allows the function to work with both direct connections and transactions.
type databaseConnection interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// buildGraphCTE constructs the recursive CTE query for traversing entity relationships.
// The query explores both upstream (dependencies) and downstream (dependents) relationships
// to build a complete graph of the entity's network.
func buildGraphCTE(rootID string, maxDepth int, edgeTypes []types.DependencyType) string {
	// Build edge type filter
	var edgeTypeFilter string
	if len(edgeTypes) > 0 {
		typeList := make([]string, len(edgeTypes))
		for i, t := range edgeTypes {
			typeList[i] = fmt.Sprintf("'%s'", t)
		}
		edgeTypeFilter = fmt.Sprintf("AND d.type IN (%s)", fmt.Sprintf("%s", typeList))
	}

	// Escape rootID for SQL
	escapedRootID := fmt.Sprintf("'%s'", rootID)

	// Recursive CTE to traverse relationships in both directions
	query := fmt.Sprintf(`
WITH RECURSIVE entity_graph AS (
  -- Base case: Start with root entity
  SELECT
    i.id,
    i.title,
    i.status,
    i.issue_type,
    i.priority,
    i.assignee,
    i.created_at,
    i.updated_at,
    0 as depth,
    i.id as path,
    NULL as related_id,
    NULL as edge_type,
    NULL as edge_metadata
  FROM issues i
  WHERE i.id = %s

  UNION ALL

  -- Recursive case: Find related entities (both dependencies and dependents)
  SELECT
    related.id,
    related.title,
    related.status,
    related.issue_type,
    related.priority,
    related.assignee,
    related.created_at,
    related.updated_at,
    eg.depth + 1,
    eg.path || ',' || related.id,
    related.id as related_id,
    d.type as edge_type,
    d.metadata as edge_metadata
  FROM entity_graph eg
  CROSS JOIN (
    -- Get entities that the current node depends on (upstream)
    SELECT d2.depends_on_id as target_id, d2.type, d2.metadata
    FROM dependencies d2
    WHERE d2.issue_id = eg.id
      %s

    UNION

    -- Get entities that depend on the current node (downstream)
    SELECT d3.issue_id as target_id, d3.type, d3.metadata
    FROM dependencies d3
    WHERE d3.depends_on_id = eg.id
      %s
  ) d
  JOIN issues related ON d.target_id = related.id
  WHERE eg.depth < %d
    -- Avoid cycles by checking if related_id is already in path
    AND ',' || eg.path || ',' NOT LIKE '%%,' || related.id || ',%%'
)
SELECT DISTINCT
  id,
  title,
  status,
  issue_type,
  priority,
  assignee,
  created_at,
  updated_at,
  depth,
  path,
  related_id,
  edge_type,
  edge_metadata
FROM entity_graph
ORDER BY depth, id;
`, escapedRootID, edgeTypeFilter, edgeTypeFilter, maxDepth)

	return query
}

// parsePath converts a comma-separated path string into a slice of IDs.
func parsePath(path string) []string {
	if path == "" {
		return nil
	}

	// Split by comma and filter empty strings
	parts := strings.Split(path, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// GetSubgraph retrieves a filtered subgraph containing only nodes that match the criteria.
// Useful for extracting specific patterns from a larger graph.
func (g *EntityGraph) GetSubgraph(depthFilter func(int) bool, nodeFilter func(*types.Issue) bool, edgeTypeFilter func(types.DependencyType) bool) *EntityGraph {
	subgraph := &EntityGraph{
		RootID: g.RootID,
		Depth:  g.Depth,
		Nodes:  make([]*GraphNode, 0),
		Edges:  make([]*GraphEdge, 0),
	}

	// Filter nodes
	nodeMap := make(map[string]bool)
	for _, node := range g.Nodes {
		if depthFilter != nil && !depthFilter(node.Depth) {
			continue
		}
		if nodeFilter != nil && !nodeFilter(node.Issue) {
			continue
		}
		subgraph.Nodes = append(subgraph.Nodes, node)
		nodeMap[node.Issue.ID] = true
	}

	// Filter edges (only include edges between filtered nodes)
	for _, edge := range g.Edges {
		if edgeTypeFilter != nil && !edgeTypeFilter(edge.Type) {
			continue
		}
		if !nodeMap[edge.From] || !nodeMap[edge.To] {
			continue
		}
		subgraph.Edges = append(subgraph.Edges, edge)
	}

	// Recalculate stats
	subgraph.Stats = calculateStats(subgraph)

	return subgraph
}

// calculateStats computes statistics for a graph.
func calculateStats(g *EntityGraph) *GraphStats {
	stats := &GraphStats{
		TotalNodes:   len(g.Nodes),
		TotalEdges:   len(g.Edges),
		NodesByDepth: make(map[int]int),
		EdgeTypes:    make(map[types.DependencyType]int),
	}

	maxDepth := 0
	for _, node := range g.Nodes {
		stats.NodesByDepth[node.Depth]++
		if node.Depth > maxDepth {
			maxDepth = node.Depth
		}
	}

	for _, edge := range g.Edges {
		stats.EdgeTypes[edge.Type]++
	}

	stats.MaxDepthReached = maxDepth
	stats.Truncated = false

	return stats
}

// GetNeighbors returns immediate neighbors of a node (depth 1 relationships).
func (g *EntityGraph) GetNeighbors(nodeID string) []*GraphNode {
	var neighbors []*GraphNode
	for _, node := range g.Nodes {
		if node.Depth == 1 && len(node.Path) > 1 {
			// Check if this node is directly connected to root
			for _, edge := range g.Edges {
				if (edge.From == g.RootID && edge.To == node.Issue.ID) ||
					(edge.To == g.RootID && edge.From == node.Issue.ID) {
					neighbors = append(neighbors, node)
					break
				}
			}
		}
	}
	return neighbors
}

// GetPath returns the path from root to the specified node.
func (g *EntityGraph) GetPath(nodeID string) []string {
	for _, node := range g.Nodes {
		if node.Issue.ID == nodeID {
			return node.Path
		}
	}
	return nil
}

// FindShortestPath finds the shortest path between two nodes in the graph.
// Returns the sequence of node IDs from start to end, or nil if no path exists.
func (g *EntityGraph) FindShortestPath(fromID, toID string) []string {
	// Build adjacency list
	adj := make(map[string][]string)
	for _, edge := range g.Edges {
		adj[edge.From] = append(adj[edge.From], edge.To)
		adj[edge.To] = append(adj[edge.To], edge.From)
	}

	// BFS to find shortest path
	queue := [][]string{{fromID}}
	visited := make(map[string]bool)
	visited[fromID] = true

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]

		current := path[len(path)-1]
		if current == toID {
			return path
		}

		for _, neighbor := range adj[current] {
			if !visited[neighbor] {
				visited[neighbor] = true
				newPath := make([]string, len(path)+1)
				copy(newPath, path)
				newPath[len(path)] = neighbor
				queue = append(queue, newPath)
			}
		}
	}

	return nil
}

// CalculateCentrality computes degree centrality for all nodes in the graph.
// Returns a map of node ID to centrality score (0.0 to 1.0).
func (g *EntityGraph) CalculateCentrality() map[string]float64 {
	centrality := make(map[string]float64)

	// Count edges per node
	edgeCount := make(map[string]int)
	for _, edge := range g.Edges {
		edgeCount[edge.From]++
		edgeCount[edge.To]++
	}

	// Normalize by max possible edges (n-1 for undirected graph)
	maxEdges := len(g.Nodes) - 1
	if maxEdges > 0 {
		for nodeID, count := range edgeCount {
			centrality[nodeID] = float64(count) / float64(maxEdges)
		}
	}

	return centrality
}

// FindClusters identifies clusters in the graph using connected components.
// Returns a list of clusters, where each cluster is a list of node IDs.
func (g *EntityGraph) FindClusters() [][]string {
	visited := make(map[string]bool)
	clusters := [][]string{}

	for _, node := range g.Nodes {
		nodeID := node.Issue.ID
		if visited[nodeID] {
			continue
		}

		// BFS to find all nodes in this cluster
		cluster := []string{}
		queue := []string{nodeID}
		visited[nodeID] = true

		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			cluster = append(cluster, current)

			// Find neighbors
			for _, edge := range g.Edges {
				var neighbor string
				if edge.From == current {
					neighbor = edge.To
				} else if edge.To == current {
					neighbor = edge.From
				} else {
					continue
				}

				if !visited[neighbor] {
					visited[neighbor] = true
					queue = append(queue, neighbor)
				}
			}
		}

		clusters = append(clusters, cluster)
	}

	return clusters
}

// GetCriticalPath finds the longest path through the graph (critical path analysis).
// Useful for identifying the sequence of dependencies that determines minimum completion time.
func (g *EntityGraph) GetCriticalPath() []string {
	// Build adjacency list and calculate depths
	adj := make(map[string][]string)
	inDegree := make(map[string]int)
	nodeSet := make(map[string]bool)

	for _, node := range g.Nodes {
		nodeSet[node.Issue.ID] = true
		inDegree[node.Issue.ID] = 0
	}

	for _, edge := range g.Edges {
		adj[edge.From] = append(adj[edge.From], edge.To)
		inDegree[edge.To]++
	}

	// Topological sort with longest path tracking
	dist := make(map[string]int)
	prev := make(map[string]string)
	for nodeID := range nodeSet {
		dist[nodeID] = 0
	}

	// Process nodes in topological order
	for i := 0; i < len(nodeSet); i++ {
		for nodeID := range nodeSet {
			if inDegree[nodeID] == 0 {
				for _, neighbor := range adj[nodeID] {
					if dist[neighbor] < dist[nodeID]+1 {
						dist[neighbor] = dist[nodeID] + 1
						prev[neighbor] = nodeID
					}
					inDegree[neighbor]--
				}
				inDegree[nodeID] = -1 // Mark as processed
			}
		}
	}

	// Find node with maximum distance
	maxDist := -1
	endNode := ""
	for nodeID, d := range dist {
		if d > maxDist {
			maxDist = d
			endNode = nodeID
		}
	}

	// Reconstruct path
	if maxDist <= 0 {
		return []string{g.RootID}
	}

	path := []string{}
	current := endNode
	for current != "" {
		path = append([]string{current}, path...)
		current = prev[current]
	}

	return path
}

// Validate checks if the graph is valid (no cycles, consistent data).
func (g *EntityGraph) Validate() error {
	// Check for duplicate nodes
	nodeIDs := make(map[string]bool)
	for _, node := range g.Nodes {
		if nodeIDs[node.Issue.ID] {
			return fmt.Errorf("duplicate node: %s", node.Issue.ID)
		}
		nodeIDs[node.Issue.ID] = true
	}

	// Check for edges referencing non-existent nodes
	for _, edge := range g.Edges {
		if !nodeIDs[edge.From] {
			return fmt.Errorf("edge references non-existent node: %s", edge.From)
		}
		if !nodeIDs[edge.To] {
			return fmt.Errorf("edge references non-existent node: %s", edge.To)
		}
	}

	// Check for cycles in the graph
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(nodeID string) bool
	hasCycle = func(nodeID string) bool {
		visited[nodeID] = true
		recStack[nodeID] = true

		for _, edge := range g.Edges {
			if edge.From == nodeID {
				if !visited[edge.To] {
					if hasCycle(edge.To) {
						return true
					}
				} else if recStack[edge.To] {
					return true
				}
			}
		}

		recStack[nodeID] = false
		return false
	}

	for _, node := range g.Nodes {
		if !visited[node.Issue.ID] {
			if hasCycle(node.Issue.ID) {
				return fmt.Errorf("graph contains cycles")
			}
		}
	}

	return nil
}

// ToDOT converts the graph to DOT format for visualization with Graphviz.
func (g *EntityGraph) ToDOT() string {
	dot := "digraph EntityGraph {\n"
	dot += "  rankdir=TB;\n"
	dot += "  node [shape=box];\n\n"

	// Add nodes
	for _, node := range g.Nodes {
		label := fmt.Sprintf("%s\\n%s", node.Issue.ID, node.Issue.Title)
		dot += fmt.Sprintf("  \"%s\" [label=\"%s\"];\n", node.Issue.ID, label)
	}

	dot += "\n"

	// Add edges
	for _, edge := range g.Edges {
		style := ""
		switch edge.Type {
		case types.DepBlocks:
			style = " [style=bold, color=red]"
		case types.DepParentChild:
			style = " [style=dashed]"
		}
		dot += fmt.Sprintf("  \"%s\" -> \"%s\" [label=\"%s\"%s];\n",
			edge.From, edge.To, edge.Type, style)
	}

	dot += "}\n"
	return dot
}

// GetNodeByID retrieves a node by its ID.
func (g *EntityGraph) GetNodeByID(id string) *GraphNode {
	for _, node := range g.Nodes {
		if node.Issue.ID == id {
			return node
		}
	}
	return nil
}

// GetEdgesByType returns all edges of a specific type.
func (g *EntityGraph) GetEdgesByType(edgeType types.DependencyType) []*GraphEdge {
	var edges []*GraphEdge
	for _, edge := range g.Edges {
		if edge.Type == edgeType {
			edges = append(edges, edge)
		}
	}
	return edges
}

// GetNodesByDepth returns all nodes at a specific depth level.
func (g *EntityGraph) GetNodesByDepth(depth int) []*GraphNode {
	var nodes []*GraphNode
	for _, node := range g.Nodes {
		if node.Depth == depth {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// GetNodesByType returns all nodes of a specific issue type.
func (g *EntityGraph) GetNodesByType(issueType types.IssueType) []*GraphNode {
	var nodes []*GraphNode
	for _, node := range g.Nodes {
		if node.Issue.IssueType == issueType {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// GetNodesByStatus returns all nodes with a specific status.
func (g *EntityGraph) GetNodesByStatus(status types.Status) []*GraphNode {
	var nodes []*GraphNode
	for _, node := range g.Nodes {
		if node.Issue.Status == status {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// FilterByDepth returns a new graph containing only nodes up to the specified depth.
func (g *EntityGraph) FilterByDepth(maxDepth int) *EntityGraph {
	return g.GetSubgraph(
		func(depth int) bool { return depth <= maxDepth },
		nil,
		nil,
	)
}

// FilterByIssueType returns a new graph containing only nodes of specified types.
func (g *EntityGraph) FilterByIssueType(types ...types.IssueType) *EntityGraph {
	typeSet := make(map[types.IssueType]bool)
	for _, t := range types {
		typeSet[t] = true
	}

	return g.GetSubgraph(
		nil,
		func(issue *types.Issue) bool { return typeSet[issue.IssueType] },
		nil,
	)
}

// FilterByEdgeType returns a new graph containing only edges of specified types.
func (g *EntityGraph) FilterByEdgeType(types ...types.DependencyType) *EntityGraph {
	typeSet := make(map[types.DependencyType]bool)
	for _, t := range types {
		typeSet[t] = true
	}

	return g.GetSubgraph(
		nil,
		nil,
		func(edgeType types.DependencyType) bool { return typeSet[edgeType] },
	)
}

// EstimateComplexity returns a heuristic estimate of graph complexity.
// Higher values indicate more complex graphs (more nodes, edges, or deeper structures).
func (g *EntityGraph) EstimateComplexity() float64 {
	if len(g.Nodes) == 0 {
		return 0
	}

	// Complexity factors:
	// - Node count (linear factor)
	// - Edge count (quadratic factor potential)
	// - Max depth (depth factor)
	// - Edge type diversity (diversity factor)

	nodeFactor := float64(len(g.Nodes))
	edgeFactor := math.Pow(float64(len(g.Edges)), 1.5) // Slightly super-linear
	depthFactor := math.Pow(float64(g.Stats.MaxDepthReached+1), 2)

	// Count unique edge types
	uniqueEdgeTypes := len(g.Stats.EdgeTypes)
	diversityFactor := float64(uniqueEdgeTypes) * 10

	complexity := nodeFactor + edgeFactor + depthFactor + diversityFactor
	return complexity
}

// FindBridges identifies bridge edges in the graph.
// A bridge is an edge whose removal increases the number of connected components.
func (g *EntityGraph) FindBridges() []*GraphEdge {
	bridges := []*GraphEdge{}

	// For each edge, test if removing it disconnects the graph
	for _, edge := range g.Edges {
		// Create a temporary graph without this edge
		tempEdges := make([]*GraphEdge, 0, len(g.Edges)-1)
		for _, e := range g.Edges {
			if e != edge {
				tempEdges = append(tempEdges, e)
			}
		}

		// Check if graph is still connected
		if !g.isConnected(tempEdges) {
			bridges = append(bridges, edge)
		}
	}

	return bridges
}

// isConnected checks if the graph is connected with the given edges.
func (g *EntityGraph) isConnected(edges []*GraphEdge) bool {
	if len(g.Nodes) == 0 {
		return true
	}

	// Build adjacency list
	adj := make(map[string][]string)
	for _, edge := range edges {
		adj[edge.From] = append(adj[edge.From], edge.To)
		adj[edge.To] = append(adj[edge.To], edge.From)
	}

	// BFS from first node
	startNode := g.Nodes[0].Issue.ID
	visited := make(map[string]bool)
	queue := []string{startNode}
	visited[startNode] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, neighbor := range adj[current] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}

	// Check if all nodes were visited
	return len(visited) == len(g.Nodes)
}

// GetInDegree returns the in-degree (number of incoming edges) for a node.
func (g *EntityGraph) GetInDegree(nodeID string) int {
	count := 0
	for _, edge := range g.Edges {
		if edge.To == nodeID {
			count++
		}
	}
	return count
}

// GetOutDegree returns the out-degree (number of outgoing edges) for a node.
func (g *EntityGraph) GetOutDegree(nodeID string) int {
	count := 0
	for _, edge := range g.Edges {
		if edge.From == nodeID {
			count++
		}
	}
	return count
}

// GetDegree returns the total degree (in + out) for a node.
func (g *EntityGraph) GetDegree(nodeID string) int {
	return g.GetInDegree(nodeID) + g.GetOutDegree(nodeID)
}
