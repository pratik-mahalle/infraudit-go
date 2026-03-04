package services

import (
	"context"

	"github.com/pratik-mahalle/infraudit/internal/domain/user"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// UserService implements user.Service
type UserService struct {
	repo   user.Repository
	logger *logger.Logger
}

// NewUserService creates a new user service
func NewUserService(repo user.Repository, log *logger.Logger) user.Service {
	return &UserService{
		repo:   repo,
		logger: log,
	}
}

// GetByID retrieves a user by internal ID
func (s *UserService) GetByID(ctx context.Context, id int64) (*user.User, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByAuthID retrieves a user by Supabase auth UUID
func (s *UserService) GetByAuthID(ctx context.Context, authID string) (*user.User, error) {
	return s.repo.GetByAuthID(ctx, authID)
}

// GetByEmail retrieves a user by email
func (s *UserService) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

// EnsureProfile creates a profile if the Supabase trigger hasn't created one yet
func (s *UserService) EnsureProfile(ctx context.Context, authID, email, fullName string) (*user.User, error) {
	// Check if profile already exists
	existing, err := s.repo.GetByAuthID(ctx, authID)
	if err == nil {
		return existing, nil
	}

	// Create profile
	var fn *string
	if fullName != "" {
		fn = &fullName
	}

	u := &user.User{
		AuthID:   authID,
		Email:    email,
		FullName: fn,
		Role:     user.RoleUser,
		PlanType: user.PlanTypeFree,
	}

	err = s.repo.Create(ctx, u)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to create user profile")
		return nil, err
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id": u.ID,
		"auth_id": authID,
		"email":   email,
	}).Info("User profile created (fallback)")

	return u, nil
}

// Update updates a user
func (s *UserService) Update(ctx context.Context, u *user.User) error {
	err := s.repo.Update(ctx, u)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to update user")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id": u.ID,
	}).Info("User updated")

	return nil
}

// GetTrialStatus gets the trial status for a user
func (s *UserService) GetTrialStatus(ctx context.Context, userID int64) (*user.TrialStatus, error) {
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Simple trial logic - can be expanded
	if u.PlanType == user.PlanTypeTrial {
		return &user.TrialStatus{
			Status:        "active",
			DaysRemaining: 14,
		}, nil
	}

	return &user.TrialStatus{
		Status:        "not_in_trial",
		DaysRemaining: 0,
	}, nil
}

// UpgradePlan upgrades a user's plan
func (s *UserService) UpgradePlan(ctx context.Context, userID int64, planType string) error {
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	u.PlanType = planType
	err = s.repo.Update(ctx, u)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to upgrade plan")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id":   userID,
		"plan_type": planType,
	}).Info("User plan upgraded")

	return nil
}
