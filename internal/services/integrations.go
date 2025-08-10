package services

import "context"

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

type InMemoryStripe struct{}

type InMemoryKubernetes struct{}

func (s *InMemoryOpenAI) AnalyzeText(ctx context.Context, prompt string) (string, error) {
	return "ok", nil
}
func (s *InMemorySlack) Send(ctx context.Context, channelID, message string) error { return nil }
func (s *InMemoryStripe) CreateCheckout(ctx context.Context, userID int64, plan string) (string, error) {
	return "https://checkout.example.com/session/abc", nil
}
func (k *InMemoryKubernetes) ListClusters(ctx context.Context, userID int64) ([]string, error) {
	return []string{"dev-cluster"}, nil
}
