package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	apiURL  string
	apiKey  string
	output  string
)

var rootCmd = &cobra.Command{
	Use:   "meshbrow",
	Short: "Managed Browser Fleet for AI Agents",
	Long: `meshbrow is a CLI tool for managing cloud-hosted stealth Chromium browsers.

Launch, manage, and monitor browser sessions with anti-detection,
unique fingerprints, and proxy rotation.

Documentation: https://docs.meshbrow.dev
Dashboard:     https://meshbrow.dev`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default $HOME/.meshbrow.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "https://api.meshbrow.dev", "API base URL")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key for authentication")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "text", "output format (text, json)")

	viper.BindPFlag("api_url", rootCmd.PersistentFlags().Lookup("api-url"))
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".meshbrow")
	}

	viper.SetEnvPrefix("MESHBROW")
	viper.AutomaticEnv()

	// Read config file if it exists (ignore error if not found)
	viper.ReadInConfig()
}

func getAPIURL() string {
	if apiURL != "https://api.meshbrow.dev" {
		return apiURL
	}
	if v := viper.GetString("api_url"); v != "" {
		return v
	}
	return "https://api.meshbrow.dev"
}

func getAPIKey() string {
	if apiKey != "" {
		return apiKey
	}
	if v := viper.GetString("api_key"); v != "" {
		return v
	}
	return ""
}
