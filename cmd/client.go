package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the HTTP client for the Meshbrow API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new API client.
func NewClient() (*Client, error) {
	key := getAPIKey()
	if key == "" {
		return nil, fmt.Errorf("API key required. Set via --api-key, MESHBROW_API_KEY env, or run 'meshbrow auth login'")
	}

	return &Client{
		baseURL: getAPIURL(),
		apiKey:  key,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *Client) do(method, path string, body interface{}) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshaling request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "meshbrow-cli/1.0.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response: %w", err)
	}

	// The API wraps successful responses in a {"data": ...} envelope.
	// Unwrap it so callers can decode payloads directly.
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var envelope struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(respBody, &envelope); err == nil && len(envelope.Data) > 0 {
			return envelope.Data, resp.StatusCode, nil
		}
	}

	return respBody, resp.StatusCode, nil
}

func (c *Client) get(path string) ([]byte, int, error) {
	return c.do(http.MethodGet, path, nil)
}

func (c *Client) post(path string, body interface{}) ([]byte, int, error) {
	return c.do(http.MethodPost, path, body)
}

func (c *Client) delete(path string) ([]byte, int, error) {
	return c.do(http.MethodDelete, path, nil)
}

// APIError represents an error response from the API.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func parseError(body []byte, statusCode int) error {
	var errResp struct {
		Error APIError `json:"error"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
		return &errResp.Error
	}
	return fmt.Errorf("API returned status %d", statusCode)
}
