package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newIaCCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iac",
		Short: "Infrastructure as Code management",
	}

	cmd.AddCommand(newIaCUploadCmd())
	cmd.AddCommand(newIaCDefinitionsCmd())
	cmd.AddCommand(newIaCDetectDriftCmd())
	cmd.AddCommand(newIaCDriftsCmd())
	cmd.AddCommand(newIaCDriftSummaryCmd())

	return cmd
}

func newIaCUploadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upload <file>",
		Short: "Upload IaC file (Terraform/CloudFormation/K8s)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			ctx := context.Background()
			body := map[string]interface{}{
				"filename": filepath.Base(filePath),
				"content":  string(content),
			}

			var result interface{}
			if err := apiClient.DoRaw(ctx, "POST", "/api/v1/iac/upload", body, &result); err != nil {
				return fmt.Errorf("failed to upload IaC file: %w", err)
			}

			fmt.Printf("Uploaded %s successfully\n", filepath.Base(filePath))
			if result != nil {
				return printOutput(result)
			}
			return nil
		},
	}
}

func newIaCDefinitionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "definitions",
		Short: "List IaC definitions",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/iac/definitions", nil, &result); err != nil {
				return fmt.Errorf("failed to list definitions: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newIaCDetectDriftCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "detect-drift",
		Short: "Detect IaC drift",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			fmt.Println("Detecting IaC drift...")

			var result interface{}
			if err := apiClient.DoRaw(ctx, "POST", "/api/v1/iac/drifts/detect", nil, &result); err != nil {
				return fmt.Errorf("IaC drift detection failed: %w", err)
			}
			fmt.Println("IaC drift detection completed")
			if result != nil {
				return printOutput(result)
			}
			return nil
		},
	}
}

func newIaCDriftsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "drifts",
		Short: "List IaC drifts",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/iac/drifts", nil, &result); err != nil {
				return fmt.Errorf("failed to list IaC drifts: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newIaCDriftSummaryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "drift-summary",
		Short: "Show IaC drift summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/iac/drifts/summary", nil, &result); err != nil {
				return fmt.Errorf("failed to get IaC drift summary: %w", err)
			}
			return printOutput(result)
		},
	}
}
