package services

import "context"

type CloudProviderService interface {
	ListAccounts(ctx context.Context, userID int64) ([]string, error)
	ConnectAccount(ctx context.Context, userID int64, provider CloudProvider, credentials map[string]string) error
	DeleteAccount(ctx context.Context, userID int64, provider CloudProvider) error
	Sync(ctx context.Context, userID int64, accountID string) error
}

type CostService interface {
	Analyze(ctx context.Context, userID int64, tr TimeRange) (map[string]any, error)
	Recommendations(ctx context.Context, userID int64) ([]map[string]any, error)
}

type InMemoryCloud struct{}

func NewInMemoryCloud() *InMemoryCloud { return &InMemoryCloud{} }

func (c *InMemoryCloud) ListAccounts(ctx context.Context, userID int64) ([]string, error) {
	return []string{"aws-acc-1"}, nil
}
func (c *InMemoryCloud) ConnectAccount(ctx context.Context, userID int64, provider CloudProvider, credentials map[string]string) error {
	return nil
}
func (c *InMemoryCloud) DeleteAccount(ctx context.Context, userID int64, provider CloudProvider) error {
	return nil
}
func (c *InMemoryCloud) Sync(ctx context.Context, userID int64, accountID string) error {
	return nil
}

type InMemoryCost struct{}

func NewInMemoryCost() *InMemoryCost { return &InMemoryCost{} }

func (c *InMemoryCost) Analyze(ctx context.Context, userID int64, tr TimeRange) (map[string]any, error) {
	return map[string]any{"total": 1234.56, "currency": "USD"}, nil
}
func (c *InMemoryCost) Recommendations(ctx context.Context, userID int64) ([]map[string]any, error) {
	return []map[string]any{{"title": "Rightsize EC2", "savings": 120.0}}, nil
}
