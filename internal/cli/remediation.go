package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newRemediationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remediation",
		Short: "Manage remediation actions",
	}

	cmd.AddCommand(newRemediationSummaryCmd())
	cmd.AddCommand(newRemediationPendingCmd())
	cmd.AddCommand(newRemediationSuggestDriftCmd())
	cmd.AddCommand(newRemediationSuggestVulnCmd())
	cmd.AddCommand(newRemediationApproveCmd())
	cmd.AddCommand(newRemediationExecuteCmd())
	cmd.AddCommand(newRemediationRollbackCmd())

	return cmd
}

func newRemediationSummaryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "summary",
		Short: "Show remediation summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/remediation/summary", nil, &result); err != nil {
				return fmt.Errorf("failed to get remediation summary: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newRemediationPendingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pending",
		Short: "List pending approvals",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/remediation/pending", nil, &result); err != nil {
				return fmt.Errorf("failed to list pending approvals: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newRemediationSuggestDriftCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "suggest-drift <drift-id>",
		Short: "Suggest remediation for a drift",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			fmt.Println("Generating remediation suggestion...")

			var result interface{}
			if err := apiClient.DoRaw(ctx, "POST", "/api/v1/remediation/suggest/drift/"+args[0], nil, &result); err != nil {
				return fmt.Errorf("failed to suggest remediation: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newRemediationSuggestVulnCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "suggest-vuln <vulnerability-id>",
		Short: "Suggest remediation for a vulnerability",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			fmt.Println("Generating remediation suggestion...")

			var result interface{}
			if err := apiClient.DoRaw(ctx, "POST", "/api/v1/remediation/suggest/vulnerability/"+args[0], nil, &result); err != nil {
				return fmt.Errorf("failed to suggest remediation: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newRemediationApproveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "approve <action-id>",
		Short: "Approve a remediation action",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			if err := apiClient.DoRaw(ctx, "POST", "/api/v1/remediation/actions/"+args[0]+"/approve", nil, nil); err != nil {
				return fmt.Errorf("failed to approve action: %w", err)
			}
			fmt.Printf("Remediation action %s approved\n", args[0])
			return nil
		},
	}
}

func newRemediationExecuteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "execute <action-id>",
		Short: "Execute a remediation action",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			fmt.Println("Executing remediation action...")

			var result interface{}
			if err := apiClient.DoRaw(ctx, "POST", "/api/v1/remediation/actions/"+args[0]+"/execute", nil, &result); err != nil {
				return fmt.Errorf("failed to execute action: %w", err)
			}
			fmt.Println("Remediation action executed")
			if result != nil {
				return printOutput(result)
			}
			return nil
		},
	}
}

func newRemediationRollbackCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rollback <action-id>",
		Short: "Rollback a remediation action",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			fmt.Println("Rolling back remediation action...")

			var result interface{}
			if err := apiClient.DoRaw(ctx, "POST", "/api/v1/remediation/actions/"+args[0]+"/rollback", nil, &result); err != nil {
				return fmt.Errorf("failed to rollback action: %w", err)
			}
			fmt.Println("Remediation action rolled back")
			if result != nil {
				return printOutput(result)
			}
			return nil
		},
	}
}
