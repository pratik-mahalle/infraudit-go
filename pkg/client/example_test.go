package client_test

import (
	"context"
	"fmt"
	"log"

	"github.com/pratik-mahalle/infraudit/pkg/client"
)

// Example demonstrates basic usage of the InfraAudit client
func Example() {
	// Create a new client
	c := client.NewClient(client.Config{
		BaseURL: "https://api.infraaudit.com",
	})

	ctx := context.Background()

	// Login
	loginResp, err := c.Login(ctx, "user@example.com", "password")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Logged in as: %s\n", loginResp.User.Email)

	// List resources
	resources, err := c.Resources().List(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d resources\n", len(resources))
}

// ExampleClient_Login demonstrates user authentication
func ExampleClient_Login() {
	c := client.NewClient(client.Config{
		BaseURL: "https://api.infraaudit.com",
	})

	loginResp, err := c.Login(context.Background(), "user@example.com", "password")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Token: %s\n", loginResp.Token)
}

// ExampleResourceService_List demonstrates listing resources with filters
func ExampleResourceService_List() {
	c := client.NewClient(client.Config{
		BaseURL: "https://api.infraaudit.com",
	})

	// Login first
	_, err := c.Login(context.Background(), "user@example.com", "password")
	if err != nil {
		log.Fatal(err)
	}

	// List resources with filters
	providerID := int64(1)
	resourceType := "ec2_instance"

	resources, err := c.Resources().List(context.Background(), &client.ResourceListOptions{
		ListOptions: client.ListOptions{
			Page:     1,
			PageSize: 20,
			Sort:     "created_at",
			Order:    "desc",
		},
		ProviderID:   &providerID,
		ResourceType: &resourceType,
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range resources {
		fmt.Printf("Resource: %s (%s)\n", r.Name, r.ResourceType)
	}
}

// ExampleProviderService_Create demonstrates creating a cloud provider connection
func ExampleProviderService_Create() {
	c := client.NewClient(client.Config{
		BaseURL: "https://api.infraaudit.com",
	})

	// Login first
	_, err := c.Login(context.Background(), "user@example.com", "password")
	if err != nil {
		log.Fatal(err)
	}

	// Create AWS provider
	provider, err := c.Providers().Create(context.Background(), client.CreateProviderRequest{
		Name:         "My AWS Account",
		ProviderType: "aws",
		Credentials: map[string]interface{}{
			"access_key_id":     "AKIAIOSFODNN7EXAMPLE",
			"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			"region":            "us-east-1",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created provider: %s (ID: %d)\n", provider.Name, provider.ID)
}

// ExampleProviderService_Sync demonstrates syncing resources from a provider
func ExampleProviderService_Sync() {
	c := client.NewClient(client.Config{
		BaseURL: "https://api.infraaudit.com",
	})

	// Login first
	_, err := c.Login(context.Background(), "user@example.com", "password")
	if err != nil {
		log.Fatal(err)
	}

	// Sync provider resources
	result, err := c.Providers().Sync(context.Background(), 1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Sync result:\n")
	fmt.Printf("  Resources found: %d\n", result.ResourcesFound)
	fmt.Printf("  Resources created: %d\n", result.ResourcesCreated)
	fmt.Printf("  Resources updated: %d\n", result.ResourcesUpdated)
	fmt.Printf("  Status: %s\n", result.Status)
}

// ExampleAlertService_List demonstrates listing critical alerts
func ExampleAlertService_List() {
	c := client.NewClient(client.Config{
		BaseURL: "https://api.infraaudit.com",
	})

	// Login first
	_, err := c.Login(context.Background(), "user@example.com", "password")
	if err != nil {
		log.Fatal(err)
	}

	// List critical alerts
	severity := "critical"
	status := "open"

	alerts, err := c.Alerts().List(context.Background(), &client.AlertListOptions{
		Severity: &severity,
		Status:   &status,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d critical alerts\n", len(alerts))
	for _, alert := range alerts {
		fmt.Printf("  - %s: %s\n", alert.Severity, alert.Title)
	}
}

// ExampleRecommendationService_List demonstrates listing cost recommendations
func ExampleRecommendationService_List() {
	c := client.NewClient(client.Config{
		BaseURL: "https://api.infraaudit.com",
	})

	// Login first
	_, err := c.Login(context.Background(), "user@example.com", "password")
	if err != nil {
		log.Fatal(err)
	}

	// List cost recommendations
	recommendationType := "cost"
	impact := "high"

	recommendations, err := c.Recommendations().List(context.Background(), &client.RecommendationListOptions{
		Type:   &recommendationType,
		Impact: &impact,
	})
	if err != nil {
		log.Fatal(err)
	}

	totalSavings := 0.0
	for _, rec := range recommendations {
		fmt.Printf("%s: Save $%.2f/month\n", rec.Title, rec.EstimatedSavings)
		totalSavings += rec.EstimatedSavings
	}
	fmt.Printf("Total potential savings: $%.2f/month\n", totalSavings)
}

// ExampleClient_Health demonstrates checking API health
func ExampleClient_Health() {
	c := client.NewClient(client.Config{
		BaseURL: "https://api.infraaudit.com",
	})

	health, err := c.Health(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("API Status: %s\n", health.Status)
	fmt.Printf("Version: %s\n", health.Version)
}

// ExampleClient_apiKey demonstrates using API key authentication
func ExampleClient_apiKey() {
	c := client.NewClient(client.Config{
		BaseURL: "https://api.infraaudit.com",
		APIKey:  "your-api-key",
	})

	// No need to login, API key is used automatically
	resources, err := c.Resources().List(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d resources\n", len(resources))
}
