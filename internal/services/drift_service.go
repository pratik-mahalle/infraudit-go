package services

import (
	"context"

	"infraaudit/backend/internal/domain/drift"
	"infraaudit/backend/internal/pkg/logger"
)

// DriftService implements drift.Service
type DriftService struct {
	repo   drift.Repository
	logger *logger.Logger
}

// NewDriftService creates a new drift service
func NewDriftService(repo drift.Repository, log *logger.Logger) drift.Service {
	return &DriftService{
		repo:   repo,
		logger: log,
	}
}

// Create creates a new drift record
func (s *DriftService) Create(ctx context.Context, d *drift.Drift) (int64, error) {
	if d.Status == "" {
		d.Status = drift.StatusDetected
	}

	id, err := s.repo.Create(ctx, d)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to create drift")
		return 0, err
	}

	s.logger.WithFields(map[string]interface{}{
		"drift_id":    id,
		"user_id":     d.UserID,
		"resource_id": d.ResourceID,
		"severity":    d.Severity,
		"drift_type":  d.DriftType,
	}).Info("Drift created")

	return id, nil
}

// GetByID retrieves a drift by ID
func (s *DriftService) GetByID(ctx context.Context, userID int64, id int64) (*drift.Drift, error) {
	return s.repo.GetByID(ctx, userID, id)
}

// Update updates a drift record
func (s *DriftService) Update(ctx context.Context, userID int64, id int64, updates map[string]interface{}) error {
	d, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return err
	}

	// Apply updates
	if resourceID, ok := updates["resource_id"].(string); ok {
		d.ResourceID = resourceID
	}
	if driftType, ok := updates["drift_type"].(string); ok {
		d.DriftType = driftType
	}
	if severity, ok := updates["severity"].(string); ok {
		d.Severity = severity
	}
	if details, ok := updates["details"].(string); ok {
		d.Details = details
	}
	if status, ok := updates["status"].(string); ok {
		d.Status = status
	}

	err = s.repo.Update(ctx, d)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to update drift")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"drift_id": id,
		"user_id":  userID,
	}).Info("Drift updated")

	return nil
}

// Delete deletes a drift record
func (s *DriftService) Delete(ctx context.Context, userID int64, id int64) error {
	err := s.repo.Delete(ctx, userID, id)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to delete drift")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"drift_id": id,
		"user_id":  userID,
	}).Info("Drift deleted")

	return nil
}

// List retrieves drifts with filters and pagination
func (s *DriftService) List(ctx context.Context, userID int64, filter drift.Filter, limit, offset int) ([]*drift.Drift, int64, error) {
	return s.repo.ListWithPagination(ctx, userID, filter, limit, offset)
}

// UpdateStatus updates drift status
func (s *DriftService) UpdateStatus(ctx context.Context, userID int64, id int64, status string) error {
	d, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return err
	}

	d.Status = status
	err = s.repo.Update(ctx, d)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to update drift status")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"drift_id": id,
		"user_id":  userID,
		"status":   status,
	}).Info("Drift status updated")

	return nil
}

// DetectDrifts detects configuration drifts for a user
func (s *DriftService) DetectDrifts(ctx context.Context, userID int64) error {
	// This would analyze resources and detect drifts
	// For now, it's a placeholder
	s.logger.WithFields(map[string]interface{}{
		"user_id": userID,
	}).Info("Detecting drifts")

	return nil
}

// GetSummary gets drift summary by severity
func (s *DriftService) GetSummary(ctx context.Context, userID int64) (map[string]int, error) {
	return s.repo.CountBySeverity(ctx, userID)
}
