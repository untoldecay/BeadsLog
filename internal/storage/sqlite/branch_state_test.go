package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/untoldecay/BeadsLog/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBranchState(t *testing.T) {
	ctx := context.Background()
	store, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("Set and Get BranchState", func(t *testing.T) {
		state := types.BranchState{
			State:         types.StatePaused,
			ScopeType:     types.ScopeBranch,
			ScopeRef:      "feature/test",
			ShortReason:   "Waiting for feedback",
			FullReasonRef: "sess-123",
			Actor:         "test-user",
			CommitSHA:     "abc123",
			BranchRef:     "refs/heads/feature/test",
			Timestamp:     time.Now().Truncate(time.Second), // SQLite DATETIME resolution
		}

		err := store.SetBranchState(ctx, state)
		require.NoError(t, err)

		got, err := store.GetBranchState(ctx, types.ScopeBranch, "feature/test")
		require.NoError(t, err)
		require.NotNil(t, got)

		assert.Equal(t, state.State, got.State)
		assert.Equal(t, state.ScopeType, got.ScopeType)
		assert.Equal(t, state.ScopeRef, got.ScopeRef)
		assert.Equal(t, state.ShortReason, got.ShortReason)
		assert.Equal(t, state.FullReasonRef, got.FullReasonRef)
		assert.Equal(t, state.Actor, got.Actor)
		assert.Equal(t, state.CommitSHA, got.CommitSHA)
		assert.Equal(t, state.BranchRef, got.BranchRef)
		assert.True(t, state.Timestamp.Equal(got.Timestamp))
	})

	t.Run("Update BranchState (Upsert)", func(t *testing.T) {
		state := types.BranchState{
			State:     types.StatePaused,
			ScopeType: types.ScopeEntity,
			ScopeRef:  "entity:123",
		}
		err := store.SetBranchState(ctx, state)
		require.NoError(t, err)

		state.State = types.StateAbandoned
		state.ShortReason = "Dead end"
		err = store.SetBranchState(ctx, state)
		require.NoError(t, err)

		got, err := store.GetBranchState(ctx, types.ScopeEntity, "entity:123")
		require.NoError(t, err)
		assert.Equal(t, types.StateAbandoned, got.State)
		assert.Equal(t, "Dead end", got.ShortReason)
	})

	t.Run("BranchCache CRUD", func(t *testing.T) {
		cache := types.BranchCache{
			BranchName:       "feature/test",
			LastValidatedSHA: "sha123",
			IsMerged:         true,
			IsDeleted:        false,
			LastCheckedAt:    time.Now().Truncate(time.Second),
		}

		err := store.UpdateBranchCache(ctx, cache)
		require.NoError(t, err)

		got, err := store.GetBranchCache(ctx, "feature/test")
		require.NoError(t, err)
		require.NotNil(t, got)

		assert.Equal(t, cache.BranchName, got.BranchName)
		assert.Equal(t, cache.LastValidatedSHA, got.LastValidatedSHA)
		assert.Equal(t, cache.IsMerged, got.IsMerged)
		assert.Equal(t, cache.IsDeleted, got.IsDeleted)
		assert.True(t, cache.LastCheckedAt.Equal(got.LastCheckedAt))
	})
}
