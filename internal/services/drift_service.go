package services

import (
	"context"

	"infraaudit/backend/internal/detector"
	"infraaudit/backend/internal/domain/baseline"
	"infraaudit/backend/internal/domain/drift"
	"infraaudit/backend/internal/domain/resource"
	"infraaudit/backend/internal/pkg/logger"
)

// DriftService implements drift.Service
type DriftService struct {
	repo         drift.Repository
	baselineRepo baseline.Repository
	resourceRepo resource.Repository
	detector     *detector.DriftDetector
	logger       *logger.Logger
}

// NewDriftService creates a new drift service
func NewDriftService(
	repo drift.Repository,
	baselineRepo baseline.Repository,
	resourceRepo resource.Repository,
	log *logger.Logger,
) drift.Service {
	return &DriftService{
		repo:         repo,
		baselineRepo: baselineRepo,
		resourceRepo: resourceRepo,
		detector:     detector.NewDriftDetector(),
		logger:       log,
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
	s.logger.WithFields(map[string]interface{}{
		"user_id": userID,
	}).Info("Starting drift detection")

	// Get all resources for the user
	resources, _, err := s.resourceRepo.List(ctx, userID, resource.Filter{}, 1000, 0)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to fetch resources for drift detection")
		return err
	}

	driftsDetected := 0
	driftsCreated := 0

	// Check each resource against its baseline
	for _, res := range resources {
		// Skip resources without configuration data
		if res.Configuration == "" {
			s.logger.WithFields(map[string]interface{}{
				"resource_id": res.ResourceID,
			}).Debug("Skipping resource without configuration")
			continue
		}

		// Get baseline for this resource
		resBaseline, err := s.baselineRepo.GetByResourceID(ctx, userID, res.ResourceID, baseline.TypeApproved)
		if err != nil {
			// If no baseline exists, create an automatic one
			if err.Error() == "Baseline not found" {
				s.logger.WithFields(map[string]interface{}{
					"resource_id": res.ResourceID,
				}).Info("Creating automatic baseline for resource")

				newBaseline := &baseline.Baseline{
					UserID:        userID,
					ResourceID:    res.ResourceID,
					Provider:      res.Provider,
					ResourceType:  res.Type,
					Configuration: res.Configuration,
					BaselineType:  baseline.TypeAutomatic,
					Description:   "Auto-created baseline on first scan",
				}
				_, err := s.baselineRepo.Create(ctx, newBaseline)
				if err != nil {
					s.logger.ErrorWithErr(err, "Failed to create automatic baseline")
				}
				continue
			}

			s.logger.ErrorWithErr(err, "Failed to get baseline")
			continue
		}

		// Detect drift by comparing configurations
		result, err := s.detector.DetectDrift(res.Type, resBaseline.Configuration, res.Configuration)
		if err != nil {
			s.logger.ErrorWithErr(err, "Failed to detect drift")
			continue
		}

		if result.HasDrift {
			driftsDetected++

			// Create drift record
			d := &drift.Drift{
				UserID:     userID,
				ResourceID: res.ResourceID,
				DriftType:  result.DriftType,
				Severity:   result.Severity,
				Details:    result.Details,
				Status:     drift.StatusDetected,
			}

			_, err := s.repo.Create(ctx, d)
			if err != nil {
				s.logger.ErrorWithErr(err, "Failed to create drift record")
				continue
			}

			driftsCreated++

			s.logger.WithFields(map[string]interface{}{
				"resource_id": res.ResourceID,
				"drift_type":  result.DriftType,
				"severity":    result.Severity,
			}).Info("Drift detected and recorded")
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id":         userID,
		"resources":       len(resources),
		"drifts_detected": driftsDetected,
		"drifts_created":  driftsCreated,
	}).Info("Drift detection completed")

	return nil
}

// GetSummary gets drift summary by severity
func (s *DriftService) GetSummary(ctx context.Context, userID int64) (map[string]int, error) {
	return s.repo.CountBySeverity(ctx, userID)
}
