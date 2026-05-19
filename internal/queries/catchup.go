package queries

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/untoldecay/BeadsLog/internal/types"
)

type CatchupDelta struct {
	Since            time.Time
	Sessions         []SearchResult
	ClosedIssues     []types.Issue
	StateChanges     []types.BranchState
}

func GetCatchupDelta(ctx context.Context, db *sql.DB, since time.Time) (*CatchupDelta, error) {
	delta := &CatchupDelta{
		Since: since,
	}

	// 1. Fetch new sessions
	sessionRows, err := db.QueryContext(ctx, `
		SELECT s.id, s.title, s.timestamp, s.narrative, COALESCE(s.branch, 'N/A')
		FROM sessions s
		WHERE s.timestamp > ?
		ORDER BY s.timestamp ASC
	`, since.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sessions for catchup: %w", err)
	}
	defer sessionRows.Close()

	for sessionRows.Next() {
		var r SearchResult
		var timestamp string
		if err := sessionRows.Scan(&r.ID, &r.Title, &timestamp, &r.Narrative, &r.Branch); err == nil {
			r.Date = timestamp
			delta.Sessions = append(delta.Sessions, r)
		}
	}

	// 2. Fetch closed issues
	issueRows, err := db.QueryContext(ctx, `
		SELECT id, title, status, issue_type, priority, COALESCE(closed_at, updated_at), COALESCE(close_reason, '')
		FROM issues
		WHERE status = 'closed' AND (closed_at > ? OR (closed_at IS NULL AND updated_at > ?))
		ORDER BY closed_at ASC
	`, since.Format(time.RFC3339), since.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issues for catchup: %w", err)
	}
	defer issueRows.Close()

	for issueRows.Next() {
		var i types.Issue
		var closedAt string
		if err := issueRows.Scan(&i.ID, &i.Title, &i.Status, &i.IssueType, &i.Priority, &closedAt, &i.CloseReason); err == nil {
			delta.ClosedIssues = append(delta.ClosedIssues, i)
		}
	}

	// 3. Fetch state changes (abandoned, paused)
	stateRows, err := db.QueryContext(ctx, `
		SELECT state, scope_type, scope_ref, short_reason, full_reason_ref, actor, timestamp
		FROM branch_states
		WHERE timestamp > ?
		ORDER BY timestamp ASC
	`, since.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch state changes for catchup: %w", err)
	}
	defer stateRows.Close()

	for stateRows.Next() {
		var s types.BranchState
		if err := stateRows.Scan(&s.State, &s.ScopeType, &s.ScopeRef, &s.ShortReason, &s.FullReasonRef, &s.Actor, &s.Timestamp); err == nil {
			delta.StateChanges = append(delta.StateChanges, s)
		}
	}

	return delta, nil
}
