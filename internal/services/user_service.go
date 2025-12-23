package services

import (
	"context"

	"github.com/pratik-mahalle/infraudit/internal/domain/user"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"golang.org/x/crypto/bcrypt"
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

// GetByID retrieves a user by ID
func (s *UserService) GetByID(ctx context.Context, id int64) (*user.User, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByEmail retrieves a user by email
func (s *UserService) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

// Create creates a new user with password
func (s *UserService) Create(ctx context.Context, email, password string) (*user.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), s.getBCryptCost())
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to hash password")
		return nil, errors.Internal("Failed to hash password", err)
	}

	u := &user.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         user.RoleUser,
		PlanType:     user.PlanTypeFree,
	}

	err = s.repo.Create(ctx, u)
	if err != nil {
		s.logger.ErrorWithErr(err, "Failed to create user")
		return nil, err
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id": u.ID,
		"email":   u.Email,
	}).Info("User created")

	return u, nil
}

// Authenticate verifies user credentials and returns the user
func (s *UserService) Authenticate(ctx context.Context, email, password string) (*user.User, error) {
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"email": email,
		}).Warn("User not found during authentication")
		return nil, errors.Unauthorized("Invalid credentials")
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"user_id": u.ID,
			"email":   email,
		}).Warn("Invalid password during authentication")
		return nil, errors.Unauthorized("Invalid credentials")
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id": u.ID,
		"email":   email,
	}).Info("User authenticated successfully")

	return u, nil
}

// getBCryptCost returns the bcrypt cost factor (default: 12)
func (s *UserService) getBCryptCost() int {
	return bcrypt.DefaultCost
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
