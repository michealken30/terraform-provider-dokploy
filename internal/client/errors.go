package client

import (
	"errors"
	"fmt"
	"net/http"
)

// APIError represents an error returned by the Dokploy API.
type APIError struct {
	StatusCode int
	Endpoint   string
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error %d on %s: %s", e.StatusCode, e.Endpoint, e.Message)
	}
	return fmt.Sprintf("API error %d on %s: %s", e.StatusCode, e.Endpoint, e.Body)
}

// NotFoundError represents a 404 error from the API.
type NotFoundError struct {
	ResourceType string
	ResourceID   string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.ResourceType, e.ResourceID)
}

// IsNotFound returns true if the error indicates a resource was not found.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	// Check for NotFoundError
	var notFoundErr *NotFoundError
	if errors.As(err, &notFoundErr) {
		return true
	}

	// Check for APIError with 404 status
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound
	}

	return false
}

// IsRetryable returns true if the error is retryable (rate limit or server error).
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		// Retry on rate limit (429) or server errors (5xx)
		return apiErr.StatusCode == http.StatusTooManyRequests ||
			(apiErr.StatusCode >= 500 && apiErr.StatusCode < 600)
	}

	return false
}

// newAPIError creates a new APIError from an HTTP response.
func newAPIError(statusCode int, endpoint string, body []byte) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Endpoint:   endpoint,
		Body:       string(body),
	}
}

// newNotFoundError creates a new NotFoundError.
func newNotFoundError(resourceType, resourceID string) *NotFoundError {
	return &NotFoundError{
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}
}
