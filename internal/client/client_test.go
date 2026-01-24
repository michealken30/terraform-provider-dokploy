package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestServer creates a test server with the given handler
func newTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// newTestClient creates a test client pointing to the test server
func newTestClient(server *httptest.Server) *Client {
	return New(server.URL, "test-api-key")
}

func TestNew(t *testing.T) {
	client := New("https://example.com", "test-key")

	if client == nil {
		t.Fatal("expected client to be created")
	}
	if client.baseURL != "https://example.com" {
		t.Errorf("baseURL = %q, want %q", client.baseURL, "https://example.com")
	}
	if client.apiKey != "test-key" {
		t.Errorf("apiKey = %q, want %q", client.apiKey, "test-key")
	}
	if client.httpClient == nil {
		t.Error("expected httpClient to be set")
	}
	if client.retryConfig.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", client.retryConfig.MaxRetries)
	}
}

func TestNewWithRetry(t *testing.T) {
	retryConfig := RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     10 * time.Second,
	}

	client := NewWithRetry("https://example.com", "test-key", retryConfig)

	if client.retryConfig.MaxRetries != 5 {
		t.Errorf("MaxRetries = %d, want 5", client.retryConfig.MaxRetries)
	}
	if client.retryConfig.InitialBackoff != 100*time.Millisecond {
		t.Errorf("InitialBackoff = %v, want 100ms", client.retryConfig.InitialBackoff)
	}
	if client.retryConfig.MaxBackoff != 10*time.Second {
		t.Errorf("MaxBackoff = %v, want 10s", client.retryConfig.MaxBackoff)
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", config.MaxRetries)
	}
	if config.InitialBackoff != 500*time.Millisecond {
		t.Errorf("InitialBackoff = %v, want 500ms", config.InitialBackoff)
	}
	if config.MaxBackoff != 30*time.Second {
		t.Errorf("MaxBackoff = %v, want 30s", config.MaxBackoff)
	}
}

func TestCalculateBackoff(t *testing.T) {
	client := New("https://example.com", "test-key")

	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{
			name:     "first attempt",
			attempt:  0,
			expected: 500 * time.Millisecond,
		},
		{
			name:     "second attempt",
			attempt:  1,
			expected: 1 * time.Second,
		},
		{
			name:     "third attempt",
			attempt:  2,
			expected: 2 * time.Second,
		},
		{
			name:     "capped at max",
			attempt:  10,
			expected: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.calculateBackoff(tt.attempt)
			if got != tt.expected {
				t.Errorf("calculateBackoff(%d) = %v, want %v", tt.attempt, got, tt.expected)
			}
		})
	}
}

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expected   bool
	}{
		{"200 OK", 200, false},
		{"201 Created", 201, false},
		{"400 Bad Request", 400, false},
		{"401 Unauthorized", 401, false},
		{"403 Forbidden", 403, false},
		{"404 Not Found", 404, false},
		{"429 Too Many Requests", 429, true},
		{"500 Internal Server Error", 500, true},
		{"502 Bad Gateway", 502, true},
		{"503 Service Unavailable", 503, true},
		{"504 Gateway Timeout", 504, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldRetry(tt.statusCode); got != tt.expected {
				t.Errorf("shouldRetry(%d) = %v, want %v", tt.statusCode, got, tt.expected)
			}
		})
	}
}

func TestDoRequest_Success(t *testing.T) {
	expected := map[string]string{"status": "ok"}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("x-api-key") != "test-api-key" {
			t.Errorf("expected x-api-key header to be set")
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("expected Accept header to be application/json")
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expected)
	})
	defer server.Close()

	client := newTestClient(server)
	data, err := client.doRequest(context.Background(), "/test.endpoint")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("expected status=ok, got %s", result["status"])
	}
}

func TestDoRequest_ClientError(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.doRequest(context.Background(), "/test.endpoint")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
	}
}

func TestDoRequest_ServerErrorNoRetry(t *testing.T) {
	// Use a client with no retries to avoid test slowdown
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	})
	defer server.Close()

	retryConfig := RetryConfig{
		MaxRetries:     0, // No retries
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
	}
	client := NewWithRetry(server.URL, "test-api-key", retryConfig)

	_, err := client.doRequest(context.Background(), "/test.endpoint")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
	}
}

func TestDoRequest_ContextCancellation(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.doRequest(ctx, "/test.endpoint")

	if err == nil {
		t.Fatal("expected error due to cancelled context")
	}
}

func TestDoPostRequest_Success(t *testing.T) {
	type testRequest struct {
		Name string `json:"name"`
	}
	type testResponse struct {
		ID string `json:"id"`
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type to be application/json")
		}
		if r.Header.Get("x-api-key") != "test-api-key" {
			t.Errorf("expected x-api-key header to be set")
		}

		var req testRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Name != "test-name" {
			t.Errorf("expected name=test-name, got %s", req.Name)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(testResponse{ID: "123"})
	})
	defer server.Close()

	client := newTestClient(server)
	data, err := client.doPostRequest(context.Background(), "/test.create", testRequest{Name: "test-name"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result testResponse
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result.ID != "123" {
		t.Errorf("expected ID=123, got %s", result.ID)
	}
}

func TestDoPostRequest_ClientError(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "validation failed"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.doPostRequest(context.Background(), "/test.create", map[string]string{"name": "test"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
	}
}

func TestDoDeleteRequest_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.doDeleteRequest(context.Background(), "/test.delete", map[string]string{"id": "123"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoDeleteRequest_NotFoundTreatedAsSuccess(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.doDeleteRequest(context.Background(), "/test.delete", map[string]string{"id": "123"})

	if err != nil {
		t.Fatalf("expected no error (404 should be success), got: %v", err)
	}
}

func TestDoDeleteRequest_ClientError(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.doDeleteRequest(context.Background(), "/test.delete", map[string]string{"id": "123"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
	}
}

func TestDoRequest_RetryOnServerError(t *testing.T) {
	attempts := 0
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})
	defer server.Close()

	retryConfig := RetryConfig{
		MaxRetries:     2,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
	}
	client := NewWithRetry(server.URL, "test-api-key", retryConfig)

	data, err := client.doRequest(context.Background(), "/test.endpoint")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}

	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("expected status=ok, got %s", result["status"])
	}
}

func TestDoPostRequest_RetryOnServerError(t *testing.T) {
	attempts := 0
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "123"}`))
	})
	defer server.Close()

	retryConfig := RetryConfig{
		MaxRetries:     2,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
	}
	client := NewWithRetry(server.URL, "test-api-key", retryConfig)

	data, err := client.doPostRequest(context.Background(), "/test.create", map[string]string{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}

	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result["id"] != "123" {
		t.Errorf("expected id=123, got %s", result["id"])
	}
}
