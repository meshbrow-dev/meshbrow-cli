package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var fleetCmd = &cobra.Command{
	Use:     "fleet",
	Aliases: []string{"f"},
	Short:   "Fleet management commands",
	Long:    "Create, monitor, and manage multi-session browser fleets.",
}

var fleetCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new fleet",
	RunE:  runFleetCreate,
}

var fleetStatusCmd = &cobra.Command{
	Use:   "status [fleet-id]",
	Short: "Get fleet status",
	Args:  cobra.ExactArgs(1),
	RunE:  runFleetStatus,
}

var fleetDestroyCmd = &cobra.Command{
	Use:     "destroy [fleet-id]",
	Aliases: []string{"rm"},
	Short:   "Destroy a fleet and all its sessions",
	Args:    cobra.ExactArgs(1),
	RunE:    runFleetDestroy,
}

var fleetListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List active fleets",
	RunE:    runFleetList,
}

func init() {
	rootCmd.AddCommand(fleetCmd)
	fleetCmd.AddCommand(fleetCreateCmd)
	fleetCmd.AddCommand(fleetStatusCmd)
	fleetCmd.AddCommand(fleetDestroyCmd)
	fleetCmd.AddCommand(fleetListCmd)

	// Create flags
	fleetCreateCmd.Flags().String("name", "", "fleet name")
	fleetCreateCmd.Flags().Int("count", 5, "number of sessions")
	fleetCreateCmd.Flags().String("stealth", "max", "stealth level")
	fleetCreateCmd.Flags().String("proxy-type", "residential", "proxy type")
	fleetCreateCmd.Flags().StringSlice("countries", nil, "distribute across countries")
	fleetCreateCmd.Flags().Int("timeout", 30, "session timeout in minutes")
}

func runFleetCreate(cmd *cobra.Command, args []string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	name, _ := cmd.Flags().GetString("name")
	count, _ := cmd.Flags().GetInt("count")
	stealth, _ := cmd.Flags().GetString("stealth")
	proxyType, _ := cmd.Flags().GetString("proxy-type")
	countries, _ := cmd.Flags().GetStringSlice("countries")
	timeout, _ := cmd.Flags().GetInt("timeout")

	reqBody := map[string]interface{}{
		"name":  name,
		"count": count,
		"config": map[string]interface{}{
			"stealth": stealth,
			"timeout": timeout,
		},
	}

	if proxyType != "" || len(countries) > 0 {
		dist := map[string]interface{}{
			"proxy_types": []string{proxyType},
		}
		if len(countries) > 0 {
			dist["countries"] = countries
		}
		reqBody["distribution"] = dist
	}

	body, status, err := client.post("/v1/fleet", reqBody)
	if err != nil {
		return fmt.Errorf("creating fleet: %w", err)
	}

	if status != 201 && status != 200 {
		return parseError(body, status)
	}

	if output == "json" {
		fmt.Println(string(body))
		return nil
	}

	var resp struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Count    int    `json:"count"`
		Status   string `json:"status"`
		Sessions []struct {
			ID          string `json:"id"`
			CDPEndpoint string `json:"cdp_endpoint"`
		} `json:"sessions"`
	}
	json.Unmarshal(body, &resp)

	fmt.Printf("✓ Fleet created\n\n")
	fmt.Printf("  ID:       %s\n", resp.ID)
	fmt.Printf("  Name:     %s\n", resp.Name)
	fmt.Printf("  Sessions: %d\n", resp.Count)
	fmt.Printf("  Status:   %s\n\n", resp.Status)

	if len(resp.Sessions) > 0 {
		fmt.Println("  Sessions:")
		for i, s := range resp.Sessions {
			fmt.Printf("    [%d] %s\n", i+1, s.CDPEndpoint)
		}
	}
	return nil
}

func runFleetStatus(cmd *cobra.Command, args []string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	fleetID := args[0]
	body, status, err := client.get("/v1/fleet/" + fleetID)
	if err != nil {
		return fmt.Errorf("getting fleet status: %w", err)
	}

	if status != 200 {
		return parseError(body, status)
	}

	if output == "json" {
		fmt.Println(string(body))
		return nil
	}

	var resp struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Status   string `json:"status"`
		Sessions []struct {
			ID      string `json:"id"`
			Status  string `json:"status"`
			Country string `json:"proxy_country"`
		} `json:"sessions"`
	}
	json.Unmarshal(body, &resp)

	fmt.Printf("Fleet: %s (%s)\n", resp.Name, resp.ID)
	fmt.Printf("Status: %s\n\n", resp.Status)

	fmt.Printf("  %-36s  %-10s  %s\n", "SESSION ID", "STATUS", "COUNTRY")
	fmt.Println("  " + strings.Repeat("-", 60))
	for _, s := range resp.Sessions {
		fmt.Printf("  %-36s  %-10s  %s\n", s.ID, s.Status, s.Country)
	}
	return nil
}

func runFleetDestroy(cmd *cobra.Command, args []string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	fleetID := args[0]
	body, status, err := client.delete("/v1/fleet/" + fleetID)
	if err != nil {
		return fmt.Errorf("destroying fleet: %w", err)
	}

	if status != 200 && status != 204 {
		return parseError(body, status)
	}

	fmt.Printf("✓ Fleet %s destroyed (all sessions terminated)\n", fleetID)
	return nil
}

func runFleetList(cmd *cobra.Command, args []string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	body, status, err := client.get("/v1/fleet")
	if err != nil {
		return fmt.Errorf("listing fleets: %w", err)
	}

	if status != 200 {
		return parseError(body, status)
	}

	if output == "json" {
		fmt.Println(string(body))
		return nil
	}

	var resp struct {
		Fleets []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Count     int    `json:"count"`
			Status    string `json:"status"`
			CreatedAt string `json:"created_at"`
		} `json:"fleets"`
	}
	json.Unmarshal(body, &resp)

	if len(resp.Fleets) == 0 {
		fmt.Println("No active fleets.")
		return nil
	}

	fmt.Printf("%-36s  %-20s  %-5s  %-10s  %s\n",
		"ID", "NAME", "SIZE", "STATUS", "CREATED")
	fmt.Println(strings.Repeat("-", 90))
	for _, f := range resp.Fleets {
		fmt.Printf("%-36s  %-20s  %-5d  %-10s  %s\n",
			f.ID, truncate(f.Name, 20), f.Count, f.Status, f.CreatedAt)
	}
	return nil
}
