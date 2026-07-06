package holded

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// TestDoJSONSuccess verifies that DoJSON decodes a 200 response with JSON
// and that the API key header is sent correctly.
func TestDoJSONSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify API key header is present and correct
		key := r.Header.Get("key")
		if key != "test-key" {
			t.Errorf("expected key header 'test-key', got '%s'", key)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}))
	defer srv.Close()

	client := &Client{
		BaseURL: srv.URL,
		Client:  srv.Client(),
		APIKey:  "test-key",
		Limiter: nil,
	}

	req, err := client.NewRequest(http.MethodGet, "/test", url.Values{}, nil)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	var result map[string]bool
	err = client.DoJSON(req, &result)
	if err != nil {
		t.Fatalf("DoJSON failed: %v", err)
	}

	if !result["ok"] {
		t.Errorf("expected ok=true, got %v", result["ok"])
	}
}

// TestDoJSONErrorIncludesBody verifies that DoJSON returns an error containing
// both the status code and the response body.
func TestDoJSONErrorIncludesBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"bad"}`)
	}))
	defer srv.Close()

	client := &Client{
		BaseURL: srv.URL,
		Client:  srv.Client(),
		APIKey:  "test-key",
		Limiter: nil,
	}

	req, err := client.NewRequest(http.MethodGet, "/test", url.Values{}, nil)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	var result map[string]string
	err = client.DoJSON(req, &result)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "400") {
		t.Errorf("error should contain '400', got: %s", errStr)
	}
	if !strings.Contains(errStr, "bad") {
		t.Errorf("error should contain 'bad', got: %s", errStr)
	}
}

// TestRetryOn503 verifies that the client retries on 503 and succeeds.
// Tracks that exactly 2 requests are made.
func TestRetryOn503(t *testing.T) {
	requestCount := atomic.Int32{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		if count == 1 {
			// First request: return 503
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		// Second request: return 200 with JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}))
	defer srv.Close()

	client := &Client{
		BaseURL: srv.URL,
		Client:  srv.Client(),
		APIKey:  "test-key",
		Limiter: nil,
	}

	req, err := client.NewRequest(http.MethodGet, "/test", url.Values{}, nil)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	var result map[string]bool
	err = client.DoJSON(req, &result)
	if err != nil {
		t.Fatalf("DoJSON failed: %v", err)
	}

	if count := requestCount.Load(); count != 2 {
		t.Errorf("expected 2 requests, got %d", count)
	}

	if !result["ok"] {
		t.Errorf("expected ok=true, got %v", result["ok"])
	}
}

// TestRetryAfterHeader verifies that the client honors Retry-After header
// and retries accordingly. Verifies that elapsed time is >= 1 second.
func TestRetryAfterHeader(t *testing.T) {
	requestCount := atomic.Int32{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		if count == 1 {
			// First request: return 429 with Retry-After header
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		// Second request: return 200 with JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}))
	defer srv.Close()

	client := &Client{
		BaseURL: srv.URL,
		Client:  srv.Client(),
		APIKey:  "test-key",
		Limiter: nil,
	}

	req, err := client.NewRequest(http.MethodGet, "/test", url.Values{}, nil)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	start := time.Now()
	var result map[string]bool
	err = client.DoJSON(req, &result)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("DoJSON failed: %v", err)
	}

	if count := requestCount.Load(); count != 2 {
		t.Errorf("expected 2 requests, got %d", count)
	}

	if elapsed < time.Second {
		t.Errorf("expected elapsed time >= 1 second, got %v", elapsed)
	}

	if !result["ok"] {
		t.Errorf("expected ok=true, got %v", result["ok"])
	}
}

// TestRetryAfterDelayParsing tests the retryAfterDelay function directly
// with various Retry-After header values.
func TestRetryAfterDelayParsing(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected time.Duration
	}{
		{
			name:     "missing header",
			header:   "",
			expected: 0,
		},
		{
			name:     "valid 5 seconds",
			header:   "5",
			expected: 5 * time.Second,
		},
		{
			name:     "capped to 30 seconds",
			header:   "9999",
			expected: 30 * time.Second,
		},
		{
			name:     "invalid non-numeric",
			header:   "abc",
			expected: 0,
		},
		{
			name:     "negative number",
			header:   "-3",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock response with the header
			resp := &http.Response{
				Header: make(http.Header),
			}
			if tt.header != "" {
				resp.Header.Set("Retry-After", tt.header)
			}

			result := retryAfterDelay(resp)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestMissingAPIKey verifies that do() returns an error when APIKey is empty
// and does not make any HTTP request.
func TestMissingAPIKey(t *testing.T) {
	requestMade := atomic.Bool{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestMade.Store(true)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := &Client{
		BaseURL: srv.URL,
		Client:  srv.Client(),
		APIKey:  "", // Empty API key
		Limiter: nil,
	}

	req, err := client.NewRequest(http.MethodGet, "/test", url.Values{}, nil)
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	var result interface{}
	err = client.DoJSON(req, &result)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "HOLDED_API_KEY") {
		t.Errorf("error should contain 'HOLDED_API_KEY', got: %s", errStr)
	}

	if requestMade.Load() {
		t.Error("expected no HTTP request to be made")
	}
}
