package remediation

import "context"

// Service defines the remediation service interface
type Service interface {
	// Suggestion Generation
	SuggestForDrift(ctx context.Context, driftID string) ([]*Suggestion, error)
	SuggestForVulnerability(ctx context.Context, vulnerabilityID string) ([]*Suggestion, error)

	// Remediation Actions
	Create(ctx context.Context, userID int64, suggestion *Suggestion) (*Action, error)
	Execute(ctx context.Context, actionID string) error
	Approve(ctx context.Context, actionID string, approverID int64) error
	Reject(ctx context.Context, actionID string, reason string) error
	Rollback(ctx context.Context, actionID string) error

	// Queries
	GetAction(ctx context.Context, id string) (*Action, error)
	ListActions(ctx context.Context, filter Filter, limit, offset int) ([]*Action, int64, error)
	GetPendingApprovals(ctx context.Context, userID int64) ([]*Action, error)

	// Statistics
	GetSummary(ctx context.Context, userID int64) (map[ActionStatus]int, error)
}
