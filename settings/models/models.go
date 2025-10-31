package models

import "time"

// ---- Profile ----

type Profile struct {
	ID        int64   `json:"id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	FullName  *string `json:"fullName,omitempty"`
	AvatarURL *string `json:"avatarUrl,omitempty"`
}

type UpdateProfileRequest struct {
	Username *string `json:"username"`
	FullName *string `json:"fullName"`
}

type UploadAvatarResponse struct {
	URL string `json:"url"`
}

// ---- Account Settings ----

type AccountSettings struct {
	Language string `json:"language"`
	Timezone string `json:"timezone"`
	Theme    string `json:"theme"`
}

type UpdateAccountSettingsRequest struct {
	Language *string `json:"language"`
	Timezone *string `json:"timezone"`
	Theme    *string `json:"theme"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// ---- Notifications ----

type NotificationSettings struct {
	EmailEnabled bool    `json:"emailEnabled"`
	SlackEnabled bool    `json:"slackEnabled"`
	WebhookURL   *string `json:"webhookUrl,omitempty"`
	Threshold    float64 `json:"threshold"`
}

type UpdateNotificationSettingsRequest struct {
	EmailEnabled *bool    `json:"emailEnabled"`
	SlackEnabled *bool    `json:"slackEnabled"`
	WebhookURL   *string  `json:"webhookUrl"`
	Threshold    *float64 `json:"threshold"`
}

type IntegrationRequest struct {
	Type   string            `json:"type"`
	Config map[string]string `json:"config"`
}

type UpdateThresholdRequest struct {
	Threshold float64 `json:"threshold"`
}

// ---- Security ----

type SecurityState struct {
	TwoFAEnabled   bool      `json:"twoFAEnabled"`
	APIKeys        []APIKey  `json:"apiKeys"`
	ActiveSessions []Session `json:"activeSessions"`
}

type TwoFASetupResponse struct {
	QRCodeURL string `json:"qrCodeUrl"`
	Secret    string `json:"secret"`
}

type CreateAPIKeyRequest struct {
	Name string `json:"name"`
}

type APIKey struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type Session struct {
	ID        string    `json:"id"`
	UserAgent string    `json:"userAgent"`
	CreatedAt time.Time `json:"createdAt"`
}

// ---- Cloud & Integrations ----

type CloudAccount struct {
	ID        string `json:"id"`
	Provider  string `json:"provider"`
	Name      string `json:"name"`
	Connected bool   `json:"connected"`
}

type CreateCloudAccountRequest struct {
	Provider    string            `json:"provider"`
	Name        string            `json:"name"`
	Credentials map[string]string `json:"credentials"`
}

type K8sCluster struct {
	Name string `json:"name"`
}

type GithubIntegrationRequest struct {
	AccessToken string `json:"accessToken"`
	Repo        string `json:"repo"`
}

// ---- Team & Access ----

type TeamMember struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

type InviteRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type UpdateMemberRequest struct {
	Role   *string `json:"role"`
	Status *string `json:"status"`
}
