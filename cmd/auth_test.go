package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func newTestClient(baseURL string) *Client {
	return &Client{baseURL: baseURL, httpClient: defaultHTTPClient()}
}

func TestPollBrowserLoginApproved(t *testing.T) {
	var polls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/cli/auth/poll" {
			t.Errorf("unexpected path %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		var req struct {
			DeviceCode string `json:"device_code"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if req.DeviceCode != "dev-123" {
			t.Errorf("device_code = %q, want dev-123", req.DeviceCode)
		}

		w.Header().Set("Content-Type", "application/json")
		if polls.Add(1) < 2 {
			json.NewEncoder(w).Encode(map[string]string{"status": "pending"})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "approved", "api_key": "mb_live_abc"})
	}))
	defer srv.Close()

	start := cliLoginStart{DeviceCode: "dev-123", ExpiresIn: 30, Interval: 1}
	key, err := pollBrowserLogin(newTestClient(srv.URL), start)
	if err != nil {
		t.Fatalf("pollBrowserLogin() error: %v", err)
	}
	if key != "mb_live_abc" {
		t.Errorf("key = %q, want mb_live_abc", key)
	}
	if polls.Load() < 2 {
		t.Errorf("polls = %d, want >= 2", polls.Load())
	}
}

func TestPollBrowserLoginExpired(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "expired"})
	}))
	defer srv.Close()

	start := cliLoginStart{DeviceCode: "dev-123", ExpiresIn: 30, Interval: 1}
	if _, err := pollBrowserLogin(newTestClient(srv.URL), start); err == nil {
		t.Fatal("pollBrowserLogin() succeeded, want expired error")
	}
}

func TestPollBrowserLoginTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "pending"})
	}))
	defer srv.Close()

	start := cliLoginStart{DeviceCode: "dev-123", ExpiresIn: 1, Interval: 1}
	if _, err := pollBrowserLogin(newTestClient(srv.URL), start); err == nil {
		t.Fatal("pollBrowserLogin() succeeded, want timeout error")
	}
}
