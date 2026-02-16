package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newWebhookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "Manage webhooks",
	}

	cmd.AddCommand(newWebhookListCmd())
	cmd.AddCommand(newWebhookCreateCmd())
	cmd.AddCommand(newWebhookGetCmd())
	cmd.AddCommand(newWebhookDeleteCmd())
	cmd.AddCommand(newWebhookTestCmd())
	cmd.AddCommand(newWebhookEventsCmd())

	return cmd
}

func newWebhookListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List webhooks",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/webhooks", nil, &result); err != nil {
				return fmt.Errorf("failed to list webhooks: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newWebhookCreateCmd() *cobra.Command {
	var name, url, secret string
	var events []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a webhook",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				name = promptInput("Webhook name: ")
			}
			if url == "" {
				url = promptInput("Webhook URL: ")
			}

			ctx := context.Background()
			body := map[string]interface{}{
				"name": name,
				"url":  url,
			}
			if secret != "" {
				body["secret"] = secret
			}
			if len(events) > 0 {
				body["events"] = events
			}

			var result interface{}
			if err := apiClient.DoRaw(ctx, "POST", "/api/v1/webhooks", body, &result); err != nil {
				return fmt.Errorf("failed to create webhook: %w", err)
			}
			fmt.Printf("Webhook '%s' created\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "webhook name")
	cmd.Flags().StringVar(&url, "url", "", "webhook URL")
	cmd.Flags().StringVar(&secret, "secret", "", "webhook secret")
	cmd.Flags().StringSliceVar(&events, "events", nil, "events to subscribe to")

	return cmd
}

func newWebhookGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get webhook details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/webhooks/"+args[0], nil, &result); err != nil {
				return fmt.Errorf("failed to get webhook: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newWebhookDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a webhook",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			if err := apiClient.DoRaw(ctx, "DELETE", "/api/v1/webhooks/"+args[0], nil, nil); err != nil {
				return fmt.Errorf("failed to delete webhook: %w", err)
			}
			fmt.Printf("Webhook %s deleted\n", args[0])
			return nil
		},
	}
}

func newWebhookTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test <id>",
		Short: "Test a webhook",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			fmt.Println("Testing webhook...")

			var result interface{}
			if err := apiClient.DoRaw(ctx, "POST", "/api/v1/webhooks/"+args[0]+"/test", nil, &result); err != nil {
				return fmt.Errorf("webhook test failed: %w", err)
			}
			fmt.Println("Webhook test successful")
			if result != nil {
				return printOutput(result)
			}
			return nil
		},
	}
}

func newWebhookEventsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "events",
		Short: "List available webhook events",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/webhooks/events", nil, &result); err != nil {
				return fmt.Errorf("failed to list webhook events: %w", err)
			}
			return printOutput(result)
		},
	}
}
