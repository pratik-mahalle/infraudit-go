package client

import (
	"context"
	"time"
)

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username,omitempty"`
	FullName string `json:"full_name,omitempty"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	User         *User     `json:"user,omitempty"`
}

// User represents a user in the system
type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username,omitempty"`
	FullName  string    `json:"full_name,omitempty"`
	Role      string    `json:"role"`
	PlanType  string    `json:"plan_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Login authenticates with email and password
func (c *Client) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	req := LoginRequest{
		Email:    email,
		Password: password,
	}

	var resp LoginResponse
	if err := c.doRequest(ctx, "POST", "/api/auth/login", req, &resp); err != nil {
		return nil, err
	}

	// Automatically set the token for future requests
	if resp.Token != "" {
		c.SetToken(resp.Token)
	}

	return &resp, nil
}

// Register creates a new user account
func (c *Client) Register(ctx context.Context, req RegisterRequest) (*LoginResponse, error) {
	var resp LoginResponse
	if err := c.doRequest(ctx, "POST", "/api/auth/register", req, &resp); err != nil {
		return nil, err
	}

	// Automatically set the token for future requests
	if resp.Token != "" {
		c.SetToken(resp.Token)
	}

	return &resp, nil
}

// GetCurrentUser retrieves the currently authenticated user
func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	var user User
	if err := c.doRequest(ctx, "GET", "/api/auth/me", nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// Logout logs out the current user
func (c *Client) Logout(ctx context.Context) error {
	if err := c.doRequest(ctx, "POST", "/api/auth/logout", nil, nil); err != nil {
		return err
	}
	// Clear the token
	c.SetToken("")
	return nil
}

// RefreshToken refreshes the authentication token
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	req := map[string]string{
		"refresh_token": refreshToken,
	}

	var resp LoginResponse
	if err := c.doRequest(ctx, "POST", "/api/auth/refresh", req, &resp); err != nil {
		return nil, err
	}

	// Update the token
	if resp.Token != "" {
		c.SetToken(resp.Token)
	}

	return &resp, nil
}
