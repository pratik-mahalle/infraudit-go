package services

import (
	"context"
	"fmt"

	"github.com/pratik-mahalle/infraudit/internal/domain/recommendation"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// RecommendationService implements recommendation.Service
type RecommendationService struct {
	repo   recommendation.Repository
	engine *RecommendationEngine
	logger *logger.Logger
}

// NewRecommendationService creates a new recommendation service
func NewRecommendationService(repo recommendation.Repository, engine *RecommendationEngine, log *logger.Logger) recommendation.Service {
	return &RecommendationService{
		repo:   repo,
		engine: engine,
		logger: log,
	}
}

// Create creates a new recommendation
func (s *RecommendationService) Create(ctx context.Context, rec *recommendation.Recommendation) (int64, error) {
	id, err := s.repo.Create(ctx, rec)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to create recommendation")
		return 0, err
	}

	s.logger.WithFields(map[string]interface{}{
		"recommendation_id": id,
		"user_id":           rec.UserID,
		"type":              rec.Type,
		"priority":          rec.Priority,
		"savings":           rec.Savings,
	}).Info("Recommendation created")

	return id, nil
}

// GetByID retrieves a recommendation by ID
func (s *RecommendationService) GetByID(ctx context.Context, userID int64, id int64) (*recommendation.Recommendation, error) {
	return s.repo.GetByID(ctx, userID, id)
}

// Update updates a recommendation
func (s *RecommendationService) Update(ctx context.Context, userID int64, id int64, updates map[string]interface{}) error {
	rec, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return err
	}

	// Apply updates
	if typ, ok := updates["type"].(string); ok {
		rec.Type = typ
	}
	if priority, ok := updates["priority"].(string); ok {
		rec.Priority = priority
	}
	if title, ok := updates["title"].(string); ok {
		rec.Title = title
	}
	if description, ok := updates["description"].(string); ok {
		rec.Description = description
	}
	if savings, ok := updates["savings"].(float64); ok {
		rec.Savings = savings
	}
	if effort, ok := updates["effort"].(string); ok {
		rec.Effort = effort
	}
	if impact, ok := updates["impact"].(string); ok {
		rec.Impact = impact
	}
	if category, ok := updates["category"].(string); ok {
		rec.Category = category
	}
	if resources, ok := updates["resources"].([]string); ok {
		rec.Resources = resources
	}

	err = s.repo.Update(ctx, rec)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to update recommendation")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"recommendation_id": id,
		"user_id":           userID,
	}).Info("Recommendation updated")

	return nil
}

// Delete deletes a recommendation
func (s *RecommendationService) Delete(ctx context.Context, userID int64, id int64) error {
	err := s.repo.Delete(ctx, userID, id)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to delete recommendation")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"recommendation_id": id,
		"user_id":           userID,
	}).Info("Recommendation deleted")

	return nil
}

// List retrieves recommendations with filters and pagination
func (s *RecommendationService) List(ctx context.Context, userID int64, filter recommendation.Filter, limit, offset int) ([]*recommendation.Recommendation, int64, error) {
	return s.repo.ListWithPagination(ctx, userID, filter, limit, offset)
}

// GetTotalSavings calculates total potential savings
func (s *RecommendationService) GetTotalSavings(ctx context.Context, userID int64) (float64, error) {
	return s.repo.GetTotalSavings(ctx, userID)
}

// GenerateRecommendations generates recommendations for a user using AI
func (s *RecommendationService) GenerateRecommendations(ctx context.Context, userID int64) error {
	s.logger.WithFields(map[string]interface{}{
		"user_id": userID,
	}).Info("Generating recommendations using AI engine")

	if s.engine == nil {
		s.logger.Warn("Recommendation engine is not configured")
		return fmt.Errorf("recommendation engine is not available")
	}

	if err := s.engine.GenerateRecommendations(ctx, userID); err != nil {
		s.logger.ErrorWithErr(err, "Failed to generate recommendations")
		return fmt.Errorf("failed to generate recommendations: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id": userID,
	}).Info("Successfully generated recommendations")

	return nil
}
