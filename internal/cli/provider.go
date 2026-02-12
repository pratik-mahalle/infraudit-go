package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pratik-mahalle/infraudit/pkg/client"
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
			providers, err := apiClient.Providers().List(ctx, nil)
			if err != nil {
				return fmt.Errorf("failed to list providers: %w", err)
			}

			format := getOutputFormat()
			if format != "table" {
				return printOutput(providers)
			}

			t := NewTable("ID", "NAME", "TYPE", "STATUS", "LAST SYNCED")
			for _, p := range providers {
				lastSync := "never"
				if p.LastSyncedAt != nil {
					lastSync = p.LastSyncedAt.Format("2006-01-02 15:04")
				}
				t.AddRow(
					strconv.FormatInt(p.ID, 10),
					p.Name,
					p.ProviderType,
					formatStatus(p.Status),
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
			credentials := map[string]interface{}{}

			switch providerType {
			case "aws":
				credentials["access_key_id"] = promptInput("AWS Access Key ID: ")
				credentials["secret_access_key"] = promptPassword("AWS Secret Access Key: ")
				region := promptInput("AWS Region [us-east-1]: ")
				if region == "" {
					region = "us-east-1"
				}
				credentials["region"] = region
			case "gcp":
				credentials["project_id"] = promptInput("GCP Project ID: ")
				credentials["service_account_json"] = promptInput("Path to service account JSON: ")
			case "azure":
				credentials["tenant_id"] = promptInput("Azure Tenant ID: ")
				credentials["client_id"] = promptInput("Azure Client ID: ")
				credentials["client_secret"] = promptPassword("Azure Client Secret: ")
				credentials["subscription_id"] = promptInput("Azure Subscription ID: ")
			default:
				return fmt.Errorf("unsupported provider type: %s (use aws, gcp, or azure)", providerType)
			}

			ctx := context.Background()
			provider, err := apiClient.Providers().Create(ctx, client.CreateProviderRequest{
				Name:         providerType + "-account",
				ProviderType: providerType,
				Credentials:  credentials,
			})
			if err != nil {
				return fmt.Errorf("failed to connect provider: %w", err)
			}

			fmt.Printf("Connected to %s successfully (ID: %d)\n", providerType, provider.ID)
			return nil
		},
	}
	return cmd
}

func newProviderSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync <provider-id>",
		Short: "Sync resources from a provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid provider ID: %s", args[0])
			}

			ctx := context.Background()
			fmt.Println("Syncing resources...")
			result, err := apiClient.Providers().Sync(ctx, id)
			if err != nil {
				return fmt.Errorf("sync failed: %w", err)
			}

			fmt.Printf("Sync complete: %d found, %d created, %d updated\n",
				result.ResourcesFound, result.ResourcesCreated, result.ResourcesUpdated)
			if len(result.Errors) > 0 {
				fmt.Printf("Errors: %d\n", len(result.Errors))
				for _, e := range result.Errors {
					fmt.Printf("  - %s\n", e)
				}
			}
			return nil
		},
	}
}

func newProviderDisconnectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disconnect <provider-id>",
		Short: "Disconnect a provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid provider ID: %s", args[0])
			}

			ctx := context.Background()
			if err := apiClient.Providers().Delete(ctx, id); err != nil {
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
			providers, err := apiClient.Providers().List(ctx, nil)
			if err != nil {
				return fmt.Errorf("failed to get provider status: %w", err)
			}

			format := getOutputFormat()
			if format != "table" {
				return printOutput(providers)
			}

			t := NewTable("ID", "TYPE", "STATUS", "LAST SYNCED")
			for _, p := range providers {
				lastSync := "never"
				if p.LastSyncedAt != nil {
					lastSync = p.LastSyncedAt.Format("2006-01-02 15:04:05")
				}
				t.AddRow(
					strconv.FormatInt(p.ID, 10),
					p.ProviderType,
					formatStatus(p.Status),
					lastSync,
				)
			}
			t.Render()
			return nil
		},
	}
}
