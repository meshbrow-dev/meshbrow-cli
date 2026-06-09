package cmd

import (
	"net/http"
	"os"
	"time"
)

// defaultHTTPClient returns a configured HTTP client for API requests.
func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

// writeFile writes data to a file.
func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// readFile reads a file's contents.
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
