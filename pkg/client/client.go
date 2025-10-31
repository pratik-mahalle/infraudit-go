package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the main InfraAudit API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
	token      string // JWT token for authenticated requests
}

// Config holds the client configuration
type Config struct {
	BaseURL    string        // API base URL (e.g., "https://api.infraaudit.com")
	APIKey     string        // Optional API key for authentication
	Timeout    time.Duration // HTTP client timeout (default: 30s)
	HTTPClient *http.Client  // Optional custom HTTP client
}

// NewClient creates a new InfraAudit API client
func NewClient(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: cfg.Timeout,
		}
	}

	return &Client{
		baseURL:    cfg.BaseURL,
		httpClient: httpClient,
		apiKey:     cfg.APIKey,
	}
}

// SetToken sets the JWT token for authenticated requests
func (c *Client) SetToken(token string) {
	c.token = token
}

// GetToken returns the current JWT token
func (c *Client) GetToken() string {
	return c.token
}

// doRequest performs an HTTP request with proper error handling
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add authentication
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	} else if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	// Perform request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err != nil {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
		}
		apiErr.StatusCode = resp.StatusCode
		return &apiErr
	}

	// Parse success response
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// Resources returns the resource management service
func (c *Client) Resources() *ResourceService {
	return &ResourceService{client: c}
}

// Providers returns the provider management service
func (c *Client) Providers() *ProviderService {
	return &ProviderService{client: c}
}

// Alerts returns the alert management service
func (c *Client) Alerts() *AlertService {
	return &AlertService{client: c}
}

// Recommendations returns the recommendation service
func (c *Client) Recommendations() *RecommendationService {
	return &RecommendationService{client: c}
}

// Drifts returns the drift detection service
func (c *Client) Drifts() *DriftService {
	return &DriftService{client: c}
}

// Anomalies returns the anomaly detection service
func (c *Client) Anomalies() *AnomalyService {
	return &AnomalyService{client: c}
}
