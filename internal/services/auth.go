package services

import "context"

type AuthService interface {
	Login(ctx context.Context, email, password string) (User, error)
	Register(ctx context.Context, username, email, password string, fullName *string, role string) (User, error)
	User(ctx context.Context, userID int64) (User, error)
	StartTrial(ctx context.Context, userID int64) (User, error)
	TrialStatus(ctx context.Context, userID int64) (TrialStatus, error)
	LinkOAuth(ctx context.Context, provider OAuthProvider, oauthID, email, displayName, avatarURL string) (User, error)
}

type InMemoryAuth struct{}

func NewInMemoryAuth() *InMemoryAuth { return &InMemoryAuth{} }

func (a *InMemoryAuth) Login(ctx context.Context, email, password string) (User, error) {
	full := "Demo User"
	return User{ID: 1, Username: "demo", Email: email, FullName: &full, Role: "user", PlanType: "free", TrialStatus: "active"}, nil
}
func (a *InMemoryAuth) Register(ctx context.Context, username, email, password string, fullName *string, role string) (User, error) {
	return User{ID: 1, Username: username, Email: email, FullName: fullName, Role: role, PlanType: "free", TrialStatus: "active"}, nil
}
func (a *InMemoryAuth) User(ctx context.Context, userID int64) (User, error) {
	full := "Demo User"
	return User{ID: userID, Username: "demo", Email: "demo@example.com", FullName: &full, Role: "user", PlanType: "free", TrialStatus: "active"}, nil
}
func (a *InMemoryAuth) StartTrial(ctx context.Context, userID int64) (User, error) {
	u, _ := a.User(ctx, userID)
	u.TrialStatus = "active"
	return u, nil
}
func (a *InMemoryAuth) TrialStatus(ctx context.Context, userID int64) (TrialStatus, error) {
	return TrialStatus{Status: "active", DaysRemaining: 7}, nil
}
func (a *InMemoryAuth) LinkOAuth(ctx context.Context, provider OAuthProvider, oauthID, email, displayName, avatarURL string) (User, error) {
	full := displayName
	return User{ID: 1, Username: "oauth", Email: email, FullName: &full, Role: "user", PlanType: "free", TrialStatus: "active"}, nil
}
