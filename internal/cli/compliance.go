package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newComplianceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compliance",
		Short: "Compliance framework management",
	}

	cmd.AddCommand(newComplianceOverviewCmd())
	cmd.AddCommand(newComplianceFrameworksCmd())
	cmd.AddCommand(newComplianceFrameworkCmd())
	cmd.AddCommand(newComplianceEnableCmd())
	cmd.AddCommand(newComplianceDisableCmd())
	cmd.AddCommand(newComplianceAssessCmd())
	cmd.AddCommand(newComplianceAssessmentsCmd())
	cmd.AddCommand(newComplianceExportCmd())
	cmd.AddCommand(newComplianceFailingControlsCmd())

	return cmd
}

func newComplianceOverviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "overview",
		Short: "Show compliance overview",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/compliance/overview", nil, &result); err != nil {
				return fmt.Errorf("failed to get compliance overview: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newComplianceFrameworksCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "frameworks",
		Short: "List compliance frameworks",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/compliance/frameworks", nil, &result); err != nil {
				return fmt.Errorf("failed to list frameworks: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newComplianceFrameworkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "framework <id>",
		Short: "Show framework details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/compliance/frameworks/"+args[0], nil, &result); err != nil {
				return fmt.Errorf("failed to get framework: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newComplianceEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable <id>",
		Short: "Enable a compliance framework",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			if err := apiClient.DoRaw(ctx, "POST", "/api/v1/compliance/frameworks/"+args[0]+"/enable", nil, nil); err != nil {
				return fmt.Errorf("failed to enable framework: %w", err)
			}
			fmt.Printf("Framework %s enabled\n", args[0])
			return nil
		},
	}
}

func newComplianceDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable <id>",
		Short: "Disable a compliance framework",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			if err := apiClient.DoRaw(ctx, "POST", "/api/v1/compliance/frameworks/"+args[0]+"/disable", nil, nil); err != nil {
				return fmt.Errorf("failed to disable framework: %w", err)
			}
			fmt.Printf("Framework %s disabled\n", args[0])
			return nil
		},
	}
}

func newComplianceAssessCmd() *cobra.Command {
	var framework string

	cmd := &cobra.Command{
		Use:   "assess",
		Short: "Run compliance assessment",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			fmt.Println("Running compliance assessment...")

			body := map[string]interface{}{}
			if framework != "" {
				body["framework_id"] = framework
			}

			var result interface{}
			if err := apiClient.DoRaw(ctx, "POST", "/api/v1/compliance/assess", body, &result); err != nil {
				return fmt.Errorf("assessment failed: %w", err)
			}
			fmt.Println("Assessment completed")
			if result != nil {
				return printOutput(result)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&framework, "framework", "", "framework ID to assess")

	return cmd
}

func newComplianceAssessmentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "assessments",
		Short: "List past assessments",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/compliance/assessments", nil, &result); err != nil {
				return fmt.Errorf("failed to list assessments: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newComplianceExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export <assessment-id>",
		Short: "Export assessment report",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/compliance/assessments/"+args[0]+"/export", nil, &result); err != nil {
				return fmt.Errorf("failed to export assessment: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newComplianceFailingControlsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "failing-controls",
		Short: "Show failing controls",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/compliance/controls/failing", nil, &result); err != nil {
				return fmt.Errorf("failed to get failing controls: %w", err)
			}
			return printOutput(result)
		},
	}
}
