# Queries Package

The `queries` package provides complex queries for traversing and analyzing entity relationships in the devlog issue tracking system.

## Overview

This package implements sophisticated graph traversal algorithms using recursive SQL Common Table Expressions (CTEs) to efficiently explore entity relationships up to specified depths. It's designed for visualizing and analyzing dependency networks, entity connections, and relationship patterns.

## Main Features

### GetEntityGraph Function

The core function `GetEntityGraph` retrieves a complete graph of entity relationships starting from a root entity.

**Signature:**
```go
func GetEntityGraph(
    ctx context.Context,
    db databaseConnection,
    rootID string,
    maxDepth int,
    edgeTypes []types.DependencyType,
    maxNodes int,
) (*EntityGraph, error)
```

**Parameters:**
- `ctx`: Context for the database operation
- `db`: Database connection (*sql.DB or *sql.Tx)
- `rootID`: The starting entity ID
- `maxDepth`: Maximum depth to traverse (0 = just root, 1 = immediate neighbors)
- `edgeTypes`: Optional filter for dependency types (nil = all types)
- `maxNodes`: Maximum number of nodes to return (0 = unlimited)

**Returns:**
- `*EntityGraph`: Complete graph structure with nodes, edges, and statistics
- `error`: Database or parsing errors

### Graph Traversal

The implementation uses a **recursive SQL CTE** that:

1. **Starts with the root entity** (base case)
2. **Recursively finds related entities** in both directions:
   - **Upstream**: Entities that the current node depends on
   - **Downstream**: Entities that depend on the current node
3. **Tracks depth and path** to prevent cycles
4. **Respects depth limits** to control traversal scope

**Key Features:**
- Bidirectional traversal (finds both dependencies and dependents)
- Cycle detection and prevention
- Depth-limited exploration
- Optional edge type filtering
- Node count limiting for safety

## Data Structures

### EntityGraph

```go
type EntityGraph struct {
    RootID string              // Starting entity ID
    Depth  int                 // Maximum depth traversed
    Nodes  []*GraphNode        // All entities in the graph
    Edges  []*GraphEdge        // All relationships
    Stats  *GraphStats         // Summary statistics
}
```

### GraphNode

```go
type GraphNode struct {
    Issue *types.Issue    // Entity data
    Depth int             // Distance from root (0 = root)
    Path  []string        // Sequence of IDs from root to this node
}
```

### GraphEdge

```go
type GraphEdge struct {
    From     string                    // Source entity ID
    To       string                    // Target entity ID
    Type     types.DependencyType      // Relationship type
    Metadata string                    // Optional edge metadata
}
```

### GraphStats

```go
type GraphStats struct {
    TotalNodes      int                               // Total nodes in graph
    TotalEdges      int                               // Total edges in graph
    NodesByDepth    map[int]int                       // Node count per depth level
    EdgeTypes       map[types.DependencyType]int      // Edge count per type
    MaxDepthReached int                               // Actual max depth reached
    Truncated       bool                              // Whether graph was truncated
}
```

## Usage Examples

### Basic Usage - Get Graph at Depth 2

```go
import (
    "context"
    "github.com/steveyegge/beads/internal/queries"
)

func main() {
    ctx := context.Background()
    db := getDatabase() // *sql.DB or *sql.Tx

    // Get graph up to depth 2 (root + 2 levels)
    graph, err := queries.GetEntityGraph(ctx, db, "issue-123", 2, nil, 0)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Graph contains %d nodes and %d edges\n",
        graph.Stats.TotalNodes, graph.Stats.TotalEdges)
}
```

### Filter by Dependency Type

```go
// Only include parent-child relationships
graph, err := queries.GetEntityGraph(
    ctx,
    db,
    "epic-456",
    5,
    []types.DependencyType{types.DepParentChild},
    0,
)
```

### Limit Node Count

```go
// Get up to 100 nodes to prevent excessive results
graph, err := queries.GetEntityGraph(
    ctx,
    db,
    "issue-789",
    10,
    nil,
    100, // Max 100 nodes
)
```

### Work with Graph Results

```go
graph, _ := queries.GetEntityGraph(ctx, db, "root", 3, nil, 0)

// Get nodes at specific depth
depth1Nodes := graph.GetNodesByDepth(1)

// Filter by issue type
bugs := graph.GetNodesByType(types.TypeBug)

// Find shortest path between two nodes
path := graph.FindShortestPath("issue-a", "issue-b")

// Export to Graphviz DOT format
dot := graph.ToDOT()
```

## Graph Analysis Methods

The `EntityGraph` type provides numerous analysis methods:

### Filtering

- `GetSubgraph()` - Extract filtered subgraph
- `FilterByDepth()` - Filter by depth level
- `FilterByIssueType()` - Filter by issue type
- `FilterByEdgeType()` - Filter by edge type

### Navigation

- `GetNeighbors()` - Get immediate neighbors
- `GetPath()` - Get path from root to node
- `FindShortestPath()` - Find shortest path between nodes
- `GetNodesByDepth()` - Get nodes at specific depth
- `GetNodesByType()` - Get nodes by issue type
- `GetNodesByStatus()` - Get nodes by status

### Analysis

- `CalculateCentrality()` - Calculate degree centrality
- `FindClusters()` - Find connected components
- `GetCriticalPath()` - Find longest dependency path
- `EstimateComplexity()` - Estimate graph complexity
- `FindBridges()` - Find bridge edges
- `GetInDegree()` - Get incoming edge count
- `GetOutDegree()` - Get outgoing edge count
- `GetDegree()` - Get total degree

### Export

- `ToDOT()` - Export to Graphviz DOT format
- `Validate()` - Validate graph structure

## Implementation Details

### Recursive CTE Query

The core traversal uses a recursive SQL CTE:

```sql
WITH RECURSIVE entity_graph AS (
  -- Base case: Start with root entity
  SELECT id, title, ..., 0 as depth, id as path, ...
  FROM issues WHERE id = ?

  UNION ALL

  -- Recursive case: Find related entities
  SELECT related.id, related.title, ..., eg.depth + 1,
         eg.path || ',' || related.id, ...
  FROM entity_graph eg
  JOIN dependencies d ON (d.issue_id = eg.id OR d.depends_on_id = eg.id)
  JOIN issues related ON (d.target_id = related.id)
  WHERE eg.depth < ?
    AND ',' || eg.path || ',' NOT LIKE '%,' || related.id || ',%'  -- Cycle prevention
)
SELECT DISTINCT * FROM entity_graph ORDER BY depth, id;
```

### Cycle Prevention

Cycles are prevented by tracking the traversal path and excluding nodes already in the path:

```sql
AND ',' || eg.path || ',' NOT LIKE '%,' || related.id || ',%'
```

This ensures we never revisit a node that's already in our current path, preventing infinite loops in cyclic graphs.

### Bidirectional Traversal

The query explores relationships in both directions:

1. **Upstream**: `WHERE d.issue_id = eg.id` - finds entities that current node depends on
2. **Downstream**: `WHERE d.depends_on_id = eg.id` - finds entities that depend on current node

This provides a complete view of the entity's relationship network.

## Performance Considerations

### Complexity

- **Time**: O(V + E) where V = vertices (nodes), E = edges
- **Space**: O(V * maxDepth) for path tracking
- **Database**: Recursive CTE with indexed lookups

### Optimization Tips

1. **Limit depth**: Use `maxDepth` to control traversal scope
2. **Filter edge types**: Use `edgeTypes` to only traverse relevant relationships
3. **Set node limits**: Use `maxNodes` to prevent excessive results
4. **Use indexes**: Ensure `dependencies(issue_id)` and `dependencies(depends_on_id)` are indexed

### Recommended Settings

- **Small graphs**: `maxDepth = 2-3`, `maxNodes = 0` (unlimited)
- **Medium graphs**: `maxDepth = 3-5`, `maxNodes = 1000`
- **Large graphs**: `maxDepth = 1-2`, `maxNodes = 100`

## Testing

The package includes comprehensive tests in `graph_test.go`:

- Unit tests for all major functions
- Integration tests with in-memory SQLite
- Edge case tests (cycles, empty graphs, etc.)
- Performance and validation tests

Run tests:
```bash
go test ./internal/queries/...
```

## Use Cases

### 1. Dependency Visualization

```go
graph, _ := queries.GetEntityGraph(ctx, db, "epic-123", 5, nil, 0)
dot := graph.ToDOT()
// Use Graphviz to render the dependency graph
```

### 2. Impact Analysis

```go
// Find all issues affected by closing an issue
graph, _ := queries.GetEntityGraph(ctx, db, "issue-456", 10, nil, 0)
downstream := graph.GetNodesByDepth(1) // Immediate dependents
```

### 3. Critical Path Analysis

```go
graph, _ := queries.GetEntityGraph(ctx, db, "feature-start", 20, nil, 0)
criticalPath := graph.GetCriticalPath()
fmt.Println("Critical path:", criticalPath)
```

### 4. Cluster Detection

```go
graph, _ := queries.GetEntityGraph(ctx, db, "root", 5, nil, 0)
clusters := graph.FindClusters()
fmt.Printf("Found %d disconnected clusters\n", len(clusters))
```

## Future Enhancements

Potential improvements for future versions:

- [ ] Support for weighted edges (priority, severity)
- [ ] Graph layout algorithms (force-directed, hierarchical)
- [ ] Export to other formats (JSON, GraphML, GEXF)
- [ ] Incremental graph updates (add/remove nodes/edges)
- [ ] Graph metrics (clustering coefficient, centrality variants)
- [ ] Subgraph pattern matching
- [ ] Temporal graph traversal (time-based relationships)

## Contributing

When adding new features:

1. Add comprehensive tests
2. Update this README
3. Consider performance implications
4. Handle edge cases (cycles, empty graphs, etc.)
5. Follow Go best practices and idioms

## License

This package is part of the devlog project and follows the same license terms.
