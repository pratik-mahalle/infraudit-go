package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	m "github.com/pratik-mahalle/infraudit/settings/models"
	u "github.com/pratik-mahalle/infraudit/settings/utils"
)

// Profile
func GetProfile(c *gin.Context) {
	profile := m.Profile{ID: 1, Username: "demo", Email: "demo@example.com"}
	u.JSON(c, http.StatusOK, profile)
}

func UpdateProfile(c *gin.Context) {
	var req m.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		u.Error(c, http.StatusBadRequest, "invalid body")
		return
	}
	u.JSON(c, http.StatusOK, gin.H{"status": "updated"})
}

func UploadAvatar(c *gin.Context) {
	u.JSON(c, http.StatusCreated, m.UploadAvatarResponse{URL: "https://example.com/avatar.png"})
}

// Account
func GetAccountSettings(c *gin.Context) {
	u.JSON(c, http.StatusOK, m.AccountSettings{Language: "en", Timezone: "UTC", Theme: "system"})
}

func UpdateAccountSettings(c *gin.Context) {
	var req m.UpdateAccountSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		u.Error(c, http.StatusBadRequest, "invalid body")
		return
	}
	u.JSON(c, http.StatusOK, gin.H{"status": "updated"})
}

func ChangePassword(c *gin.Context) {
	var req m.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.NewPassword == "" {
		u.Error(c, http.StatusBadRequest, "invalid body")
		return
	}
	u.JSON(c, http.StatusOK, gin.H{"status": "changed"})
}

// Notifications
func GetNotificationSettings(c *gin.Context) {
	u.JSON(c, http.StatusOK, m.NotificationSettings{EmailEnabled: true, SlackEnabled: false, Threshold: 80})
}

func UpdateNotificationSettings(c *gin.Context) {
	var req m.UpdateNotificationSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		u.Error(c, http.StatusBadRequest, "invalid body")
		return
	}
	u.JSON(c, http.StatusOK, gin.H{"status": "updated"})
}

func CreateNotificationIntegration(c *gin.Context) {
	var req m.IntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Type == "" {
		u.Error(c, http.StatusBadRequest, "invalid integration")
		return
	}
	u.JSON(c, http.StatusCreated, gin.H{"status": "connected", "type": req.Type})
}

func UpdateNotificationThreshold(c *gin.Context) {
	var req m.UpdateThresholdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		u.Error(c, http.StatusBadRequest, "invalid body")
		return
	}
	u.JSON(c, http.StatusOK, gin.H{"status": "updated", "threshold": req.Threshold})
}

// Security
func GetSecurity(c *gin.Context) {
	now := time.Now()
	u.JSON(c, http.StatusOK, m.SecurityState{
		TwoFAEnabled:   false,
		APIKeys:        []m.APIKey{{ID: "key_123", Name: "default", CreatedAt: now.AddDate(0, -1, 0)}},
		ActiveSessions: []m.Session{{ID: "sess_1", UserAgent: "Mozilla/5.0", CreatedAt: now.Add(-2 * time.Hour)}},
	})
}

func Setup2FA(c *gin.Context) {
	u.JSON(c, http.StatusCreated, m.TwoFASetupResponse{QRCodeURL: "https://example.com/qr.png", Secret: "ABC123"})
}

func CreateAPIKey(c *gin.Context) {
	var req m.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		u.Error(c, http.StatusBadRequest, "invalid name")
		return
	}
	u.JSON(c, http.StatusCreated, m.APIKey{ID: "key_", Name: req.Name, CreatedAt: time.Now()})
}

func DeleteAPIKey(c *gin.Context) {
	u.JSON(c, http.StatusNoContent, nil)
}

func LogoutSession(c *gin.Context) {
	u.JSON(c, http.StatusOK, gin.H{"status": "logged_out"})
}

// Cloud & Integrations
func ListCloudAccounts(c *gin.Context) {
	u.JSON(c, http.StatusOK, []m.CloudAccount{{ID: "aws-1", Provider: "aws", Name: "AWS Prod", Connected: true}})
}

func CreateCloudAccount(c *gin.Context) {
	var req m.CreateCloudAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Provider == "" || req.Name == "" {
		u.Error(c, http.StatusBadRequest, "invalid account")
		return
	}
	u.JSON(c, http.StatusCreated, m.CloudAccount{ID: "acc_123", Provider: req.Provider, Name: req.Name, Connected: true})
}

func DeleteCloudAccount(c *gin.Context) {
	u.JSON(c, http.StatusNoContent, nil)
}

func ListK8sClusters(c *gin.Context) {
	u.JSON(c, http.StatusOK, []m.K8sCluster{{Name: "dev-cluster"}})
}

func IntegrateGithub(c *gin.Context) {
	var req m.GithubIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.AccessToken == "" || req.Repo == "" {
		u.Error(c, http.StatusBadRequest, "invalid github config")
		return
	}
	u.JSON(c, http.StatusCreated, gin.H{"status": "connected", "repo": req.Repo})
}

// Team & Access
func ListTeamMembers(c *gin.Context) {
	u.JSON(c, http.StatusOK, []m.TeamMember{{ID: "u1", Email: "demo@example.com", Role: "owner", Status: "active"}})
}

func InviteTeamMember(c *gin.Context) {
	var req m.InviteRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" || req.Role == "" {
		u.Error(c, http.StatusBadRequest, "invalid invite")
		return
	}
	u.JSON(c, http.StatusCreated, gin.H{"status": "invited", "email": req.Email})
}

func UpdateTeamMember(c *gin.Context) {
	var req m.UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		u.Error(c, http.StatusBadRequest, "invalid body")
		return
	}
	u.JSON(c, http.StatusOK, gin.H{"status": "updated"})
}

func DeleteTeamMember(c *gin.Context) {
	u.JSON(c, http.StatusNoContent, nil)
}
