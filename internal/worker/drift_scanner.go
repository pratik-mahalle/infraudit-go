package worker

import (
	"context"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/domain/drift"
	"github.com/pratik-mahalle/infraudit/internal/domain/provider"
	"github.com/pratik-mahalle/infraudit/internal/domain/user"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// DriftScanner handles periodic drift detection scans
type DriftScanner struct {
	driftService    drift.Service
	providerService provider.Service
	userService     user.Service
	userRepo        user.Repository
	interval        time.Duration
	logger          *logger.Logger
}

// NewDriftScanner creates a new drift scanner worker
func NewDriftScanner(
	driftService drift.Service,
	providerService provider.Service,
	userService user.Service,
	userRepo user.Repository,
	interval time.Duration,
	log *logger.Logger,
) *DriftScanner {
	return &DriftScanner{
		driftService:    driftService,
		providerService: providerService,
		userService:     userService,
		userRepo:        userRepo,
		interval:        interval,
		logger:          log,
	}
}

// Start begins the periodic drift scanning process
func (s *DriftScanner) Start(ctx context.Context) {
	s.logger.Info("Starting drift scanner worker")

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run initial scan
	s.scanAllUsers(ctx)

	for {
		select {
		case <-ticker.C:
			s.scanAllUsers(ctx)
		case <-ctx.Done():
			s.logger.Info("Drift scanner worker stopped")
			return
		}
	}
}

// scanAllUsers performs drift detection for all users with connected providers
func (s *DriftScanner) scanAllUsers(ctx context.Context) {
	s.logger.Info("Starting drift detection scan for all users")

	// Get all users
	users, _, err := s.userRepo.List(ctx, 1000, 0) // Get up to 1000 users
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to get users for drift scanning")
		return
	}

	for _, user := range users {
		if err := s.scanUser(ctx, user.ID); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"user_id": user.ID,
				"email":   user.Email,
			}).ErrorWithErr(err, "Failed to scan user for drifts")
		}
	}

	s.logger.Info("Completed drift detection scan for all users")
}

// scanUser performs drift detection for a specific user
func (s *DriftScanner) scanUser(ctx context.Context, userID int64) error {
	s.logger.WithFields(map[string]interface{}{
		"user_id": userID,
	}).Info("Starting drift detection scan for user")

	// Get user's connected providers
	providers, err := s.providerService.List(ctx, userID)
	if err != nil {
		return err
	}

	if len(providers) == 0 {
		s.logger.WithFields(map[string]interface{}{
			"user_id": userID,
		}).Info("No connected providers found for user")
		return nil
	}

	// Sync resources from all providers
	for _, provider := range providers {
		if !provider.IsConnected {
			s.logger.WithFields(map[string]interface{}{
				"user_id":  userID,
				"provider": provider.Provider,
				"status":   "disconnected",
			}).Warn("Provider not connected, skipping")
			continue
		}

		if err := s.providerService.Sync(ctx, userID, provider.Provider); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"user_id":  userID,
				"provider": provider.Provider,
			}).ErrorWithErr(err, "Failed to sync provider")
			continue
		}
	}

	// Run drift detection
	if err := s.driftService.DetectDrifts(ctx, userID); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"user_id": userID,
		}).ErrorWithErr(err, "Failed to detect drifts")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id": userID,
	}).Info("Completed drift detection scan for user")

	return nil
}

// SetInterval updates the scanning interval
func (s *DriftScanner) SetInterval(interval time.Duration) {
	s.interval = interval
	s.logger.WithFields(map[string]interface{}{
		"interval": interval.String(),
	}).Info("Updated drift scanner interval")
}
