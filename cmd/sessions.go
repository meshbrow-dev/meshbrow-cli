package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var sessionsCmd = &cobra.Command{
	Use:     "sessions",
	Aliases: []string{"session", "s"},
	Short:   "Manage browser sessions",
	Long:    "Create, list, and manage browser sessions.",
}

var sessionsCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"launch", "new"},
	Short:   "Create a new browser session",
	RunE:    runSessionsCreate,
}

var sessionsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List active sessions",
	RunE:    runSessionsList,
}

var sessionsGetCmd = &cobra.Command{
	Use:   "get [session-id]",
	Short: "Get session details",
	Args:  cobra.ExactArgs(1),
	RunE:  runSessionsGet,
}

var sessionsDestroyCmd = &cobra.Command{
	Use:     "destroy [session-id]",
	Aliases: []string{"rm", "delete", "kill"},
	Short:   "Destroy a session",
	Args:    cobra.ExactArgs(1),
	RunE:    runSessionsDestroy,
}

func init() {
	rootCmd.AddCommand(sessionsCmd)
	sessionsCmd.AddCommand(sessionsCreateCmd)
	sessionsCmd.AddCommand(sessionsListCmd)
	sessionsCmd.AddCommand(sessionsGetCmd)
	sessionsCmd.AddCommand(sessionsDestroyCmd)

	// Create flags
	sessionsCreateCmd.Flags().String("stealth", "max", "stealth level (none, basic, max)")
	sessionsCreateCmd.Flags().String("proxy-type", "residential", "proxy type (residential, datacenter, isp, mobile)")
	sessionsCreateCmd.Flags().String("proxy-country", "", "proxy country (ISO 3166-1 alpha-2)")
	sessionsCreateCmd.Flags().Bool("proxy-sticky", false, "use sticky proxy (same IP)")
	sessionsCreateCmd.Flags().String("viewport", "1920x1080", "viewport dimensions (WxH)")
	sessionsCreateCmd.Flags().String("locale", "", "browser locale (e.g., en-US)")
	sessionsCreateCmd.Flags().String("timezone", "", "browser timezone (e.g., America/New_York)")
	sessionsCreateCmd.Flags().Int("timeout", 30, "session timeout in minutes")
	sessionsCreateCmd.Flags().String("profile", "", "profile ID to restore session from")
	sessionsCreateCmd.Flags().Bool("save-profile", false, "save session state on destroy")
}

func runSessionsCreate(cmd *cobra.Command, args []string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	stealth, _ := cmd.Flags().GetString("stealth")
	proxyType, _ := cmd.Flags().GetString("proxy-type")
	proxyCountry, _ := cmd.Flags().GetString("proxy-country")
	proxySticky, _ := cmd.Flags().GetBool("proxy-sticky")
	viewport, _ := cmd.Flags().GetString("viewport")
	locale, _ := cmd.Flags().GetString("locale")
	timezone, _ := cmd.Flags().GetString("timezone")
	timeout, _ := cmd.Flags().GetInt("timeout")
	profile, _ := cmd.Flags().GetString("profile")

	// Parse viewport
	var width, height int
	fmt.Sscanf(viewport, "%dx%d", &width, &height)
	if width == 0 || height == 0 {
		width, height = 1920, 1080
	}

	reqBody := map[string]interface{}{
		"stealth": stealth,
		"viewport": map[string]int{
			"width":  width,
			"height": height,
		},
		"timeout": timeout,
	}

	if proxyType != "" || proxyCountry != "" {
		reqBody["proxy"] = map[string]interface{}{
			"type":    proxyType,
			"country": proxyCountry,
			"sticky":  proxySticky,
		}
	}

	if locale != "" {
		reqBody["locale"] = locale
	}
	if timezone != "" {
		reqBody["timezone"] = timezone
	}
	if profile != "" {
		reqBody["profile_id"] = profile
	}

	body, status, err := client.post("/v1/sessions", reqBody)
	if err != nil {
		return fmt.Errorf("creating session: %w", err)
	}

	if status != 201 && status != 200 {
		return parseError(body, status)
	}

	var resp struct {
		ID          string `json:"id"`
		Status      string `json:"status"`
		CDPEndpoint string `json:"cdp_endpoint"`
		Token       string `json:"token"`
		ExpiresAt   string `json:"expires_at"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	if output == "json" {
		fmt.Println(string(body))
		return nil
	}

	fmt.Printf("✓ Session created\n\n")
	fmt.Printf("  ID:       %s\n", resp.ID)
	fmt.Printf("  Status:   %s\n", resp.Status)
	fmt.Printf("  CDP:      %s\n", resp.CDPEndpoint)
	fmt.Printf("  Expires:  %s\n", resp.ExpiresAt)
	fmt.Printf("\n  Connect with Playwright:\n")
	fmt.Printf("    const browser = await chromium.connectOverCDP('%s');\n", resp.CDPEndpoint)
	return nil
}

func runSessionsList(cmd *cobra.Command, args []string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	body, status, err := client.get("/v1/sessions")
	if err != nil {
		return fmt.Errorf("listing sessions: %w", err)
	}

	if status != 200 {
		return parseError(body, status)
	}

	if output == "json" {
		fmt.Println(string(body))
		return nil
	}

	var resp struct {
		Sessions []struct {
			ID        string `json:"id"`
			Status    string `json:"status"`
			Stealth   string `json:"stealth"`
			Proxy     string `json:"proxy_type"`
			Country   string `json:"proxy_country"`
			CreatedAt string `json:"created_at"`
		} `json:"sessions"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	if resp.Total == 0 {
		fmt.Println("No active sessions.")
		return nil
	}

	fmt.Printf("%-36s  %-10s  %-8s  %-12s  %-4s  %s\n",
		"ID", "STATUS", "STEALTH", "PROXY", "GEO", "CREATED")
	fmt.Println(strings.Repeat("-", 100))

	for _, s := range resp.Sessions {
		fmt.Printf("%-36s  %-10s  %-8s  %-12s  %-4s  %s\n",
			s.ID, s.Status, s.Stealth, s.Proxy, s.Country, s.CreatedAt)
	}

	fmt.Printf("\nTotal: %d sessions\n", resp.Total)
	return nil
}

func runSessionsGet(cmd *cobra.Command, args []string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	sessionID := args[0]
	body, status, err := client.get("/v1/sessions/" + sessionID)
	if err != nil {
		return fmt.Errorf("getting session: %w", err)
	}

	if status != 200 {
		return parseError(body, status)
	}

	if output == "json" {
		fmt.Println(string(body))
		return nil
	}

	var resp struct {
		ID          string `json:"id"`
		Status      string `json:"status"`
		CDPEndpoint string `json:"cdp_endpoint"`
		Stealth     string `json:"stealth"`
		ProxyType   string `json:"proxy_type"`
		Country     string `json:"proxy_country"`
		CreatedAt   string `json:"created_at"`
		ExpiresAt   string `json:"expires_at"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fmt.Printf("Session: %s\n\n", resp.ID)
	fmt.Printf("  Status:   %s\n", resp.Status)
	fmt.Printf("  Stealth:  %s\n", resp.Stealth)
	fmt.Printf("  Proxy:    %s (%s)\n", resp.ProxyType, resp.Country)
	fmt.Printf("  CDP:      %s\n", resp.CDPEndpoint)
	fmt.Printf("  Created:  %s\n", resp.CreatedAt)
	fmt.Printf("  Expires:  %s\n", resp.ExpiresAt)
	return nil
}

func runSessionsDestroy(cmd *cobra.Command, args []string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	sessionID := args[0]
	body, status, err := client.delete("/v1/sessions/" + sessionID)
	if err != nil {
		return fmt.Errorf("destroying session: %w", err)
	}

	if status != 200 && status != 204 {
		return parseError(body, status)
	}

	fmt.Printf("✓ Session %s destroyed\n", sessionID)
	return nil
}
