package notification

import "context"

// Repository defines the notification repository interface
type Repository interface {
	// Preferences
	CreatePreference(ctx context.Context, preference *Preference) error
	GetPreference(ctx context.Context, userID int64, channel Channel) (*Preference, error)
	UpdatePreference(ctx context.Context, preference *Preference) error
	ListPreferences(ctx context.Context, userID int64) ([]*Preference, error)
	DeletePreference(ctx context.Context, id string) error

	// Logs
	CreateLog(ctx context.Context, log *Log) error
	UpdateLog(ctx context.Context, log *Log) error
	ListLogs(ctx context.Context, filter LogFilter, limit, offset int) ([]*Log, int64, error)
	GetPendingLogs(ctx context.Context) ([]*Log, error)

	// Webhooks
	CreateWebhook(ctx context.Context, webhook *Webhook) error
	GetWebhook(ctx context.Context, id string) (*Webhook, error)
	UpdateWebhook(ctx context.Context, webhook *Webhook) error
	DeleteWebhook(ctx context.Context, id string) error
	ListWebhooks(ctx context.Context, userID int64, limit, offset int) ([]*Webhook, int64, error)
	GetWebhooksForEvent(ctx context.Context, userID int64, eventType EventType) ([]*Webhook, error)

	// Webhook Deliveries
	CreateDelivery(ctx context.Context, delivery *WebhookDelivery) error
	UpdateDelivery(ctx context.Context, delivery *WebhookDelivery) error
	GetPendingDeliveries(ctx context.Context) ([]*WebhookDelivery, error)
	ListDeliveries(ctx context.Context, webhookID string, limit, offset int) ([]*WebhookDelivery, int64, error)
}
