package queries

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/untoldecay/BeadsLog/internal/types"
)

func TestGroupCatchupDelta(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	_, _ = db.Exec("INSERT INTO entities (id, name) VALUES ('e1', 'paymentgateway')")
	for _, id := range []string{"s3", "s4"} {
		_, _ = db.Exec("INSERT INTO sessions (id, title, timestamp) VALUES (?, ?, '2026-07-01T10:00:00Z')", id, id)
		_, _ = db.Exec("INSERT INTO session_entities (session_id, entity_id) VALUES (?, 'e1')", id)
	}

	delta := &CatchupDelta{
		Since: time.Now().Add(-24 * time.Hour),
		Sessions: []SearchResult{
			{ID: "s1", Title: "[feature] Pay retry", Branch: "feat/x (local)", Author: "alice",
				Narrative: "Retries were dropped silently.\n\n# Log\n**Final Status:** Retry queue shipped and verified."},
			{ID: "s2", Title: "[fix] Pay timeout", Branch: "feat/x", Author: "bob",
				Narrative: "Timeouts on slow gateways.\n\n# Log"},
			{ID: "s3", Title: "[fix] Gateway auth", Branch: "main", Author: "alice", Narrative: "Auth tokens expired."},
			{ID: "s4", Title: "[enhance] Gateway logs", Branch: "main", Author: "alice", Narrative: "Logs lacked request IDs."},
			{ID: "s5", Title: "[docs] Readme", Branch: "main", Author: "carol", Narrative: "Readme was stale."},
		},
		StateChanges: []types.BranchState{
			{State: types.StatePaused, ScopeType: types.ScopeBranch, ScopeRef: "feat/x", ShortReason: "waiting on review"},
			{State: types.StateAbandoned, ScopeType: types.ScopeEntity, ScopeRef: "oldcomponent", ShortReason: "superseded"},
		},
	}

	groups, unattached := GroupCatchupDelta(ctx, db, delta)

	if len(groups) != 3 {
		t.Fatalf("expected 3 groups (branch, entity, general), got %d: %+v", len(groups), groups)
	}

	branch := groups[0]
	if branch.Key != "feat/x" || branch.Kind != "branch" || len(branch.Sessions) != 2 {
		t.Errorf("branch group wrong: %+v", branch)
	}
	if len(branch.Authors) != 2 {
		t.Errorf("expected deduped authors [alice bob], got %v", branch.Authors)
	}
	if len(branch.StateChanges) != 1 || branch.StateChanges[0].ScopeRef != "feat/x" {
		t.Errorf("pause should attach to branch group: %+v", branch.StateChanges)
	}
	if !strings.Contains(branch.Narrative, "Started with: Retries were dropped silently.") {
		t.Errorf("narrative missing first problem: %q", branch.Narrative)
	}
	// Latest session (s2) has no Final Status → falls back to its problem
	if !strings.Contains(branch.Narrative, "Timeouts on slow gateways") {
		t.Errorf("narrative missing latest state: %q", branch.Narrative)
	}

	entityGroup := groups[1]
	if entityGroup.Kind != "entity" || entityGroup.Key != "paymentgateway" || len(entityGroup.Sessions) != 2 {
		t.Errorf("entity group wrong: %+v", entityGroup)
	}

	general := groups[2]
	if general.Kind != "general" || len(general.Sessions) != 1 || general.Sessions[0].ID != "s5" {
		t.Errorf("general group wrong: %+v", general)
	}

	if len(unattached) != 1 || unattached[0].ScopeRef != "oldcomponent" {
		t.Errorf("unrelated state change should stay unattached: %+v", unattached)
	}
}

func TestArcNarrativeFinalStatus(t *testing.T) {
	sessions := []SearchResult{
		{Narrative: "The cache was stale.\n\ncontent"},
		{Narrative: "Follow-up tuning.\n\n**Final Status:** Cache invalidation shipped."},
	}
	n := arcNarrative(sessions)
	if !strings.Contains(n, "Started with: The cache was stale.") {
		t.Errorf("missing start: %q", n)
	}
	if !strings.Contains(n, "Currently: Cache invalidation shipped.") {
		t.Errorf("missing final status: %q", n)
	}
}
