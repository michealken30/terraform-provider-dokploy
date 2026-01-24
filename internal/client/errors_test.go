package client

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name: "with message",
			err: &APIError{
				StatusCode: 400,
				Endpoint:   "/project.create",
				Message:    "invalid project name",
			},
			expected: "API error 400 on /project.create: invalid project name",
		},
		{
			name: "with body only",
			err: &APIError{
				StatusCode: 500,
				Endpoint:   "/server.all",
				Body:       `{"error": "internal error"}`,
			},
			expected: `API error 500 on /server.all: {"error": "internal error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("APIError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNotFoundError_Error(t *testing.T) {
	err := &NotFoundError{
		ResourceType: "project",
		ResourceID:   "abc123",
	}

	expected := "project not found: abc123"
	if got := err.Error(); got != expected {
		t.Errorf("NotFoundError.Error() = %q, want %q", got, expected)
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "NotFoundError",
			err:      &NotFoundError{ResourceType: "project", ResourceID: "123"},
			expected: true,
		},
		{
			name:     "wrapped NotFoundError",
			err:      fmt.Errorf("failed: %w", &NotFoundError{ResourceType: "server", ResourceID: "456"}),
			expected: true,
		},
		{
			name:     "APIError with 404",
			err:      &APIError{StatusCode: http.StatusNotFound, Endpoint: "/test"},
			expected: true,
		},
		{
			name:     "wrapped APIError with 404",
			err:      fmt.Errorf("request failed: %w", &APIError{StatusCode: 404, Endpoint: "/test"}),
			expected: true,
		},
		{
			name:     "APIError with 500",
			err:      &APIError{StatusCode: http.StatusInternalServerError, Endpoint: "/test"},
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("something went wrong"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.expected {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "rate limit 429",
			err:      &APIError{StatusCode: http.StatusTooManyRequests, Endpoint: "/test"},
			expected: true,
		},
		{
			name:     "server error 500",
			err:      &APIError{StatusCode: http.StatusInternalServerError, Endpoint: "/test"},
			expected: true,
		},
		{
			name:     "server error 502",
			err:      &APIError{StatusCode: http.StatusBadGateway, Endpoint: "/test"},
			expected: true,
		},
		{
			name:     "server error 503",
			err:      &APIError{StatusCode: http.StatusServiceUnavailable, Endpoint: "/test"},
			expected: true,
		},
		{
			name:     "client error 400",
			err:      &APIError{StatusCode: http.StatusBadRequest, Endpoint: "/test"},
			expected: false,
		},
		{
			name:     "not found 404",
			err:      &APIError{StatusCode: http.StatusNotFound, Endpoint: "/test"},
			expected: false,
		},
		{
			name:     "unauthorized 401",
			err:      &APIError{StatusCode: http.StatusUnauthorized, Endpoint: "/test"},
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("network error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryable(tt.err); got != tt.expected {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewAPIError(t *testing.T) {
	err := newAPIError(500, "/project.all", []byte(`{"error": "failed"}`))

	if err.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", err.StatusCode)
	}
	if err.Endpoint != "/project.all" {
		t.Errorf("Endpoint = %q, want /project.all", err.Endpoint)
	}
	if err.Body != `{"error": "failed"}` {
		t.Errorf("Body = %q, want %q", err.Body, `{"error": "failed"}`)
	}
}

func TestNewNotFoundError(t *testing.T) {
	err := newNotFoundError("environment", "env-123")

	if err.ResourceType != "environment" {
		t.Errorf("ResourceType = %q, want environment", err.ResourceType)
	}
	if err.ResourceID != "env-123" {
		t.Errorf("ResourceID = %q, want env-123", err.ResourceID)
	}
}
