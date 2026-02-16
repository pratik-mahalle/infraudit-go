package cli

import (
	"context"
	"fmt"

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

			// Build query parameters
			path := "/api/resources?"
			params := []string{}
			if provider != "" {
				params = append(params, "provider="+provider)
			}
			if resourceType != "" {
				params = append(params, "type="+resourceType)
			}
			if region != "" {
				params = append(params, "region="+region)
			}
			if status != "" {
				params = append(params, "status="+status)
			}
			
			for i, p := range params {
				if i > 0 {
					path += "&"
				}
				path += p
			}

			// Use DoRaw to get paginated response
			var response struct {
				Data []struct {
					ID            int64             `json:"id"`
					ResourceID    string            `json:"resourceId,omitempty"`
					Provider      string            `json:"provider"`
					Name          string            `json:"name"`
					Type          string            `json:"type"`
					Region        string            `json:"region"`
					Status        string            `json:"status"`
					Tags          map[string]string `json:"tags,omitempty"`
					Cost          float64           `json:"cost"`
				} `json:"data"`
				Page     int `json:"page"`
				PageSize int `json:"pageSize"`
				Total    int `json:"total"`
			}
			
			if err := apiClient.DoRaw(ctx, "GET", path, nil, &response); err != nil {
				return fmt.Errorf("failed to list resources: %w", err)
			}

			format := getOutputFormat()
			if format != "table" {
				return printOutput(response.Data)
			}

			t := NewTable("RESOURCE ID", "NAME", "TYPE", "REGION", "STATUS", "PROVIDER")
			for _, r := range response.Data {
				t.AddRow(
					truncate(r.ResourceID, 20),
					truncate(r.Name, 30),
					r.Type,
					r.Region,
					formatStatus(r.Status),
					r.Provider,
				)
			}
			t.Render()
			fmt.Printf("\nShowing %d of %d resources\n", len(response.Data), response.Total)
			return nil
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "", "filter by provider type (aws, gcp, azure)")
	cmd.Flags().StringVar(&resourceType, "type", "", "filter by resource type")
	cmd.Flags().StringVar(&region, "region", "", "filter by region")
	cmd.Flags().StringVar(&status, "status", "", "filter by status")

	return cmd
}

func newResourceGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <resource-id>",
		Short: "Get resource details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceID := args[0]

			ctx := context.Background()
			
			var resource struct {
				ID            int64             `json:"id"`
				ResourceID    string            `json:"resourceId,omitempty"`
				Provider      string            `json:"provider"`
				Name          string            `json:"name"`
				Type          string            `json:"type"`
				Region        string            `json:"region"`
				Status        string            `json:"status"`
				Tags          map[string]string `json:"tags,omitempty"`
				Cost          float64           `json:"cost"`
				Configuration string            `json:"configuration,omitempty"`
			}
			
			path := fmt.Sprintf("/api/resources/%s", resourceID)
			if err := apiClient.DoRaw(ctx, "GET", path, nil, &resource); err != nil {
				return fmt.Errorf("failed to get resource: %w", err)
			}

			format := getOutputFormat()
			if format != "table" {
				return printOutput(resource)
			}

			fmt.Printf("Resource ID: %s\n", resource.ResourceID)
			fmt.Printf("Name:        %s\n", resource.Name)
			fmt.Printf("Type:        %s\n", resource.Type)
			fmt.Printf("Region:      %s\n", resource.Region)
			fmt.Printf("Status:      %s\n", resource.Status)
			fmt.Printf("Provider:    %s\n", resource.Provider)
			fmt.Printf("Cost:        $%.2f\n", resource.Cost)
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
		Use:   "delete <resource-id>",
		Short: "Delete a resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceID := args[0]

			ctx := context.Background()
			path := fmt.Sprintf("/api/resources/%s", resourceID)
			if err := apiClient.DoRaw(ctx, "DELETE", path, nil, nil); err != nil {
				return fmt.Errorf("failed to delete resource: %w", err)
			}

			fmt.Println("Resource deleted successfully")
			return nil
		},
	}
}
