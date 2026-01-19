package remediation

import "context"

// Repository defines the remediation repository interface
type Repository interface {
	Create(ctx context.Context, action *Action) error
	GetByID(ctx context.Context, id string) (*Action, error)
	Update(ctx context.Context, action *Action) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter Filter, limit, offset int) ([]*Action, int64, error)
	GetByDriftID(ctx context.Context, driftID string) ([]*Action, error)
	GetByVulnerabilityID(ctx context.Context, vulnerabilityID string) ([]*Action, error)
	GetPendingApprovals(ctx context.Context, userID int64) ([]*Action, error)
	CountByStatus(ctx context.Context, userID int64) (map[ActionStatus]int, error)
}
