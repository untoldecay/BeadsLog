package queries

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/untoldecay/BeadsLog/internal/types"
)

type CatchupDelta struct {
	Since        time.Time
	Sessions     []SearchResult
	ClosedIssues []types.Issue
	StateChanges []types.BranchState
}

func GetCatchupDelta(ctx context.Context, db *sql.DB, since time.Time) (*CatchupDelta, error) {
	delta := &CatchupDelta{
		Since: since,
	}

	// 1. Fetch new sessions
	sessionRows, err := db.QueryContext(ctx, `
		SELECT s.id, s.title, s.timestamp, s.narrative, COALESCE(s.branch, 'N/A'), COALESCE(s.author, 'Unknown')
		FROM sessions s
		WHERE s.is_ghost = 0 AND s.timestamp > ?
		ORDER BY s.timestamp ASC
	`, since.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sessions for catchup: %w", err)
	}
	defer sessionRows.Close()

	for sessionRows.Next() {
		var r SearchResult
		var timestamp string
		if err := sessionRows.Scan(&r.ID, &r.Title, &timestamp, &r.Narrative, &r.Branch, &r.Author); err == nil {
			r.Date = timestamp
			delta.Sessions = append(delta.Sessions, r)
		}
	}

	// 2. Fetch closed issues
	issueRows, err := db.QueryContext(ctx, `
		SELECT id, title, status, issue_type, priority, COALESCE(closed_at, updated_at), COALESCE(close_reason, ''), COALESCE(actor, owner)
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
		var actor string
		if err := issueRows.Scan(&i.ID, &i.Title, &i.Status, &i.IssueType, &i.Priority, &closedAt, &i.CloseReason, &actor); err == nil {
			i.Owner = actor // Using Owner field to store the actor for catchup display
			delta.ClosedIssues = append(delta.ClosedIssues, i)
		}
	}

	// 3. Fetch state changes (abandoned, paused). RFC3339Nano: stored
	// timestamps carry sub-second precision and compare lexicographically.
	stateRows, err := db.QueryContext(ctx, `
		SELECT state, scope_type, scope_ref, short_reason, full_reason_ref, actor, timestamp
		FROM branch_states
		WHERE timestamp > ?
		ORDER BY timestamp ASC
	`, since.Format(time.RFC3339Nano))
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

// FeatureGroup is one "feature arc" of the catchup digest: sessions sharing a
// work branch, or (for trunk work) a dominant entity.
type FeatureGroup struct {
	Key          string              `json:"key"`
	Kind         string              `json:"kind"` // branch | entity | general
	Narrative    string              `json:"narrative"`
	Sessions     []SearchResult      `json:"sessions"`
	Authors      []string            `json:"authors"`
	Entities     []string            `json:"entities"`
	StateChanges []types.BranchState `json:"state_changes"`
}

var trunkBranches = map[string]bool{"main": true, "master": true, "develop": true, "N/A": true, "": true}

// GroupCatchupDelta organizes the delta's sessions into feature arcs.
// Branch is the primary arc key; trunk sessions are grouped by their most
// shared entity; leftovers form a "General" group. State changes matching an
// arc's key or entities attach to it; the rest are returned separately.
func GroupCatchupDelta(ctx context.Context, db *sql.DB, delta *CatchupDelta) ([]FeatureGroup, []types.BranchState) {
	sessionEntities := fetchSessionEntities(ctx, db, delta.Sessions)

	var groups []FeatureGroup
	branchIdx := make(map[string]int)
	var trunkPool []SearchResult

	for _, s := range delta.Sessions {
		branch := strings.TrimSuffix(s.Branch, " (local)")
		if trunkBranches[branch] {
			trunkPool = append(trunkPool, s)
			continue
		}
		i, ok := branchIdx[branch]
		if !ok {
			groups = append(groups, FeatureGroup{Key: branch, Kind: "branch"})
			i = len(groups) - 1
			branchIdx[branch] = i
		}
		groups[i].Sessions = append(groups[i].Sessions, s)
	}

	// Greedy entity grouping for trunk sessions: the most shared entity forms
	// an arc, repeat while an entity still covers at least 2 sessions.
	for len(trunkPool) > 0 {
		counts := make(map[string]int)
		for _, s := range trunkPool {
			for _, e := range sessionEntities[s.ID] {
				counts[e]++
			}
		}
		best, bestCount := "", 0
		for e, c := range counts {
			if c > bestCount || (c == bestCount && e < best) {
				best, bestCount = e, c
			}
		}
		if bestCount < 2 {
			break
		}
		g := FeatureGroup{Key: best, Kind: "entity"}
		var rest []SearchResult
		for _, s := range trunkPool {
			member := false
			for _, e := range sessionEntities[s.ID] {
				if e == best {
					member = true
					break
				}
			}
			if member {
				g.Sessions = append(g.Sessions, s)
			} else {
				rest = append(rest, s)
			}
		}
		groups = append(groups, g)
		trunkPool = rest
	}
	if len(trunkPool) > 0 {
		groups = append(groups, FeatureGroup{Key: "General", Kind: "general", Sessions: trunkPool})
	}

	// Enrich each group: authors, top entities, narrative
	for i := range groups {
		seenAuthor := make(map[string]bool)
		entityCount := make(map[string]int)
		for _, s := range groups[i].Sessions {
			if s.Author != "" && !seenAuthor[s.Author] {
				seenAuthor[s.Author] = true
				groups[i].Authors = append(groups[i].Authors, s.Author)
			}
			for _, e := range sessionEntities[s.ID] {
				entityCount[e]++
			}
		}
		groups[i].Entities = topKeys(entityCount, 5)
		groups[i].Narrative = arcNarrative(groups[i].Sessions)
	}

	// Attach state changes to their arc
	var unattached []types.BranchState
	for _, sc := range delta.StateChanges {
		attached := false
		for i := range groups {
			if sc.ScopeRef == groups[i].Key || containsString(groups[i].Entities, sc.ScopeRef) {
				groups[i].StateChanges = append(groups[i].StateChanges, sc)
				attached = true
				break
			}
		}
		if !attached {
			unattached = append(unattached, sc)
		}
	}

	return groups, unattached
}

// arcNarrative builds a 2-3 line human summary from real session content:
// why the arc started (first session's problem statement) and where it stands
// (latest session's Final Status, falling back to its problem).
func arcNarrative(sessions []SearchResult) string {
	if len(sessions) == 0 {
		return ""
	}
	first, last := sessions[0], sessions[len(sessions)-1]

	started := firstProblemLine(first.Narrative)
	var parts []string
	if started != "" {
		parts = append(parts, "Started with: "+trimAtBoundary(started, 200))
	}
	if status := finalStatusLine(last.Narrative); status != "" {
		parts = append(parts, "Currently: "+trimAtBoundary(status, 200))
	} else if len(sessions) > 1 {
		if latest := firstProblemLine(last.Narrative); latest != "" && latest != started {
			parts = append(parts, "Latest: "+trimAtBoundary(latest, 200))
		}
	}
	return strings.Join(parts, " ")
}

// firstProblemLine returns the first meaningful prose line of a narrative
// (SyncSession stores narrative as "<problem>\n\n<content>").
func firstProblemLine(narrative string) string {
	for _, line := range strings.Split(narrative, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "none" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "---") {
			continue
		}
		return strings.ReplaceAll(line, "**", "")
	}
	return ""
}

// finalStatusLine extracts the "**Final Status:**" line a devlog closes with.
func finalStatusLine(narrative string) string {
	for _, line := range strings.Split(narrative, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "**Final Status:**") {
			return strings.TrimSpace(strings.TrimPrefix(strings.ReplaceAll(trimmed, "**", ""), "Final Status:"))
		}
	}
	return ""
}

func trimAtBoundary(s string, max int) string {
	if len(s) <= max {
		return s
	}
	cut := s[:max]
	if i := strings.LastIndex(cut, " "); i > max/2 {
		cut = cut[:i]
	}
	return cut + "…"
}

func topKeys(counts map[string]int, limit int) []string {
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if counts[keys[i]] != counts[keys[j]] {
			return counts[keys[i]] > counts[keys[j]]
		}
		return keys[i] < keys[j]
	})
	if len(keys) > limit {
		keys = keys[:limit]
	}
	return keys
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// fetchSessionEntities bulk-loads entity display names for the given sessions.
func fetchSessionEntities(ctx context.Context, db *sql.DB, sessions []SearchResult) map[string][]string {
	result := make(map[string][]string)
	if len(sessions) == 0 {
		return result
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(sessions)), ",")
	args := make([]interface{}, len(sessions))
	for i, s := range sessions {
		args[i] = s.ID
	}
	rows, err := db.QueryContext(ctx, `
		SELECT se.session_id, COALESCE(e.preferred_name, e.name)
		FROM session_entities se
		JOIN entities e ON se.entity_id = e.id
		WHERE se.session_id IN (`+placeholders+`)`, args...)
	if err != nil {
		return result
	}
	defer rows.Close()
	for rows.Next() {
		var id, name string
		if rows.Scan(&id, &name) == nil {
			result[id] = append(result[id], name)
		}
	}
	return result
}
