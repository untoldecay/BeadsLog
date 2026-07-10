package queries

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// AliasEntities collapses one or more alias entities into a target canonical entity.
// It merges all relationships and session mentions in a single transaction and 
// records the mapping in the registry.
func AliasEntities(ctx context.Context, db *sql.DB, targetID string, aliases []ResolvedEntity) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, alias := range aliases {
		if alias.ID == targetID {
			continue
		}

		// 1. Merge session_entities
		_, err = tx.ExecContext(ctx, `
			INSERT OR IGNORE INTO session_entities (session_id, entity_id, relevance)
			SELECT session_id, ?, relevance FROM session_entities WHERE entity_id = ?
		`, targetID, alias.ID)
		if err != nil {
			return fmt.Errorf("failed to merge session_entities: %w", err)
		}
		_, err = tx.ExecContext(ctx, "DELETE FROM session_entities WHERE entity_id = ?", alias.ID)
		if err != nil {
			return fmt.Errorf("failed to cleanup alias session_entities: %w", err)
		}

		// 2. Merge entity_deps (Outgoing)
		_, err = tx.ExecContext(ctx, `
			INSERT OR IGNORE INTO entity_deps (from_entity, to_entity, relationship, discovered_in)
			SELECT ?, to_entity, relationship, discovered_in FROM entity_deps WHERE from_entity = ?
		`, targetID, alias.ID)
		if err != nil {
			return fmt.Errorf("failed to merge outgoing entity_deps: %w", err)
		}

		// 3. Merge entity_deps (Incoming)
		_, err = tx.ExecContext(ctx, `
			INSERT OR IGNORE INTO entity_deps (from_entity, to_entity, relationship, discovered_in)
			SELECT from_entity, ?, relationship, discovered_in FROM entity_deps WHERE to_entity = ?
		`, targetID, alias.ID)
		if err != nil {
			return fmt.Errorf("failed to merge incoming entity_deps: %w", err)
		}

		_, err = tx.ExecContext(ctx, "DELETE FROM entity_deps WHERE from_entity = ? OR to_entity = ?", alias.ID, alias.ID)
		if err != nil {
			return fmt.Errorf("failed to cleanup alias entity_deps: %w", err)
		}

		// 4. Update Target Entity Stats
		var aliasMentions int
		var aliasFirstSeen string
		err = tx.QueryRowContext(ctx, "SELECT mention_count, first_seen FROM entities WHERE id = ?", alias.ID).Scan(&aliasMentions, &aliasFirstSeen)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to get alias stats: %w", err)
		}

		if err == nil {
			_, err = tx.ExecContext(ctx, `
				UPDATE entities 
				SET mention_count = mention_count + ?,
				    first_seen = MIN(first_seen, ?)
				WHERE id = ?
			`, aliasMentions, aliasFirstSeen, targetID)
			if err != nil {
				return fmt.Errorf("failed to update target entity stats: %w", err)
			}
		}

		// 5. Record in Registry
		_, err = tx.ExecContext(ctx, `
			INSERT OR REPLACE INTO entity_aliases (alias_name, canonical_id)
			VALUES (?, ?)
		`, alias.Name, targetID)
		if err != nil {
			return fmt.Errorf("failed to record alias in registry: %w", err)
		}

		// 6. Delete Alias Entity
		_, err = tx.ExecContext(ctx, "DELETE FROM entities WHERE id = ?", alias.ID)
		if err != nil {
			return fmt.Errorf("failed to delete alias entity: %w", err)
		}
	}

	return tx.Commit()
}

// UnaliasEntity removes a mapping from the registry.
func UnaliasEntity(ctx context.Context, db *sql.DB, aliasName string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM entity_aliases WHERE alias_name = ?", aliasName)
	return err
}

// normalizeEntityName strips separators so spelling variants of the same
// component compare equal: "ollama-extractor" == "ollamaextractor" == "Ollama Extractor".
func normalizeEntityName(name string) string {
	return strings.ToLower(strings.NewReplacer("-", "", "_", "", ".", "", " ", "").Replace(name))
}

// AutoAliasDuplicates merges entities whose names are identical after
// normalization. Only exact normalized matches are merged — near matches
// ("ollama" vs "ollamaextractor") remain suggestions, never auto-merges.
// The canonical entity is the one with the highest mention_count; variant
// names stay in the entity_aliases registry so future extractions resolve
// to the canonical. Returns the number of entities merged away.
func AutoAliasDuplicates(ctx context.Context, db *sql.DB) (int, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, name, mention_count FROM entities")
	if err != nil {
		return 0, err
	}
	groups := make(map[string][]ResolvedEntity)
	counts := make(map[string]int)
	for rows.Next() {
		var e ResolvedEntity
		var mentions int
		if rows.Scan(&e.ID, &e.Name, &mentions) != nil {
			continue
		}
		key := normalizeEntityName(e.Name)
		groups[key] = append(groups[key], e)
		counts[e.ID] = mentions
	}
	rows.Close()

	merged := 0
	for _, group := range groups {
		if len(group) < 2 {
			continue
		}
		canonical := group[0]
		for _, e := range group[1:] {
			if counts[e.ID] > counts[canonical.ID] {
				canonical = e
			}
		}
		var aliases []ResolvedEntity
		for _, e := range group {
			if e.ID != canonical.ID {
				aliases = append(aliases, e)
			}
		}
		if err := AliasEntities(ctx, db, canonical.ID, aliases); err != nil {
			return merged, err
		}
		merged += len(aliases)
	}
	return merged, nil
}

type AliasSuggestion struct {
	EntityA    string
	EntityB    string
	Similarity float64
}

// GetAliasSuggestions finds pairs of entities that frequently appear in the same sessions.
// It uses Jaccard similarity: intersection / union of sessions.
func GetAliasSuggestions(ctx context.Context, db *sql.DB, threshold float64) ([]AliasSuggestion, error) {
	query := `
		WITH EntitySessions AS (
			SELECT entity_id, session_id FROM session_entities
		),
		Overlap AS (
			SELECT 
				es1.entity_id as e1, 
				es2.entity_id as e2, 
				COUNT(*) as intersection_count
			FROM EntitySessions es1
			JOIN EntitySessions es2 ON es1.session_id = es2.session_id
			WHERE es1.entity_id < es2.entity_id
			GROUP BY es1.entity_id, es2.entity_id
		),
		Stats AS (
			SELECT entity_id, COUNT(*) as session_count FROM EntitySessions GROUP BY entity_id
		)
		SELECT 
			COALESCE(e1_info.preferred_name, e1_info.name), 
			COALESCE(e2_info.preferred_name, e2_info.name), 
			CAST(o.intersection_count AS REAL) / (s1.session_count + s2.session_count - o.intersection_count) as similarity
		FROM Overlap o
		JOIN Stats s1 ON o.e1 = s1.entity_id
		JOIN Stats s2 ON o.e2 = s2.entity_id
		JOIN entities e1_info ON o.e1 = e1_info.id
		JOIN entities e2_info ON o.e2 = e2_info.id
		WHERE similarity >= ?
		ORDER BY similarity DESC
	`

	rows, err := db.QueryContext(ctx, query, threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suggestions []AliasSuggestion
	for rows.Next() {
		var s AliasSuggestion
		if err := rows.Scan(&s.EntityA, &s.EntityB, &s.Similarity); err != nil {
			return nil, err
		}
		suggestions = append(suggestions, s)
	}

	return suggestions, nil
}
