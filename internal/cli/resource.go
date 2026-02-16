package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pratik-mahalle/infraudit/pkg/client"
	"github.com/spf13/cobra"
)

func newResourceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resource",
		Short: "Manage cloud resources",
	}

	cmd.AddCommand(newResourceListCmd())
	cmd.AddCommand(newResourceGetCmd())
	cmd.AddCommand(newResourceDeleteCmd())

	return cmd
}

func newResourceListCmd() *cobra.Command {
	var provider, resourceType, region, status string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			opts := &client.ResourceListOptions{}
			if provider != "" {
				id, err := strconv.ParseInt(provider, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid provider ID: %s", provider)
				}
				opts.ProviderID = &id
			}
			if resourceType != "" {
				opts.ResourceType = &resourceType
			}
			if region != "" {
				opts.Region = &region
			}
			if status != "" {
				opts.Status = &status
			}

			resources, err := apiClient.Resources().List(ctx, opts)
			if err != nil {
				return fmt.Errorf("failed to list resources: %w", err)
			}

			format := getOutputFormat()
			if format != "table" {
				return printOutput(resources)
			}

			t := NewTable("ID", "NAME", "TYPE", "REGION", "STATUS", "PROVIDER")
			for _, r := range resources {
				t.AddRow(
					strconv.FormatInt(r.ID, 10),
					truncate(r.Name, 30),
					r.ResourceType,
					r.Region,
					formatStatus(r.Status),
					strconv.FormatInt(r.ProviderID, 10),
				)
			}
			t.Render()
			return nil
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "", "filter by provider ID")
	cmd.Flags().StringVar(&resourceType, "type", "", "filter by resource type")
	cmd.Flags().StringVar(&region, "region", "", "filter by region")
	cmd.Flags().StringVar(&status, "status", "", "filter by status")

	return cmd
}

func newResourceGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get resource details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid resource ID: %s", args[0])
			}

			ctx := context.Background()
			resource, err := apiClient.Resources().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("failed to get resource: %w", err)
			}

			format := getOutputFormat()
			if format != "table" {
				return printOutput(resource)
			}

			fmt.Printf("ID:       %d\n", resource.ID)
			fmt.Printf("Name:     %s\n", resource.Name)
			fmt.Printf("Type:     %s\n", resource.ResourceType)
			fmt.Printf("Region:   %s\n", resource.Region)
			fmt.Printf("Status:   %s\n", resource.Status)
			fmt.Printf("Provider: %d\n", resource.ProviderID)
			if len(resource.Tags) > 0 {
				fmt.Println("Tags:")
				for k, v := range resource.Tags {
					fmt.Printf("  %s: %s\n", k, v)
				}
			}
			return nil
		},
	}
}

func newResourceDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid resource ID: %s", args[0])
			}

			ctx := context.Background()
			if err := apiClient.Resources().Delete(ctx, id); err != nil {
				return fmt.Errorf("failed to delete resource: %w", err)
			}

			fmt.Println("Resource deleted successfully")
			return nil
		},
	}
}
