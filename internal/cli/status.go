package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show dashboard summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			format := getOutputFormat()
			if format != "table" {
				summary := map[string]interface{}{}

				providers, err := apiClient.Providers().List(ctx, nil)
				if err == nil {
					summary["providers"] = len(providers)
				}
				resources, err := apiClient.Resources().List(ctx, nil)
				if err == nil {
					summary["resources"] = len(resources)
				}
				drifts, err := apiClient.Drifts().List(ctx, nil)
				if err == nil {
					summary["drifts"] = len(drifts)
				}
				alerts, err := apiClient.Alerts().List(ctx, nil)
				if err == nil {
					summary["alerts"] = len(alerts)
				}
				return printOutput(summary)
			}

			fmt.Println("InfraAudit Dashboard")
			fmt.Println(strings.Repeat("=", 40))

			// Providers
			providers, err := apiClient.Providers().List(ctx, nil)
			if err != nil {
				fmt.Printf("  Providers:     (error: %v)\n", err)
			} else {
				connected := 0
				for _, p := range providers {
					if p.Status == "connected" {
						connected++
					}
				}
				fmt.Printf("  Providers:     %d connected (%d total)\n", connected, len(providers))
			}

			// Resources
			resources, err := apiClient.Resources().List(ctx, nil)
			if err != nil {
				fmt.Printf("  Resources:     (error: %v)\n", err)
			} else {
				fmt.Printf("  Resources:     %d synced\n", len(resources))
			}

			// Drifts
			drifts, err := apiClient.Drifts().List(ctx, nil)
			if err != nil {
				fmt.Printf("  Drifts:        (error: %v)\n", err)
			} else {
				critical := 0
				for _, d := range drifts {
					if d.Severity == "critical" {
						critical++
					}
				}
				fmt.Printf("  Drifts:        %d detected", len(drifts))
				if critical > 0 {
					fmt.Printf(" (%d critical)", critical)
				}
				fmt.Println()
			}

			// Alerts
			alerts, err := apiClient.Alerts().List(ctx, nil)
			if err != nil {
				fmt.Printf("  Alerts:        (error: %v)\n", err)
			} else {
				high := 0
				for _, a := range alerts {
					if a.Severity == "high" || a.Severity == "critical" {
						high++
					}
				}
				fmt.Printf("  Alerts:        %d open", len(alerts))
				if high > 0 {
					fmt.Printf(" (%d high severity)", high)
				}
				fmt.Println()
			}

			return nil
		},
	}
}
