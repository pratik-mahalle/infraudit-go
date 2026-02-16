package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newJobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Manage scheduled jobs",
	}

	cmd.AddCommand(newJobListCmd())
	cmd.AddCommand(newJobGetCmd())
	cmd.AddCommand(newJobCreateCmd())
	cmd.AddCommand(newJobDeleteCmd())
	cmd.AddCommand(newJobRunCmd())
	cmd.AddCommand(newJobExecutionsCmd())
	cmd.AddCommand(newJobTypesCmd())

	return cmd
}

func newJobListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List scheduled jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/jobs", nil, &result); err != nil {
				return fmt.Errorf("failed to list jobs: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newJobGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get job details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/jobs/"+args[0], nil, &result); err != nil {
				return fmt.Errorf("failed to get job: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newJobCreateCmd() *cobra.Command {
	var name, jobType, schedule string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a scheduled job",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				name = promptInput("Job name: ")
			}
			if jobType == "" {
				jobType = promptInput("Job type: ")
			}
			if schedule == "" {
				schedule = promptInput("Cron schedule: ")
			}

			ctx := context.Background()
			body := map[string]interface{}{
				"name":     name,
				"type":     jobType,
				"schedule": schedule,
			}

			var result interface{}
			if err := apiClient.DoRaw(ctx, "POST", "/api/v1/jobs", body, &result); err != nil {
				return fmt.Errorf("failed to create job: %w", err)
			}
			fmt.Printf("Job '%s' created\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "job name")
	cmd.Flags().StringVar(&jobType, "type", "", "job type")
	cmd.Flags().StringVar(&schedule, "schedule", "", "cron schedule")

	return cmd
}

func newJobDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			if err := apiClient.DoRaw(ctx, "DELETE", "/api/v1/jobs/"+args[0], nil, nil); err != nil {
				return fmt.Errorf("failed to delete job: %w", err)
			}
			fmt.Printf("Job %s deleted\n", args[0])
			return nil
		},
	}
}

func newJobRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run <id>",
		Short: "Trigger job execution",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			fmt.Println("Triggering job execution...")

			var result interface{}
			if err := apiClient.DoRaw(ctx, "POST", "/api/v1/jobs/"+args[0]+"/run", nil, &result); err != nil {
				return fmt.Errorf("failed to run job: %w", err)
			}
			fmt.Println("Job execution started")
			if result != nil {
				return printOutput(result)
			}
			return nil
		},
	}
}

func newJobExecutionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "executions <job-id>",
		Short: "List job executions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/jobs/"+args[0]+"/executions", nil, &result); err != nil {
				return fmt.Errorf("failed to list executions: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newJobTypesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "types",
		Short: "List available job types",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/jobs/types", nil, &result); err != nil {
				return fmt.Errorf("failed to list job types: %w", err)
			}
			return printOutput(result)
		},
	}
}
