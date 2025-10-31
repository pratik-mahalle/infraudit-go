package routes

import (
	"github.com/gin-gonic/gin"

	c "infraaudit/backend/settings/controllers"
)

func RegisterSettingsRoutes(r *gin.Engine) {
	// Profile
	r.GET("/profile", c.GetProfile)
	r.PUT("/profile", c.UpdateProfile)
	r.POST("/profile/avatar", c.UploadAvatar)

	// Account
	r.GET("/account/settings", c.GetAccountSettings)
	r.PUT("/account/settings", c.UpdateAccountSettings)
	r.POST("/account/password", c.ChangePassword)

	// Notifications
	r.GET("/notifications/settings", c.GetNotificationSettings)
	r.PUT("/notifications/settings", c.UpdateNotificationSettings)
	r.POST("/notifications/integration", c.CreateNotificationIntegration)
	r.PUT("/notifications/threshold", c.UpdateNotificationThreshold)

	// Security
	r.GET("/security", c.GetSecurity)
	r.POST("/security/2fa/setup", c.Setup2FA)
	r.POST("/security/api-key", c.CreateAPIKey)
	r.DELETE("/security/api-key/:id", c.DeleteAPIKey)
	r.POST("/security/logout-session", c.LogoutSession)

	// Cloud & Integrations
	r.GET("/cloud/accounts", c.ListCloudAccounts)
	r.POST("/cloud/account", c.CreateCloudAccount)
	r.DELETE("/cloud/account/:id", c.DeleteCloudAccount)
	r.GET("/cloud/k8s-clusters", c.ListK8sClusters)
	r.POST("/integrations/github", c.IntegrateGithub)

	// Team & Access
	r.GET("/team/members", c.ListTeamMembers)
	r.POST("/team/invite", c.InviteTeamMember)
	r.PUT("/team/member/:id", c.UpdateTeamMember)
	r.DELETE("/team/member/:id", c.DeleteTeamMember)
}
