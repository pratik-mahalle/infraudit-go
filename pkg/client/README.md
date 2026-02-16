# InfraAudit Go Client SDK

The official Go client library for the InfraAudit API. This package provides a simple and idiomatic Go interface to interact with InfraAudit's cloud infrastructure auditing and monitoring platform.

## Installation

```bash
go get github.com/pratik-mahalle/infraudit-go/pkg/client
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/pratik-mahalle/infraudit-go/pkg/client"
)

func main() {
    // Create a new client
    c := client.NewClient(client.Config{
        BaseURL: "https://api.infraudit.com",
    })

    // Login
    ctx := context.Background()
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
```

## Features

- **Authentication**: Login, register, logout, token refresh
- **Resource Management**: CRUD operations for cloud resources
- **Provider Management**: Connect and manage cloud providers (AWS, GCP, Azure)
- **Alerts**: Monitor and manage security and operational alerts
- **Recommendations**: Get cost and performance optimization suggestions
- **Drift Detection**: Track security configuration drifts
- **Anomaly Detection**: Detect cost anomalies and unusual usage patterns
- **Health Checks**: Monitor API health and connectivity

## Usage Examples

### Authentication

#### Login with Email and Password

```go
c := client.NewClient(client.Config{
    BaseURL: "https://api.infraudit.com",
})

loginResp, err := c.Login(context.Background(), "user@example.com", "password")
if err != nil {
    log.Fatal(err)
}

// Token is automatically set for future requests
fmt.Printf("Token: %s\n", loginResp.Token)
```

#### Register a New User

```go
resp, err := c.Register(context.Background(), client.RegisterRequest{
    Email:    "newuser@example.com",
    Password: "securepassword",
    Username: "newuser",
    FullName: "New User",
})
if err != nil {
    log.Fatal(err)
}
```

#### Get Current User

```go
user, err := c.GetCurrentUser(context.Background())
if err != nil {
    log.Fatal(err)
}

fmt.Printf("User: %s (%s)\n", user.Username, user.Email)
```

#### Using API Key Authentication

```go
c := client.NewClient(client.Config{
    BaseURL: "https://api.infraudit.com",
    APIKey:  "your-api-key",
})

// No need to login, API key is used automatically
resources, err := c.Resources().List(context.Background(), nil)
```

### Resource Management

#### List Resources

```go
// List all resources
resources, err := c.Resources().List(context.Background(), nil)

// List with filters
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
```

#### Get a Single Resource

```go
resource, err := c.Resources().Get(context.Background(), 123)
if err != nil {
    var apiErr *client.APIError
    if errors.As(err, &apiErr) && apiErr.IsNotFound() {
        fmt.Println("Resource not found")
        return
    }
    log.Fatal(err)
}

fmt.Printf("Resource: %s (%s)\n", resource.Name, resource.ResourceType)
```

#### Create a Resource

```go
resource, err := c.Resources().Create(context.Background(), client.CreateResourceRequest{
    ProviderID:   1,
    ResourceType: "s3_bucket",
    ResourceID:   "my-bucket-id",
    Name:         "my-bucket",
    Region:       "us-east-1",
    Tags: map[string]string{
        "Environment": "production",
        "Team":        "platform",
    },
})
```

#### Update a Resource

```go
name := "updated-name"
resource, err := c.Resources().Update(context.Background(), 123, client.UpdateResourceRequest{
    Name: &name,
    Tags: map[string]string{
        "Environment": "staging",
    },
})
```

#### Delete a Resource

```go
err := c.Resources().Delete(context.Background(), 123)
```

#### Get Resource Cost

```go
cost, err := c.Resources().GetCost(context.Background(), 123)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Monthly cost: $%.2f\n", cost.Monthly)
```

### Provider Management

#### List Providers

```go
providers, err := c.Providers().List(context.Background(), nil)

// Filter by type
providerType := "aws"
providers, err := c.Providers().List(context.Background(), &client.ProviderListOptions{
    ProviderType: &providerType,
})
```

#### Create a Provider

```go
provider, err := c.Providers().Create(context.Background(), client.CreateProviderRequest{
    Name:         "My AWS Account",
    ProviderType: "aws",
    Credentials: map[string]interface{}{
        "access_key_id":     "AKIAIOSFODNN7EXAMPLE",
        "secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
        "region":            "us-east-1",
    },
})
```

#### Test Provider Connection

```go
success, err := c.Providers().TestConnection(context.Background(), 1)
if err != nil {
    log.Fatal(err)
}

if success {
    fmt.Println("Connection successful!")
}
```

#### Sync Provider Resources

```go
result, err := c.Providers().Sync(context.Background(), 1)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d resources, created %d, updated %d\n",
    result.ResourcesFound, result.ResourcesCreated, result.ResourcesUpdated)
```

### Alert Management

#### List Alerts

```go
severity := "critical"
alerts, err := c.Alerts().List(context.Background(), &client.AlertListOptions{
    Severity: &severity,
    ListOptions: client.ListOptions{
        PageSize: 50,
    },
})
```

#### Create an Alert

```go
alert, err := c.Alerts().Create(context.Background(), client.CreateAlertRequest{
    Type:        "security",
    Severity:    "high",
    Title:       "Unencrypted S3 Bucket",
    Description: "S3 bucket 'my-bucket' is not encrypted",
})
```

#### Acknowledge an Alert

```go
alert, err := c.Alerts().Acknowledge(context.Background(), 123)
```

#### Resolve an Alert

```go
alert, err := c.Alerts().Resolve(context.Background(), 123)
```

### Recommendations

#### List Recommendations

```go
recommendationType := "cost"
recommendations, err := c.Recommendations().List(context.Background(), &client.RecommendationListOptions{
    Type: &recommendationType,
})

for _, rec := range recommendations {
    fmt.Printf("%s: Save $%.2f/month\n", rec.Title, rec.EstimatedSavings)
}
```

#### Apply a Recommendation

```go
rec, err := c.Recommendations().Apply(context.Background(), 123)
```

#### Dismiss a Recommendation

```go
rec, err := c.Recommendations().Dismiss(context.Background(), 123)
```

### Drift Detection

#### List Drifts

```go
status := "detected"
drifts, err := c.Drifts().List(context.Background(), &client.DriftListOptions{
    Status: &status,
})

for _, drift := range drifts {
    fmt.Printf("Drift: %s (Severity: %s)\n", drift.Description, drift.Severity)
}
```

#### Investigate a Drift

```go
drift, err := c.Drifts().Investigate(context.Background(), 123)
```

#### Resolve a Drift

```go
drift, err := c.Drifts().Resolve(context.Background(), 123)
```

### Anomaly Detection

#### List Anomalies

```go
anomalies, err := c.Anomalies().List(context.Background(), nil)

for _, anomaly := range anomalies {
    fmt.Printf("Anomaly: %s (%.2f%% deviation)\n",
        anomaly.Description, anomaly.Deviation)
}
```

#### Investigate an Anomaly

```go
anomaly, err := c.Anomalies().Investigate(context.Background(), 123)
```

### Health Check

```go
health, err := c.Health(context.Background())
if err != nil {
    log.Fatal(err)
}

fmt.Printf("API Status: %s\n", health.Status)

// Simple ping
if err := c.Ping(context.Background()); err != nil {
    log.Fatal("API is down")
}
```

## Error Handling

The client provides rich error information through the `APIError` type:

```go
resource, err := c.Resources().Get(context.Background(), 123)
if err != nil {
    var apiErr *client.APIError
    if errors.As(err, &apiErr) {
        // Check specific error types
        if apiErr.IsNotFound() {
            fmt.Println("Resource not found")
        } else if apiErr.IsUnauthorized() {
            fmt.Println("Authentication required")
        } else if apiErr.IsForbidden() {
            fmt.Println("Permission denied")
        } else if apiErr.IsValidationError() {
            fmt.Printf("Validation error: %v\n", apiErr.Details)
        } else if apiErr.IsServerError() {
            fmt.Println("Server error")
        }

        fmt.Printf("Error code: %s\n", apiErr.Code)
        fmt.Printf("Error message: %s\n", apiErr.Message)
        fmt.Printf("Status code: %d\n", apiErr.StatusCode)
    }
}
```

## Configuration

### Custom HTTP Client

```go
import "net/http"

httpClient := &http.Client{
    Timeout: 60 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:    10,
        IdleConnTimeout: 30 * time.Second,
    },
}

c := client.NewClient(client.Config{
    BaseURL:    "https://api.infraudit.com",
    HTTPClient: httpClient,
})
```

### Custom Timeout

```go
c := client.NewClient(client.Config{
    BaseURL: "https://api.infraudit.com",
    Timeout: 60 * time.Second,
})
```

### Manual Token Management

```go
c := client.NewClient(client.Config{
    BaseURL: "https://api.infraudit.com",
})

// Set token manually
c.SetToken("your-jwt-token")

// Get current token
token := c.GetToken()
```

## Context Support

All methods accept a `context.Context` parameter for timeout and cancellation:

```go
import "time"

// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

resources, err := c.Resources().List(ctx, nil)

// With cancellation
ctx, cancel := context.WithCancel(context.Background())

go func() {
    time.Sleep(5 * time.Second)
    cancel()
}()

resources, err := c.Resources().List(ctx, nil)
```

## Best Practices

1. **Reuse the Client**: Create one client instance and reuse it throughout your application
2. **Use Context**: Always pass a context with appropriate timeout
3. **Handle Errors**: Check for specific error types using the APIError helpers
4. **Token Management**: The client automatically manages tokens after login
5. **Pagination**: Use pagination for large lists to avoid memory issues

## Thread Safety

The client is safe for concurrent use by multiple goroutines.

## License

See the main repository for license information.

## Support

For issues and questions, please visit: https://github.com/pratik-mahalle/infraudit-go/issues
