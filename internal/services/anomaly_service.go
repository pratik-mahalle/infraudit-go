package services

import (
	"context"

	"github.com/pratik-mahalle/infraudit/internal/domain/anomaly"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// AnomalyService implements anomaly.Service
type AnomalyService struct {
	repo   anomaly.Repository
	logger *logger.Logger
}

// NewAnomalyService creates a new anomaly service
func NewAnomalyService(repo anomaly.Repository, log *logger.Logger) anomaly.Service {
	return &AnomalyService{
		repo:   repo,
		logger: log,
	}
}

// Create creates a new anomaly record
func (s *AnomalyService) Create(ctx context.Context, a *anomaly.Anomaly) (int64, error) {
	if a.Status == "" {
		a.Status = anomaly.StatusDetected
	}

	id, err := s.repo.Create(ctx, a)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to create anomaly")
		return 0, err
	}

	s.logger.WithFields(map[string]interface{}{
		"anomaly_id":   id,
		"user_id":      a.UserID,
		"resource_id":  a.ResourceID,
		"severity":     a.Severity,
		"anomaly_type": a.AnomalyType,
		"percentage":   a.Percentage,
	}).Info("Anomaly created")

	return id, nil
}

// GetByID retrieves an anomaly by ID
func (s *AnomalyService) GetByID(ctx context.Context, userID int64, id int64) (*anomaly.Anomaly, error) {
	return s.repo.GetByID(ctx, userID, id)
}

// Update updates an anomaly record
func (s *AnomalyService) Update(ctx context.Context, userID int64, id int64, updates map[string]interface{}) error {
	a, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return err
	}

	// Apply updates
	if resourceID, ok := updates["resource_id"].(string); ok {
		a.ResourceID = resourceID
	}
	if anomalyType, ok := updates["anomaly_type"].(string); ok {
		a.AnomalyType = anomalyType
	}
	if severity, ok := updates["severity"].(string); ok {
		a.Severity = severity
	}
	if percentage, ok := updates["percentage"].(int); ok {
		a.Percentage = percentage
	}
	if previousCost, ok := updates["previous_cost"].(int); ok {
		a.PreviousCost = previousCost
	}
	if currentCost, ok := updates["current_cost"].(int); ok {
		a.CurrentCost = currentCost
	}
	if status, ok := updates["status"].(string); ok {
		a.Status = status
	}

	err = s.repo.Update(ctx, a)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to update anomaly")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"anomaly_id": id,
		"user_id":    userID,
	}).Info("Anomaly updated")

	return nil
}

// Delete deletes an anomaly record
func (s *AnomalyService) Delete(ctx context.Context, userID int64, id int64) error {
	err := s.repo.Delete(ctx, userID, id)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to delete anomaly")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"anomaly_id": id,
		"user_id":    userID,
	}).Info("Anomaly deleted")

	return nil
}

// List retrieves anomalies with filters and pagination
func (s *AnomalyService) List(ctx context.Context, userID int64, filter anomaly.Filter, limit, offset int) ([]*anomaly.Anomaly, int64, error) {
	return s.repo.ListWithPagination(ctx, userID, filter, limit, offset)
}

// UpdateStatus updates anomaly status
func (s *AnomalyService) UpdateStatus(ctx context.Context, userID int64, id int64, status string) error {
	a, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return err
	}

	a.Status = status
	err = s.repo.Update(ctx, a)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to update anomaly status")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"anomaly_id": id,
		"user_id":    userID,
		"status":     status,
	}).Info("Anomaly status updated")

	return nil
}

// DetectAnomalies detects cost anomalies for a user
func (s *AnomalyService) DetectAnomalies(ctx context.Context, userID int64) error {
	// This would analyze cost data and detect anomalies
	// For now, it's a placeholder
	s.logger.WithFields(map[string]interface{}{
		"user_id": userID,
	}).Info("Detecting anomalies")

	return nil
}

// GetSummary gets anomaly summary by severity
func (s *AnomalyService) GetSummary(ctx context.Context, userID int64) (map[string]int, error) {
	return s.repo.CountBySeverity(ctx, userID)
}
