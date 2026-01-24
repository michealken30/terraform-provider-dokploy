package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetOrganizations_Success(t *testing.T) {
	orgs := []Organization{
		{OrganizationID: "org-1", Name: "Org One"},
		{OrganizationID: "org-2", Name: "Org Two"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/organization.all" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(orgs)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetOrganizations(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 organizations, got %d", len(result))
	}
	if result[0].Name != "Org One" {
		t.Errorf("expected name 'Org One', got %s", result[0].Name)
	}
}

func TestGetOrganizations_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	})
	defer server.Close()

	client := NewWithRetry(server.URL, "test-api-key", RetryConfig{MaxRetries: 0})
	_, err := client.GetOrganizations(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetDeployments_Application_Success(t *testing.T) {
	deployments := []Deployment{
		{DeploymentID: "deploy-1", Status: "success"},
		{DeploymentID: "deploy-2", Status: "running"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/deployment.all" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("applicationId") != "app-123" {
			t.Errorf("expected applicationId=app-123, got %s", r.URL.Query().Get("applicationId"))
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(deployments)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetDeployments(context.Background(), "app-123", "application")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 deployments, got %d", len(result))
	}
}

func TestGetDeployments_Compose_Success(t *testing.T) {
	deployments := []Deployment{
		{DeploymentID: "deploy-1", Status: "success"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/deployment.allByCompose" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("composeId") != "compose-123" {
			t.Errorf("expected composeId=compose-123, got %s", r.URL.Query().Get("composeId"))
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(deployments)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetDeployments(context.Background(), "compose-123", "compose")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 deployment, got %d", len(result))
	}
}

func TestGetDeployments_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "application not found"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetDeployments(context.Background(), "non-existent", "application")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetDeployments_EmptyResult(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]Deployment{})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetDeployments(context.Background(), "app-no-deployments", "application")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 deployments, got %d", len(result))
	}
}

func TestGetGitProviders_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	})
	defer server.Close()

	client := NewWithRetry(server.URL, "test-api-key", RetryConfig{MaxRetries: 0})
	_, err := client.GetGitProviders(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetSSHKeys_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetSSHKeys(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetRegistries_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "forbidden"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetRegistries(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetCertificates_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	})
	defer server.Close()

	client := NewWithRetry(server.URL, "test-api-key", RetryConfig{MaxRetries: 0})
	_, err := client.GetCertificates(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetDestinations_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	})
	defer server.Close()

	client := NewWithRetry(server.URL, "test-api-key", RetryConfig{MaxRetries: 0})
	_, err := client.GetDestinations(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetNotifications_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	})
	defer server.Close()

	client := NewWithRetry(server.URL, "test-api-key", RetryConfig{MaxRetries: 0})
	_, err := client.GetNotifications(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
