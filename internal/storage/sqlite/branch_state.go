package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/untoldecay/BeadsLog/internal/types"
)

// SetBranchState stores or updates a lifecycle state for a specific scope
func (s *SQLiteStorage) SetBranchState(ctx context.Context, state types.BranchState) error {
	query := `
		INSERT INTO branch_states (
			state, scope_type, scope_ref, short_reason, full_reason_ref, actor, commit_sha, branch_ref, timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(scope_type, scope_ref) DO UPDATE SET
			state = excluded.state,
			short_reason = excluded.short_reason,
			full_reason_ref = excluded.full_reason_ref,
			actor = excluded.actor,
			commit_sha = excluded.commit_sha,
			branch_ref = excluded.branch_ref,
			timestamp = excluded.timestamp
	`
	
	if state.Timestamp.IsZero() {
		state.Timestamp = time.Now()
	}

	_, err := s.db.ExecContext(ctx, query,
		state.State, state.ScopeType, state.ScopeRef, state.ShortReason,
		state.FullReasonRef, state.Actor, state.CommitSHA, state.BranchRef, state.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("failed to set branch state: %w", err)
	}
	return nil
}

// GetBranchState retrieves the lifecycle state for a specific scope
func (s *SQLiteStorage) GetBranchState(ctx context.Context, scopeType types.ScopeType, scopeRef string) (*types.BranchState, error) {
	query := `
		SELECT id, state, scope_type, scope_ref, short_reason, full_reason_ref, actor, timestamp, commit_sha, branch_ref
		FROM branch_states
		WHERE scope_type = ? AND scope_ref = ?
	`
	
	var state types.BranchState
	err := s.db.QueryRowContext(ctx, query, scopeType, scopeRef).Scan(
		&state.ID, &state.State, &state.ScopeType, &state.ScopeRef, &state.ShortReason,
		&state.FullReasonRef, &state.Actor, &state.Timestamp, &state.CommitSHA, &state.BranchRef,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get branch state: %w", err)
	}
	return &state, nil
}

// UpdateBranchCache updates the cached Git facts for a branch
func (s *SQLiteStorage) UpdateBranchCache(ctx context.Context, cache types.BranchCache) error {
	query := `
		INSERT INTO branch_cache (
			branch_name, last_validated_sha, is_merged, is_deleted, last_checked_at
		) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(branch_name) DO UPDATE SET
			last_validated_sha = excluded.last_validated_sha,
			is_merged = excluded.is_merged,
			is_deleted = excluded.is_deleted,
			last_checked_at = excluded.last_checked_at
	`
	
	if cache.LastCheckedAt.IsZero() {
		cache.LastCheckedAt = time.Now()
	}

	_, err := s.db.ExecContext(ctx, query,
		cache.BranchName, cache.LastValidatedSHA, cache.IsMerged, cache.IsDeleted, cache.LastCheckedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update branch cache: %w", err)
	}
	return nil
}

// GetBranchCache retrieves cached Git facts for a branch
func (s *SQLiteStorage) GetBranchCache(ctx context.Context, branchName string) (*types.BranchCache, error) {
	query := `
		SELECT branch_name, last_validated_sha, is_merged, is_deleted, last_checked_at
		FROM branch_cache
		WHERE branch_name = ?
	`
	
	var cache types.BranchCache
	err := s.db.QueryRowContext(ctx, query, branchName).Scan(
		&cache.BranchName, &cache.LastValidatedSHA, &cache.IsMerged, &cache.IsDeleted, &cache.LastCheckedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get branch cache: %w", err)
	}
	return &cache, nil
}
