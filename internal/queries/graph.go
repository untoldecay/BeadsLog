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
	// Breadth-First Search for shortest path
	/*
	query := `
	WITH RECURSIVE path(entity_id, path_json, depth) AS (
		SELECT ?, json_array(?), 0
		UNION ALL
		SELECT 
			se2.entity_id, 
			json_insert(p.path_json, '$[' || json_array_length(p.path_json) || ']', json_object('entity_id', se2.entity_id, 'session_id', se1.session_id)),
			p.depth + 1
		FROM path p
		JOIN session_entities se1 ON p.entity_id = se1.entity_id
		JOIN session_entities se2 ON se1.session_id = se2.session_id
		WHERE p.depth < 5 AND p.entity_id != se2.entity_id
	)
	SELECT path_json FROM path WHERE entity_id = ? ORDER BY depth LIMIT 1;
	`
	*/
	// Note: Complex recursive path finding in SQLite might be tricky without json extensions or specific versions.
	return nil, fmt.Errorf("path search not yet implemented in v1")
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
