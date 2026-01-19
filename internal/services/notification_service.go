package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pratik-mahalle/infraudit/internal/domain/notification"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// NotificationService implements notification.Service
type NotificationService struct {
	repo            notification.Repository
	logger          *logger.Logger
	slackWebhookURL string
	httpClient      *http.Client
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	repo notification.Repository,
	log *logger.Logger,
	slackWebhookURL string,
) notification.Service {
	return &NotificationService{
		repo:            repo,
		logger:          log,
		slackWebhookURL: slackWebhookURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Send sends a notification to all enabled channels based on priority
func (s *NotificationService) Send(ctx context.Context, n *notification.Notification) error {
	// Get user preferences
	prefs, err := s.repo.ListPreferences(ctx, n.UserID)
	if err != nil {
		return fmt.Errorf("failed to get preferences: %w", err)
	}

	// Determine which channels to use based on priority
	targetChannels := notification.GetChannelsForPriority(n.Priority)

	// Create preference map for quick lookup
	prefMap := make(map[notification.Channel]*notification.Preference)
	for _, p := range prefs {
		prefMap[p.Channel] = p
	}

	var errors []error
	for _, channel := range targetChannels {
		pref, exists := prefMap[channel]
		if !exists || !pref.IsEnabled {
			continue
		}

		if err := s.SendToChannel(ctx, n.UserID, channel, n); err != nil {
			errors = append(errors, err)
			s.logger.WithFields(map[string]interface{}{
				"user_id": n.UserID,
				"channel": channel,
			}).ErrorWithErr(err, "Failed to send notification to channel")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send to %d channels", len(errors))
	}

	return nil
}

// SendToChannel sends a notification to a specific channel
func (s *NotificationService) SendToChannel(ctx context.Context, userID int64, channel notification.Channel, n *notification.Notification) error {
	// Create log entry
	payloadJSON, _ := json.Marshal(map[string]interface{}{
		"title":   n.Title,
		"message": n.Message,
		"data":    n.Data,
	})

	log := &notification.Log{
		ID:               uuid.New().String(),
		UserID:           userID,
		Channel:          channel,
		NotificationType: n.Type,
		Status:           notification.DeliveryStatusPending,
		Priority:         n.Priority,
		Payload:          payloadJSON,
	}

	if err := s.repo.CreateLog(ctx, log); err != nil {
		return fmt.Errorf("failed to create log: %w", err)
	}

	var err error
	switch channel {
	case notification.ChannelSlack:
		err = s.sendSlack(ctx, userID, n)
	case notification.ChannelEmail:
		err = s.sendEmail(ctx, userID, n)
	case notification.ChannelWebhook:
		// Webhooks are handled separately via TriggerEvent
		return nil
	default:
		err = fmt.Errorf("unsupported channel: %s", channel)
	}

	now := time.Now()
	if err != nil {
		log.Status = notification.DeliveryStatusFailed
		log.ErrorMessage = err.Error()
	} else {
		log.Status = notification.DeliveryStatusSent
		log.SentAt = &now
	}

	s.repo.UpdateLog(ctx, log)

	return err
}

// SendImmediate sends a notification immediately without checking preferences
func (s *NotificationService) SendImmediate(ctx context.Context, n *notification.Notification) error {
	// Send to all configured channels immediately
	var errors []error

	// Try Slack first for critical notifications
	if s.slackWebhookURL != "" {
		if err := s.sendSlack(ctx, n.UserID, n); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send immediate notification: %v", errors)
	}

	return nil
}

// sendSlack sends a notification via Slack
func (s *NotificationService) sendSlack(ctx context.Context, userID int64, n *notification.Notification) error {
	// Get user's Slack config
	pref, err := s.repo.GetPreference(ctx, userID, notification.ChannelSlack)
	if err != nil {
		return fmt.Errorf("failed to get Slack preference: %w", err)
	}

	webhookURL := s.slackWebhookURL
	if pref != nil && len(pref.Config) > 0 {
		var config notification.SlackConfig
		if err := json.Unmarshal(pref.Config, &config); err == nil && config.WebhookURL != "" {
			webhookURL = config.WebhookURL
		}
	}

	if webhookURL == "" {
		return fmt.Errorf("no Slack webhook URL configured")
	}

	// Build Slack message
	message := s.buildSlackMessage(n)
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Slack API error: %s", string(body))
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id": userID,
		"type":    n.Type,
	}).Info("Slack notification sent")

	return nil
}

// buildSlackMessage builds a Slack message payload
func (s *NotificationService) buildSlackMessage(n *notification.Notification) map[string]interface{} {
	// Determine color based on priority
	color := "#36a64f" // default green
	switch n.Priority {
	case notification.PriorityCritical:
		color = "#ff0000" // red
	case notification.PriorityHigh:
		color = "#ff8c00" // orange
	case notification.PriorityMedium:
		color = "#ffcc00" // yellow
	case notification.PriorityLow:
		color = "#36a64f" // green
	}

	// Determine emoji based on notification type
	emoji := ":bell:"
	switch n.Type {
	case notification.NotificationTypeDriftAlert:
		emoji = ":warning:"
	case notification.NotificationTypeVulnerabilityAlert:
		emoji = ":rotating_light:"
	case notification.NotificationTypeCostAlert:
		emoji = ":money_with_wings:"
	case notification.NotificationTypeComplianceAlert:
		emoji = ":clipboard:"
	case notification.NotificationTypeRemediationApproval:
		emoji = ":question:"
	case notification.NotificationTypeRemediationComplete:
		emoji = ":white_check_mark:"
	case notification.NotificationTypeDailySummary, notification.NotificationTypeWeeklySummary:
		emoji = ":chart_with_upwards_trend:"
	}

	return map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color":       color,
				"title":       fmt.Sprintf("%s %s", emoji, n.Title),
				"text":        n.Message,
				"footer":      "InfraAudit",
				"footer_icon": "https://example.com/infraaudit-icon.png",
				"ts":          time.Now().Unix(),
			},
		},
	}
}

// sendEmail sends a notification via email
func (s *NotificationService) sendEmail(ctx context.Context, userID int64, n *notification.Notification) error {
	// TODO: Implement email sending with SendGrid or SMTP
	s.logger.WithFields(map[string]interface{}{
		"user_id": userID,
		"type":    n.Type,
	}).Info("Email notification queued (not implemented)")

	return nil
}

// GetPreferences retrieves notification preferences for a user
func (s *NotificationService) GetPreferences(ctx context.Context, userID int64) ([]*notification.Preference, error) {
	return s.repo.ListPreferences(ctx, userID)
}

// UpdatePreference updates a notification preference
func (s *NotificationService) UpdatePreference(ctx context.Context, userID int64, channel notification.Channel, isEnabled bool, config interface{}) error {
	pref, err := s.repo.GetPreference(ctx, userID, channel)
	if err != nil {
		return err
	}

	configJSON, _ := json.Marshal(config)

	if pref == nil {
		// Create new preference
		pref = &notification.Preference{
			ID:        uuid.New().String(),
			UserID:    userID,
			Channel:   channel,
			IsEnabled: isEnabled,
			Config:    configJSON,
		}
		return s.repo.CreatePreference(ctx, pref)
	}

	// Update existing preference
	pref.IsEnabled = isEnabled
	if config != nil {
		pref.Config = configJSON
	}

	return s.repo.UpdatePreference(ctx, pref)
}

// GetHistory retrieves notification history
func (s *NotificationService) GetHistory(ctx context.Context, userID int64, filter notification.LogFilter, limit, offset int) ([]*notification.Log, int64, error) {
	filter.UserID = userID
	return s.repo.ListLogs(ctx, filter, limit, offset)
}

// CreateWebhook creates a new webhook
func (s *NotificationService) CreateWebhook(ctx context.Context, userID int64, name, url, secret string, events []notification.EventType) (*notification.Webhook, error) {
	if secret == "" {
		secret = generateWebhookSecret()
	}

	retryConfig := notification.DefaultWebhookRetryConfig()
	retryJSON, _ := json.Marshal(retryConfig)

	webhook := &notification.Webhook{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        name,
		URL:         url,
		Secret:      secret,
		Events:      events,
		IsEnabled:   true,
		RetryConfig: retryJSON,
	}

	if err := s.repo.CreateWebhook(ctx, webhook); err != nil {
		return nil, err
	}

	s.logger.WithFields(map[string]interface{}{
		"webhook_id": webhook.ID,
		"user_id":    userID,
		"events":     events,
	}).Info("Webhook created")

	return webhook, nil
}

// GetWebhook retrieves a webhook
func (s *NotificationService) GetWebhook(ctx context.Context, id string) (*notification.Webhook, error) {
	return s.repo.GetWebhook(ctx, id)
}

// UpdateWebhook updates a webhook
func (s *NotificationService) UpdateWebhook(ctx context.Context, id string, updates map[string]interface{}) (*notification.Webhook, error) {
	webhook, err := s.repo.GetWebhook(ctx, id)
	if err != nil {
		return nil, err
	}

	if name, ok := updates["name"].(string); ok {
		webhook.Name = name
	}
	if url, ok := updates["url"].(string); ok {
		webhook.URL = url
	}
	if secret, ok := updates["secret"].(string); ok {
		webhook.Secret = secret
	}
	if events, ok := updates["events"].([]notification.EventType); ok {
		webhook.Events = events
	}
	if isEnabled, ok := updates["is_enabled"].(bool); ok {
		webhook.IsEnabled = isEnabled
	}

	if err := s.repo.UpdateWebhook(ctx, webhook); err != nil {
		return nil, err
	}

	return webhook, nil
}

// DeleteWebhook deletes a webhook
func (s *NotificationService) DeleteWebhook(ctx context.Context, id string) error {
	return s.repo.DeleteWebhook(ctx, id)
}

// ListWebhooks lists webhooks for a user
func (s *NotificationService) ListWebhooks(ctx context.Context, userID int64, limit, offset int) ([]*notification.Webhook, int64, error) {
	return s.repo.ListWebhooks(ctx, userID, limit, offset)
}

// TestWebhook sends a test event to a webhook
func (s *NotificationService) TestWebhook(ctx context.Context, id string) error {
	webhook, err := s.repo.GetWebhook(ctx, id)
	if err != nil {
		return err
	}

	testPayload := map[string]interface{}{
		"event":     "webhook.test",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"data": map[string]interface{}{
			"message": "This is a test webhook event from InfraAudit",
		},
	}

	return s.deliverWebhook(ctx, webhook, notification.EventType("webhook.test"), testPayload)
}

// TriggerEvent triggers a webhook event for a user
func (s *NotificationService) TriggerEvent(ctx context.Context, userID int64, eventType notification.EventType, data map[string]interface{}) error {
	// Get webhooks subscribed to this event
	webhooks, err := s.repo.GetWebhooksForEvent(ctx, userID, eventType)
	if err != nil {
		return fmt.Errorf("failed to get webhooks: %w", err)
	}

	payload := map[string]interface{}{
		"event":     eventType,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"data":      data,
	}

	for _, webhook := range webhooks {
		// Create delivery record
		delivery := &notification.WebhookDelivery{
			ID:        uuid.New().String(),
			WebhookID: webhook.ID,
			EventType: eventType,
			Status:    notification.DeliveryStatusPending,
		}
		payloadJSON, _ := json.Marshal(payload)
		delivery.Payload = string(payloadJSON)

		if err := s.repo.CreateDelivery(ctx, delivery); err != nil {
			s.logger.ErrorWithErr(err, "Failed to create webhook delivery record")
			continue
		}

		// Deliver asynchronously
		go func(w *notification.Webhook, d *notification.WebhookDelivery) {
			deliverCtx := context.Background()
			if err := s.deliverWebhook(deliverCtx, w, eventType, payload); err != nil {
				d.Status = notification.DeliveryStatusFailed
				d.ResponseBody = err.Error()
			} else {
				d.Status = notification.DeliveryStatusSent
				now := time.Now()
				d.DeliveredAt = &now
			}
			s.repo.UpdateDelivery(deliverCtx, d)
		}(webhook, delivery)
	}

	return nil
}

// deliverWebhook delivers a webhook event
func (s *NotificationService) deliverWebhook(ctx context.Context, webhook *notification.Webhook, eventType notification.EventType, payload map[string]interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook.URL, bytes.NewReader(payloadJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Event", string(eventType))
	req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))

	// Add HMAC signature if secret is configured
	if webhook.Secret != "" {
		signature := s.signPayload(payloadJSON, webhook.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to deliver webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook returned error status %d: %s", resp.StatusCode, string(body))
	}

	// Update last triggered
	now := time.Now()
	webhook.LastTriggered = &now
	s.repo.UpdateWebhook(ctx, webhook)

	s.logger.WithFields(map[string]interface{}{
		"webhook_id": webhook.ID,
		"event_type": eventType,
		"status":     resp.StatusCode,
	}).Info("Webhook delivered")

	return nil
}

// signPayload signs the payload with HMAC-SHA256
func (s *NotificationService) signPayload(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

// ProcessPendingNotifications processes pending notifications
func (s *NotificationService) ProcessPendingNotifications(ctx context.Context) error {
	logs, err := s.repo.GetPendingLogs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending logs: %w", err)
	}

	for _, log := range logs {
		// Retry sending
		n := &notification.Notification{
			Type:     log.NotificationType,
			Priority: log.Priority,
			UserID:   log.UserID,
		}

		// Parse payload
		var payload map[string]interface{}
		json.Unmarshal(log.Payload, &payload)
		if title, ok := payload["title"].(string); ok {
			n.Title = title
		}
		if message, ok := payload["message"].(string); ok {
			n.Message = message
		}

		if err := s.SendToChannel(ctx, log.UserID, log.Channel, n); err != nil {
			log.RetryCount++
			if log.RetryCount >= 3 {
				log.Status = notification.DeliveryStatusFailed
			} else {
				log.Status = notification.DeliveryStatusRetrying
			}
			log.ErrorMessage = err.Error()
		} else {
			log.Status = notification.DeliveryStatusSent
			now := time.Now()
			log.SentAt = &now
		}

		s.repo.UpdateLog(ctx, log)
	}

	return nil
}

// ProcessPendingWebhookDeliveries processes pending webhook deliveries
func (s *NotificationService) ProcessPendingWebhookDeliveries(ctx context.Context) error {
	deliveries, err := s.repo.GetPendingDeliveries(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending deliveries: %w", err)
	}

	for _, d := range deliveries {
		webhook, err := s.repo.GetWebhook(ctx, d.WebhookID)
		if err != nil {
			continue
		}

		var payload map[string]interface{}
		json.Unmarshal([]byte(d.Payload), &payload)

		if err := s.deliverWebhook(ctx, webhook, d.EventType, payload); err != nil {
			d.RetryCount++
			if d.RetryCount >= 3 {
				d.Status = notification.DeliveryStatusFailed
			}
			d.ResponseBody = err.Error()
		} else {
			d.Status = notification.DeliveryStatusSent
			now := time.Now()
			d.DeliveredAt = &now
		}

		s.repo.UpdateDelivery(ctx, d)
	}

	return nil
}

// generateWebhookSecret generates a random webhook secret
func generateWebhookSecret() string {
	return uuid.New().String()
}
