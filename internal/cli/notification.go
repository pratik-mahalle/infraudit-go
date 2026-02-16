package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newNotificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "notification",
		Aliases: []string{"notif"},
		Short:   "Manage notifications",
	}

	cmd.AddCommand(newNotifPreferencesCmd())
	cmd.AddCommand(newNotifUpdateCmd())
	cmd.AddCommand(newNotifHistoryCmd())
	cmd.AddCommand(newNotifSendCmd())

	return cmd
}

func newNotifPreferencesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "preferences",
		Short: "Show notification preferences",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/notifications/preferences", nil, &result); err != nil {
				return fmt.Errorf("failed to get preferences: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newNotifUpdateCmd() *cobra.Command {
	var enabled string

	cmd := &cobra.Command{
		Use:   "update <channel>",
		Short: "Update notification preference",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			channel := args[0]

			body := map[string]interface{}{}
			if enabled != "" {
				body["enabled"] = enabled == "true"
			}

			if err := apiClient.DoRaw(ctx, "PUT", "/api/v1/notifications/preferences/"+channel, body, nil); err != nil {
				return fmt.Errorf("failed to update preference: %w", err)
			}
			fmt.Printf("Notification preference for '%s' updated\n", channel)
			return nil
		},
	}

	cmd.Flags().StringVar(&enabled, "enabled", "", "enable or disable (true/false)")

	return cmd
}

func newNotifHistoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "history",
		Short: "Show notification history",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/notifications/history", nil, &result); err != nil {
				return fmt.Errorf("failed to get notification history: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newNotifSendCmd() *cobra.Command {
	var channel, message string

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send test notification",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			body := map[string]interface{}{}
			if channel != "" {
				body["channel"] = channel
			}
			if message != "" {
				body["message"] = message
			}

			if err := apiClient.DoRaw(ctx, "POST", "/api/v1/notifications/send", body, nil); err != nil {
				return fmt.Errorf("failed to send notification: %w", err)
			}
			fmt.Println("Notification sent")
			return nil
		},
	}

	cmd.Flags().StringVar(&channel, "channel", "", "notification channel")
	cmd.Flags().StringVar(&message, "message", "", "notification message")

	return cmd
}
