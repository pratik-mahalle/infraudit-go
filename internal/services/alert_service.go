package services

import (
	"context"

	"github.com/pratik-mahalle/infraudit/internal/domain/alert"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// AlertService implements alert.Service
type AlertService struct {
	repo   alert.Repository
	logger *logger.Logger
}

// NewAlertService creates a new alert service
func NewAlertService(repo alert.Repository, log *logger.Logger) alert.Service {
	return &AlertService{
		repo:   repo,
		logger: log,
	}
}

// Create creates a new alert
func (s *AlertService) Create(ctx context.Context, a *alert.Alert) (int64, error) {
	if a.Status == "" {
		a.Status = alert.StatusOpen
	}

	id, err := s.repo.Create(ctx, a)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to create alert")
		return 0, err
	}

	s.logger.WithFields(map[string]interface{}{
		"alert_id": id,
		"user_id":  a.UserID,
		"severity": a.Severity,
		"type":     a.Type,
	}).Info("Alert created")

	return id, nil
}

// GetByID retrieves an alert by ID
func (s *AlertService) GetByID(ctx context.Context, userID int64, id int64) (*alert.Alert, error) {
	return s.repo.GetByID(ctx, userID, id)
}

// Update updates an alert
func (s *AlertService) Update(ctx context.Context, userID int64, id int64, updates map[string]interface{}) error {
	a, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return err
	}

	// Apply updates
	if typ, ok := updates["type"].(string); ok {
		a.Type = typ
	}
	if severity, ok := updates["severity"].(string); ok {
		a.Severity = severity
	}
	if title, ok := updates["title"].(string); ok {
		a.Title = title
	}
	if description, ok := updates["description"].(string); ok {
		a.Description = description
	}
	if resource, ok := updates["resource"].(string); ok {
		a.Resource = resource
	}
	if status, ok := updates["status"].(string); ok {
		a.Status = status
	}

	err = s.repo.Update(ctx, a)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to update alert")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"alert_id": id,
		"user_id":  userID,
	}).Info("Alert updated")

	return nil
}

// Delete deletes an alert
func (s *AlertService) Delete(ctx context.Context, userID int64, id int64) error {
	err := s.repo.Delete(ctx, userID, id)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to delete alert")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"alert_id": id,
		"user_id":  userID,
	}).Info("Alert deleted")

	return nil
}

// List retrieves alerts with filters and pagination
func (s *AlertService) List(ctx context.Context, userID int64, filter alert.Filter, limit, offset int) ([]*alert.Alert, int64, error) {
	return s.repo.ListWithPagination(ctx, userID, filter, limit, offset)
}

// UpdateStatus updates alert status
func (s *AlertService) UpdateStatus(ctx context.Context, userID int64, id int64, status string) error {
	a, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return err
	}

	a.Status = status
	err = s.repo.Update(ctx, a)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to update alert status")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"alert_id": id,
		"user_id":  userID,
		"status":   status,
	}).Info("Alert status updated")

	return nil
}

// GetSummary gets alert summary by status
func (s *AlertService) GetSummary(ctx context.Context, userID int64) (map[string]int, error) {
	return s.repo.CountByStatus(ctx, userID)
}
