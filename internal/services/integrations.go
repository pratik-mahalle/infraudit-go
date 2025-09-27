package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type OpenAIService interface {
	AnalyzeText(ctx context.Context, prompt string) (string, error)
}

type SlackService interface {
	Send(ctx context.Context, channelID, message string) error
}

type StripeService interface {
	CreateCheckout(ctx context.Context, userID int64, plan string) (string, error)
}

type KubernetesService interface {
	ListClusters(ctx context.Context, userID int64) ([]string, error)
}

type InMemoryOpenAI struct{}

type InMemorySlack struct{}
type WebhookSlack struct {
	WebhookURL string
	HTTPClient *http.Client
}

type InMemoryStripe struct{}

type InMemoryKubernetes struct{}

func (s *InMemoryOpenAI) AnalyzeText(ctx context.Context, prompt string) (string, error) {
	return "ok", nil
}
func (s *InMemorySlack) Send(ctx context.Context, channelID, message string) error { return nil }

// WebhookSlack posts a message to Slack Incoming Webhook. channelID is optional; if provided,
// it will be passed via payload to override the default channel configured for the webhook.
func (s *WebhookSlack) Send(ctx context.Context, channelID, message string) error {
	if s == nil || s.WebhookURL == "" {
		return errors.New("slack webhook not configured")
	}
	client := s.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	payload := map[string]any{
		"text": message,
	}
	if channelID != "" {
		payload["channel"] = channelID
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("slack webhook returned non-2xx status")
	}
	return nil
}
func (s *InMemoryStripe) CreateCheckout(ctx context.Context, userID int64, plan string) (string, error) {
	return "https://checkout.example.com/session/abc", nil
}
func (k *InMemoryKubernetes) ListClusters(ctx context.Context, userID int64) ([]string, error) {
	return []string{"dev-cluster"}, nil
}
