package queries

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/untoldecay/BeadsLog/internal/utils"
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
	EntityA        string  `json:"entity_a"`
	EntityB        string  `json:"entity_b"`
	Similarity     float64 `json:"session_overlap"`
	NameSimilarity float64 `json:"name_similarity"`
	Score          float64 `json:"score"`
}
// It uses Jaccard similarity: intersection / union of sessions.
// GetAliasSuggestions returns entity pairs that are likely the same thing.
// Candidates need BOTH high session co-occurrence (Jaccard >= threshold) AND
// name similarity: co-occurrence alone conflates "related" with "identical"
// (a command and the file it lives in overlap 100%), and single-session pairs
// are pure noise (1/1 = perfect overlap from one observation) — hence the
// >= 2 sessions floor. Pairs recorded in alias_dismissals are never returned.
// limit <= 0 returns all matches; results are ranked by combined score.
func GetAliasSuggestions(ctx context.Context, db *sql.DB, threshold float64, limit int) ([]AliasSuggestion, error) {
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
		AND s1.session_count >= 2 AND s2.session_count >= 2
		ORDER BY similarity DESC
	`

	// Load dismissals before opening the main result set: nested queries on
	// one *sql.DB grab a second pooled connection (breaks :memory: databases).
	dismissed, err := loadDismissedPairs(ctx, db)
	if err != nil {
		return nil, err
	}

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
		if dismissed[dismissalKey(s.EntityA, s.EntityB)] {
			continue
		}
		s.NameSimilarity = nameSimilarity(s.EntityA, s.EntityB)
		if s.NameSimilarity < minNameSimilarity {
			continue
		}
		s.Score = s.Similarity * s.NameSimilarity
		suggestions = append(suggestions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(suggestions, func(i, j int) bool { return suggestions[i].Score > suggestions[j].Score })
	if limit > 0 && len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}
	return suggestions, nil
}

// minNameSimilarity gates suggestions: pairs must look alike, not just
// co-occur. 0.55 keeps levenshtein-close spellings and containment variants.
const minNameSimilarity = 0.55

// nameSimilarity scores how alike two entity names are, on normalized forms.
// ponytail: heuristic — exact=1.0, substring containment (>=4 chars)=0.7,
// else levenshtein ratio. Upgrade to token-set similarity if this misranks.
func nameSimilarity(a, b string) float64 {
	na, nb := normalizeEntityName(a), normalizeEntityName(b)
	if na == nb {
		return 1.0
	}
	shorter, longer := na, nb
	if len(shorter) > len(longer) {
		shorter, longer = longer, shorter
	}
	if len(shorter) >= 4 && strings.Contains(longer, shorter) {
		return 0.7
	}
	if len(longer) == 0 {
		return 0
	}
	return 1.0 - float64(utils.ComputeDistance(na, nb))/float64(len(longer))
}

// dismissalKey returns the canonical storage key for a pair: normalized
// names in lexicographic order, joined — insensitive to argument order.
func dismissalKey(a, b string) string {
	na, nb := normalizeEntityName(a), normalizeEntityName(b)
	if na > nb {
		na, nb = nb, na
	}
	return na + "\x00" + nb
}

func loadDismissedPairs(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, "SELECT name_a, name_b FROM alias_dismissals")
	if err != nil {
		// Table may predate migration on read-only paths; treat as none dismissed
		return map[string]bool{}, nil //nolint:nilerr // absence of table = no dismissals
	}
	defer rows.Close()
	dismissed := make(map[string]bool)
	for rows.Next() {
		var a, b string
		if err := rows.Scan(&a, &b); err != nil {
			return nil, err
		}
		dismissed[dismissalKey(a, b)] = true
	}
	return dismissed, rows.Err()
}

// DismissAliasPair records that a suggested pair was reviewed and rejected,
// so it is never suggested again.
func DismissAliasPair(ctx context.Context, db *sql.DB, a, b, actor string) error {
	na, nb := normalizeEntityName(a), normalizeEntityName(b)
	if na > nb {
		na, nb = nb, na
	}
	_, err := db.ExecContext(ctx,
		"INSERT OR REPLACE INTO alias_dismissals (name_a, name_b, dismissed_by) VALUES (?, ?, ?)",
		na, nb, actor)
	return err
}
