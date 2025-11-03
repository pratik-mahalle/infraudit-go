package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GeminiClient is a client for the Google Gemini API
type GeminiClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	model      string
}

// GeminiRequest represents a request to the Gemini API
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

// GeminiContent represents content in a Gemini request
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

// GeminiPart represents a part of the content
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiResponse represents the response from Gemini API
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

// NewGeminiClient creates a new Gemini API client
func NewGeminiClient(apiKey string) *GeminiClient {
	return &GeminiClient{
		apiKey:  apiKey,
		baseURL: "https://generativelanguage.googleapis.com/v1beta/models",
		model:   "gemini-2.0-flash-exp", // Using Gemini 2.0 Flash (latest available)
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GenerateContent generates content using the Gemini API
func (c *GeminiClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("Gemini API key is not configured")
	}

	url := fmt.Sprintf("%s/%s:generateContent", c.baseURL, c.model)

	requestBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// AnalyzeResourceForRecommendations analyzes a resource and generates recommendations
func (c *GeminiClient) AnalyzeResourceForRecommendations(ctx context.Context, resourceData map[string]interface{}) (string, error) {
	jsonData, err := json.MarshalIndent(resourceData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal resource data: %w", err)
	}

	prompt := fmt.Sprintf(`Analyze the following cloud infrastructure resource and provide actionable recommendations for:
1. Cost optimization (rightsizing, unused resources, reserved instances)
2. Security improvements (encryption, access controls, network security)
3. Compliance (industry standards, best practices)

Resource Data:
%s

Provide recommendations in JSON format with the following structure for each recommendation:
{
  "type": "cost_optimization|security_improvement|compliance",
  "priority": "critical|high|medium|low",
  "title": "Brief title",
  "description": "Detailed description of the issue and recommendation",
  "category": "specific category (e.g., encryption, rightsizing, access_control, compliance_standard)",
  "savings": estimated_monthly_savings_in_dollars (only for cost recommendations, 0 otherwise),
  "effort": "low|medium|high",
  "impact": "high|medium|low",
  "resources": ["resource_id"]
}

Return ONLY a JSON array of recommendations, no additional text.`, string(jsonData))

	return c.GenerateContent(ctx, prompt)
}

// AnalyzeVulnerabilitiesForRecommendations analyzes vulnerabilities and generates security recommendations
func (c *GeminiClient) AnalyzeVulnerabilitiesForRecommendations(ctx context.Context, vulnerabilities []map[string]interface{}) (string, error) {
	jsonData, err := json.MarshalIndent(vulnerabilities, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal vulnerabilities: %w", err)
	}

	prompt := fmt.Sprintf(`Analyze the following security vulnerabilities and provide actionable security recommendations:

Vulnerabilities:
%s

Provide recommendations in JSON format focusing on:
1. Critical vulnerabilities that need immediate attention
2. Patterns of security issues
3. Preventive measures

Use this JSON structure for each recommendation:
{
  "type": "security_improvement",
  "priority": "critical|high|medium|low",
  "title": "Brief title",
  "description": "Detailed description and remediation steps",
  "category": "vulnerability_management|patching|security_hardening",
  "savings": 0,
  "effort": "low|medium|high",
  "impact": "high|medium|low",
  "resources": ["affected_resource_ids"]
}

Return ONLY a JSON array of recommendations, no additional text.`, string(jsonData))

	return c.GenerateContent(ctx, prompt)
}

// AnalyzeDriftsForRecommendations analyzes configuration drifts and generates compliance recommendations
func (c *GeminiClient) AnalyzeDriftsForRecommendations(ctx context.Context, drifts []map[string]interface{}) (string, error) {
	jsonData, err := json.MarshalIndent(drifts, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal drifts: %w", err)
	}

	prompt := fmt.Sprintf(`Analyze the following configuration drifts and provide actionable recommendations for security and compliance:

Configuration Drifts:
%s

Provide recommendations in JSON format focusing on:
1. Security configuration changes that need attention
2. Compliance violations
3. Drift remediation strategies

Use this JSON structure for each recommendation:
{
  "type": "security_improvement|compliance",
  "priority": "critical|high|medium|low",
  "title": "Brief title",
  "description": "Detailed description and remediation steps",
  "category": "configuration_management|compliance_standard|security_baseline",
  "savings": 0,
  "effort": "low|medium|high",
  "impact": "high|medium|low",
  "resources": ["affected_resource_ids"]
}

Return ONLY a JSON array of recommendations, no additional text.`, string(jsonData))

	return c.GenerateContent(ctx, prompt)
}
