package services

import (
	"context"

	"github.com/pratik-mahalle/infraudit/internal/domain/resource"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// ResourceService implements resource.Service
type ResourceService struct {
	repo   resource.Repository
	logger *logger.Logger
}

// NewResourceService creates a new resource service
func NewResourceService(repo resource.Repository, log *logger.Logger) resource.Service {
	return &ResourceService{
		repo:   repo,
		logger: log,
	}
}

// GetByID retrieves a resource by ID
func (s *ResourceService) GetByID(ctx context.Context, userID int64, resourceID string) (*resource.Resource, error) {
	return s.repo.GetByID(ctx, userID, resourceID)
}

// List retrieves resources with filters and pagination
func (s *ResourceService) List(ctx context.Context, userID int64, filter resource.Filter, limit, offset int) ([]*resource.Resource, int64, error) {
	return s.repo.List(ctx, userID, filter, limit, offset)
}

// Create creates a new resource
func (s *ResourceService) Create(ctx context.Context, res *resource.Resource) error {
	err := s.repo.Create(ctx, res)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to create resource")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id":     res.UserID,
		"resource_id": res.ResourceID,
		"provider":    res.Provider,
		"type":        res.Type,
	}).Info("Resource created")

	return nil
}

// Update updates a resource
func (s *ResourceService) Update(ctx context.Context, userID int64, resourceID string, updates map[string]interface{}) error {
	res, err := s.repo.GetByID(ctx, userID, resourceID)
	if err != nil {
		return err
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		res.Name = name
	}
	if typ, ok := updates["type"].(string); ok {
		res.Type = typ
	}
	if region, ok := updates["region"].(string); ok {
		res.Region = region
	}
	if status, ok := updates["status"].(string); ok {
		res.Status = status
	}

	err = s.repo.Update(ctx, res)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to update resource")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id":     userID,
		"resource_id": resourceID,
	}).Info("Resource updated")

	return nil
}

// Delete deletes a resource
func (s *ResourceService) Delete(ctx context.Context, userID int64, resourceID string) error {
	err := s.repo.Delete(ctx, userID, resourceID)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to delete resource")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id":     userID,
		"resource_id": resourceID,
	}).Info("Resource deleted")

	return nil
}

// SyncProviderResources syncs resources from a cloud provider
func (s *ResourceService) SyncProviderResources(ctx context.Context, userID int64, provider string) error {
	// This would integrate with the cloud provider SDK
	// For now, it's a placeholder
	s.logger.WithFields(map[string]interface{}{
		"user_id":  userID,
		"provider": provider,
	}).Info("Syncing resources from provider")

	return errors.Internal("Provider sync not yet implemented", nil)
}
