package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pratik-mahalle/infraudit/internal/domain/notification"
)

// NotificationRepository implements notification.Repository for PostgreSQL
type NotificationRepository struct {
	db *sql.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// ===== Preferences =====

// CreatePreference creates a new notification preference
func (r *NotificationRepository) CreatePreference(ctx context.Context, p *notification.Preference) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}

	configJSON, err := json.Marshal(p.Config)
	if err != nil {
		configJSON = []byte("{}")
	}

	query := `
		INSERT INTO notification_preferences (id, user_id, channel, is_enabled, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	now := time.Now()
	_, err = r.db.ExecContext(ctx, query,
		p.ID,
		p.UserID,
		string(p.Channel),
		p.IsEnabled,
		string(configJSON),
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create notification preference: %w", err)
	}

	p.CreatedAt = now
	p.UpdatedAt = now
	return nil
}

// GetPreference retrieves a notification preference
func (r *NotificationRepository) GetPreference(ctx context.Context, userID int64, channel notification.Channel) (*notification.Preference, error) {
	query := `
		SELECT id, user_id, channel, is_enabled, config, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1 AND channel = $2
	`

	var p notification.Preference
	var configStr sql.NullString
	var ch string

	err := r.db.QueryRowContext(ctx, query, userID, string(channel)).Scan(
		&p.ID,
		&p.UserID,
		&ch,
		&p.IsEnabled,
		&configStr,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get notification preference: %w", err)
	}

	p.Channel = notification.Channel(ch)
	if configStr.Valid {
		p.Config = json.RawMessage(configStr.String)
	}

	return &p, nil
}

// UpdatePreference updates a notification preference
func (r *NotificationRepository) UpdatePreference(ctx context.Context, p *notification.Preference) error {
	configJSON, err := json.Marshal(p.Config)
	if err != nil {
		configJSON = []byte("{}")
	}

	query := `
		UPDATE notification_preferences
		SET is_enabled = $1, config = $2, updated_at = $3
		WHERE id = $4
	`

	p.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		p.IsEnabled,
		string(configJSON),
		p.UpdatedAt,
		p.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update notification preference: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("notification preference not found")
	}

	return nil
}

// ListPreferences lists all preferences for a user
func (r *NotificationRepository) ListPreferences(ctx context.Context, userID int64) ([]*notification.Preference, error) {
	query := `
		SELECT id, user_id, channel, is_enabled, config, created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1
		ORDER BY channel
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list notification preferences: %w", err)
	}
	defer rows.Close()

	var preferences []*notification.Preference
	for rows.Next() {
		var p notification.Preference
		var configStr sql.NullString
		var ch string

		err := rows.Scan(
			&p.ID,
			&p.UserID,
			&ch,
			&p.IsEnabled,
			&configStr,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification preference: %w", err)
		}

		p.Channel = notification.Channel(ch)
		if configStr.Valid {
			p.Config = json.RawMessage(configStr.String)
		}

		preferences = append(preferences, &p)
	}

	return preferences, nil
}

// DeletePreference deletes a notification preference
func (r *NotificationRepository) DeletePreference(ctx context.Context, id string) error {
	query := `DELETE FROM notification_preferences WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification preference: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("notification preference not found")
	}

	return nil
}

// ===== Logs =====

// CreateLog creates a new notification log
func (r *NotificationRepository) CreateLog(ctx context.Context, l *notification.Log) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}

	payloadJSON, err := json.Marshal(l.Payload)
	if err != nil {
		payloadJSON = []byte("{}")
	}

	query := `
		INSERT INTO notification_logs (id, user_id, channel, event_type, status, priority, message, error_message, retry_count, sent_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	now := time.Now()
	_, err = r.db.ExecContext(ctx, query,
		l.ID,
		l.UserID,
		string(l.Channel),
		string(l.NotificationType),
		string(l.Status),
		string(l.Priority),
		string(payloadJSON),
		l.ErrorMessage,
		l.RetryCount,
		l.SentAt,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create notification log: %w", err)
	}

	l.CreatedAt = now
	return nil
}

// UpdateLog updates a notification log
func (r *NotificationRepository) UpdateLog(ctx context.Context, l *notification.Log) error {
	payloadJSON, err := json.Marshal(l.Payload)
	if err != nil {
		payloadJSON = []byte("{}")
	}

	query := `
		UPDATE notification_logs
		SET status = $1, error_message = $2, retry_count = $3, sent_at = $4
		WHERE id = $5
	`

	_, err = r.db.ExecContext(ctx, query,
		string(l.Status),
		l.ErrorMessage,
		l.RetryCount,
		l.SentAt,
		l.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update notification log: %w", err)
	}

	// Update message (payload) if needed
	if len(l.Payload) > 0 {
		updateQuery := `UPDATE notification_logs SET message = $1 WHERE id = $2`
		r.db.ExecContext(ctx, updateQuery, string(payloadJSON), l.ID)
	}

	return nil
}

// ListLogs lists notification logs with filtering
func (r *NotificationRepository) ListLogs(ctx context.Context, filter notification.LogFilter, limit, offset int) ([]*notification.Log, int64, error) {
	baseSelect := `
		SELECT id, user_id, channel, event_type, status, priority, message, error_message, retry_count, sent_at, created_at
		FROM notification_logs
		WHERE 1=1
	`
	countBase := `SELECT COUNT(*) FROM notification_logs WHERE 1=1`
	var args []interface{}
	paramN := 1

	queryFilters := ""
	if filter.UserID > 0 {
		queryFilters += fmt.Sprintf(" AND user_id = $%d", paramN)
		args = append(args, filter.UserID)
		paramN++
	}
	if filter.Channel != "" {
		queryFilters += fmt.Sprintf(" AND channel = $%d", paramN)
		args = append(args, string(filter.Channel))
		paramN++
	}
	if filter.NotificationType != "" {
		queryFilters += fmt.Sprintf(" AND event_type = $%d", paramN)
		args = append(args, string(filter.NotificationType))
		paramN++
	}
	if filter.Status != "" {
		queryFilters += fmt.Sprintf(" AND status = $%d", paramN)
		args = append(args, string(filter.Status))
		paramN++
	}
	if filter.From != nil {
		queryFilters += fmt.Sprintf(" AND created_at >= $%d", paramN)
		args = append(args, *filter.From)
		paramN++
	}
	if filter.To != nil {
		queryFilters += fmt.Sprintf(" AND created_at <= $%d", paramN)
		args = append(args, *filter.To)
		paramN++
	}

	countQuery := countBase + queryFilters
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count notification logs: %w", err)
	}

	query := baseSelect + queryFilters + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", paramN, paramN+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list notification logs: %w", err)
	}
	defer rows.Close()

	var logs []*notification.Log
	for rows.Next() {
		var l notification.Log
		var payloadStr sql.NullString
		var sentAt sql.NullTime
		var channel, notifType, status, priority string

		err := rows.Scan(
			&l.ID,
			&l.UserID,
			&channel,
			&notifType,
			&status,
			&priority,
			&payloadStr,
			&l.ErrorMessage,
			&l.RetryCount,
			&sentAt,
			&l.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan notification log: %w", err)
		}

		l.Channel = notification.Channel(channel)
		l.NotificationType = notification.NotificationType(notifType)
		l.Status = notification.DeliveryStatus(status)
		l.Priority = notification.Priority(priority)
		if payloadStr.Valid {
			l.Payload = json.RawMessage(payloadStr.String)
		}
		if sentAt.Valid {
			l.SentAt = &sentAt.Time
		}

		logs = append(logs, &l)
	}

	return logs, total, nil
}

// GetPendingLogs retrieves pending notification logs
func (r *NotificationRepository) GetPendingLogs(ctx context.Context) ([]*notification.Log, error) {
	query := `
		SELECT id, user_id, channel, event_type, status, priority, message, error_message, retry_count, sent_at, created_at
		FROM notification_logs
		WHERE status IN ('pending', 'retrying')
		ORDER BY priority DESC, created_at ASC
		LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending logs: %w", err)
	}
	defer rows.Close()

	var logs []*notification.Log
	for rows.Next() {
		var l notification.Log
		var payloadStr sql.NullString
		var sentAt sql.NullTime
		var channel, notifType, status, priority string

		err := rows.Scan(
			&l.ID,
			&l.UserID,
			&channel,
			&notifType,
			&status,
			&priority,
			&payloadStr,
			&l.ErrorMessage,
			&l.RetryCount,
			&sentAt,
			&l.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification log: %w", err)
		}

		l.Channel = notification.Channel(channel)
		l.NotificationType = notification.NotificationType(notifType)
		l.Status = notification.DeliveryStatus(status)
		l.Priority = notification.Priority(priority)
		if payloadStr.Valid {
			l.Payload = json.RawMessage(payloadStr.String)
		}
		if sentAt.Valid {
			l.SentAt = &sentAt.Time
		}

		logs = append(logs, &l)
	}

	return logs, nil
}

// ===== Webhooks =====

// CreateWebhook creates a new webhook
func (r *NotificationRepository) CreateWebhook(ctx context.Context, w *notification.Webhook) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}

	eventsJSON, err := json.Marshal(w.Events)
	if err != nil {
		eventsJSON = []byte("[]")
	}

	retryJSON, _ := json.Marshal(w.RetryConfig)

	query := `
		INSERT INTO webhooks (id, user_id, name, url, secret, event_types, is_enabled, retry_config, last_triggered_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	now := time.Now()
	_, err = r.db.ExecContext(ctx, query,
		w.ID,
		w.UserID,
		w.Name,
		w.URL,
		w.Secret,
		string(eventsJSON),
		w.IsEnabled,
		string(retryJSON),
		w.LastTriggered,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	w.CreatedAt = now
	w.UpdatedAt = now
	return nil
}

// GetWebhook retrieves a webhook by ID
func (r *NotificationRepository) GetWebhook(ctx context.Context, id string) (*notification.Webhook, error) {
	query := `
		SELECT id, user_id, name, url, secret, event_types, is_enabled, retry_config, last_triggered_at, created_at, updated_at
		FROM webhooks
		WHERE id = $1
	`

	var w notification.Webhook
	var eventsStr, retryStr, secret sql.NullString
	var lastTriggered sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&w.ID,
		&w.UserID,
		&w.Name,
		&w.URL,
		&secret,
		&eventsStr,
		&w.IsEnabled,
		&retryStr,
		&lastTriggered,
		&w.CreatedAt,
		&w.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("webhook not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	if secret.Valid {
		w.Secret = secret.String
	}
	if eventsStr.Valid {
		json.Unmarshal([]byte(eventsStr.String), &w.Events)
	}
	if retryStr.Valid {
		w.RetryConfig = json.RawMessage(retryStr.String)
	}
	if lastTriggered.Valid {
		w.LastTriggered = &lastTriggered.Time
	}

	return &w, nil
}

// UpdateWebhook updates a webhook
func (r *NotificationRepository) UpdateWebhook(ctx context.Context, w *notification.Webhook) error {
	eventsJSON, err := json.Marshal(w.Events)
	if err != nil {
		eventsJSON = []byte("[]")
	}

	retryJSON, _ := json.Marshal(w.RetryConfig)

	query := `
		UPDATE webhooks
		SET name = $1, url = $2, secret = $3, event_types = $4, is_enabled = $5, retry_config = $6, last_triggered_at = $7, updated_at = $8
		WHERE id = $9
	`

	w.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		w.Name,
		w.URL,
		w.Secret,
		string(eventsJSON),
		w.IsEnabled,
		string(retryJSON),
		w.LastTriggered,
		w.UpdatedAt,
		w.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update webhook: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("webhook not found")
	}

	return nil
}

// DeleteWebhook deletes a webhook
func (r *NotificationRepository) DeleteWebhook(ctx context.Context, id string) error {
	query := `DELETE FROM webhooks WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("webhook not found")
	}

	return nil
}

// ListWebhooks lists webhooks for a user
func (r *NotificationRepository) ListWebhooks(ctx context.Context, userID int64, limit, offset int) ([]*notification.Webhook, int64, error) {
	countQuery := `SELECT COUNT(*) FROM webhooks WHERE user_id = $1`
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count webhooks: %w", err)
	}

	query := `
		SELECT id, user_id, name, url, secret, event_types, is_enabled, retry_config, last_triggered_at, created_at, updated_at
		FROM webhooks
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list webhooks: %w", err)
	}
	defer rows.Close()

	var webhooks []*notification.Webhook
	for rows.Next() {
		var w notification.Webhook
		var eventsStr, retryStr, secret sql.NullString
		var lastTriggered sql.NullTime

		err := rows.Scan(
			&w.ID,
			&w.UserID,
			&w.Name,
			&w.URL,
			&secret,
			&eventsStr,
			&w.IsEnabled,
			&retryStr,
			&lastTriggered,
			&w.CreatedAt,
			&w.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan webhook: %w", err)
		}

		if secret.Valid {
			w.Secret = secret.String
		}
		if eventsStr.Valid {
			json.Unmarshal([]byte(eventsStr.String), &w.Events)
		}
		if retryStr.Valid {
			w.RetryConfig = json.RawMessage(retryStr.String)
		}
		if lastTriggered.Valid {
			w.LastTriggered = &lastTriggered.Time
		}

		webhooks = append(webhooks, &w)
	}

	return webhooks, total, nil
}

// GetWebhooksForEvent retrieves webhooks subscribed to an event
func (r *NotificationRepository) GetWebhooksForEvent(ctx context.Context, userID int64, eventType notification.EventType) ([]*notification.Webhook, error) {
	query := `
		SELECT id, user_id, name, url, secret, event_types, is_enabled, retry_config, last_triggered_at, created_at, updated_at
		FROM webhooks
		WHERE user_id = $1 AND is_enabled = true
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query webhooks: %w", err)
	}
	defer rows.Close()

	var webhooks []*notification.Webhook
	for rows.Next() {
		var w notification.Webhook
		var eventsStr, retryStr, secret sql.NullString
		var lastTriggered sql.NullTime

		err := rows.Scan(
			&w.ID,
			&w.UserID,
			&w.Name,
			&w.URL,
			&secret,
			&eventsStr,
			&w.IsEnabled,
			&retryStr,
			&lastTriggered,
			&w.CreatedAt,
			&w.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook: %w", err)
		}

		if secret.Valid {
			w.Secret = secret.String
		}
		if eventsStr.Valid {
			json.Unmarshal([]byte(eventsStr.String), &w.Events)
		}
		if retryStr.Valid {
			w.RetryConfig = json.RawMessage(retryStr.String)
		}
		if lastTriggered.Valid {
			w.LastTriggered = &lastTriggered.Time
		}

		// Filter by event type
		for _, e := range w.Events {
			if e == eventType {
				webhooks = append(webhooks, &w)
				break
			}
		}
	}

	return webhooks, nil
}

// ===== Webhook Deliveries =====

// CreateDelivery creates a new webhook delivery
func (r *NotificationRepository) CreateDelivery(ctx context.Context, d *notification.WebhookDelivery) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}

	query := `
		INSERT INTO webhook_deliveries (id, webhook_id, event_type, payload, status, response_status, response_body, retry_count, attempted_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		d.ID,
		d.WebhookID,
		string(d.EventType),
		d.Payload,
		string(d.Status),
		d.ResponseStatus,
		d.ResponseBody,
		d.RetryCount,
		d.DeliveredAt,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to create webhook delivery: %w", err)
	}

	d.CreatedAt = now
	return nil
}

// UpdateDelivery updates a webhook delivery
func (r *NotificationRepository) UpdateDelivery(ctx context.Context, d *notification.WebhookDelivery) error {
	query := `
		UPDATE webhook_deliveries
		SET status = $1, response_status = $2, response_body = $3, retry_count = $4, attempted_at = $5
		WHERE id = $6
	`

	_, err := r.db.ExecContext(ctx, query,
		string(d.Status),
		d.ResponseStatus,
		d.ResponseBody,
		d.RetryCount,
		d.DeliveredAt,
		d.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update webhook delivery: %w", err)
	}

	return nil
}

// GetPendingDeliveries retrieves pending webhook deliveries
func (r *NotificationRepository) GetPendingDeliveries(ctx context.Context) ([]*notification.WebhookDelivery, error) {
	query := `
		SELECT id, webhook_id, event_type, payload, status, response_status, response_body, retry_count, attempted_at, created_at
		FROM webhook_deliveries
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []*notification.WebhookDelivery
	for rows.Next() {
		var d notification.WebhookDelivery
		var eventType, status string
		var deliveredAt sql.NullTime
		var responseBody sql.NullString

		err := rows.Scan(
			&d.ID,
			&d.WebhookID,
			&eventType,
			&d.Payload,
			&status,
			&d.ResponseStatus,
			&responseBody,
			&d.RetryCount,
			&deliveredAt,
			&d.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook delivery: %w", err)
		}

		d.EventType = notification.EventType(eventType)
		d.Status = notification.DeliveryStatus(status)
		if responseBody.Valid {
			d.ResponseBody = responseBody.String
		}
		if deliveredAt.Valid {
			d.DeliveredAt = &deliveredAt.Time
		}

		deliveries = append(deliveries, &d)
	}

	return deliveries, nil
}

// ListDeliveries lists webhook deliveries
func (r *NotificationRepository) ListDeliveries(ctx context.Context, webhookID string, limit, offset int) ([]*notification.WebhookDelivery, int64, error) {
	countQuery := `SELECT COUNT(*) FROM webhook_deliveries WHERE webhook_id = $1`
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, webhookID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count webhook deliveries: %w", err)
	}

	query := `
		SELECT id, webhook_id, event_type, payload, status, response_status, response_body, retry_count, attempted_at, created_at
		FROM webhook_deliveries
		WHERE webhook_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, webhookID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list webhook deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []*notification.WebhookDelivery
	for rows.Next() {
		var d notification.WebhookDelivery
		var eventType, status string
		var deliveredAt sql.NullTime
		var responseBody sql.NullString

		err := rows.Scan(
			&d.ID,
			&d.WebhookID,
			&eventType,
			&d.Payload,
			&status,
			&d.ResponseStatus,
			&responseBody,
			&d.RetryCount,
			&deliveredAt,
			&d.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan webhook delivery: %w", err)
		}

		d.EventType = notification.EventType(eventType)
		d.Status = notification.DeliveryStatus(status)
		if responseBody.Valid {
			d.ResponseBody = responseBody.String
		}
		if deliveredAt.Valid {
			d.DeliveredAt = &deliveredAt.Time
		}

		deliveries = append(deliveries, &d)
	}

	return deliveries, total, nil
}
