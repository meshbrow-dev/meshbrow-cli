package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show system status and usage",
	RunE:  runStatus,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print CLI version",
	Run:   runVersion,
}

// Version is set at build time via ldflags.
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func init() {
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(versionCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	body, status, err := client.get("/v1/status")
	if err != nil {
		return fmt.Errorf("getting status: %w", err)
	}

	if status != 200 {
		return parseError(body, status)
	}

	if output == "json" {
		fmt.Println(string(body))
		return nil
	}

	var resp struct {
		Status  string `json:"status"`
		Version string `json:"version"`
		Usage   struct {
			ActiveSessions int `json:"active_sessions"`
			TotalToday     int `json:"total_today"`
			TotalMonth     int `json:"total_month"`
			Limit          int `json:"limit"`
		} `json:"usage"`
		Plan string `json:"plan"`
	}
	json.Unmarshal(body, &resp)

	fmt.Printf("Meshbrow Status\n\n")
	fmt.Printf("  API:      %s (%s)\n", resp.Status, getAPIURL())
	fmt.Printf("  Version:  %s\n", resp.Version)
	fmt.Printf("  Plan:     %s\n\n", resp.Plan)
	fmt.Printf("  Usage:\n")
	fmt.Printf("    Active sessions: %d\n", resp.Usage.ActiveSessions)
	fmt.Printf("    Today:           %d\n", resp.Usage.TotalToday)
	fmt.Printf("    This month:      %d / %d\n", resp.Usage.TotalMonth, resp.Usage.Limit)
	return nil
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("meshbrow %s\n", Version)
	fmt.Printf("  commit:  %s\n", Commit)
	fmt.Printf("  built:   %s\n", BuildDate)
}
