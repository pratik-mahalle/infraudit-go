package recommendation

import "context"

// Service defines the interface for recommendation business logic
type Service interface {
	// Create creates a new recommendation
	Create(ctx context.Context, rec *Recommendation) (int64, error)

	// GetByID retrieves a recommendation by ID
	GetByID(ctx context.Context, userID int64, id int64) (*Recommendation, error)

	// Update updates a recommendation
	Update(ctx context.Context, userID int64, id int64, updates map[string]interface{}) error

	// Delete deletes a recommendation
	Delete(ctx context.Context, userID int64, id int64) error

	// List retrieves recommendations with filters and pagination
	List(ctx context.Context, userID int64, filter Filter, limit, offset int) ([]*Recommendation, int64, error)

	// GetTotalSavings calculates total potential savings
	GetTotalSavings(ctx context.Context, userID int64) (float64, error)

	// GenerateRecommendations generates recommendations for a user
	GenerateRecommendations(ctx context.Context, userID int64) error
}
