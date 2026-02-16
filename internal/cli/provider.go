package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func newProviderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Manage cloud providers",
	}

	cmd.AddCommand(newProviderListCmd())
	cmd.AddCommand(newProviderConnectCmd())
	cmd.AddCommand(newProviderSyncCmd())
	cmd.AddCommand(newProviderDisconnectCmd())
	cmd.AddCommand(newProviderStatusCmd())

	return cmd
}

func newProviderListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List connected providers",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			
			// Use DoRaw to match the API response structure
			var providers []struct {
				Provider    string     `json:"provider"`
				IsConnected bool       `json:"is_connected"`
				LastSynced  *time.Time `json:"last_synced,omitempty"`
			}
			if err := apiClient.DoRaw(ctx, "GET", "/api/providers", nil, &providers); err != nil {
				return fmt.Errorf("failed to list providers: %w", err)
			}

			format := getOutputFormat()
			if format != "table" {
				return printOutput(providers)
			}

			t := NewTable("PROVIDER", "STATUS", "LAST SYNCED")
			for _, p := range providers {
				status := "disconnected"
				if p.IsConnected {
					status = "connected"
				}
				lastSync := "never"
				if p.LastSynced != nil {
					lastSync = p.LastSynced.Format("2006-01-02 15:04")
				}
				t.AddRow(
					p.Provider,
					status,
					lastSync,
				)
			}
			t.Render()
			return nil
		},
	}
}

func newProviderConnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect <aws|gcp|azure>",
		Short: "Connect a cloud provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			providerType := args[0]
			
			// Build the request body matching the ConnectProviderRequest DTO
			req := map[string]interface{}{
				"provider": providerType,
			}

			switch providerType {
			case "aws":
				accessKeyID := promptInput("AWS Access Key ID: ")
				secretKey, err := promptPassword("AWS Secret Access Key: ")
				if err != nil {
					return err
				}
				region := promptInput("AWS Region [us-east-1]: ")
				if region == "" {
					region = "us-east-1"
				}
				req["aws_access_key_id"] = accessKeyID
				req["aws_secret_access_key"] = secretKey
				req["aws_region"] = region
			case "gcp":
				projectID := promptInput("GCP Project ID: ")
				serviceAccountJSON := promptInput("Path to service account JSON: ")
				req["gcp_project_id"] = projectID
				req["gcp_service_account_json"] = serviceAccountJSON
			case "azure":
				tenantID := promptInput("Azure Tenant ID: ")
				clientID := promptInput("Azure Client ID: ")
				clientSecret, err := promptPassword("Azure Client Secret: ")
				if err != nil {
					return err
				}
				subscriptionID := promptInput("Azure Subscription ID: ")
				req["azure_tenant_id"] = tenantID
				req["azure_client_id"] = clientID
				req["azure_client_secret"] = clientSecret
				req["azure_subscription_id"] = subscriptionID
			default:
				return fmt.Errorf("unsupported provider type: %s (use aws, gcp, or azure)", providerType)
			}

			ctx := context.Background()
			var result map[string]interface{}
			path := fmt.Sprintf("/api/providers/%s/connect", providerType)
			if err := apiClient.DoRaw(ctx, "POST", path, req, &result); err != nil {
				return fmt.Errorf("failed to connect provider: %w", err)
			}

			fmt.Printf("Connected to %s successfully\n", providerType)
			return nil
		},
	}
	return cmd
}

func newProviderSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync <aws|gcp|azure>",
		Short: "Sync resources from a provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			providerType := args[0]
			
			// Validate provider type
			if providerType != "aws" && providerType != "gcp" && providerType != "azure" {
				return fmt.Errorf("invalid provider type: %s (use aws, gcp, or azure)", providerType)
			}

			ctx := context.Background()
			fmt.Printf("Syncing %s resources...\n", providerType)
			
			var result map[string]interface{}
			path := fmt.Sprintf("/api/providers/%s/sync", providerType)
			if err := apiClient.DoRaw(ctx, "POST", path, nil, &result); err != nil {
				return fmt.Errorf("sync failed: %w", err)
			}
			
			fmt.Printf("Sync initiated for %s\n", providerType)
			return nil
		},
	}
}

func newProviderDisconnectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disconnect <aws|gcp|azure>",
		Short: "Disconnect a provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			providerType := args[0]
			
			// Validate provider type
			if providerType != "aws" && providerType != "gcp" && providerType != "azure" {
				return fmt.Errorf("invalid provider type: %s (use aws, gcp, or azure)", providerType)
			}

			ctx := context.Background()
			path := fmt.Sprintf("/api/providers/%s", providerType)
			if err := apiClient.DoRaw(ctx, "DELETE", path, nil, nil); err != nil {
				return fmt.Errorf("failed to disconnect provider: %w", err)
			}

			fmt.Println("Provider disconnected successfully")
			return nil
		},
	}
}

func newProviderStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show provider sync status",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			
			var providers []struct {
				Provider    string     `json:"provider"`
				IsConnected bool       `json:"is_connected"`
				LastSynced  *time.Time `json:"last_synced,omitempty"`
			}
			if err := apiClient.DoRaw(ctx, "GET", "/api/providers", nil, &providers); err != nil {
				return fmt.Errorf("failed to get provider status: %w", err)
			}

			format := getOutputFormat()
			if format != "table" {
				return printOutput(providers)
			}

			t := NewTable("PROVIDER", "STATUS", "LAST SYNCED")
			for _, p := range providers {
				status := "disconnected"
				if p.IsConnected {
					status = "connected"
				}
				lastSync := "never"
				if p.LastSynced != nil {
					lastSync = p.LastSynced.Format("2006-01-02 15:04:05")
				}
				t.AddRow(
					p.Provider,
					status,
					lastSync,
				)
			}
			t.Render()
			return nil
		},
	}
}
