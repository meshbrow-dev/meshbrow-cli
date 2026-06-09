package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var profilesCmd = &cobra.Command{
	Use:     "profiles",
	Aliases: []string{"profile", "p"},
	Short:   "Manage browser profiles",
	Long:    "Create, list, and manage persistent browser profiles.",
}

var profilesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new profile",
	RunE:  runProfilesCreate,
}

var profilesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List profiles",
	RunE:    runProfilesList,
}

var profilesGetCmd = &cobra.Command{
	Use:   "get [profile-id]",
	Short: "Get profile details",
	Args:  cobra.ExactArgs(1),
	RunE:  runProfilesGet,
}

var profilesDeleteCmd = &cobra.Command{
	Use:     "delete [profile-id]",
	Aliases: []string{"rm"},
	Short:   "Delete a profile",
	Args:    cobra.ExactArgs(1),
	RunE:    runProfilesDelete,
}

func init() {
	rootCmd.AddCommand(profilesCmd)
	profilesCmd.AddCommand(profilesCreateCmd)
	profilesCmd.AddCommand(profilesListCmd)
	profilesCmd.AddCommand(profilesGetCmd)
	profilesCmd.AddCommand(profilesDeleteCmd)

	// Create flags
	profilesCreateCmd.Flags().String("name", "", "profile name (required)")
	profilesCreateCmd.Flags().String("platform", "Win32", "OS platform")
	profilesCreateCmd.Flags().String("locale", "en-US", "browser locale")
	profilesCreateCmd.Flags().String("timezone", "", "browser timezone")
	profilesCreateCmd.Flags().String("proxy-type", "", "default proxy type")
	profilesCreateCmd.Flags().String("proxy-country", "", "default proxy country")
	profilesCreateCmd.Flags().StringSlice("tags", nil, "profile tags")
	profilesCreateCmd.MarkFlagRequired("name")
}

func runProfilesCreate(cmd *cobra.Command, args []string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	name, _ := cmd.Flags().GetString("name")
	platform, _ := cmd.Flags().GetString("platform")
	locale, _ := cmd.Flags().GetString("locale")
	timezone, _ := cmd.Flags().GetString("timezone")
	proxyType, _ := cmd.Flags().GetString("proxy-type")
	proxyCountry, _ := cmd.Flags().GetString("proxy-country")
	tags, _ := cmd.Flags().GetStringSlice("tags")

	reqBody := map[string]interface{}{
		"name": name,
		"fingerprint": map[string]string{
			"platform": platform,
			"locale":   locale,
			"timezone": timezone,
		},
	}

	if proxyType != "" || proxyCountry != "" {
		reqBody["proxy"] = map[string]interface{}{
			"type":    proxyType,
			"country": proxyCountry,
		}
	}

	if len(tags) > 0 {
		reqBody["tags"] = tags
	}

	body, status, err := client.post("/v1/profiles", reqBody)
	if err != nil {
		return fmt.Errorf("creating profile: %w", err)
	}

	if status != 201 && status != 200 {
		return parseError(body, status)
	}

	if output == "json" {
		fmt.Println(string(body))
		return nil
	}

	var resp struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	json.Unmarshal(body, &resp)

	fmt.Printf("✓ Profile created\n\n")
	fmt.Printf("  ID:   %s\n", resp.ID)
	fmt.Printf("  Name: %s\n", resp.Name)
	return nil
}

func runProfilesList(cmd *cobra.Command, args []string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	body, status, err := client.get("/v1/profiles")
	if err != nil {
		return fmt.Errorf("listing profiles: %w", err)
	}

	if status != 200 {
		return parseError(body, status)
	}

	if output == "json" {
		fmt.Println(string(body))
		return nil
	}

	var resp struct {
		Profiles []struct {
			ID        string   `json:"id"`
			Name      string   `json:"name"`
			Platform  string   `json:"platform"`
			Tags      []string `json:"tags"`
			CreatedAt string   `json:"created_at"`
		} `json:"profiles"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	if resp.Total == 0 {
		fmt.Println("No profiles found.")
		return nil
	}

	fmt.Printf("%-36s  %-20s  %-10s  %-20s  %s\n",
		"ID", "NAME", "PLATFORM", "TAGS", "CREATED")
	fmt.Println(strings.Repeat("-", 100))

	for _, p := range resp.Profiles {
		tags := strings.Join(p.Tags, ", ")
		fmt.Printf("%-36s  %-20s  %-10s  %-20s  %s\n",
			p.ID, truncate(p.Name, 20), p.Platform, truncate(tags, 20), p.CreatedAt)
	}

	fmt.Printf("\nTotal: %d profiles\n", resp.Total)
	return nil
}

func runProfilesGet(cmd *cobra.Command, args []string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	profileID := args[0]
	body, status, err := client.get("/v1/profiles/" + profileID)
	if err != nil {
		return fmt.Errorf("getting profile: %w", err)
	}

	if status != 200 {
		return parseError(body, status)
	}

	if output == "json" {
		fmt.Println(string(body))
		return nil
	}

	var resp struct {
		ID        string   `json:"id"`
		Name      string   `json:"name"`
		Platform  string   `json:"platform"`
		Locale    string   `json:"locale"`
		Timezone  string   `json:"timezone"`
		Tags      []string `json:"tags"`
		CreatedAt string   `json:"created_at"`
		UpdatedAt string   `json:"updated_at"`
	}
	json.Unmarshal(body, &resp)

	fmt.Printf("Profile: %s\n\n", resp.Name)
	fmt.Printf("  ID:        %s\n", resp.ID)
	fmt.Printf("  Platform:  %s\n", resp.Platform)
	fmt.Printf("  Locale:    %s\n", resp.Locale)
	fmt.Printf("  Timezone:  %s\n", resp.Timezone)
	fmt.Printf("  Tags:      %s\n", strings.Join(resp.Tags, ", "))
	fmt.Printf("  Created:   %s\n", resp.CreatedAt)
	fmt.Printf("  Updated:   %s\n", resp.UpdatedAt)
	return nil
}

func runProfilesDelete(cmd *cobra.Command, args []string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	profileID := args[0]
	body, status, err := client.delete("/v1/profiles/" + profileID)
	if err != nil {
		return fmt.Errorf("deleting profile: %w", err)
	}

	if status != 200 && status != 204 {
		return parseError(body, status)
	}

	fmt.Printf("✓ Profile %s deleted\n", profileID)
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
