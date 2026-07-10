package types

import "time"

// LifecycleState represents the state of a devlog session or branch
type LifecycleState string

const (
	StateActive    LifecycleState = "active"
	StateOngoing   LifecycleState = "ongoing" // alive but not currently in front of anyone (context switch)
	StatePaused    LifecycleState = "paused"  // explicitly parked with a reason
	StateAbandoned LifecycleState = "abandoned"
)

// ScopeType represents the scope of a lifecycle state change
type ScopeType string

const (
	ScopeBranch  ScopeType = "branch"
	ScopeEntity  ScopeType = "entity"
	ScopeFile    ScopeType = "file"
	ScopeTask    ScopeType = "task"
	ScopeSession ScopeType = "session"
)

// BranchState represents a manually set state for a specific scope (e.g. branch, entity)
type BranchState struct {
	ID            int64          `json:"id"`
	State         LifecycleState `json:"state"`
	ScopeType     ScopeType      `json:"scope_type"`
	ScopeRef      string         `json:"scope_ref"`
	ShortReason   string         `json:"short_reason"`
	FullReasonRef string         `json:"full_reason_ref"` // Link to the special devlog entry ID
	Actor         string         `json:"actor"`
	Timestamp     time.Time      `json:"timestamp"`
	CommitSHA     string         `json:"commit_sha"`
	BranchRef     string         `json:"branch_ref"`
}

// BranchCache represents cached Git facts about a branch
type BranchCache struct {
	BranchName       string    `json:"branch_name"`
	LastValidatedSHA string    `json:"last_validated_sha"`
	IsMerged         bool      `json:"is_merged"`
	IsDeleted        bool      `json:"is_deleted"`
	LastCheckedAt    time.Time `json:"last_checked_at"`
}

// AliasRecord represents a sticky entity alias for repository-wide sharing
type AliasRecord struct {
	AliasName     string `json:"alias_name"`
	CanonicalName string `json:"canonical_name"`
}
