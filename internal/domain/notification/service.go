package notification

import "context"

// Service defines the notification service interface
type Service interface {
	// Send Notifications
	Send(ctx context.Context, notification *Notification) error
	SendToChannel(ctx context.Context, userID int64, channel Channel, notification *Notification) error
	SendImmediate(ctx context.Context, notification *Notification) error

	// Preferences
	GetPreferences(ctx context.Context, userID int64) ([]*Preference, error)
	UpdatePreference(ctx context.Context, userID int64, channel Channel, isEnabled bool, config interface{}) error

	// Logs
	GetHistory(ctx context.Context, userID int64, filter LogFilter, limit, offset int) ([]*Log, int64, error)

	// Webhooks
	CreateWebhook(ctx context.Context, userID int64, name, url, secret string, events []EventType) (*Webhook, error)
	GetWebhook(ctx context.Context, id string) (*Webhook, error)
	UpdateWebhook(ctx context.Context, id string, updates map[string]interface{}) (*Webhook, error)
	DeleteWebhook(ctx context.Context, id string) error
	ListWebhooks(ctx context.Context, userID int64, limit, offset int) ([]*Webhook, int64, error)
	TestWebhook(ctx context.Context, id string) error

	// Webhook Events
	TriggerEvent(ctx context.Context, userID int64, eventType EventType, data map[string]interface{}) error

	// Background Processing
	ProcessPendingNotifications(ctx context.Context) error
	ProcessPendingWebhookDeliveries(ctx context.Context) error
}

// Sender defines the interface for sending notifications via a specific channel
type Sender interface {
	Send(ctx context.Context, notification *Notification, config interface{}) error
	Channel() Channel
}
