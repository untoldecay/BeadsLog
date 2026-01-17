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
