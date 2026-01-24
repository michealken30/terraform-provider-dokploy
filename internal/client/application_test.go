package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetApplication_Success(t *testing.T) {
	projects := []Project{
		{
			ProjectID: "proj-1",
			Name:      "Project 1",
			Environments: []Environment{
				{
					EnvironmentID: "env-1",
					Applications: []Application{
						{ApplicationID: "app-1", Name: "App One"},
						{ApplicationID: "app-2", Name: "App Two"},
					},
				},
			},
		},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(projects)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetApplication(context.Background(), "app-2")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ApplicationID != "app-2" {
		t.Errorf("expected application ID app-2, got %s", result.ApplicationID)
	}
	if result.Name != "App Two" {
		t.Errorf("expected name 'App Two', got %s", result.Name)
	}
}

func TestGetApplication_NotFound(t *testing.T) {
	projects := []Project{
		{
			ProjectID: "proj-1",
			Environments: []Environment{
				{
					EnvironmentID: "env-1",
					Applications: []Application{
						{ApplicationID: "app-1", Name: "App One"},
					},
				},
			},
		},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(projects)
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetApplication(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestGetApplication_EmptyProjects(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]Project{})
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetApplication(context.Background(), "app-1")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestCreateApplication_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/application.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Name != "New App" {
			t.Errorf("expected name 'New App', got %s", req.Name)
		}
		if req.EnvironmentID != "env-123" {
			t.Errorf("expected environment ID 'env-123', got %s", req.EnvironmentID)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateApplicationResponse{ApplicationID: "new-app-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateApplication(context.Background(), CreateApplicationRequest{
		Name:          "New App",
		EnvironmentID: "env-123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ApplicationID != "new-app-id" {
		t.Errorf("expected application ID new-app-id, got %s", result.ApplicationID)
	}
}

func TestCreateApplication_WithOptionalFields(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		var req CreateApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.AppName == nil || *req.AppName != "my-app" {
			t.Errorf("expected appName 'my-app', got %v", req.AppName)
		}
		if req.Description == nil || *req.Description != "Test description" {
			t.Errorf("expected description 'Test description', got %v", req.Description)
		}
		if req.ServerID == nil || *req.ServerID != "srv-123" {
			t.Errorf("expected server ID 'srv-123', got %v", req.ServerID)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateApplicationResponse{ApplicationID: "app-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	appName := "my-app"
	desc := "Test description"
	serverID := "srv-123"
	result, err := client.CreateApplication(context.Background(), CreateApplicationRequest{
		Name:          "New App",
		AppName:       &appName,
		Description:   &desc,
		EnvironmentID: "env-123",
		ServerID:      &serverID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ApplicationID != "app-id" {
		t.Errorf("expected application ID app-id, got %s", result.ApplicationID)
	}
}

func TestCreateApplication_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "name already exists"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateApplication(context.Background(), CreateApplicationRequest{
		Name:          "Duplicate",
		EnvironmentID: "env-123",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateApplication_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/application.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req UpdateApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.ApplicationID != "app-1" {
			t.Errorf("expected application ID app-1, got %s", req.ApplicationID)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	newName := "Updated App"
	err := client.UpdateApplication(context.Background(), UpdateApplicationRequest{
		ApplicationID: "app-1",
		Name:          &newName,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateApplication_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "application not found"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdateApplication(context.Background(), UpdateApplicationRequest{
		ApplicationID: "non-existent",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteApplication_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/application.delete" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req DeleteApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.ApplicationID != "app-to-delete" {
			t.Errorf("expected application ID app-to-delete, got %s", req.ApplicationID)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteApplication(context.Background(), "app-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteApplication_NotFoundTreatedAsSuccess(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteApplication(context.Background(), "already-deleted")

	if err != nil {
		t.Fatalf("expected no error (404 should be success for delete), got: %v", err)
	}
}

func TestDeleteApplication_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "application is running"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteApplication(context.Background(), "running-app")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetEnvironment_Success(t *testing.T) {
	projects := []Project{
		{
			ProjectID: "proj-1",
			Name:      "Project 1",
			Environments: []Environment{
				{EnvironmentID: "env-1", Name: "Production"},
				{EnvironmentID: "env-2", Name: "Staging"},
			},
		},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(projects)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetEnvironment(context.Background(), "env-2")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.EnvironmentID != "env-2" {
		t.Errorf("expected environment ID env-2, got %s", result.EnvironmentID)
	}
	if result.Name != "Staging" {
		t.Errorf("expected name 'Staging', got %s", result.Name)
	}
}

func TestGetEnvironment_NotFound(t *testing.T) {
	projects := []Project{
		{
			ProjectID: "proj-1",
			Environments: []Environment{
				{EnvironmentID: "env-1", Name: "Production"},
			},
		},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(projects)
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetEnvironment(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestGetEnvironmentsByProjectID_Success(t *testing.T) {
	projects := []Project{
		{
			ProjectID: "proj-1",
			Name:      "Project 1",
			Environments: []Environment{
				{EnvironmentID: "env-1", Name: "Production"},
				{EnvironmentID: "env-2", Name: "Staging"},
			},
		},
		{
			ProjectID: "proj-2",
			Environments: []Environment{
				{EnvironmentID: "env-3", Name: "Development"},
			},
		},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(projects)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetEnvironmentsByProjectID(context.Background(), "proj-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 environments, got %d", len(result))
	}
}

func TestGetEnvironmentsByProjectID_ProjectNotFound(t *testing.T) {
	projects := []Project{
		{ProjectID: "proj-1"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(projects)
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetEnvironmentsByProjectID(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateEnvironment_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/environment.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateEnvironmentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Name != "New Environment" {
			t.Errorf("expected name 'New Environment', got %s", req.Name)
		}
		if req.ProjectID != "proj-123" {
			t.Errorf("expected project ID 'proj-123', got %s", req.ProjectID)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateEnvironmentResponse{EnvironmentID: "new-env-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateEnvironment(context.Background(), CreateEnvironmentRequest{
		Name:      "New Environment",
		ProjectID: "proj-123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.EnvironmentID != "new-env-id" {
		t.Errorf("expected environment ID new-env-id, got %s", result.EnvironmentID)
	}
}

func TestUpdateEnvironment_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/environment.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req UpdateEnvironmentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.EnvironmentID != "env-1" {
			t.Errorf("expected environment ID env-1, got %s", req.EnvironmentID)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	newName := "Updated Environment"
	err := client.UpdateEnvironment(context.Background(), UpdateEnvironmentRequest{
		EnvironmentID: "env-1",
		Name:          &newName,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteEnvironment_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/environment.remove" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteEnvironment(context.Background(), "env-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
