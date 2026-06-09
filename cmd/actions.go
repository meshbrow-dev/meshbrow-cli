package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var screenshotCmd = &cobra.Command{
	Use:   "screenshot [session-id]",
	Short: "Capture a screenshot from a session",
	Args:  cobra.ExactArgs(1),
	RunE:  runScreenshot,
}

var navigateCmd = &cobra.Command{
	Use:   "navigate [session-id] [url]",
	Short: "Navigate a session to a URL",
	Args:  cobra.ExactArgs(2),
	RunE:  runNavigate,
}

var execCmd = &cobra.Command{
	Use:   "exec [session-id] [script]",
	Short: "Execute JavaScript in a session",
	Args:  cobra.ExactArgs(2),
	RunE:  runExec,
}

func init() {
	rootCmd.AddCommand(screenshotCmd)
	rootCmd.AddCommand(navigateCmd)
	rootCmd.AddCommand(execCmd)

	screenshotCmd.Flags().StringP("file", "f", "", "save screenshot to file (default: stdout as base64)")
	screenshotCmd.Flags().Bool("full-page", false, "capture full page screenshot")

	navigateCmd.Flags().String("wait", "load", "wait condition (load, domcontentloaded, networkidle)")
	navigateCmd.Flags().Int("timeout", 30, "navigation timeout in seconds")
}

func runScreenshot(cmd *cobra.Command, args []string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	sessionID := args[0]
	fullPage, _ := cmd.Flags().GetBool("full-page")
	file, _ := cmd.Flags().GetString("file")

	reqBody := map[string]interface{}{
		"full_page": fullPage,
	}

	body, status, err := client.post("/v1/sessions/"+sessionID+"/screenshot", reqBody)
	if err != nil {
		return fmt.Errorf("capturing screenshot: %w", err)
	}

	if status != 200 {
		return parseError(body, status)
	}

	var resp struct {
		Data   string `json:"data"`
		Format string `json:"format"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	imgData, err := base64.StdEncoding.DecodeString(resp.Data)
	if err != nil {
		return fmt.Errorf("decoding screenshot: %w", err)
	}

	if file != "" {
		if err := os.WriteFile(file, imgData, 0644); err != nil {
			return fmt.Errorf("saving screenshot: %w", err)
		}
		fmt.Printf("✓ Screenshot saved to %s (%d bytes)\n", file, len(imgData))
	} else {
		// Write raw PNG to stdout for piping
		os.Stdout.Write(imgData)
	}

	return nil
}

func runNavigate(cmd *cobra.Command, args []string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	sessionID := args[0]
	url := args[1]
	wait, _ := cmd.Flags().GetString("wait")
	timeout, _ := cmd.Flags().GetInt("timeout")

	reqBody := map[string]interface{}{
		"url":     url,
		"wait":    wait,
		"timeout": timeout,
	}

	body, status, err := client.post("/v1/sessions/"+sessionID+"/navigate", reqBody)
	if err != nil {
		return fmt.Errorf("navigating: %w", err)
	}

	if status != 200 {
		return parseError(body, status)
	}

	if output == "json" {
		fmt.Println(string(body))
		return nil
	}

	fmt.Printf("✓ Navigated to %s\n", url)
	return nil
}

func runExec(cmd *cobra.Command, args []string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	sessionID := args[0]
	script := args[1]

	reqBody := map[string]interface{}{
		"script": script,
	}

	body, status, err := client.post("/v1/sessions/"+sessionID+"/execute", reqBody)
	if err != nil {
		return fmt.Errorf("executing script: %w", err)
	}

	if status != 200 {
		return parseError(body, status)
	}

	if output == "json" {
		fmt.Println(string(body))
		return nil
	}

	var resp struct {
		Result interface{} `json:"result"`
	}
	json.Unmarshal(body, &resp)

	resultJSON, _ := json.MarshalIndent(resp.Result, "", "  ")
	fmt.Println(string(resultJSON))
	return nil
}
