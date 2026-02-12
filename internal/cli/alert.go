package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pratik-mahalle/infraudit/pkg/client"
	"github.com/spf13/cobra"
)

func newAlertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alert",
		Short: "Manage alerts",
	}

	cmd.AddCommand(newAlertListCmd())
	cmd.AddCommand(newAlertGetCmd())
	cmd.AddCommand(newAlertSummaryCmd())
	cmd.AddCommand(newAlertAcknowledgeCmd())
	cmd.AddCommand(newAlertResolveCmd())

	return cmd
}

func newAlertListCmd() *cobra.Command {
	var severity, status, alertType string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List alerts",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			opts := &client.AlertListOptions{}
			if severity != "" {
				opts.Severity = &severity
			}
			if status != "" {
				opts.Status = &status
			}
			if alertType != "" {
				opts.Type = &alertType
			}

			alerts, err := apiClient.Alerts().List(ctx, opts)
			if err != nil {
				return fmt.Errorf("failed to list alerts: %w", err)
			}

			format := getOutputFormat()
			if format != "table" {
				return printOutput(alerts)
			}

			t := NewTable("ID", "TYPE", "SEVERITY", "STATUS", "TITLE")
			for _, a := range alerts {
				t.AddRow(
					strconv.FormatInt(a.ID, 10),
					a.Type,
					formatSeverity(a.Severity),
					formatStatus(a.Status),
					truncate(a.Title, 50),
				)
			}
			t.Render()
			return nil
		},
	}

	cmd.Flags().StringVar(&severity, "severity", "", "filter by severity")
	cmd.Flags().StringVar(&status, "status", "", "filter by status")
	cmd.Flags().StringVar(&alertType, "type", "", "filter by type")

	return cmd
}

func newAlertGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get alert details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid alert ID: %s", args[0])
			}

			ctx := context.Background()
			alert, err := apiClient.Alerts().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("failed to get alert: %w", err)
			}

			format := getOutputFormat()
			if format != "table" {
				return printOutput(alert)
			}

			fmt.Printf("ID:          %d\n", alert.ID)
			fmt.Printf("Type:        %s\n", alert.Type)
			fmt.Printf("Severity:    %s\n", formatSeverity(alert.Severity))
			fmt.Printf("Status:      %s\n", alert.Status)
			fmt.Printf("Title:       %s\n", alert.Title)
			fmt.Printf("Description: %s\n", alert.Description)
			fmt.Printf("Created:     %s\n", alert.CreatedAt.Format("2006-01-02 15:04:05"))
			return nil
		},
	}
}

func newAlertSummaryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "summary",
		Short: "Show alert summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			var summary interface{}
			err := apiClient.DoRaw(ctx, "GET", "/api/v1/alerts/summary", nil, &summary)
			if err != nil {
				return fmt.Errorf("failed to get alert summary: %w", err)
			}

			return printOutput(summary)
		},
	}
}

func newAlertAcknowledgeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "acknowledge <id>",
		Short: "Acknowledge an alert",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid alert ID: %s", args[0])
			}

			ctx := context.Background()
			if _, err := apiClient.Alerts().Acknowledge(ctx, id); err != nil {
				return fmt.Errorf("failed to acknowledge alert: %w", err)
			}

			fmt.Printf("Alert %d acknowledged\n", id)
			return nil
		},
	}
}

func newAlertResolveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resolve <id>",
		Short: "Resolve an alert",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid alert ID: %s", args[0])
			}

			ctx := context.Background()
			if _, err := apiClient.Alerts().Resolve(ctx, id); err != nil {
				return fmt.Errorf("failed to resolve alert: %w", err)
			}

			fmt.Printf("Alert %d resolved\n", id)
			return nil
		},
	}
}
