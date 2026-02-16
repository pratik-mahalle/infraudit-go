package cli

import (
	"fmt"
	"os"

	"github.com/pratik-mahalle/infraudit/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile      string
	outputFormat string
	noColor      bool
	serverURL    string
	apiClient    *client.Client
)

var rootCmd = &cobra.Command{
	Use:   "infraudit",
	Short: "InfraAudit CLI - Cloud Infrastructure Auditing and Security Platform",
	Long: `InfraAudit CLI provides command-line access to the InfraAudit platform
for managing cloud infrastructure, detecting security drifts, monitoring costs,
running compliance assessments, and automating remediation.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip client init for config and auth login commands
		if cmd.Name() == "init" || cmd.Name() == "set" || cmd.Name() == "get" ||
			(cmd.Parent() != nil && cmd.Parent().Name() == "config") {
			return nil
		}
		if cmd.Name() == "login" || cmd.Name() == "register" {
			return initClient()
		}
		return initAuthenticatedClient()
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default $HOME/.infraudit/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "output format: table, json, yaml")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "", "server URL (overrides config)")

	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	_ = viper.BindPFlag("server_url", rootCmd.PersistentFlags().Lookup("server"))

	// Register all subcommands
	rootCmd.AddCommand(newAuthCmd())
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newStatusCmd())
	rootCmd.AddCommand(newProviderCmd())
	rootCmd.AddCommand(newResourceCmd())
	rootCmd.AddCommand(newDriftCmd())
	rootCmd.AddCommand(newAlertCmd())
	rootCmd.AddCommand(newVulnerabilityCmd())
	rootCmd.AddCommand(newCostCmd())
	rootCmd.AddCommand(newComplianceCmd())
	rootCmd.AddCommand(newKubernetesCmd())
	rootCmd.AddCommand(newIaCCmd())
	rootCmd.AddCommand(newJobCmd())
	rootCmd.AddCommand(newRemediationCmd())
	rootCmd.AddCommand(newRecommendationCmd())
	rootCmd.AddCommand(newNotificationCmd())
	rootCmd.AddCommand(newWebhookCmd())
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return
		}
		configDir := home + "/.infraudit"
		_ = os.MkdirAll(configDir, 0700)
		viper.AddConfigPath(configDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("INFRAUDIT")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("server_url", "http://localhost:8080")
	viper.SetDefault("output", "table")

	_ = viper.ReadInConfig()
}

func initClient() error {
	url := viper.GetString("server_url")
	if serverURL != "" {
		url = serverURL
	}

	apiClient = client.NewClient(client.Config{
		BaseURL: url,
	})
	return nil
}

func initAuthenticatedClient() error {
	if err := initClient(); err != nil {
		return err
	}

	token := viper.GetString("auth.token")
	if token == "" {
		return fmt.Errorf("not authenticated. Run 'infraudit auth login' first")
	}

	apiClient.SetToken(token)
	return nil
}

func getOutputFormat() string {
	if outputFormat != "" && outputFormat != "table" {
		return outputFormat
	}
	return viper.GetString("output")
}
