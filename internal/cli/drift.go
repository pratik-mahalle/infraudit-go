package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pratik-mahalle/infraudit/pkg/client"
	"github.com/spf13/cobra"
)

func newDriftCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drift",
		Short: "Manage security drifts",
	}

	cmd.AddCommand(newDriftListCmd())
	cmd.AddCommand(newDriftGetCmd())
	cmd.AddCommand(newDriftDetectCmd())
	cmd.AddCommand(newDriftSummaryCmd())
	cmd.AddCommand(newDriftResolveCmd())

	return cmd
}

func newDriftListCmd() *cobra.Command {
	var severity, status, driftType string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List drifts",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			opts := &client.DriftListOptions{}
			if severity != "" {
				opts.Severity = &severity
			}
			if status != "" {
				opts.Status = &status
			}
			if driftType != "" {
				opts.DriftType = &driftType
			}

			drifts, err := apiClient.Drifts().List(ctx, opts)
			if err != nil {
				return fmt.Errorf("failed to list drifts: %w", err)
			}

			format := getOutputFormat()
			if format != "table" {
				return printOutput(drifts)
			}

			t := NewTable("ID", "TYPE", "SEVERITY", "STATUS", "RESOURCE", "DESCRIPTION")
			for _, d := range drifts {
				t.AddRow(
					strconv.FormatInt(d.ID, 10),
					d.DriftType,
					formatSeverity(d.Severity),
					formatStatus(d.Status),
					strconv.FormatInt(d.ResourceID, 10),
					truncate(d.Description, 40),
				)
			}
			t.Render()
			return nil
		},
	}

	cmd.Flags().StringVar(&severity, "severity", "", "filter by severity")
	cmd.Flags().StringVar(&status, "status", "", "filter by status")
	cmd.Flags().StringVar(&driftType, "type", "", "filter by drift type")

	return cmd
}

func newDriftGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get drift details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid drift ID: %s", args[0])
			}

			ctx := context.Background()
			drift, err := apiClient.Drifts().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("failed to get drift: %w", err)
			}

			format := getOutputFormat()
			if format != "table" {
				return printOutput(drift)
			}

			fmt.Printf("ID:          %d\n", drift.ID)
			fmt.Printf("Type:        %s\n", drift.DriftType)
			fmt.Printf("Severity:    %s\n", formatSeverity(drift.Severity))
			fmt.Printf("Status:      %s\n", drift.Status)
			fmt.Printf("Resource:    %d\n", drift.ResourceID)
			fmt.Printf("Description: %s\n", drift.Description)
			fmt.Printf("Detected:    %s\n", drift.DetectedAt.Format("2006-01-02 15:04:05"))
			return nil
		},
	}
}

func newDriftDetectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "detect",
		Short: "Trigger drift detection",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			fmt.Println("Starting drift detection...")

			var result interface{}
			err := apiClient.DoRaw(ctx, "POST", "/api/v1/drifts/detect", nil, &result)
			if err != nil {
				return fmt.Errorf("drift detection failed: %w", err)
			}

			fmt.Println("Drift detection completed")
			if result != nil {
				return printOutput(result)
			}
			return nil
		},
	}
}

func newDriftSummaryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "summary",
		Short: "Show drift summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			var summary interface{}
			err := apiClient.DoRaw(ctx, "GET", "/api/v1/drifts/summary", nil, &summary)
			if err != nil {
				return fmt.Errorf("failed to get drift summary: %w", err)
			}

			return printOutput(summary)
		},
	}
}

func newDriftResolveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resolve <id>",
		Short: "Mark drift as resolved",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid drift ID: %s", args[0])
			}

			ctx := context.Background()
			if _, err := apiClient.Drifts().Resolve(ctx, id); err != nil {
				return fmt.Errorf("failed to resolve drift: %w", err)
			}

			fmt.Printf("Drift %d marked as resolved\n", id)
			return nil
		},
	}
}
