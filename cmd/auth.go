package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

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
	Short: "Log in to Meshbrow",
	Long: `Authenticate the CLI with your Meshbrow account.

By default this opens your browser to complete authentication — once you
approve the request on the web, the CLI finishes automatically.

To authenticate without a browser, pass an API key directly:

  meshbrow auth login --key YOUR_API_KEY`,
	RunE: runAuthLogin,
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

	authLoginCmd.Flags().String("key", "", "API key to store (skips the browser flow)")
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	key, _ := cmd.Flags().GetString("key")

	if key == "" {
		browserKey, err := browserLogin()
		if err != nil {
			return err
		}
		key = browserKey
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

// --- Browser-based login (device authorization flow) ---

type cliLoginStart struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type cliLoginPoll struct {
	Status string `json:"status"`
	APIKey string `json:"api_key"`
}

// openBrowser is a package variable so tests can stub it.
var openBrowser = func(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}

// browserLogin runs the device authorization flow and returns an API key.
func browserLogin() (string, error) {
	client := &Client{
		baseURL:    getAPIURL(),
		httpClient: defaultHTTPClient(),
	}

	body, status, err := client.post("/v1/cli/auth/start", nil)
	if err != nil {
		return "", fmt.Errorf("starting login: %w", err)
	}
	if status != 200 {
		return "", fmt.Errorf("starting login failed (status %d)", status)
	}

	var start cliLoginStart
	if err := json.Unmarshal(body, &start); err != nil {
		return "", fmt.Errorf("parsing login response: %w", err)
	}

	fmt.Printf("First, confirm this code matches your browser: %s\n\n", start.UserCode)
	fmt.Printf("Opening %s\n", start.VerificationURL)
	if err := openBrowser(start.VerificationURL); err != nil {
		fmt.Println("Could not open a browser automatically — open the URL above manually.")
	}
	fmt.Println("\nWaiting for approval...")

	return pollBrowserLogin(client, start)
}

// pollBrowserLogin polls until the login is approved, denied, or expired.
func pollBrowserLogin(client *Client, start cliLoginStart) (string, error) {
	interval := time.Duration(start.Interval) * time.Second
	if interval <= 0 {
		interval = 2 * time.Second
	}
	expiresIn := time.Duration(start.ExpiresIn) * time.Second
	if expiresIn <= 0 {
		expiresIn = 10 * time.Minute
	}
	deadline := time.Now().Add(expiresIn)

	for time.Now().Before(deadline) {
		time.Sleep(interval)

		body, status, err := client.post("/v1/cli/auth/poll", map[string]string{
			"device_code": start.DeviceCode,
		})
		if err != nil || status != 200 {
			continue // transient error, keep polling until the deadline
		}

		var poll cliLoginPoll
		if err := json.Unmarshal(body, &poll); err != nil {
			continue
		}

		switch poll.Status {
		case "approved":
			return poll.APIKey, nil
		case "expired":
			return "", fmt.Errorf("login request expired — run 'meshbrow auth login' again")
		}
	}

	return "", fmt.Errorf("login timed out — run 'meshbrow auth login' again")
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
