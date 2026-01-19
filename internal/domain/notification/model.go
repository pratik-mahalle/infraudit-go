package notification

import (
	"encoding/json"
	"time"
)

// Preference represents a user's notification preferences for a channel
type Preference struct {
	ID        string          `json:"id"`
	UserID    int64           `json:"user_id"`
	Channel   Channel         `json:"channel"`
	IsEnabled bool            `json:"is_enabled"`
	Config    json.RawMessage `json:"config,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// Log represents a notification log entry
type Log struct {
	ID               string           `json:"id"`
	UserID           int64            `json:"user_id"`
	Channel          Channel          `json:"channel"`
	NotificationType NotificationType `json:"notification_type"`
	Status           DeliveryStatus   `json:"status"`
	Priority         Priority         `json:"priority"`
	Payload          json.RawMessage  `json:"payload,omitempty"`
	ErrorMessage     string           `json:"error_message,omitempty"`
	RetryCount       int              `json:"retry_count"`
	SentAt           *time.Time       `json:"sent_at,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
}

// Webhook represents a configured webhook
type Webhook struct {
	ID            string          `json:"id"`
	UserID        int64           `json:"user_id"`
	Name          string          `json:"name"`
	URL           string          `json:"url"`
	Secret        string          `json:"secret,omitempty"`
	Events        []EventType     `json:"events"`
	IsEnabled     bool            `json:"is_enabled"`
	RetryConfig   json.RawMessage `json:"retry_config,omitempty"`
	LastTriggered *time.Time      `json:"last_triggered,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID             string         `json:"id"`
	WebhookID      string         `json:"webhook_id"`
	EventType      EventType      `json:"event_type"`
	Payload        string         `json:"payload"`
	Status         DeliveryStatus `json:"status"`
	ResponseStatus int            `json:"response_status,omitempty"`
	ResponseBody   string         `json:"response_body,omitempty"`
	RetryCount     int            `json:"retry_count"`
	DeliveredAt    *time.Time     `json:"delivered_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}

// Channel represents a notification channel
type Channel string

const (
	ChannelSlack   Channel = "slack"
	ChannelEmail   Channel = "email"
	ChannelWebhook Channel = "webhook"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeDriftAlert          NotificationType = "drift_alert"
	NotificationTypeVulnerabilityAlert  NotificationType = "vulnerability_alert"
	NotificationTypeCostAlert           NotificationType = "cost_alert"
	NotificationTypeComplianceAlert     NotificationType = "compliance_alert"
	NotificationTypeRemediationApproval NotificationType = "remediation_approval"
	NotificationTypeRemediationComplete NotificationType = "remediation_complete"
	NotificationTypeDailySummary        NotificationType = "daily_summary"
	NotificationTypeWeeklySummary       NotificationType = "weekly_summary"
	NotificationTypeJobComplete         NotificationType = "job_complete"
	NotificationTypeJobFailed           NotificationType = "job_failed"
)

// Priority represents notification priority
type Priority string

const (
	PriorityCritical Priority = "critical"
	PriorityHigh     Priority = "high"
	PriorityMedium   Priority = "medium"
	PriorityLow      Priority = "low"
)

// DeliveryStatus represents the status of a notification delivery
type DeliveryStatus string

const (
	DeliveryStatusPending  DeliveryStatus = "pending"
	DeliveryStatusSent     DeliveryStatus = "sent"
	DeliveryStatusFailed   DeliveryStatus = "failed"
	DeliveryStatusRetrying DeliveryStatus = "retrying"
)

// EventType represents webhook event types
type EventType string

const (
	EventDriftDetected        EventType = "drift.detected"
	EventVulnerabilityFound   EventType = "vulnerability.found"
	EventComplianceFailed     EventType = "compliance.failed"
	EventCostAnomalyDetected  EventType = "cost.anomaly_detected"
	EventRemediationCompleted EventType = "remediation.completed"
	EventScanCompleted        EventType = "scan.completed"
	EventJobCompleted         EventType = "job.completed"
	EventJobFailed            EventType = "job.failed"
)

// Notification represents a notification to be sent
type Notification struct {
	Type     NotificationType       `json:"type"`
	Priority Priority               `json:"priority"`
	Title    string                 `json:"title"`
	Message  string                 `json:"message"`
	Data     map[string]interface{} `json:"data,omitempty"`
	UserID   int64                  `json:"user_id"`
}

// SlackConfig represents Slack channel configuration
type SlackConfig struct {
	WebhookURL   string   `json:"webhook_url"`
	Channel      string   `json:"channel"`
	Username     string   `json:"username,omitempty"`
	IconEmoji    string   `json:"icon_emoji,omitempty"`
	MentionUsers []string `json:"mention_users,omitempty"`
}

// EmailConfig represents email notification configuration
type EmailConfig struct {
	To         []string `json:"to"`
	DigestMode bool     `json:"digest_mode"`
	DigestTime string   `json:"digest_time,omitempty"` // e.g., "08:00"
}

// WebhookRetryConfig represents retry configuration for webhooks
type WebhookRetryConfig struct {
	MaxRetries     int `json:"max_retries"`
	InitialDelayMs int `json:"initial_delay_ms"`
	MaxDelayMs     int `json:"max_delay_ms"`
	BackoffFactor  int `json:"backoff_factor"`
}

// PreferenceFilter contains preference filtering options
type PreferenceFilter struct {
	UserID  int64
	Channel Channel
}

// LogFilter contains log filtering options
type LogFilter struct {
	UserID           int64
	Channel          Channel
	NotificationType NotificationType
	Status           DeliveryStatus
	From             *time.Time
	To               *time.Time
}

// WebhookFilter contains webhook filtering options
type WebhookFilter struct {
	UserID    int64
	IsEnabled *bool
	EventType EventType
}

// IsValid checks if the channel is valid
func (c Channel) IsValid() bool {
	switch c {
	case ChannelSlack, ChannelEmail, ChannelWebhook:
		return true
	default:
		return false
	}
}

// GetChannelForPriority returns the appropriate channels for a priority
func GetChannelsForPriority(priority Priority) []Channel {
	switch priority {
	case PriorityCritical:
		return []Channel{ChannelSlack, ChannelEmail, ChannelWebhook}
	case PriorityHigh:
		return []Channel{ChannelSlack, ChannelEmail, ChannelWebhook}
	case PriorityMedium:
		return []Channel{ChannelEmail, ChannelWebhook}
	case PriorityLow:
		return []Channel{ChannelEmail}
	default:
		return []Channel{ChannelEmail}
	}
}

// AllEventTypes returns all available event types
func AllEventTypes() []EventType {
	return []EventType{
		EventDriftDetected,
		EventVulnerabilityFound,
		EventComplianceFailed,
		EventCostAnomalyDetected,
		EventRemediationCompleted,
		EventScanCompleted,
		EventJobCompleted,
		EventJobFailed,
	}
}

// DefaultWebhookRetryConfig returns default retry configuration
func DefaultWebhookRetryConfig() WebhookRetryConfig {
	return WebhookRetryConfig{
		MaxRetries:     3,
		InitialDelayMs: 1000,
		MaxDelayMs:     30000,
		BackoffFactor:  2,
	}
}
