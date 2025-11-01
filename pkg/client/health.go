package client

import "context"

// Health checks the health of the API
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	var health HealthResponse
	if err := c.doRequest(ctx, "GET", "/healthz", nil, &health); err != nil {
		return nil, err
	}
	return &health, nil
}

// Ping is a simple connectivity test
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.Health(ctx)
	return err
}
