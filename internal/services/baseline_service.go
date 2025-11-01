package services

import (
	"context"

	"infraudit/backend/internal/domain/baseline"
	"infraudit/backend/internal/pkg/logger"
)

// BaselineService implements baseline.Service
type BaselineService struct {
	repo   baseline.Repository
	logger *logger.Logger
}

// NewBaselineService creates a new baseline service
func NewBaselineService(repo baseline.Repository, log *logger.Logger) baseline.Service {
	return &BaselineService{
		repo:   repo,
		logger: log,
	}
}

// CreateBaseline creates a new baseline for a resource
func (s *BaselineService) CreateBaseline(ctx context.Context, b *baseline.Baseline) (int64, error) {
	if b.BaselineType == "" {
		b.BaselineType = baseline.TypeManual
	}

	id, err := s.repo.Create(ctx, b)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to create baseline")
		return 0, err
	}

	s.logger.WithFields(map[string]interface{}{
		"baseline_id":   id,
		"user_id":       b.UserID,
		"resource_id":   b.ResourceID,
		"baseline_type": b.BaselineType,
	}).Info("Baseline created")

	return id, nil
}

// GetBaseline retrieves baseline for a resource
func (s *BaselineService) GetBaseline(ctx context.Context, userID int64, resourceID string, baselineType string) (*baseline.Baseline, error) {
	return s.repo.GetByResourceID(ctx, userID, resourceID, baselineType)
}

// UpdateBaseline updates an existing baseline
func (s *BaselineService) UpdateBaseline(ctx context.Context, userID int64, id int64, updates map[string]interface{}) error {
	// Get existing baseline
	b, err := s.repo.GetByResourceID(ctx, userID, "", "")
	if err != nil {
		return err
	}

	// Apply updates
	if config, ok := updates["configuration"].(string); ok {
		b.Configuration = config
	}
	if desc, ok := updates["description"].(string); ok {
		b.Description = desc
	}

	err = s.repo.Update(ctx, b)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to update baseline")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"baseline_id": id,
		"user_id":     userID,
	}).Info("Baseline updated")

	return nil
}

// DeleteBaseline deletes a baseline
func (s *BaselineService) DeleteBaseline(ctx context.Context, userID int64, id int64) error {
	err := s.repo.Delete(ctx, userID, id)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to delete baseline")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"baseline_id": id,
		"user_id":     userID,
	}).Info("Baseline deleted")

	return nil
}

// ListBaselines lists all baselines for a user
func (s *BaselineService) ListBaselines(ctx context.Context, userID int64) ([]*baseline.Baseline, error) {
	return s.repo.List(ctx, userID)
}
