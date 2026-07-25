package sqlite

import (
	"context"
	"fmt"

	"github.com/untoldecay/BeadsLog/internal/types"
)

// GetAllAliases returns all sticky aliases from the registry.
// It joins with entities to get the canonical name for sharing.
func (s *SQLiteStorage) GetAllAliases(ctx context.Context) ([]types.AliasRecord, error) {
	query := `
		SELECT ea.alias_name, e.name as canonical_name
		FROM entity_aliases ea
		JOIN entities e ON ea.canonical_id = e.id
		ORDER BY e.name, ea.alias_name
	`
	
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get aliases: %w", err)
	}
	defer rows.Close()

	var aliases []types.AliasRecord
	for rows.Next() {
		var a types.AliasRecord
		if err := rows.Scan(&a.AliasName, &a.CanonicalName); err != nil {
			return nil, err
		}
		aliases = append(aliases, a)
	}
	return aliases, nil
}

// SaveAliases batch inserts aliases into the registry.
// Used during sync/import to restore shared aliases.
func (s *SQLiteStorage) SaveAliases(ctx context.Context, aliases []types.AliasRecord) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, a := range aliases {
		// We need the ID of the canonical name. 
		// If it doesn't exist, we create a placeholder entity.
		var canonicalID string
		err := tx.QueryRowContext(ctx, "SELECT id FROM entities WHERE name = ?", a.CanonicalName).Scan(&canonicalID)
		if err != nil {
			// Create placeholder (using hash name as ID for stability if needed, 
			// but here we just need A canonical ID that exists)
			canonicalID = fmt.Sprintf("ent-%s", a.CanonicalName) // Simple placeholder ID
			_, err = tx.ExecContext(ctx, `
				INSERT OR IGNORE INTO entities (id, name, type, source)
				VALUES (?, ?, 'component', 'registry')
			`, canonicalID, a.CanonicalName)
			if err != nil {
				return fmt.Errorf("failed to create placeholder entity for alias: %w", err)
			}
		}

		_, err = tx.ExecContext(ctx, `
			INSERT OR REPLACE INTO entity_aliases (alias_name, canonical_id)
			VALUES (?, ?)
		`, a.AliasName, canonicalID)
		if err != nil {
			return fmt.Errorf("failed to save alias %s: %w", a.AliasName, err)
		}
	}

	return tx.Commit()
}
