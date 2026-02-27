package dto

import (
	"encoding/json"
	"time"
)

// ======= Job DTOs =======

// CreateJobRequest represents a request to create a scheduled job
// Frontend sends ScheduledJob with "type" field
type CreateJobRequest struct {
	JobType  string                 `json:"type" validate:"required"`
	Schedule string                 `json:"schedule" validate:"required"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

// UpdateJobRequest represents a request to update a scheduled job
// Frontend ScheduledJob uses "enabled" not "isEnabled"
type UpdateJobRequest struct {
	Schedule  *string                 `json:"schedule,omitempty"`
	IsEnabled *bool                   `json:"enabled,omitempty"`
	Config    *map[string]interface{} `json:"config,omitempty"`
}

// JobResponse represents a scheduled job response
type JobResponse struct {
	ID        string          `json:"id"`
	UserID    int64           `json:"userId"`
	JobType   string          `json:"type"`
	Schedule  string          `json:"schedule"`
	IsEnabled bool            `json:"enabled"`
	Config    json.RawMessage `json:"config,omitempty"`
	LastRun   *time.Time      `json:"lastRun,omitempty"`
	NextRun   *time.Time      `json:"nextRun,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

// JobExecutionResponse represents a job execution response
type JobExecutionResponse struct {
	ID           string          `json:"id"`
	JobID        string          `json:"jobId"`
	UserID       int64           `json:"userId"`
	JobType      string          `json:"jobType"`
	Status       string          `json:"status"`
	StartedAt    *time.Time      `json:"startedAt,omitempty"`
	CompletedAt  *time.Time      `json:"completedAt,omitempty"`
	DurationMs   int             `json:"durationMs,omitempty"`
	Result       json.RawMessage `json:"result,omitempty"`
	ErrorMessage string          `json:"errorMessage,omitempty"`
	CreatedAt    time.Time       `json:"createdAt"`
}

// TriggerJobRequest represents a request to trigger a job
type TriggerJobRequest struct {
	Config map[string]interface{} `json:"config,omitempty"`
}

// ListJobsResponse represents a list of jobs response
type ListJobsResponse struct {
	Jobs  []JobResponse `json:"jobs"`
	Total int64         `json:"total"`
}

// ListJobExecutionsResponse represents a list of job executions response
type ListJobExecutionsResponse struct {
	Executions []JobExecutionResponse `json:"executions"`
	Total      int64                  `json:"total"`
}

// ======= Remediation DTOs =======

// RemediationSuggestionResponse represents a remediation suggestion
type RemediationSuggestionResponse struct {
	ID              string                  `json:"id"`
	IssueType       string                  `json:"issueType"`
	IssueID         string                  `json:"issueId"`
	Title           string                  `json:"title"`
	Description     string                  `json:"description"`
	Severity        string                  `json:"severity"`
	RemediationType string                  `json:"remediationType"`
	Strategy        *RemediationStrategyDTO `json:"strategy"`
	Risk            string                  `json:"risk"`
	Impact          string                  `json:"impact"`
	EstimatedTime   string                  `json:"estimatedTime"`
}

// RemediationStrategyDTO represents a remediation strategy
type RemediationStrategyDTO struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Steps       []RemediationStepDTO   `json:"steps"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// RemediationStepDTO represents a remediation step
type RemediationStepDTO struct {
	Order       int    `json:"order"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status,omitempty"`
}

// CreateRemediationRequest represents a request to create a remediation action
type CreateRemediationRequest struct {
	SuggestionID string `json:"suggestionId" validate:"required"`
}

// ApproveRemediationRequest represents a request to approve a remediation
type ApproveRemediationRequest struct {
	Approved bool   `json:"approved"`
	Reason   string `json:"reason,omitempty"`
}

// RemediationActionResponse represents a remediation action response
type RemediationActionResponse struct {
	ID               string                  `json:"id"`
	UserID           int64                   `json:"userId"`
	DriftID          *string                 `json:"driftId,omitempty"`
	VulnerabilityID  *string                 `json:"vulnerabilityId,omitempty"`
	RemediationType  string                  `json:"remediationType"`
	Status           string                  `json:"status"`
	Strategy         *RemediationStrategyDTO `json:"strategy,omitempty"`
	ApprovalRequired bool                    `json:"approvalRequired"`
	ApprovedBy       *int64                  `json:"approvedBy,omitempty"`
	ApprovedAt       *time.Time              `json:"approvedAt,omitempty"`
	StartedAt        *time.Time              `json:"startedAt,omitempty"`
	CompletedAt      *time.Time              `json:"completedAt,omitempty"`
	Result           json.RawMessage         `json:"result,omitempty"`
	ErrorMessage     string                  `json:"errorMessage,omitempty"`
	CreatedAt        time.Time               `json:"createdAt"`
	UpdatedAt        time.Time               `json:"updatedAt"`
}

// ListRemediationActionsResponse represents a list of remediation actions
type ListRemediationActionsResponse struct {
	Actions []RemediationActionResponse `json:"actions"`
	Total   int64                       `json:"total"`
}

// RemediationSummaryResponse represents a remediation summary
type RemediationSummaryResponse struct {
	Pending    int `json:"pending"`
	Approved   int `json:"approved"`
	InProgress int `json:"inProgress"`
	Completed  int `json:"completed"`
	Failed     int `json:"failed"`
	Rejected   int `json:"rejected"`
	RolledBack int `json:"rolledBack"`
}

// ======= Notification DTOs =======

// NotificationPreferenceResponse represents a notification preference
type NotificationPreferenceResponse struct {
	ID        string          `json:"id"`
	Channel   string          `json:"channel"`
	IsEnabled bool            `json:"isEnabled"`
	Config    json.RawMessage `json:"config,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

// UpdatePreferenceRequest represents a request to update a notification preference
// Supports both frontend formats: { isEnabled, config } or { enabled, settings }
type UpdatePreferenceRequest struct {
	IsEnabled bool                   `json:"isEnabled"`
	Enabled   bool                   `json:"enabled"`
	Config    map[string]interface{} `json:"config,omitempty"`
	Settings  map[string]interface{} `json:"settings,omitempty"`
}

// NotificationLogResponse represents a notification log entry
type NotificationLogResponse struct {
	ID               string          `json:"id"`
	Channel          string          `json:"channel"`
	NotificationType string          `json:"notificationType"`
	Status           string          `json:"status"`
	Priority         string          `json:"priority"`
	Payload          json.RawMessage `json:"payload,omitempty"`
	ErrorMessage     string          `json:"errorMessage,omitempty"`
	SentAt           *time.Time      `json:"sentAt,omitempty"`
	CreatedAt        time.Time       `json:"createdAt"`
}

// ListNotificationLogsResponse represents a list of notification logs
type ListNotificationLogsResponse struct {
	Logs  []NotificationLogResponse `json:"logs"`
	Total int64                     `json:"total"`
}

// ======= Webhook DTOs =======

// CreateWebhookRequest represents a request to create a webhook
type CreateWebhookRequest struct {
	Name   string   `json:"name" validate:"required"`
	URL    string   `json:"url" validate:"required,url"`
	Secret string   `json:"secret,omitempty"`
	Events []string `json:"events" validate:"required,min=1"`
}

// UpdateWebhookRequest represents a request to update a webhook
type UpdateWebhookRequest struct {
	Name      *string   `json:"name,omitempty"`
	URL       *string   `json:"url,omitempty"`
	Secret    *string   `json:"secret,omitempty"`
	Events    *[]string `json:"events,omitempty"`
	IsEnabled *bool     `json:"isEnabled,omitempty"`
}

// WebhookResponse represents a webhook response
type WebhookResponse struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	URL           string     `json:"url"`
	Events        []string   `json:"events"`
	IsEnabled     bool       `json:"isEnabled"`
	LastTriggered *time.Time `json:"lastTriggered,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// ListWebhooksResponse represents a list of webhooks
type ListWebhooksResponse struct {
	Webhooks []WebhookResponse `json:"webhooks"`
	Total    int64             `json:"total"`
}

// WebhookDeliveryResponse represents a webhook delivery
type WebhookDeliveryResponse struct {
	ID             string     `json:"id"`
	EventType      string     `json:"eventType"`
	Status         string     `json:"status"`
	ResponseStatus int        `json:"responseStatus,omitempty"`
	RetryCount     int        `json:"retryCount"`
	DeliveredAt    *time.Time `json:"deliveredAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
}

// ListWebhookDeliveriesResponse represents a list of webhook deliveries
type ListWebhookDeliveriesResponse struct {
	Deliveries []WebhookDeliveryResponse `json:"deliveries"`
	Total      int64                     `json:"total"`
}

// AvailableWebhookEventsResponse represents available webhook events
type AvailableWebhookEventsResponse struct {
	Events []WebhookEventInfo `json:"events"`
}

// WebhookEventInfo represents information about a webhook event
type WebhookEventInfo struct {
	Event       string `json:"event"`
	Description string `json:"description"`
}

// SendNotificationRequest represents a request to send a notification
// Supports both structured (type/priority/title/message) and simple (channel/message) formats
type SendNotificationRequest struct {
	Type     string                 `json:"type,omitempty"`
	Channel  string                 `json:"channel,omitempty"`
	Priority string                 `json:"priority,omitempty"`
	Title    string                 `json:"title,omitempty"`
	Message  string                 `json:"message" validate:"required"`
	Data     map[string]interface{} `json:"data,omitempty"`
}
