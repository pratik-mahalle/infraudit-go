package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newCostCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cost",
		Short: "Cloud cost analytics",
	}

	cmd.AddCommand(newCostOverviewCmd())
	cmd.AddCommand(newCostTrendsCmd())
	cmd.AddCommand(newCostForecastCmd())
	cmd.AddCommand(newCostSyncCmd())
	cmd.AddCommand(newCostAnomaliesCmd())
	cmd.AddCommand(newCostDetectAnomaliesCmd())
	cmd.AddCommand(newCostOptimizationsCmd())
	cmd.AddCommand(newCostSavingsCmd())

	return cmd
}

func newCostOverviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "overview",
		Short: "Show cost overview",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/costs", nil, &result); err != nil {
				return fmt.Errorf("failed to get cost overview: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newCostTrendsCmd() *cobra.Command {
	var provider, period string

	cmd := &cobra.Command{
		Use:   "trends",
		Short: "Show cost trends",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			path := "/api/v1/costs/trends"
			params := buildQueryParams(map[string]string{
				"provider": provider,
				"period":   period,
			})
			if params != "" {
				path += "?" + params
			}

			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", path, nil, &result); err != nil {
				return fmt.Errorf("failed to get cost trends: %w", err)
			}
			return printOutput(result)
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "", "filter by provider")
	cmd.Flags().StringVar(&period, "period", "", "time period (7d, 30d, 90d)")

	return cmd
}

func newCostForecastCmd() *cobra.Command {
	var provider, days string

	cmd := &cobra.Command{
		Use:   "forecast",
		Short: "Show cost forecast",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			path := "/api/v1/costs/forecast"
			params := buildQueryParams(map[string]string{
				"provider": provider,
				"days":     days,
			})
			if params != "" {
				path += "?" + params
			}

			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", path, nil, &result); err != nil {
				return fmt.Errorf("failed to get cost forecast: %w", err)
			}
			return printOutput(result)
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "", "filter by provider")
	cmd.Flags().StringVar(&days, "days", "", "forecast days (30, 60, 90)")

	return cmd
}

func newCostSyncCmd() *cobra.Command {
	var provider string

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync cost data",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			fmt.Println("Syncing cost data...")

			body := map[string]interface{}{}
			if provider != "" {
				body["provider"] = provider
			}

			var result interface{}
			if err := apiClient.DoRaw(ctx, "POST", "/api/v1/costs/sync", body, &result); err != nil {
				return fmt.Errorf("failed to sync costs: %w", err)
			}
			fmt.Println("Cost sync completed")
			if result != nil {
				return printOutput(result)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "", "sync specific provider")

	return cmd
}

func newCostAnomaliesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "anomalies",
		Short: "List cost anomalies",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/costs/anomalies", nil, &result); err != nil {
				return fmt.Errorf("failed to list cost anomalies: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newCostDetectAnomaliesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "detect-anomalies",
		Short: "Trigger anomaly detection",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			fmt.Println("Running anomaly detection...")

			var result interface{}
			if err := apiClient.DoRaw(ctx, "POST", "/api/v1/costs/anomalies/detect", nil, &result); err != nil {
				return fmt.Errorf("anomaly detection failed: %w", err)
			}
			fmt.Println("Anomaly detection completed")
			if result != nil {
				return printOutput(result)
			}
			return nil
		},
	}
}

func newCostOptimizationsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "optimizations",
		Short: "List cost optimizations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/costs/optimizations", nil, &result); err != nil {
				return fmt.Errorf("failed to list optimizations: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newCostSavingsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "savings",
		Short: "Show potential savings",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/costs/savings", nil, &result); err != nil {
				return fmt.Errorf("failed to get savings: %w", err)
			}
			return printOutput(result)
		},
	}
}

// buildQueryParams builds a URL query string from a map, skipping empty values.
func buildQueryParams(params map[string]string) string {
	result := ""
	for k, v := range params {
		if v == "" {
			continue
		}
		if result != "" {
			result += "&"
		}
		result += k + "=" + v
	}
	return result
}
