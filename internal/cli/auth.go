package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/pratik-mahalle/infraudit/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
	}

	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(newAuthRegisterCmd())
	cmd.AddCommand(newAuthLogoutCmd())
	cmd.AddCommand(newAuthWhoamiCmd())

	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var email, password string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login with email and password",
		RunE: func(cmd *cobra.Command, args []string) error {
			if email == "" {
				email = promptInput("Email: ")
			}
			if password == "" {
				password = promptPassword("Password: ")
			}

			ctx := context.Background()
			resp, err := apiClient.Login(ctx, email, password)
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			// Store credentials
			viper.Set("auth.token", resp.Token)
			if resp.RefreshToken != "" {
				viper.Set("auth.refresh_token", resp.RefreshToken)
			}
			if resp.User != nil {
				viper.Set("auth.email", resp.User.Email)
			}

			if err := writeConfig(); err != nil {
				return fmt.Errorf("failed to save credentials: %w", err)
			}

			name := email
			if resp.User != nil && resp.User.FullName != "" {
				name = resp.User.FullName
			}
			fmt.Printf("Logged in as %s\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "email address")
	cmd.Flags().StringVar(&password, "password", "", "password")

	return cmd
}

func newAuthRegisterCmd() *cobra.Command {
	var email, password, fullName string

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new account",
		RunE: func(cmd *cobra.Command, args []string) error {
			if email == "" {
				email = promptInput("Email: ")
			}
			if fullName == "" {
				fullName = promptInput("Full name: ")
			}
			if password == "" {
				password = promptPassword("Password: ")
				confirm := promptPassword("Confirm password: ")
				if password != confirm {
					return fmt.Errorf("passwords do not match")
				}
			}

			ctx := context.Background()
			resp, err := apiClient.Register(ctx, client.RegisterRequest{
				Email:    email,
				Password: password,
				FullName: fullName,
			})
			if err != nil {
				return fmt.Errorf("registration failed: %w", err)
			}

			viper.Set("auth.token", resp.Token)
			if resp.RefreshToken != "" {
				viper.Set("auth.refresh_token", resp.RefreshToken)
			}
			viper.Set("auth.email", email)

			if err := writeConfig(); err != nil {
				return fmt.Errorf("failed to save credentials: %w", err)
			}

			fmt.Printf("Account created. Logged in as %s\n", email)
			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "email address")
	cmd.Flags().StringVar(&password, "password", "", "password")
	cmd.Flags().StringVar(&fullName, "name", "", "full name")

	return cmd
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear stored credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.Set("auth.token", "")
			viper.Set("auth.refresh_token", "")
			viper.Set("auth.email", "")

			if err := writeConfig(); err != nil {
				return fmt.Errorf("failed to clear credentials: %w", err)
			}

			fmt.Println("Logged out successfully")
			return nil
		},
	}
}

func newAuthWhoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show current user info",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			user, err := apiClient.GetCurrentUser(ctx)
			if err != nil {
				return fmt.Errorf("failed to get user info: %w", err)
			}

			format := getOutputFormat()
			if format != "table" {
				return printOutput(user)
			}

			fmt.Printf("Email:    %s\n", user.Email)
			if user.FullName != "" {
				fmt.Printf("Name:     %s\n", user.FullName)
			}
			if user.Username != "" {
				fmt.Printf("Username: %s\n", user.Username)
			}
			fmt.Printf("Role:     %s\n", user.Role)
			fmt.Printf("Plan:     %s\n", user.PlanType)
			fmt.Printf("ID:       %d\n", user.ID)
			return nil
		},
	}
}

func promptInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func promptPassword(prompt string) string {
	fmt.Print(prompt)
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return ""
	}
	return string(password)
}
