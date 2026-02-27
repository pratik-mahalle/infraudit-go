package dto

// UserDTO represents a user in API responses
type UserDTO struct {
	ID       int64   `json:"id"`
	Email    string  `json:"email"`
	Username string  `json:"username,omitempty"`
	FullName *string `json:"fullName,omitempty"`
	Role     string  `json:"role"`
	PlanType string  `json:"planType"`
}

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	FullName *string `json:"fullName,omitempty"`
}
