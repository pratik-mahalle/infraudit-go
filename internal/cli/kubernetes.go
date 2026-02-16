package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newKubernetesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kubernetes",
		Aliases: []string{"k8s"},
		Short:   "Kubernetes cluster management",
	}

	cmd.AddCommand(newK8sClustersCmd())
	cmd.AddCommand(newK8sRegisterCmd())
	cmd.AddCommand(newK8sDeleteCmd())
	cmd.AddCommand(newK8sSyncCmd())
	cmd.AddCommand(newK8sNamespacesCmd())
	cmd.AddCommand(newK8sDeploymentsCmd())
	cmd.AddCommand(newK8sPodsCmd())
	cmd.AddCommand(newK8sServicesCmd())
	cmd.AddCommand(newK8sStatsCmd())

	return cmd
}

func newK8sClustersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clusters",
		Short: "List Kubernetes clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/kubernetes/clusters", nil, &result); err != nil {
				return fmt.Errorf("failed to list clusters: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newK8sRegisterCmd() *cobra.Command {
	var name, kubeconfig string

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				name = promptInput("Cluster name: ")
			}

			ctx := context.Background()
			body := map[string]interface{}{
				"name": name,
			}
			if kubeconfig != "" {
				body["kubeconfig"] = kubeconfig
			}

			var result interface{}
			if err := apiClient.DoRaw(ctx, "POST", "/api/v1/kubernetes/clusters", body, &result); err != nil {
				return fmt.Errorf("failed to register cluster: %w", err)
			}
			fmt.Printf("Cluster '%s' registered\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "cluster name")
	cmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")

	return cmd
}

func newK8sDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			if err := apiClient.DoRaw(ctx, "DELETE", "/api/v1/kubernetes/clusters/"+args[0], nil, nil); err != nil {
				return fmt.Errorf("failed to delete cluster: %w", err)
			}
			fmt.Printf("Cluster %s deleted\n", args[0])
			return nil
		},
	}
}

func newK8sSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync <id>",
		Short: "Sync cluster resources",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			fmt.Println("Syncing cluster...")
			var result interface{}
			if err := apiClient.DoRaw(ctx, "POST", "/api/v1/kubernetes/clusters/"+args[0]+"/sync", nil, &result); err != nil {
				return fmt.Errorf("failed to sync cluster: %w", err)
			}
			fmt.Println("Cluster sync completed")
			return nil
		},
	}
}

func newK8sNamespacesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "namespaces <cluster-id>",
		Short: "List namespaces in a cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/kubernetes/clusters/"+args[0]+"/namespaces", nil, &result); err != nil {
				return fmt.Errorf("failed to list namespaces: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newK8sDeploymentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "deployments <cluster-id>",
		Short: "List deployments in a cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/kubernetes/clusters/"+args[0]+"/deployments", nil, &result); err != nil {
				return fmt.Errorf("failed to list deployments: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newK8sPodsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pods <cluster-id>",
		Short: "List pods in a cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/kubernetes/clusters/"+args[0]+"/pods", nil, &result); err != nil {
				return fmt.Errorf("failed to list pods: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newK8sServicesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "services <cluster-id>",
		Short: "List services in a cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/kubernetes/clusters/"+args[0]+"/services", nil, &result); err != nil {
				return fmt.Errorf("failed to list services: %w", err)
			}
			return printOutput(result)
		},
	}
}

func newK8sStatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show cluster statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			var result interface{}
			if err := apiClient.DoRaw(ctx, "GET", "/api/v1/kubernetes/stats", nil, &result); err != nil {
				return fmt.Errorf("failed to get cluster stats: %w", err)
			}
			return printOutput(result)
		},
	}
}
