package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var cookiesCmd = &cobra.Command{
	Use:   "cookies",
	Short: "Export and import session cookies",
}

var cookiesExportCmd = &cobra.Command{
	Use:   "export [session-id]",
	Short: "Export cookies from a session",
	Args:  cobra.ExactArgs(1),
	RunE:  runCookiesExport,
}

var cookiesImportCmd = &cobra.Command{
	Use:   "import [session-id] [file]",
	Short: "Import cookies into a session",
	Args:  cobra.ExactArgs(2),
	RunE:  runCookiesImport,
}

func init() {
	rootCmd.AddCommand(cookiesCmd)
	cookiesCmd.AddCommand(cookiesExportCmd)
	cookiesCmd.AddCommand(cookiesImportCmd)

	cookiesExportCmd.Flags().String("format", "json", "export format (json, netscape)")
	cookiesExportCmd.Flags().StringP("file", "f", "", "save to file (default: stdout)")

	cookiesImportCmd.Flags().String("format", "json", "import format (json, netscape)")
	cookiesImportCmd.Flags().Bool("merge", false, "merge with existing cookies")
}

func runCookiesExport(cmd *cobra.Command, args []string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	sessionID := args[0]
	format, _ := cmd.Flags().GetString("format")
	file, _ := cmd.Flags().GetString("file")

	body, status, err := client.get(fmt.Sprintf("/v1/sessions/%s/cookies?format=%s", sessionID, format))
	if err != nil {
		return fmt.Errorf("exporting cookies: %w", err)
	}

	if status != 200 {
		return parseError(body, status)
	}

	if file != "" {
		if err := writeFile(file, body); err != nil {
			return fmt.Errorf("saving cookies: %w", err)
		}
		fmt.Printf("✓ Cookies exported to %s\n", file)
		return nil
	}

	fmt.Print(string(body))
	return nil
}

func runCookiesImport(cmd *cobra.Command, args []string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	sessionID := args[0]
	filePath := args[1]
	format, _ := cmd.Flags().GetString("format")
	merge, _ := cmd.Flags().GetBool("merge")

	data, err := readFile(filePath)
	if err != nil {
		return fmt.Errorf("reading cookie file: %w", err)
	}

	reqBody := map[string]interface{}{
		"cookies": string(data),
		"format":  format,
		"merge":   merge,
	}

	body, status, err := client.post("/v1/sessions/"+sessionID+"/cookies/import", reqBody)
	if err != nil {
		return fmt.Errorf("importing cookies: %w", err)
	}

	if status != 200 {
		return parseError(body, status)
	}

	var resp struct {
		Imported int `json:"imported"`
	}
	json.Unmarshal(body, &resp)

	fmt.Printf("✓ Imported %d cookies into session %s\n", resp.Imported, sessionID)
	return nil
}
