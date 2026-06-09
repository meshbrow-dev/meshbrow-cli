package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  "Manage API keys and authentication for the Meshbrow CLI.",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Save API key for CLI usage",
	Long:  "Store your API key in ~/.meshbrow.yaml for future CLI commands.",
	RunE:  runAuthLogin,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	RunE:  runAuthStatus,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored API key",
	RunE:  runAuthLogout,
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)

	authLoginCmd.Flags().String("key", "", "API key to store")
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	key, _ := cmd.Flags().GetString("key")
	if key == "" {
		fmt.Print("Enter your API key: ")
		fmt.Scanln(&key)
	}

	if key == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Validate the key by calling the API
	tmpClient := &Client{
		baseURL:    getAPIURL(),
		apiKey:     key,
		httpClient: defaultHTTPClient(),
	}

	body, status, err := tmpClient.get("/v1/auth/me")
	if err != nil {
		return fmt.Errorf("validating API key: %w", err)
	}

	if status != 200 {
		return fmt.Errorf("invalid API key (status %d)", status)
	}

	var user struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	json.Unmarshal(body, &user)

	// Save to config file
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	configPath := filepath.Join(home, ".meshbrow.yaml")
	viper.Set("api_key", key)
	viper.Set("api_url", getAPIURL())

	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("✓ Authenticated as %s\n", user.Email)
	fmt.Printf("  Config saved to %s\n", configPath)
	return nil
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	key := getAPIKey()
	if key == "" {
		fmt.Println("Not authenticated. Run 'meshbrow auth login' to configure.")
		return nil
	}

	client, err := NewClient()
	if err != nil {
		return err
	}

	body, status, err := client.get("/v1/auth/me")
	if err != nil {
		return fmt.Errorf("checking auth status: %w", err)
	}

	if status != 200 {
		fmt.Println("✗ API key is invalid or expired")
		return nil
	}

	var user struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Plan  string `json:"plan"`
	}
	json.Unmarshal(body, &user)

	fmt.Printf("✓ Authenticated\n")
	fmt.Printf("  Email: %s\n", user.Email)
	fmt.Printf("  Plan:  %s\n", user.Plan)
	fmt.Printf("  API:   %s\n", getAPIURL())
	return nil
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	configPath := filepath.Join(home, ".meshbrow.yaml")
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing config: %w", err)
	}

	fmt.Println("✓ Logged out. Config removed.")
	return nil
}
