package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pratik-mahalle/infraudit/internal/api/dto"
	"github.com/pratik-mahalle/infraudit/internal/domain/notification"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	notificationService notification.Service
	logger              *logger.Logger
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationService notification.Service, log *logger.Logger) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		logger:              log,
	}
}

// GetPreferences handles GET /api/v1/notifications/preferences
func (h *NotificationHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	prefs, err := h.notificationService.GetPreferences(r.Context(), userID)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to get notification preferences")
		respondError(w, http.StatusInternalServerError, "failed to get preferences")
		return
	}

	response := make([]dto.NotificationPreferenceResponse, 0, len(prefs))
	for _, p := range prefs {
		response = append(response, dto.NotificationPreferenceResponse{
			ID:        p.ID,
			Channel:   string(p.Channel),
			IsEnabled: p.IsEnabled,
			Config:    p.Config,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"preferences": response})
}

// UpdatePreference handles PUT /api/v1/notifications/preferences/{channel}
func (h *NotificationHandler) UpdatePreference(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	channel := chi.URLParam(r, "channel")
	if channel == "" {
		respondError(w, http.StatusBadRequest, "channel is required")
		return
	}

	ch := notification.Channel(channel)
	if !ch.IsValid() {
		respondError(w, http.StatusBadRequest, "invalid channel")
		return
	}

	var req dto.UpdatePreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.notificationService.UpdatePreference(r.Context(), userID, ch, req.IsEnabled, req.Config); err != nil {
		h.logger.ErrorWithErr(err, "Failed to update notification preference")
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "preference updated"})
}

// GetHistory handles GET /api/v1/notifications/history
func (h *NotificationHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	filter := notification.LogFilter{
		UserID: userID,
	}

	if channel := r.URL.Query().Get("channel"); channel != "" {
		filter.Channel = notification.Channel(channel)
	}
	if notifType := r.URL.Query().Get("type"); notifType != "" {
		filter.NotificationType = notification.NotificationType(notifType)
	}
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = notification.DeliveryStatus(status)
	}

	logs, total, err := h.notificationService.GetHistory(r.Context(), userID, filter, limit, offset)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to get notification history")
		respondError(w, http.StatusInternalServerError, "failed to get history")
		return
	}

	response := dto.ListNotificationLogsResponse{
		Logs:  make([]dto.NotificationLogResponse, 0, len(logs)),
		Total: total,
	}

	for _, l := range logs {
		response.Logs = append(response.Logs, dto.NotificationLogResponse{
			ID:               l.ID,
			Channel:          string(l.Channel),
			NotificationType: string(l.NotificationType),
			Status:           string(l.Status),
			Priority:         string(l.Priority),
			Payload:          l.Payload,
			ErrorMessage:     l.ErrorMessage,
			SentAt:           l.SentAt,
			CreatedAt:        l.CreatedAt,
		})
	}

	respondJSON(w, http.StatusOK, response)
}

// SendNotification handles POST /api/v1/notifications/send
func (h *NotificationHandler) SendNotification(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.SendNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	n := &notification.Notification{
		Type:     notification.NotificationType(req.Type),
		Priority: notification.Priority(req.Priority),
		Title:    req.Title,
		Message:  req.Message,
		Data:     req.Data,
		UserID:   userID,
	}

	if err := h.notificationService.Send(r.Context(), n); err != nil {
		h.logger.ErrorWithErr(err, "Failed to send notification")
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]string{"message": "notification sent"})
}

// ===== Webhook Handlers =====

// ListWebhooks handles GET /api/v1/webhooks
func (h *NotificationHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	webhooks, total, err := h.notificationService.ListWebhooks(r.Context(), userID, limit, offset)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to list webhooks")
		respondError(w, http.StatusInternalServerError, "failed to list webhooks")
		return
	}

	response := dto.ListWebhooksResponse{
		Webhooks: make([]dto.WebhookResponse, 0, len(webhooks)),
		Total:    total,
	}

	for _, wh := range webhooks {
		response.Webhooks = append(response.Webhooks, mapWebhookToResponse(wh))
	}

	respondJSON(w, http.StatusOK, response)
}

// CreateWebhook handles POST /api/v1/webhooks
func (h *NotificationHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.URL == "" || len(req.Events) == 0 {
		respondError(w, http.StatusBadRequest, "name, url, and events are required")
		return
	}

	events := make([]notification.EventType, 0, len(req.Events))
	for _, e := range req.Events {
		events = append(events, notification.EventType(e))
	}

	webhook, err := h.notificationService.CreateWebhook(r.Context(), userID, req.Name, req.URL, req.Secret, events)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to create webhook")
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, mapWebhookToResponse(webhook))
}

// GetWebhook handles GET /api/v1/webhooks/{id}
func (h *NotificationHandler) GetWebhook(w http.ResponseWriter, r *http.Request) {
	webhookID := chi.URLParam(r, "id")
	if webhookID == "" {
		respondError(w, http.StatusBadRequest, "webhook id is required")
		return
	}

	webhook, err := h.notificationService.GetWebhook(r.Context(), webhookID)
	if err != nil {
		respondError(w, http.StatusNotFound, "webhook not found")
		return
	}

	respondJSON(w, http.StatusOK, mapWebhookToResponse(webhook))
}

// UpdateWebhook handles PUT /api/v1/webhooks/{id}
func (h *NotificationHandler) UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	webhookID := chi.URLParam(r, "id")
	if webhookID == "" {
		respondError(w, http.StatusBadRequest, "webhook id is required")
		return
	}

	var req dto.UpdateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.URL != nil {
		updates["url"] = *req.URL
	}
	if req.Secret != nil {
		updates["secret"] = *req.Secret
	}
	if req.Events != nil {
		events := make([]notification.EventType, 0, len(*req.Events))
		for _, e := range *req.Events {
			events = append(events, notification.EventType(e))
		}
		updates["events"] = events
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}

	webhook, err := h.notificationService.UpdateWebhook(r.Context(), webhookID, updates)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to update webhook")
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, mapWebhookToResponse(webhook))
}

// DeleteWebhook handles DELETE /api/v1/webhooks/{id}
func (h *NotificationHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	webhookID := chi.URLParam(r, "id")
	if webhookID == "" {
		respondError(w, http.StatusBadRequest, "webhook id is required")
		return
	}

	if err := h.notificationService.DeleteWebhook(r.Context(), webhookID); err != nil {
		h.logger.ErrorWithErr(err, "Failed to delete webhook")
		respondError(w, http.StatusNotFound, "webhook not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "webhook deleted"})
}

// TestWebhook handles POST /api/v1/webhooks/{id}/test
func (h *NotificationHandler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	webhookID := chi.URLParam(r, "id")
	if webhookID == "" {
		respondError(w, http.StatusBadRequest, "webhook id is required")
		return
	}

	if err := h.notificationService.TestWebhook(r.Context(), webhookID); err != nil {
		h.logger.ErrorWithErr(err, "Failed to test webhook")
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "test webhook sent"})
}

// GetAvailableEvents handles GET /api/v1/webhooks/events
func (h *NotificationHandler) GetAvailableEvents(w http.ResponseWriter, _ *http.Request) {
	events := []dto.WebhookEventInfo{
		{Event: "drift.detected", Description: "Triggered when a configuration drift is detected"},
		{Event: "vulnerability.found", Description: "Triggered when a new vulnerability is discovered"},
		{Event: "compliance.failed", Description: "Triggered when a compliance check fails"},
		{Event: "cost.anomaly_detected", Description: "Triggered when a cost anomaly is detected"},
		{Event: "remediation.completed", Description: "Triggered when a remediation action completes"},
		{Event: "scan.completed", Description: "Triggered when a security scan completes"},
		{Event: "job.completed", Description: "Triggered when a scheduled job completes"},
		{Event: "job.failed", Description: "Triggered when a scheduled job fails"},
	}

	respondJSON(w, http.StatusOK, dto.AvailableWebhookEventsResponse{Events: events})
}

// Helper functions

func mapWebhookToResponse(wh *notification.Webhook) dto.WebhookResponse {
	events := make([]string, 0, len(wh.Events))
	for _, e := range wh.Events {
		events = append(events, string(e))
	}

	return dto.WebhookResponse{
		ID:            wh.ID,
		Name:          wh.Name,
		URL:           wh.URL,
		Events:        events,
		IsEnabled:     wh.IsEnabled,
		LastTriggered: wh.LastTriggered,
		CreatedAt:     wh.CreatedAt,
		UpdatedAt:     wh.UpdatedAt,
	}
}
