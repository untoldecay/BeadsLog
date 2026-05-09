package queries

import (
	"context"
	"database/sql"
	"fmt"
)

type EntityNode struct {
	ID           string
	Name         string
	Relationship string
	Depth        int
	Path         string
}

type EntityGraph struct {
	Nodes []EntityNode
}

func GetEntityGraphExact(ctx context.Context, db *sql.DB, entityName string, depth int) (*EntityGraph, error) {
	query := `
	WITH RECURSIVE graph(id, name, rel_type, depth, path) AS (
		SELECT e.id, e.name, '', 0, e.name
		FROM entities e WHERE e.name = ? 
		
		UNION ALL 
		
		SELECT e.id, e.name, ed.relationship, g.depth+1, g.path || ' → ' || e.name
		FROM entities e 
		JOIN entity_deps ed ON e.id = ed.to_entity 
		JOIN graph g ON ed.from_entity = g.id
		WHERE g.depth < ? AND g.path NOT LIKE '%' || e.name || '%'
	)
	SELECT id, name, rel_type, depth, path FROM graph ORDER BY depth;
	`

	rows, err := db.QueryContext(ctx, query, entityName, depth)
	if err != nil {
		return nil, fmt.Errorf("failed to query entity graph: %w", err)
	}
	defer rows.Close()

	return parseGraph(rows)
}

type CooccurrenceNode struct {
	Name  string
	Count int
}

// GetRelatedEntitiesByCooccurrence finds entities that frequently appear in the same sessions
func GetRelatedEntitiesByCooccurrence(ctx context.Context, db *sql.DB, entityID string, limit int) ([]CooccurrenceNode, error) {
	query := `
		SELECT e.name, COUNT(*) as co_count
		FROM session_entities se1
		JOIN session_entities se2 ON se1.session_id = se2.session_id
		JOIN entities e ON se2.entity_id = e.id
		WHERE se1.entity_id = ? AND se2.entity_id != ?
		GROUP BY e.name
		ORDER BY co_count DESC, e.name ASC
		LIMIT ?
	`
	rows, err := db.QueryContext(ctx, query, entityID, entityID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []CooccurrenceNode
	for rows.Next() {
		var node CooccurrenceNode
		if err := rows.Scan(&node.Name, &node.Count); err == nil {
			results = append(results, node)
		}
	}
	return results, nil
}

type PathStep struct {
	EntityName  string
	SessionID   string
	SessionTitle string
}

// GetPathBetweenEntities finds the shortest path between two entities via devlog sessions
func GetPathBetweenEntities(ctx context.Context, db *sql.DB, startID, endID string) ([]PathStep, error) {
	// 1. Get entity names for the path
	nameMap := make(map[string]string)
	eRows, err := db.QueryContext(ctx, "SELECT id, name FROM entities WHERE id IN (?, ?)", startID, endID)
	if err == nil {
		defer eRows.Close()
		for eRows.Next() {
			var id, name string
			if err := eRows.Scan(&id, &name); err == nil {
				nameMap[id] = name
			}
		}
	}

	// 2. Shortest Path via BFS (Breadth-First Search) using session_entities
	// path(current_entity, session_id, parent_index, depth)
	type queueItem struct {
		entityID  string
		sessionID string
		parentIdx int // Index in the visited list
		depth     int
	}

	queue := []queueItem{{entityID: startID, depth: 0, parentIdx: -1}}
	visited := []queueItem{}
	seenEntities := map[string]bool{startID: true}

	var finalItem *queueItem
	found := false

	// Limit depth to 5 to prevent infinite loops or massive memory usage
	head := 0
	for head < len(queue) {
		curr := queue[head]
		visited = append(visited, curr)
		currIdx := len(visited) - 1
		head++

		if curr.entityID == endID {
			finalItem = &visited[currIdx]
			found = true
			break
		}

		if curr.depth >= 5 {
			continue
		}

		// Find all sessions this entity appears in
		rows, err := db.QueryContext(ctx, `
			SELECT se2.entity_id, se1.session_id
			FROM session_entities se1
			JOIN session_entities se2 ON se1.session_id = se2.session_id
			WHERE se1.entity_id = ? AND se2.entity_id != ?
		`, curr.entityID, curr.entityID)
		if err != nil {
			continue
		}

		for rows.Next() {
			var nextEntity, viaSession string
			if err := rows.Scan(&nextEntity, &viaSession); err == nil {
				if !seenEntities[nextEntity] {
					seenEntities[nextEntity] = true
					queue = append(queue, queueItem{
						entityID:  nextEntity,
						sessionID: viaSession,
						parentIdx: currIdx,
						depth:     curr.depth + 1,
					})
				}
			}
		}
		rows.Close()
	}

	if !found {
		return nil, fmt.Errorf("no path found between entities within 5 hops")
	}

	// 3. Reconstruct Path
	var path []PathStep
	curr := finalItem
	for curr != nil {
		// Get entity name
		name := nameMap[curr.entityID]
		if name == "" {
			db.QueryRowContext(ctx, "SELECT name FROM entities WHERE id = ?", curr.entityID).Scan(&name)
			nameMap[curr.entityID] = name
		}

		// The sessionID in the curr item is the one that GOT us to this entity from its parent
		title := ""
		if curr.sessionID != "" {
			db.QueryRowContext(ctx, "SELECT title FROM sessions WHERE id = ?", curr.sessionID).Scan(&title)
		}

		path = append([]PathStep{{
			EntityName:   name,
			SessionID:    curr.sessionID,
			SessionTitle: title,
		}}, path...)

		if curr.parentIdx == -1 {
			break
		}
		curr = &visited[curr.parentIdx]
	}

	// Adjust: The session info belongs between A and B.
	// Current structure:
	// A (session: "")
	// B (session: S1)
	// C (session: S2)
	// This means A --[S1]--> B --[S2]--> C. This is correct for the RenderPath logic.

	return path, nil
}


func parseGraph(rows *sql.Rows) (*EntityGraph, error) {
	graph := &EntityGraph{}
	for rows.Next() {
		var node EntityNode
		if err := rows.Scan(&node.ID, &node.Name, &node.Relationship, &node.Depth, &node.Path); err != nil {
			return nil, fmt.Errorf("failed to scan graph node: %w", err)
		}
		graph.Nodes = append(graph.Nodes, node)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating graph rows: %w", err)
	}
	return graph, nil
}
