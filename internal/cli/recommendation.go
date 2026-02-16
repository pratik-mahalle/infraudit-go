package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pratik-mahalle/infraudit/pkg/client"
	"github.com/spf13/cobra"
)

func newRecommendationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "recommendation",
		Aliases: []string{"rec"},
		Short:   "Manage recommendations",
	}

	cmd.AddCommand(newRecListCmd())
	cmd.AddCommand(newRecGetCmd())
	cmd.AddCommand(newRecGenerateCmd())
	cmd.AddCommand(newRecSavingsCmd())
	cmd.AddCommand(newRecApplyCmd())
	cmd.AddCommand(newRecDismissCmd())

	return cmd
}

func newRecListCmd() *cobra.Command {
	var recType, impact, status string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recommendations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			opts := &client.RecommendationListOptions{}
			if recType != "" {
				opts.Type = &recType
			}
			if impact != "" {
				opts.Impact = &impact
			}
			if status != "" {
				opts.Status = &status
			}

			recs, err := apiClient.Recommendations().List(ctx, opts)
			if err != nil {
				return fmt.Errorf("failed to list recommendations: %w", err)
			}

			format := getOutputFormat()
			if format != "table" {
				return printOutput(recs)
			}

			t := NewTable("ID", "TYPE", "IMPACT", "STATUS", "SAVINGS", "TITLE")
			for _, r := range recs {
				savings := ""
				if r.EstimatedSavings > 0 {
					savings = fmt.Sprintf("$%.2f", r.EstimatedSavings)
				}
				t.AddRow(
					strconv.FormatInt(r.ID, 10),
					r.Type,
					r.Impact,
					formatStatus(r.Status),
					savings,
					truncate(r.Title, 40),
				)
			}
			t.Render()
			return nil
		},
	}

	cmd.Flags().StringVar(&recType, "type", "", "filter by type (cost, performance, security)")
	cmd.Flags().StringVar(&impact, "impact", "", "filter by impact (high, medium, low)")
	cmd.Flags().StringVar(&status, "status", "", "filter by status")

	return cmd
}

func newRecGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get recommendation details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid recommendation ID: %s", args[0])
			}

			ctx := context.Background()
			rec, err := apiClient.Recommendations().Get(ctx, id)
			if err != nil {
				return fmt.Errorf("failed to get recommendation: %w", err)
			}

			format := getOutputFormat()
			if format != "table" {
				return printOutput(rec)
			}

			fmt.Printf("ID:          %d\n", rec.ID)
			fmt.Printf("Type:        %s\n", rec.Type)
			fmt.Printf("Impact:      %s\n", rec.Impact)
			fmt.Printf("Status:      %s\n", rec.Status)
			fmt.Printf("Title:       %s\n", rec.Title)
			fmt.Printf("Description: %s\n", rec.Description)
			if rec.EstimatedSavings > 0 {
				fmt.Printf("Savings:     $%.2f\n", rec.EstimatedSavings)
			}
			return nil
		},
	}
}

func newRecGenerateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "generate",
		Short: "Generate AI recommendations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			fmt.Println("Generating recommendations...")

			var result interface{}
			if err := apiClient.DoRaw(ctx, "POST", "/api/v1/recommendations/generate", nil, &result); err != nil {
				return fmt.Errorf("failed to generate recommendations: %w", err)
			}
			fmt.Println("Recommendations generated")
			if result != nil {
				return printOutput(result)
			}
			return nil
		},
	}
}

func newRecSavingsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "savings",
		Short: "Show potential savings",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/recommendations/savings", nil, &result); err != nil {
				return fmt.Errorf("failed to get savings: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newRecApplyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "apply <id>",
		Short: "Mark recommendation as applied",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid recommendation ID: %s", args[0])
			}

			ctx := context.Background()
			if _, err := apiClient.Recommendations().Apply(ctx, id); err != nil {
				return fmt.Errorf("failed to apply recommendation: %w", err)
			}
			fmt.Printf("Recommendation %d marked as applied\n", id)
			return nil
		},
	}
}

func newRecDismissCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dismiss <id>",
		Short: "Dismiss a recommendation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid recommendation ID: %s", args[0])
			}

			ctx := context.Background()
			if _, err := apiClient.Recommendations().Dismiss(ctx, id); err != nil {
				return fmt.Errorf("failed to dismiss recommendation: %w", err)
			}
			fmt.Printf("Recommendation %d dismissed\n", id)
			return nil
		},
	}
}
